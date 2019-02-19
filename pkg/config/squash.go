package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	sqOpts "github.com/solo-io/squash/pkg/options"
	squashkube "github.com/solo-io/squash/pkg/platforms/kubernetes"
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

	createdPod, err := s.getClientSet().CoreV1().Pods(s.getDebuggerPodNamespace()).Create(dbgpod)
	if err != nil {
		return nil, err
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
	portSpec := fmt.Sprintf("%v:%v", s.LocalPort, sqOpts.DebuggerPort)
	cmd1 := exec.Command("kubectl", "port-forward", createdPod.ObjectMeta.Name, portSpec, "-n", s.getDebuggerPodNamespace())
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	cmd1.Stdin = os.Stdin
	err := cmd1.Start()
	if err != nil {
		s.showLogs(err, createdPod)
		return err
	}
	// kill the kubectl port-forward process on exit to free the port
	// this defer must be called after Start() initializes Process
	defer cmd1.Process.Kill()

	// Delaying to allow port forwarding to complete.
	duration := time.Duration(5) * time.Second
	time.Sleep(duration)

	// TODO(mitchdraft) dlv only atm - check if dlv before doing this
	cmd2 := exec.Command("dlv", "connect", fmt.Sprintf("127.0.0.1:%v", s.LocalPort))
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	cmd2.Stdin = os.Stdin
	err = cmd2.Run()
	if err != nil {
		log.Warn("failed, printing logs")
		log.Warn(err)
		s.showLogs(err, createdPod)
		return err
	}
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

func (s *Squash) debugPodFor(debugger string, in *v1.Pod, containername string) (*v1.Pod, error) {
	const crisockvolume = "crisock"

	// this is our convention for naming the container images that contain specific debuggers
	fullParticularContainerName := fmt.Sprintf("%v-%v", sqOpts.ParticularContainerRootName, debugger)
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
			HostPID:       true,
			RestartPolicy: v1.RestartPolicyNever,
			NodeName:      in.Spec.NodeName,
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
