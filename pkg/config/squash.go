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
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	squashv1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/debuggers/local"
	sqOpts "github.com/solo-io/squash/pkg/options"
	squashkube "github.com/solo-io/squash/pkg/platforms/kubernetes"
	"github.com/solo-io/squash/pkg/utils"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	SquashNamespace string
}

func NewSquashConfig() Squash {
	s := Squash{}
	s.clientset = s.getClientSet()
	return s
}

type DebugTarget struct {
	Pod       *v1.Pod
	Container *v1.Container
}

func StartDebugContainer(s Squash, dbt DebugTarget) (*v1.Pod, error) {

	dbgpod, err := s.debugPodFor()
	if err != nil {
		return nil, err
	}
	// create namespace. ignore errors as it most likely exists and will error
	s.getClientSet().CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: s.SquashNamespace}})

	// create debugger pod
	createdPod, err := s.getClientSet().CoreV1().Pods(s.SquashNamespace).Create(dbgpod)
	if err != nil {
		return nil, fmt.Errorf("Could not create pod: %v", err)
	}

	if !s.Machine && !s.NoClean {
		// do not remove the pod on a debug server as it is waiting for a
		// connection
		defer s.deletePod(createdPod)
	}

	// wait for running state
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.TimeoutSeconds)*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	err = <-s.waitForPod(ctx, createdPod)
	cancel()
	// ctx, cancel = context.WithTimeout(context.Background(), time.Duration(s.TimeoutSeconds)*time.Second)
	// err <-s.waitForDebugAttachment(ctx)
	if err != nil {
		// s.printError(createdPodName)
		return nil, fmt.Errorf("Waiting for pod: %v", err)
	}

	if err := s.ReportOrConnectToCreatedDebuggerPod(); err != nil {
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

func (s *Squash) ReportOrConnectToCreatedDebuggerPod() error {
	if s.Machine {
		return s.printEditorExtensionData()
	}
	return s.connectUser()
}

func (s *Squash) printEditorExtensionData() error {
	da, err := s.getDebugAttachment()
	if err != nil {
		return err
	}
	fmt.Printf("pod.name: %v", da.PlankName)
	return nil
}

// TODO - remove this when V2 api is ready
func (s *Squash) getIntent() squashv1.Intent {
	return squashv1.Intent{
		Debugger: s.Debugger,
		Pod: &core.ResourceRef{
			Name:      s.Pod,
			Namespace: s.Namespace,
		},
		ContainerName: s.Container,
	}
}

func (s *Squash) getDebugAttachment() (*squashv1.DebugAttachment, error) {
	// Refactor - eventually Intent will be created during config/user entry
	intent := s.getIntent()
	daClient, err := utils.GetDebugAttachmentClient(context.Background())
	if err != nil {
		return &squashv1.DebugAttachment{}, err
	}
	return intent.GetDebugAttachment(daClient)
}

func (s *Squash) connectUser() error {
	if s.Machine {
		return nil
	}
	da, err := s.getDebugAttachment()
	if err != nil {
		return err
	}
	remoteDbgPort, err := local.GetDebugPortFromCrd(da.Metadata.Name, s.Namespace)
	if err != nil {
		return err
	}
	debugger := local.GetParticularDebugger(s.Debugger)
	kubectlCmd := debugger.GetRemoteConnectionCmd(
		da.PlankName,
		s.SquashNamespace,
		s.Pod,
		s.Namespace,
		s.LocalPort,
		remoteDbgPort,
	)
	// Starting port forward in background.
	if err := kubectlCmd.Start(); err != nil {
		// s.printError(createdPodName)
		return err
	}
	// kill the kubectl port-forward process on exit to free the port
	// this defer must be called after Start() initializes Process
	defer kubectlCmd.Process.Kill()

	// Delaying to allow port forwarding to complete.
	time.Sleep(5 * time.Second)
	if os.Getenv("DEBUG_SELF") != "" {
		fmt.Println("FOR DEBUGGING SQUASH'S DEBUGGER CONTAINER:")
		fmt.Println("TODO")
		// s.printError(createdPod)
	}

	dbgCmd := debugger.GetDebugCmd(s.LocalPort)
	if err := ptyWrap(dbgCmd); err != nil {
		// s.printError(createdPodName)
		return err
	}
	return nil
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
				createdPod, err = s.getClientSet().CoreV1().Pods(s.SquashNamespace).Get(name, options)

				if createdPod.Status.Phase == v1.PodPending {
					if !s.Machine {
						fmt.Println("Pod creating")
					}
					continue
				}
				if err != nil {
					errchan <- errors.Wrap(err, "Error during read")
					return
				}
				// TODO - consider refactor such that GetParticularDebugger is only ever called once per session
				if !local.GetParticularDebugger(s.Debugger).ExpectRunningPlank() {
					return
				}
				if createdPod.Status.Phase == v1.PodRunning {
					return
				}
				if createdPod.Status.Phase != v1.PodPending {
					// err := s.printError(createdPod)
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

func (s *Squash) printError(podName string) error {
	var options v1.PodLogOptions
	req := s.getClientSet().Core().Pods(s.Namespace).GetLogs(podName, &options)

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

func (s *Squash) debugPodFor() (*v1.Pod, error) {
	it := s.getIntent()
	const crisockvolume = "crisock"
	targetPod, err := s.getClientSet().CoreV1().Pods(it.Pod.Namespace).Get(it.Pod.Name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// get debugAttachment name so Plank knows where to find it
	daClient, err := utils.GetDebugAttachmentClient(context.Background())
	if err != nil {
		return nil, err
	}
	da, err := it.GetDebugAttachment(daClient)
	if err != nil {
		return nil, err
	}

	// this is our convention for naming the container images that contain specific debuggers
	fullParticularContainerName := containerNameFromSpec(it.Debugger)
	// repoRoot/containerName:tag
	targetImage := fmt.Sprintf("%v/%v:%v", s.DebugContainerRepo, fullParticularContainerName, s.DebugContainerVersion)
	templatePod := &v1.Pod{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			GenerateName: sqOpts.PlankContainerName,
			Labels:       sqOpts.GeneratePlankLabels(it.Pod),
		},
		Spec: v1.PodSpec{
			ServiceAccountName: sqOpts.PlankServiceAccountName,
			HostPID:            true,
			RestartPolicy:      v1.RestartPolicyNever,
			NodeName:           targetPod.Spec.NodeName,
			Containers: []v1.Container{{
				Name:      sqOpts.PlankContainerName,
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
					Name:  sqOpts.PlankEnvDebugAttachmentNamespace,
					Value: it.Pod.Namespace,
				}, {
					Name:  sqOpts.PlankEnvDebugAttachmentName,
					Value: da.Metadata.Name,
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
