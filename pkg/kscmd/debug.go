package kscmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	skaffkubeapi "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	survey "gopkg.in/AlecAivazis/survey.v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	squashkube "github.com/solo-io/squash/pkg/platforms/kubernetes"
)

var ImageVersion string
var ImageRepo string

const DebuggerPort = "1235"

const (
	ImageContainer = "kubesquash-container"
	ContainerName  = "kubesquash-container"
	namespace      = "squash"
	skaffoldFile   = "skaffold.yaml"
)

type SquashConfig struct {
	ChooseDebugger        bool
	NoClean               bool
	ChoosePod             bool
	NoDetectSkaffold      bool
	TimeoutSeconds        int
	DebugContainerVersion string
	DebugContainerRepo    string
	DebugServer           bool

	Debugger  string
	Namespace string
	Pod       string
	Container string
	Machine   bool

	CRISock string
}

func StartDebugContainer(config SquashConfig) error {
	// find the container from skaffold, or ask the user to chose one.

	dp := DebugPrepare{
		config: config,
	}

	debugger, err := dp.chooseDebugger()
	if err != nil {
		return err
	}
	ns, podname, image := config.Namespace, config.Pod, config.Container
	if podname == "" && image == "" {
		if !config.NoDetectSkaffold {
			image, podname, _ = SkaffoldConfigToPod(skaffoldFile)
		}
	}

	dbg, err := dp.GetMissing(ns, podname, image)
	if err != nil {
		return err
	}

	if !config.Machine {
		confirmed := false
		prompt := &survey.Confirm{
			Message: "Going to attach " + debugger + " to pod " + dbg.Pod.ObjectMeta.Name + ". continue?",
			Default: true,
		}
		survey.AskOne(prompt, &confirmed, nil)
		if !confirmed {
			return errors.New("user aborted")
		}
	}

	dbgpod, err := dp.debugPodFor(debugger, dbg.Pod, dbg.Container.Name)
	if err != nil {
		return err
	}
	// create namespace. ignore errors as it most likely exists and will error
	dp.getClientSet().CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}})

	createdPod, err := dp.getClientSet().CoreV1().Pods(namespace).Create(dbgpod)
	if err != nil {
		return err
	}

	// TODO: we may be able to delete with DebugServer. see TODO below
	if (!dp.config.DebugServer) && (!config.NoClean) {
		// do not remove the pod on a debug server as it is waiting for a
		// connection
		defer dp.deletePod(createdPod)
	}

	// wait for running state
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
	err = <-dp.waitForPod(ctx, createdPod)
	cancel()
	if err != nil {
		dp.showLogs(err, createdPod)
		return err
	}

	if dp.config.DebugServer {
		// TODO: do we want to delete the pod on successful completion?
		// that would require us to track the lifetime of the session

		// print the pod name and exit
		fmt.Printf("pod.name: %v", createdPod.Name)
	} else {

		// Starting port forward in background.
		cmd1 := exec.Command("kubectl", "port-forward", createdPod.ObjectMeta.Name, DebuggerPort, "-n", namespace)
		cmd1.Stdout = os.Stdout
		cmd1.Stderr = os.Stderr
		cmd1.Stdin = os.Stdin
		err = cmd1.Start()
		if err != nil {
			dp.showLogs(err, createdPod)
			return err
		}

		// Delaying to allow port forwarding to complete.
		duration := time.Duration(5) * time.Second
		time.Sleep(duration)

		cmd2 := exec.Command("dlv", "connect", "127.0.0.1:"+DebuggerPort)
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		cmd2.Stdin = os.Stdin
		err = cmd2.Run()
		if err != nil {
			dp.showLogs(err, createdPod)
			return err
		}

	}
	return nil
}
func (dp *DebugPrepare) deletePod(createdPod *v1.Pod) {
	var options metav1.DeleteOptions
	dp.getClientSet().CoreV1().Pods(namespace).Delete(createdPod.ObjectMeta.Name, &options)
}
func (dp *DebugPrepare) showLogs(err error, createdPod *v1.Pod) {

	cmd := exec.Command("kubectl", "-n", namespace, "logs", createdPod.ObjectMeta.Name, ContainerName)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Can't get logs from errored pod")
		return
	}

	fmt.Printf("Pod errored with: %v\n Logs:\n %s", err, string(buf))
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
				createdPod, err = dp.getClientSet().CoreV1().Pods(namespace).Get(name, options)
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
	req := dp.getClientSet().Core().Pods(namespace).GetLogs(pod.ObjectMeta.Name, &options)

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
	Namespace string
	Pod       *v1.Pod
	Container *v1.Container
}

type DebugPrepare struct {
	clientset kubernetes.Interface
	config    SquashConfig
}

func GetSkaffoldConfig(filename string) (*config.SkaffoldConfig, error) {

	buf, err := util.ReadConfiguration(filename)
	if err != nil {
		return nil, errors.Wrap(err, "read skaffold config")
	}

	apiVersion := &config.ApiVersion{}
	if err := yaml.Unmarshal(buf, apiVersion); err != nil {
		return nil, errors.Wrap(err, "parsing api version")
	}

	if apiVersion.Version != config.LatestVersion {
		return nil, errors.New("Config version out of date.`")
	}

	cfg, err := config.GetConfig(buf, true, false)
	if err != nil {
		return nil, errors.Wrap(err, "parsing skaffold config")
	}

	// we already ensured that the versions match in the previous block,
	// so this type assertion is safe.
	latestConfig, ok := cfg.(*config.SkaffoldConfig)
	if !ok {
		return nil, errors.Wrap(err, "can't use skaffold config")
	}
	return latestConfig, nil
}

func SkaffoldConfigToPod(filename string) (string, string, error) {
	latestConfig, err := GetSkaffoldConfig(filename)

	if err != nil {
		return "", "", err
	}
	if len(latestConfig.Build.Artifacts) == 0 {
		return "", "", errors.New("no artifacts")
	}
	image := latestConfig.Build.Artifacts[0].ImageName
	podname := "" //latestConfig.Deploy.Name
	return image, podname, nil
}

func (dp *DebugPrepare) getClientSet() kubernetes.Interface {
	if dp.clientset != nil {
		return dp.clientset
	}
	clientset, err := skaffkubeapi.GetClientset()
	if err != nil {
		panic(err)
	}
	dp.clientset = clientset
	return dp.clientset

}

func (dp *DebugPrepare) GetMissing(ns, podname, image string) (*Debugee, error) {

	//	clientset.CoreV1().Namespace().
	// see if namespace exist, and if not prompot for one.
	var options metav1.GetOptions
	var debuggee Debugee
	debuggee.Namespace = ns
	if debuggee.Namespace == "" {
		var err error
		debuggee.Namespace, err = dp.chooseNamespace()
		if err != nil {
			return nil, errors.Wrap(err, "choosing namespace")
		}
	}

	if podname == "" {
		var err error
		debuggee.Pod, err = dp.choosePod(debuggee.Namespace, image)
		if err != nil {
			return nil, errors.Wrap(err, "choosing pod")
		}
	} else {
		var err error
		debuggee.Pod, err = dp.getClientSet().CoreV1().Pods(debuggee.Namespace).Get(podname, options)
		if err != nil {
			return nil, errors.Wrap(err, "fetching pod")
		}
	}

	if image == "" {
		var err error
		debuggee.Container, err = dp.chooseContainer(debuggee.Pod)
		if err != nil {
			return nil, errors.Wrap(err, "choosing container")
		}
	} else {
		for _, podContainer := range debuggee.Pod.Spec.Containers {
			if strings.HasPrefix(podContainer.Image, image) {
				debuggee.Container = &podContainer
				break
			}
		}
		if debuggee.Container == nil {
			return nil, errors.New("no such container image")
		}
	}
	return &debuggee, nil
}

func (dp *DebugPrepare) chooseContainer(pod *v1.Pod) (*v1.Container, error) {
	if len(pod.Spec.Containers) == 0 {
		return nil, errors.New("no container to choose from")

	}
	if len(pod.Spec.Containers) == 1 {
		return &pod.Spec.Containers[0], nil
	}

	containerNames := make([]string, 0, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		contname := container.Name
		containerNames = append(containerNames, contname)
	}

	question := &survey.Select{
		Message: "Select a container",
		Options: containerNames,
	}
	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		return nil, err
	}

	for _, container := range pod.Spec.Containers {
		if choice == container.Name {
			return &container, nil
		}
	}

	return nil, errors.New("selected container not found")
}

func (dp *DebugPrepare) detectLang() string {
	if dp.config.ChooseDebugger {
		// manual mode
		return ""
	}
	// TODO: find some decent huristics to make this work
	return "dlv"
}

func (dp *DebugPrepare) chooseDebugger() (string, error) {
	if len(dp.config.Debugger) != 0 {
		return dp.config.Debugger, nil
	}

	availableDebuggers := []string{"dlv", "gdb"}
	debugger := dp.detectLang()

	if debugger == "" {
		question := &survey.Select{
			Message: "Select a debugger",
			Options: availableDebuggers,
		}
		var choice string
		if err := survey.AskOne(question, &choice, survey.Required); err != nil {
			return "", err
		}
		return choice, nil
	}
	return debugger, nil
}

func (dp *DebugPrepare) chooseNamespace() (string, error) {

	var options metav1.ListOptions
	namespaces, err := dp.getClientSet().CoreV1().Namespaces().List(options)
	if err != nil {
		return "", errors.Wrap(err, "reading namesapces")
	}
	namespaceNames := make([]string, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		nsname := ns.ObjectMeta.Name
		if nsname == "squash" {
			continue
		}
		if strings.HasPrefix(nsname, "kube-") {
			continue
		}
		namespaceNames = append(namespaceNames, nsname)
	}
	if len(namespaceNames) == 0 {
		return "", errors.New("no namespaces available!")
	}

	if len(namespaceNames) == 1 {
		return namespaceNames[0], nil
	}

	question := &survey.Select{
		Message: "Select a namespace",
		Options: namespaceNames,
	}
	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		return "", err
	}
	return choice, nil
}

func (dp *DebugPrepare) choosePod(ns, image string) (*v1.Pod, error) {

	var options metav1.ListOptions
	pods, err := dp.getClientSet().CoreV1().Pods(ns).List(options)
	if err != nil {
		return nil, errors.Wrap(err, "reading namesapces")
	}
	podName := make([]string, 0, len(pods.Items))
	for _, pod := range pods.Items {
		if dp.config.ChoosePod || image == "" {
			podName = append(podName, pod.ObjectMeta.Name)
		} else {
			for _, podContainer := range pod.Spec.Containers {
				if strings.HasPrefix(podContainer.Image, image) {
					podName = append(podName, pod.ObjectMeta.Name)
					break
				}
			}
		}
	}

	var choice string
	if len(podName) == 1 {
		choice = podName[0]
	} else {
		question := &survey.Select{
			Message: "Select a pod",
			Options: podName,
		}
		if err := survey.AskOne(question, &choice, survey.Required); err != nil {
			return nil, err
		}
	}
	for _, pod := range pods.Items {
		if choice == pod.ObjectMeta.Name {
			return &pod, nil
		}
	}

	return nil, errors.New("pod not found")
}

func (dp *DebugPrepare) debugPodFor(debugger string, in *v1.Pod, containername string) (*v1.Pod, error) {
	trueVar := true
	const crisockvolume = "crisock"
	isDebugServer := ""
	if dp.config.DebugServer {
		isDebugServer = "1"
	}
	templatePod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: ContainerName,
			Labels:       map[string]string{"squash": "kubesquash-container"},
		},
		Spec: v1.PodSpec{
			HostPID:       true,
			RestartPolicy: v1.RestartPolicyNever,
			NodeName:      in.Spec.NodeName,
			Containers: []v1.Container{{
				Name:      ContainerName,
				Image:     dp.config.DebugContainerRepo + "/" + ImageContainer + "-" + debugger + ":" + dp.config.DebugContainerVersion,
				Stdin:     true,
				StdinOnce: true,
				TTY:       true,
				VolumeMounts: []v1.VolumeMount{{
					Name:      crisockvolume,
					MountPath: squashkube.CriRuntime,
				}},
				SecurityContext: &v1.SecurityContext{
					Privileged: &trueVar,
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
					Name:  "DEBUGGER_SERVER",
					Value: fmt.Sprintf("%s", isDebugServer),
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
