package kscmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/squash/pkg/config"
	sqOpts "github.com/solo-io/squash/pkg/options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	squashkube "github.com/solo-io/squash/pkg/platforms/kubernetes"
)

func StartDebugContainer(cfg config.Squash, dbg Debugee) (*v1.Pod, error) {

	dp := DebugPrepare{
		config: cfg,
	}

	dbgpod, err := dp.debugPodFor(cfg.Debugger, dbg.Pod, dbg.Container.Name)
	if err != nil {
		return nil, err
	}
	debuggerPodNamespace := dp.config.Namespace
	// create namespace. ignore errors as it most likely exists and will error
	dp.getClientSet().CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: debuggerPodNamespace}})

	createdPod, err := dp.getClientSet().CoreV1().Pods(debuggerPodNamespace).Create(dbgpod)
	if err != nil {
		return nil, err
	}

	if !cfg.NoClean {
		// do not remove the pod on a debug server as it is waiting for a
		// connection
		defer dp.deletePod(createdPod)
	}

	// wait for running state
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	err = <-dp.waitForPod(ctx, createdPod)
	cancel()
	if err != nil {
		dp.showLogs(err, createdPod)
		return nil, err
	}

	if err := dp.connectUser(debuggerPodNamespace, createdPod); err != nil {
		return nil, err
	}

	return createdPod, nil
}

func (dp *DebugPrepare) connectUser(debuggerPodNamespace string, createdPod *v1.Pod) error {
	if dp.config.Machine {
		return nil
	}
	// Starting port forward in background.
	portSpec := sqOpts.DebuggerPort
	localConnectPort := sqOpts.DebuggerPort
	if dp.config.LocalPort != 0 {
		portSpec = fmt.Sprintf("%v:%v", dp.config.LocalPort, sqOpts.DebuggerPort)
		localConnectPort = fmt.Sprintf("%v", dp.config.LocalPort)
	}
	cmd1 := exec.Command("kubectl", "port-forward", createdPod.ObjectMeta.Name, portSpec, "-n", debuggerPodNamespace)
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	cmd1.Stdin = os.Stdin
	err := cmd1.Start()
	if err != nil {
		dp.showLogs(err, createdPod)
		return err
	}

	// Delaying to allow port forwarding to complete.
	duration := time.Duration(5) * time.Second
	time.Sleep(duration)

	// TODO(mitchdraft) dlv only atm - check if dlv before doing this
	cmd2 := exec.Command("dlv", "connect", fmt.Sprintf("127.0.0.1:%v", localConnectPort))
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	cmd2.Stdin = os.Stdin
	err = cmd2.Run()
	if err != nil {
		log.Warn("failed, printing logs")
		log.Warn(err)
		dp.showLogs(err, createdPod)
		return err
	}
	return nil
}

func (dp *DebugPrepare) deletePod(createdPod *v1.Pod) {
	var options metav1.DeleteOptions
	dp.getClientSet().CoreV1().Pods(dp.config.Namespace).Delete(createdPod.ObjectMeta.Name, &options)
}
func (dp *DebugPrepare) showLogs(err error, createdPod *v1.Pod) {

	cmd := exec.Command("kubectl", "-n", dp.config.Namespace, "logs", createdPod.ObjectMeta.Name, sqOpts.ContainerName)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Can't get logs from errored pod")
		return
	}

	fmt.Printf("Pod errored with: %v\n Logs:\n %s", err, string(buf))
	log.Warn(fmt.Sprintf("Pod errored with: %v\n Logs:\n %s", err, string(buf)))
}

func (dp *DebugPrepare) waitForPod(ctx context.Context, createdPod *v1.Pod) <-chan error {
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

				var options metav1.GetOptions
				options.ResourceVersion = createdPod.ResourceVersion
				var err error
				createdPod, err = dp.getClientSet().CoreV1().Pods(dp.config.Namespace).Get(name, options)
				if err != nil {
					errchan <- err
					return
				}
				if createdPod.Status.Phase == v1.PodRunning {
					return
				}
				if createdPod.Status.Phase != v1.PodPending {
					err := dp.printError(createdPod)
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

func (dp *DebugPrepare) printError(pod *v1.Pod) error {
	var options v1.PodLogOptions
	req := dp.getClientSet().Core().Pods(dp.config.Namespace).GetLogs(pod.ObjectMeta.Name, &options)

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

type Debugee struct {
	Pod       *v1.Pod
	Container *v1.Container
}

type DebugPrepare struct {
	clientset kubernetes.Interface
	config    config.Squash
}

func (dp *DebugPrepare) getClientSet() kubernetes.Interface {
	if dp.clientset != nil {
		return dp.clientset
	}
	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		panic(err)
	}
	cs, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		panic(err)
	}
	dp.clientset = cs
	return dp.clientset

}

func (dp *DebugPrepare) debugPodFor(debugger string, in *v1.Pod, containername string) (*v1.Pod, error) {
	const crisockvolume = "crisock"

	// this is our convention for naming the container images that contain specific debuggers
	fullParticularContainerName := fmt.Sprintf("%v-%v", sqOpts.ParticularContainerRootName, debugger)
	// repoRoot/containerName:tag
	targetImage := fmt.Sprintf("%v/%v:%v", dp.config.DebugContainerRepo, fullParticularContainerName, dp.config.DebugContainerVersion)
	templatePod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
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
						Path: dp.config.CRISock,
					},
				},
			}},
		}}

	return templatePod, nil
}
