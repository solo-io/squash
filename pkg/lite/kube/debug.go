package kube

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	skaffkubeapi "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	squashkube "github.com/solo-io/squash/pkg/platforms/kubernetes"
	survey "gopkg.in/AlecAivazis/survey.v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

func Debug() error {
	cfg := GetConfig()

	containerProcess := squashkube.NewContainerProcess()
	info, err := containerProcess.GetContainerInfoKube(nil, &cfg.Attachment)
	if err != nil {
		return err
	}

	pid := info.Pids[0]

	// exec into dlv
	log.WithField("pid", pid).Info("attaching with dlv")
	fulldlv, err := exec.LookPath("dlv")
	if err != nil {
		return err
	}
	err = syscall.Exec(fulldlv, []string{"dlv", "attach", fmt.Sprintf("%d", pid)}, nil)
	log.WithField("err", err).Info("exec failed!")

	return errors.New("can't start dlv")
}

func StartDebugContainer() error {

	// find the container from skaffold, or ask the user to chose one.

	var dp DebugPrepare
	dbg, err := dp.GetMissing("", "", "")
	if err != nil {
		return err
	}

	dbgpod, err := dp.debugPodFor(dbg.Pod, dbg.Container.Name)
	if err != nil {
		return err
	}
	createdPod, err := dp.getClientSet().CoreV1().Pods("squash").Create(dbgpod)
	if err != nil {
		return err
	}

	// wait for runnign state
	name := createdPod.ObjectMeta.Name
	for {
		var options metav1.GetOptions

		createdPod, err := dp.getClientSet().CoreV1().Pods("squash").Get(name, options)
		if err != nil {
			return err
		}
		if createdPod.Status.Phase == v1.PodRunning {
			break
		}
		if createdPod.Status.Phase != v1.PodPending {
			return errors.New("pod is not running and not pending")
		}
		time.Sleep(time.Second)
	}

	// attach to the created
	cmd := exec.Command("kubectl", "attach", "-n", "squash", "-i", "-t", createdPod.ObjectMeta.Name)

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
	latestConfig := cfg.(*config.SkaffoldConfig)
	return latestConfig, nil
}

func SkaffoldConfigToPod(filename string) (string, string, error) {
	latestConfig, err := GetSkaffoldConfig(filename)

	if err != nil {
		return "", "", err
	}
	image := latestConfig.Build.Artifacts[0].ImageName
	podname := latestConfig.Deploy.Name
	return image, podname, nil
}

func (dp *DebugPrepare) trySkaffold() error {
	filename := "skaffold.yaml"
	image, podname, err := SkaffoldConfigToPod(filename)

	if err != nil {
		return err
	}

	dp.GetMissing("default", podname, image)
	panic("TODO")
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
		debuggee.Pod, err = dp.choosePod(debuggee.Namespace)
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
		if len(debuggee.Pod.Spec.Containers) == 1 {
			debuggee.Container = &debuggee.Pod.Spec.Containers[0]
		}
	}

	return &debuggee, nil
}

func (dp *DebugPrepare) chooseNamespace() (string, error) {

	var options metav1.ListOptions
	namespaces, err := dp.getClientSet().CoreV1().Namespaces().List(options)
	if err != nil {
		return "", errors.Wrap(err, "reading namesapces")
	}
	namespaceNames := make([]string, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.ObjectMeta.Name)
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

func (dp *DebugPrepare) choosePod(ns string) (*v1.Pod, error) {

	var options metav1.ListOptions
	pods, err := dp.getClientSet().CoreV1().Pods(ns).List(options)
	if err != nil {
		return nil, errors.Wrap(err, "reading namesapces")
	}
	podName := make([]string, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podName = append(podName, pod.ObjectMeta.Name)
	}
	question := &survey.Select{
		Message: "Select a pod",
		Options: podName,
	}
	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		if choice == pod.ObjectMeta.Name {
			return &pod, nil
		}
	}
	return nil, errors.New("pod not found")
}

func (dp *DebugPrepare) debugPodFor(in *v1.Pod, containername string) (*v1.Pod, error) {
	obj, _, err := scheme.Codecs.UniversalDecoder(schema.GroupVersion{Version: "v1"}).Decode([]byte(podTemplate), nil, nil)
	if err != nil {
		return nil, err
	}
	templatePod := obj.(*v1.Pod)
	templatePod.Spec.NodeName = in.Spec.NodeName
	templatePod.Spec.Containers[0].Env[0].Value = in.ObjectMeta.Namespace
	templatePod.Spec.Containers[0].Env[1].Value = in.ObjectMeta.Name
	templatePod.Spec.Containers[0].Env[2].Value = containername

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
    image: soloio/squash-lite-container:v0.2.1-63-gee29fd98
    imagePullPolicy: Never
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
  volumes:
  - name: crisock
    hostPath:
      path: /var/run/dockershim.sock
`
