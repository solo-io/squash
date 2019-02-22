package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kr/pty"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	squashv1 "github.com/solo-io/squash/pkg/api/v1"
	sqOpts "github.com/solo-io/squash/pkg/options"
	squashkube "github.com/solo-io/squash/pkg/platforms/kubernetes"
	"github.com/solo-io/squash/pkg/utils"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	plankServiceAccountName     = "squash-plank"
	plankClusterRoleName        = "squash-plank-cr"
	plankClusterRoleBindingName = "squash-plank-crb"
)

type Squash struct {
	ChooseDebugger        bool
	NoClean               bool
	ChoosePod             bool
	TimeoutSeconds        int
	DebugContainerVersion string
	DebugContainerRepo    string
	LocalPort             int

	Debugger           string
	Namespace          string
	Pod                string
	Container          string
	Machine            bool
	DebugServerAddress string

	CRISock string

	clientset kubernetes.Interface
}

type DebugTarget struct {
	Pod       *v1.Pod
	Container *v1.Container
}

func StartDebugContainer(s Squash, dbt DebugTarget) (*v1.Pod, error) {

	dbgpod, err := s.debugPodFor(s.Debugger, dbt.Pod, dbt.Container.Name)
	if err != nil {
		return nil, err
	}
	// create namespace. ignore errors as it most likely exists and will error
	s.getClientSet().CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: s.getDebuggerPodNamespace()}})

	// grant permissions required by debugger pod
	if err := s.createPermissions(); err != nil {
		return nil, err
	}

	// create debugger pod
	createdPod, err := s.getClientSet().CoreV1().Pods(s.getDebuggerPodNamespace()).Create(dbgpod)
	if err != nil {
		return nil, fmt.Errorf("Could not create pod: %v", err)
	}

	if !s.Machine && !s.NoClean {
		// do not remove the pod on a debug server as it is waiting for a
		// connection
		defer s.deletePod(createdPod)
	}

	// wait for running state
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.TimeoutSeconds)*time.Second)
	err = <-s.waitForPod(ctx, createdPod)
	cancel()
	if err != nil {
		s.showLogs(err, createdPod)
		return nil, err
	}

	if err := s.ReportOrConnectToCreatedDebuggerPod(createdPod); err != nil {
		return nil, err
	}

	return createdPod, nil
}

// for the debug controller, this function finds the debug target
// from the squash spec that it recieves
// If it is able to find a unique target, it applies the target
// values to the DebugTarget argument. Otherwise it errors.
func (s *Squash) ExpectToGetUniqueDebugTargetFromSpec(dbt *DebugTarget) error {
	if err := s.GetDebugTargetPodFromSpec(dbt); err != nil {
		return err
	}
	if err := s.GetDebugTargetContainerFromSpec(dbt); err != nil {
		return err
	}
	return nil
}

func (s *Squash) GetDebugTargetPodFromSpec(dbt *DebugTarget) error {
	var err error
	dbt.Pod, err = s.getClientSet().CoreV1().Pods(s.Namespace).Get(s.Pod, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "fetching pod")
	}
	return nil
}

func (s *Squash) GetDebugTargetContainerFromSpec(dbt *DebugTarget) error {
	for _, podContainer := range dbt.Pod.Spec.Containers {
		log.Debug(podContainer.Image)
		log.Info(podContainer.Image)
		if strings.HasPrefix(podContainer.Image, s.Container) {
			dbt.Container = &podContainer
			break
		}
	}
	if dbt.Container == nil {
		return errors.New(fmt.Sprintf("no such container image: %v", s.Container))
	}
	return nil
}

func (s *Squash) getDebuggerPodNamespace() string {
	return s.Namespace
}

func (s *Squash) ReportOrConnectToCreatedDebuggerPod(createdPod *v1.Pod) error {
	if s.Machine {
		fmt.Printf("pod.name: %v", createdPod.Name)
	} else {
		return s.connectUser(createdPod)
	}
	return nil
}

func (s *Squash) connectUser(createdPod *v1.Pod) error {
	if s.Machine {
		return nil
	}
	// Starting port forward in background.
	remoteDbgPort, err := s.getDebugPortFromCrd()
	if err != nil {
		return err
	}
	kubectlCmd := s.getPortForwardCmd(createdPod.ObjectMeta.Name, remoteDbgPort)
	if err := kubectlCmd.Start(); err != nil {
		s.showLogs(err, createdPod)
		return err
	}
	// kill the kubectl port-forward process on exit to free the port
	// this defer must be called after Start() initializes Process
	defer kubectlCmd.Process.Kill()

	// Delaying to allow port forwarding to complete.
	time.Sleep(5 * time.Second)
	if os.Getenv("DEBUG_SELF") != "" {
		fmt.Println("FOR DEBUGGING SQUASH'S DEBUGGER CONTAINER:")
		s.printError(createdPod)
	}

	dbgCmd := s.getDebugCmd()
	if err := ptyWrap(dbgCmd); err != nil {
		// if err := dbgCmd.Run(); err != nil {
		log.Warn("failed, printing logs")
		log.Warn(err)
		s.showLogs(err, createdPod)
		return err
	}
	return nil
}

func (s *Squash) getPortForwardCmd(dbgPodName string, dbgPodPort int) *exec.Cmd {
	targetPodName, targetNamespace := "", ""
	targetRemotePort := dbgPodPort
	switch s.Debugger {
	case "dlv":
		// for dlv, we proxy through the debug container
		targetPodName = dbgPodName
		targetNamespace = s.getDebuggerPodNamespace()
	case "java":
		// for java, we connect directly to the container we are debugging
		targetPodName = s.Pod
		targetNamespace = s.Namespace
	default:
		// TODO - log/error
		fmt.Println("Unsupported debugger")
	}
	portSpec := fmt.Sprintf("%v:%v", s.LocalPort, targetRemotePort)
	cmd := exec.Command("kubectl", "port-forward", targetPodName, portSpec, "-n", targetNamespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func (s *Squash) getDebugCmd() *exec.Cmd {
	cmd := &exec.Cmd{}
	switch s.Debugger {
	case "dlv":
		cmd = exec.Command("dlv", "connect", fmt.Sprintf("127.0.0.1:%v", s.LocalPort))
	case "java":
		cmd = exec.Command("jdb", "-attach", fmt.Sprintf("127.0.0.1:%v", s.LocalPort))
	default:
		log.Warn(fmt.Errorf("debugger not recognized %v", s.Debugger))
		return nil
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func (s *Squash) getDebugPortFromCrd() (int, error) {
	// TODO - all of our ports should be gotten from the crd. As is, it is possible that the random port chosen from ip_addr:0 could return 1236 - slim chance but may as well handle it
	// TODO - should eventually just read all values from crd
	port := sqOpts.OutPort
	if s.Debugger == "java" {
		// Give debug container time to create the CRD
		// TODO - reduce this sleep time
		time.Sleep(5 * time.Second)
		ctx := context.Background()
		daClient, err := utils.GetDebugAttachmentClient(ctx)
		if err != nil {
			log.WithField("err", err).Error("getting debug attachment client")
			return 0, err
		}
		daName := squashv1.GenDebugAttachmentName(s.Pod, s.Container)
		da, err := (*daClient).Read(s.Namespace, daName, clients.ReadOpts{Ctx: ctx})
		if err != nil {
			return 0, fmt.Errorf("Could not read debug attachment %v in namespace %v: %v", daName, s.Namespace, err)
		}
		port, err = da.GetPortFromDebugServerAddress()
		if err != nil {
			return 0, err
		}
	}
	return port, nil
}

func ptyWrap(c *exec.Cmd) error {

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)

	return nil
}

func (s *Squash) showLogs(err error, createdPod *v1.Pod) {

	cmd := exec.Command("kubectl", "-n", s.Namespace, "logs", createdPod.ObjectMeta.Name, sqOpts.ContainerName)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Can't get logs from errored pod")
		return
	}

	fmt.Printf("Pod errored with: %v\n Logs:\n %s", err, string(buf))
	log.Warn(fmt.Sprintf("Pod errored with: %v\n Logs:\n %s", err, string(buf)))
}

func (s *Squash) deletePod(createdPod *v1.Pod) {
	var options meta_v1.DeleteOptions
	s.getClientSet().CoreV1().Pods(s.Namespace).Delete(createdPod.ObjectMeta.Name, &options)
}

func (s *Squash) waitForPod(ctx context.Context, createdPod *v1.Pod) <-chan error {
	errchan := make(chan error, 1)
	go func() {
		defer close(errchan)
		name := createdPod.ObjectMeta.Name

		for {
			select {
			case <-ctx.Done():
				errchan <- ctx.Err()
				return
			case <-time.After(time.Second):

				var options meta_v1.GetOptions
				options.ResourceVersion = createdPod.ResourceVersion
				var err error
				createdPod, err = s.getClientSet().CoreV1().Pods(s.Namespace).Get(name, options)
				if err != nil {
					errchan <- err
					return
				}
				if !s.expectRunningPod() {
					return
				}
				if createdPod.Status.Phase == v1.PodRunning {
					return
				}
				if createdPod.Status.Phase != v1.PodPending {
					err := s.printError(createdPod)
					if err != nil {
						errchan <- errors.Wrap(err, "pod is not running and not pending")
					} else {
						errchan <- errors.New("pod is not running and not pending")
					}
					return
				}
			}
		}
	}()
	return errchan
}

func (s *Squash) expectRunningPod() bool {
	switch s.Debugger {
	case "dlv":
		return true
	case "java":
		return false
	default:
		// TODO - remove this when debugger name validation is in place
		return true
	}
}

func (s *Squash) printError(pod *v1.Pod) error {
	var options v1.PodLogOptions
	req := s.getClientSet().Core().Pods(s.Namespace).GetLogs(pod.ObjectMeta.Name, &options)

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	_, err = io.Copy(os.Stderr, readCloser)
	if err != nil {
		return err
	}
	return nil
}

func (s *Squash) getClientSet() kubernetes.Interface {
	if s.clientset != nil {
		return s.clientset
	}
	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		panic(err)
	}
	cs, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		panic(err)
	}
	s.clientset = cs
	return s.clientset

}

// Containers have a common name, suffixed by the particular debugger that they have installed
// TODO(mitchdraft) - implement more specific debug containers (for example, bare containers for debuggers that don't need a specific process)
// for now, default to the gdb variant
func containerNameFromSpec(debugger string) string {
	containerVariant := "gdb"
	if debugger == "dlv" {
		containerVariant = "dlv"
	}
	return fmt.Sprintf("%v-%v", sqOpts.ParticularContainerRootName, containerVariant)
}

func (s *Squash) debugPodFor(debugger string, in *v1.Pod, containername string) (*v1.Pod, error) {
	const crisockvolume = "crisock"

	// this is our convention for naming the container images that contain specific debuggers
	fullParticularContainerName := containerNameFromSpec(debugger)
	// repoRoot/containerName:tag
	targetImage := fmt.Sprintf("%v/%v:%v", s.DebugContainerRepo, fullParticularContainerName, s.DebugContainerVersion)
	templatePod := &v1.Pod{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			GenerateName: sqOpts.ContainerName,
			Labels:       map[string]string{sqOpts.SquashLabelSelectorKey: sqOpts.SquashLabelSelectorValue},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: plankServiceAccountName,
			HostPID:            true,
			RestartPolicy:      v1.RestartPolicyNever,
			NodeName:           in.Spec.NodeName,
			Containers: []v1.Container{{
				Name:      sqOpts.ContainerName,
				Image:     targetImage,
				Stdin:     true,
				StdinOnce: true,
				TTY:       true,
				VolumeMounts: []v1.VolumeMount{{
					Name:      crisockvolume,
					MountPath: squashkube.CriRuntime,
				}},
				SecurityContext: &v1.SecurityContext{
					Capabilities: &v1.Capabilities{
						Add: []v1.Capability{
							"SYS_PTRACE",
						},
					},
				},
				Env: []v1.EnvVar{{
					Name:  "SQUASH_NAMESPACE",
					Value: in.ObjectMeta.Namespace,
				}, {
					Name:  "SQUASH_POD",
					Value: in.ObjectMeta.Name,
				}, {
					Name:  "SQUASH_CONTAINER",
					Value: containername,
				}, {
					Name:  "DEBUGGER_NAME",
					Value: s.Debugger,
				},
				}},
			},
			Volumes: []v1.Volume{{
				Name: crisockvolume,
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: s.CRISock,
					},
				},
			}},
		}}

	return templatePod, nil
}

// TODO - call this once from squashctl once during config instead of each time a pod is created
// for now - just print errors since these resources may have already been created
// we need them to exist in each namespace
func (s *Squash) createPermissions() error {
	namespace := s.Namespace
	cs := s.clientset

	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: plankServiceAccountName,
		},
	}
	if _, err := cs.CoreV1().ServiceAccounts(namespace).Create(&sa); err != nil {
		fmt.Println(err)
	}

	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: plankClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
				Resources: []string{"pods"},
				APIGroups: []string{""},
			},
			{
				Verbs:     []string{"list"},
				Resources: []string{"namespaces"},
				APIGroups: []string{""},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
				Resources: []string{"debugattachments"},
				APIGroups: []string{"squash.solo.io"},
			},
			{
				// TODO remove the register permission when solo-kit is updated
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete", "register"},
				Resources: []string{"customresourcedefinitions"},
				APIGroups: []string{"apiextensions.k8s.io"},
			},
		},
	}
	if _, err := cs.Rbac().ClusterRoles().Create(cr); err != nil {
		fmt.Println(err)
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      plankClusterRoleBindingName,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      plankServiceAccountName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name: plankClusterRoleName,
			Kind: "ClusterRole",
		},
	}
	if _, err := cs.Rbac().ClusterRoleBindings().Create(crb); err != nil {
		fmt.Println(err)
	}
	return nil
}
