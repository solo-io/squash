package kube

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
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
)

var ImageVersion string
var ImageRepo string

const (
	ImageContainer = "squash-lite-container"
	namespace      = "squash"
	skaffoldFile   = "skaffold.yaml"
)

type SquashConfig struct {
	ChooseDebugger bool
	NoClean        bool
	ChoosePod      bool
	TimeoutSeconds int
	// InClusterMode disables interactive prompts
	InClusterMode bool
	Debugger      string
}

func StartDebugContainer(config SquashConfig, clientset *kubernetes.Interface) error {
	// find the container from skaffold, or ask the user to chose one.

	dp := DebugPrepare{
		config: config,
	}
	if clientset != nil {
		dp.clientset = *clientset
	}

	si, err := dp.getClientSet().Discovery().ServerVersion()
	if err != nil {
		return err
	}
	minoirver, err := strconv.Atoi(si.Minor)
	if err != nil {
		return err
	}
	if minoirver < 10 {
		return fmt.Errorf("squash lite requires kube 1.10 or higher. your version is %s.%s;", si.Major, si.Minor)
	}

	debugger := config.Debugger
	if !config.InClusterMode {
		debugger, err = dp.chooseDebugger()
		if err != nil {
			return err
		}
	}

	image, podname, _ := SkaffoldConfigToPod(skaffoldFile)

	dbg, err := dp.GetMissing("", podname, image)
	if err != nil {
		return err
	}

	if !config.InClusterMode {
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

	// wait for runnign state
	name := createdPod.ObjectMeta.Name
	if !config.NoClean {
		defer func() {
			var options metav1.DeleteOptions
			dp.getClientSet().CoreV1().Pods(namespace).Delete(name, &options)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
	err = <-dp.waitForPod(ctx, createdPod)
	cancel()
	if err != nil {
		return err
	}

	// attach to the created
	cmd := exec.Command("kubectl", "attach", "-n", namespace, "-i", "-t", createdPod.ObjectMeta.Name, "-c", "squash-lite-container")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
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

func (dp *DebugPrepare) GetMissing(ns, podname, container string) (*Debugee, error) {

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
		debuggee.Pod, err = dp.choosePod(debuggee.Namespace, container)
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

	if container == "" {
		var err error
		debuggee.Container, err = dp.chooseContainer(debuggee.Pod)
		if err != nil {
			return nil, errors.Wrap(err, "choosing container")
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

func (dp *DebugPrepare) choosePod(ns, container string) (*v1.Pod, error) {

	var options metav1.ListOptions
	pods, err := dp.getClientSet().CoreV1().Pods(ns).List(options)
	if err != nil {
		return nil, errors.Wrap(err, "reading namesapces")
	}
	podName := make([]string, 0, len(pods.Items))
	for _, pod := range pods.Items {
		if dp.config.ChoosePod || container == "" {
			podName = append(podName, pod.ObjectMeta.Name)
		} else {
			for _, podContainer := range pod.Spec.Containers {
				if strings.HasPrefix(podContainer.Image, container) {
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
	templatePod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "squash-lite-container",
			Labels:       map[string]string{"squash": "squash-lite-container"},
		},
		Spec: v1.PodSpec{
			HostPID:       true,
			RestartPolicy: v1.RestartPolicyNever,
			NodeName:      in.Spec.NodeName,
			Containers: []v1.Container{{
				Name:      "squash-lite-container",
				Image:     ImageRepo + "/" + ImageContainer + "-" + debugger + ":" + ImageVersion,
				Stdin:     true,
				StdinOnce: true,
				TTY:       true,
				VolumeMounts: []v1.VolumeMount{{
					Name:      crisockvolume,
					MountPath: "/var/run/cri.sock",
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
				},
				}},
			},
			Volumes: []v1.Volume{{
				Name: crisockvolume,
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "/var/run/dockershim.sock",
					},
				},
			}},
		}}

	return templatePod, nil
}
