package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	squashkubeutils "github.com/solo-io/squash/pkg/utils/kubeutils"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	squashv1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/debuggers/local"
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
	ProcessName        string
	KubeConfig         string

	CRISock string

	clientset kubernetes.Interface
	daClient  *squashv1.DebugAttachmentClient

	SquashNamespace string
}

func NewSquashConfig(daClient *squashv1.DebugAttachmentClient) Squash {
	return Squash{
		daClient: daClient,
	}
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

	cs, err := s.getClientSet()
	if err != nil {
		return nil, err
	}

	// create namespace. ignore errors as it most likely exists and will error
	cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: s.SquashNamespace}})

	// create debugger pod
	createdPod, err := cs.CoreV1().Pods(s.SquashNamespace).Create(dbgpod)
	if err != nil {
		return nil, fmt.Errorf("Could not create pod: %v", err)
	}

	if !s.Machine && !s.NoClean {
		// do not remove the pod on a debug server as it is waiting for a
		// connection
		// TODO: handle returned error
		defer s.deletePod(createdPod)
	}

	// wait for running state
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	err = <-s.waitForPod(ctx, createdPod)
	cancel()
	if err != nil {
		// s.printError(createdPodName)
		return nil, fmt.Errorf("Waiting for pod: %v", err)
	}

	if err := s.ReportOrConnectToCreatedDebuggerPod(); err != nil {
		return nil, err
	}

	return createdPod, nil
}

type noDebugAttachmentClient struct{}

func (noDebugAttachmentClient) Error() string {
	return "no debug attachment has been provided"
}

func IsNoDebugAttachmentClientError(err error) bool {
	switch err.(type) {
	case *noDebugAttachmentClient:
		return true
	}
	return false
}

// need to check this in a function since this code is used by the cli and a client
// may not have been created yet
func (s *Squash) GetClient() (*squashv1.DebugAttachmentClient, error) {
	if s.daClient == nil {
		return nil, &noDebugAttachmentClient{}
	}
	return s.daClient, nil
}
func (s *Squash) SetClient(daClient *squashv1.DebugAttachmentClient) {
	s.daClient = daClient
}

// for the debug controller, this function finds the debug target
// from the squash spec that it receives
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
	cs, err := s.getClientSet()
	if err != nil {
		return err
	}
	dbt.Pod, err = cs.CoreV1().Pods(s.Namespace).Get(s.Pod, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "fetching pod")
	}
	return nil
}

func (s *Squash) GetDebugTargetContainerFromSpec(dbt *DebugTarget) error {
	for _, podContainer := range dbt.Pod.Spec.Containers {
		log.Debug(podContainer.Image)
		log.Info(podContainer.Image)
		if strings.HasPrefix(podContainer.Name, s.Container) {
			dbt.Container = &podContainer
			break
		}
	}
	if dbt.Container == nil {
		return errors.New(fmt.Sprintf("no such container name: %v", s.Container))
	}
	return nil
}

func (s *Squash) getDebuggerPodNamespace() string {
	return s.Namespace
}

func (s *Squash) ReportOrConnectToCreatedDebuggerPod() error {
	da, err := s.getDebugAttachment()
	if err != nil {
		return err
	}
	remoteDbgPort, err := local.GetDebugPortFromCrd(da.Metadata.Name, s.Namespace)
	if err != nil {
		return err
	}
	if s.Machine {
		return s.printEditorExtensionData(remoteDbgPort)
	}
	return s.connectUser(da, remoteDbgPort)
}

type EditorData struct {
	PortForwardCmd string
}

func (s *Squash) connectUser(da *squashv1.DebugAttachment, remoteDbgPort int) error {
	if s.Machine {
		return nil
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
	return s.callLocalDebuggerCommand(dbgCmd)
}

func (s *Squash) printEditorExtensionData(remoteDbgPort int) error {
	da, err := s.getDebugAttachment()
	if err != nil {
		return err
	}

	debugger := local.GetParticularDebugger(s.Debugger)
	kubectlCmd := debugger.GetEditorRemoteConnectionCmd(
		da.PlankName,
		s.SquashNamespace,
		s.Pod,
		s.Namespace,
		remoteDbgPort,
	)
	ed := EditorData{
		PortForwardCmd: kubectlCmd,
	}
	json, err := json.Marshal(ed)
	if err != nil {
		return err
	}
	fmt.Println(string(json))
	return nil
}

// TODO - remove this when V2 api is ready
func (s *Squash) GetIntent() squashv1.Intent {
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
	intent := s.GetIntent()
	daClient, err := s.GetClient()
	if err != nil {
		return nil, err
	}
	return intent.GetDebugAttachment(*daClient)
}

func (s *Squash) deletePod(createdPod *v1.Pod) error {
	var options meta_v1.DeleteOptions
	cs, err := s.getClientSet()
	if err != nil {
		return err
	}
	return cs.CoreV1().Pods(s.Namespace).Delete(createdPod.ObjectMeta.Name, &options)
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

				cs, err := s.getClientSet()
				if err != nil {
					errchan <- err
					return
				}
				createdPod, err = cs.CoreV1().Pods(s.SquashNamespace).Get(name, options)

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
						errchan <- errors.Wrapf(err, "pod is not running and not pending, status: %v", createdPod.Status.Phase)
					} else {
						errchan <- errors.New(fmt.Sprintf("pod is not running and not pending, status: %v", createdPod.Status.Phase))
					}
					return
				}
			}
		}
	}()
	return errchan
}

func (s *Squash) printError(podName string) error {
	cs, err := s.getClientSet()
	if err != nil {
		return err
	}

	var options v1.PodLogOptions

	req := cs.CoreV1().Pods(s.Namespace).GetLogs(podName, &options)

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
	it := s.GetIntent()
	const crisockvolume = "crisock"
	cs, err := s.getClientSet()
	if err != nil {
		return nil, err
	}
	targetPod, err := cs.CoreV1().Pods(it.Pod.Namespace).Get(it.Pod.Name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// get debugAttachment name so Plank knows where to find it
	da, err := s.getDebugAttachment()
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
			ImagePullSecrets: []v1.LocalObjectReference{{
				Name: sqOpts.SquashServiceAccountImagePullSecretName,
			}},
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
				}, {
					Name:  sqOpts.PlankEnvDebugSquashNamespace,
					Value: s.SquashNamespace,
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

func (s *Squash) getClientSet() (kubernetes.Interface, error) {
	if s.clientset == nil {
		cs, err := squashkubeutils.GetKubeClient()
		if err != nil {
			return nil, err
		}
		s.clientset = cs
	}

	return s.clientset, nil
}

// DeletePlankPod deletes the plank pod that was created for the debug session
// represented by the Squash object. This should be called when a debugging session
// is terminated.
func (s *Squash) DeletePlankPod() error {
	da, err := s.getDebugAttachment()
	if err != nil {
		return err
	}

	cs, err := s.getClientSet()
	if err != nil {
		return err
	}

	return cs.CoreV1().Pods(s.SquashNamespace).Delete(da.PlankName, &meta_v1.DeleteOptions{})
}
