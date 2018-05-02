package kube

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	skaffkubeapi "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	survey "gopkg.in/AlecAivazis/survey.v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

var ImageVersion string
var ImageRepo string

const (
	ImageContainer = "squash-lite-container"
	namespace      = "squash"
	skaffoldFile   = "skaffold.yaml"
)

func (dp *DebugPrepare) trySkaffold() error {
	image, podname, err := SkaffoldConfigToPod(skaffoldFile)

	if err != nil {
		return err
	}

	dp.GetMissing("default", podname, image)
	panic("TODO")
}

func StartDebugContainer() error {
	// find the container from skaffold, or ask the user to chose one.

	var dp DebugPrepare

	debugger, err := dp.chooseDebugger()
	if err != nil {
		return err
	}

	image, podname, _ := SkaffoldConfigToPod(skaffoldFile)

	dbg, err := dp.GetMissing("", podname, image)
	if err != nil {
		return err
	}

	confirmed := false
	prompt := &survey.Confirm{
		Message: "Going to attach " + debugger + " to pod " + dbg.Pod.ObjectMeta.Name + ". continue?",
		Default: true,
	}
	survey.AskOne(prompt, &confirmed, nil)
	if !confirmed {
		return errors.New("user aborted")
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
	if os.Getenv("NO_CLEAN") != "1" {
		defer func() {
			var options metav1.DeleteOptions
			dp.getClientSet().CoreV1().Pods(namespace).Delete(name, &options)
		}()
	}

	for {
		var options metav1.GetOptions

		createdPod, err := dp.getClientSet().CoreV1().Pods(namespace).Get(name, options)
		if err != nil {
			return err
		}
		if createdPod.Status.Phase == v1.PodRunning {
			break
		}
		if createdPod.Status.Phase != v1.PodPending {
			// TODO: print logs from the pod
			return errors.New("pod is not running and not pending")
		}
		time.Sleep(time.Second)
	}

	// attach to the created
	cmd := exec.Command("kubectl", "attach", "-n", namespace, "-i", "-t", createdPod.ObjectMeta.Name)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
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
	// TODO: should we make this a user choice?
	return &pod.Spec.Containers[0], nil
}

func (dp *DebugPrepare) detectLang() string {
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
		if container == "" {
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
	obj, _, err := scheme.Codecs.UniversalDecoder(schema.GroupVersion{Version: "v1"}).Decode([]byte(podTemplate), nil, nil)
	if err != nil {
		return nil, err
	}
	templatePod := obj.(*v1.Pod)
	templatePod.Spec.NodeName = in.Spec.NodeName
	templatePod.Spec.Containers[0].Image = ImageRepo + "/" + ImageContainer + "-" + debugger + ":" + ImageVersion
	templatePod.Spec.Containers[0].Env[0].Value = in.ObjectMeta.Namespace
	templatePod.Spec.Containers[0].Env[1].Value = in.ObjectMeta.Name
	templatePod.Spec.Containers[0].Env[2].Value = containername
	templatePod.Spec.Containers[0].Env[3].Value = debugger

	return templatePod, nil
}

var podTemplate = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    squash: squash-lite-container
  generateName: squash-lite-container
spec:
  hostPID: true
  restartPolicy: Never
  nodeName: placeholder
  containers:
  - name: squash-lite-container
    image: placeholder/squash-lite-container:placeholder
    stdin: true
    stdinOnce: true
    tty: true
    volumeMounts:
    - mountPath: /var/run/cri.sock
      name: crisock
    securityContext:
      privileged: true
    ports:
    - containerPort: 1234
      protocol: TCP
    env:
    - name: SQUASH_NAMESPACE
      value: placeholder
    - name: SQUASH_POD
      value: placeholder
    - name: SQUASH_CONTAINER
      value: placeholder
    - name: DEBUGGER
      value: placeholder
  volumes:
  - name: crisock
    hostPath:
      path: /var/run/dockershim.sock
`
