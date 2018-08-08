package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/solo-io/squash/pkg/platforms"

	log "github.com/sirupsen/logrus"

	k8models "github.com/solo-io/squash/pkg/platforms/kubernetes/models"
	criapi "k8s.io/kubernetes/pkg/kubelet/apis/cri"
	kubeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"

	"k8s.io/kubernetes/pkg/kubelet/remote"
)

const (
	defaultTimeout = 10 * time.Second
)

type CRIContainerProcess struct{}

var _ platforms.ContainerProcess = &CRIContainerProcess{}

func NewContainerProcess() (*CRIContainerProcess, error) {
	// test that we have access to the runtime service
	r, err := remote.NewRemoteRuntimeService(criRuntime, defaultTimeout)
	if err != nil {
		return nil, err
	}

	_, err = r.Status()
	if err != nil {
		return nil, err
	}

	return &CRIContainerProcess{}, nil
}

func (c *CRIContainerProcess) GetContainerInfo(maincontext context.Context, attachment interface{}) (*platforms.ContainerInfo, error) {

	log.WithField("attachment", attachment).Debug("Cri GetPid called")

	ka, err := k8models.GenericToKubeAttachment(attachment)

	if err != nil {
		return nil, errors.New("bad attachment format")
	}
	return c.GetContainerInfoKube(maincontext, ka)
}

func (c *CRIContainerProcess) GetContainerInfoKube(maincontext context.Context, ka *k8models.KubeAttachment) (*platforms.ContainerInfo, error) {

	if maincontext == nil {
		maincontext = context.Background()
	}

	// contact the local CRI and get the container
	runtimeService, err := remote.NewRemoteRuntimeService(criRuntime, defaultTimeout)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string)
	labels["io.kubernetes.pod.name"] = ka.Pod
	labels["io.kubernetes.pod.namespace"] = ka.Namespace
	st := kubeapi.PodSandboxStateValue{State: kubeapi.PodSandboxState_SANDBOX_READY}
	inpod := &kubeapi.PodSandboxFilter{
		LabelSelector: labels,
		State:         &st,
	}

	log.WithField("inpod", spew.Sdump(inpod)).Debug("Cri GetPid ListPodSandbox")

	resp, err := runtimeService.ListPodSandbox(inpod)
	if err != nil {
		log.WithField("err", err).Warn("ListPodSandbox error")
		return nil, err
	}
	if len(resp) != 1 {
		log.WithField("items", spew.Sdump(resp)).Warn("Invalid number of pods")
		return nil, errors.New("Invalid number of pods")
	}
	pod := resp[0]

	labels = make(map[string]string)
	labels["io.kubernetes.container.name"] = ka.Container
	incont := &kubeapi.ContainerFilter{
		PodSandboxId:  pod.Id,
		LabelSelector: labels,
	}
	log.WithField("incont", spew.Sdump(incont)).Debug("Cri GetPid ListContainers")

	respcont, err := runtimeService.ListContainers(incont)

	if err != nil {
		log.WithField("err", err).Warn("ListContainers error")
		return nil, err
	}
	log.WithField("respcont", spew.Sdump(respcont)).Debug("Cri GetPid ListContainers - got response")

	var containers []*kubeapi.Container
	for _, cont := range respcont {
		if cont.State == kubeapi.ContainerState_CONTAINER_RUNNING {
			containers = append(containers, cont)
		}
	}
	log.WithField("containers", spew.Sdump(containers)).Debug("Cri GetPid ListContainers - filtered response")

	if len(containers) != 1 {
		log.WithField("containers", containers).Warn("Invalid number of containers")
		return nil, errors.New("Invalid number of containers")
	}
	container := containers[0]
	containerid := container.Id

	// we check the mnt namespace cause this is the one that cannot be shared with the host...
	nstocheck := "mnt"
	// get pids
	nsinod, err := getNS(maincontext, runtimeService, nstocheck, containerid)
	if err != nil {
		log.WithField("err", err).Warn("getNS error")
		return nil, err
	}

	potentialpids, err := FindPidsInNS(nsinod, nstocheck)
	if err != nil {
		log.WithField("err", err).Warn("FindPidsInNS error")
		return nil, err
	}

	log.WithField("potentialpids", potentialpids).Info("found some pids")
	return &platforms.ContainerInfo{Pids: potentialpids, Name: fmt.Sprintf("%s.%s", ka.Pod, ka.Namespace)}, nil
}

func getNS(origctx context.Context, cli criapi.RuntimeService, ns string, containerid string) (uint64, error) {

	cmd := []string{"ls", "-l", "/proc/self/ns/"}

	stdout, _, err := cli.ExecSync(containerid, cmd, time.Second)
	if err != nil {
		log.WithField("err", err).Warn("Error exec sync to get pid ns!")
		return 0, err
	}
	/* output looks like:
	lrwxrwxrwx 1 root root 0 Jul 28 16:39 /proc/1/ns/pid -> pid:[4026532605]
	...
	*/
	output := stdout
	regex := regexp.MustCompile(ns + `:\[(\d+)\]`)
	matches := regex.FindStringSubmatch(string(output))
	if len(matches) != 2 {
		return 0, errors.New("mnt namespace not found")
	}

	inod, err := strconv.ParseInt(matches[1], 10, 0)
	return uint64(inod), err
}
