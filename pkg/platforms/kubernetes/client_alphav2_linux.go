package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/platforms"

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
	r, err := remote.NewRemoteRuntimeService(CriRuntime, defaultTimeout)
	if err != nil {
		return nil, err
	}

	_, err = r.Status()
	if err != nil {
		return nil, err
	}

	return &CRIContainerProcess{}, nil
}

func (c *CRIContainerProcess) GetContainerInfo(maincontext context.Context, attachment *v1.DebugAttachment) (*platforms.ContainerInfo, error) {

	fmt.Println("v2")
	logger := contextutils.LoggerFrom(maincontext)
	logger.Debugw("Cri GetPid called", "attachment", attachment)

	ka, err := k8models.DebugAttachmentToKubeAttachment(attachment)

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
	runtimeService, err := remote.NewRemoteRuntimeService("unix://"+CriRuntime, defaultTimeout)
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

	logger := contextutils.LoggerFrom(maincontext)
	logger.Debugw("Cri GetPid ListPodSandbox", "inpod", spew.Sdump(inpod))

	resp, err := runtimeService.ListPodSandbox(inpod)
	if err != nil {
		logger.Warnw("ListPodSandbox error", "err", err)
		return nil, err
	}
	if len(resp) != 1 {
		logger.Warnw("Invalid number of pods", "items", spew.Sdump(resp))
		return nil, errors.New("Invalid number of pods")
	}
	pod := resp[0]

	labels = make(map[string]string)
	labels["io.kubernetes.container.name"] = ka.Container
	incont := &kubeapi.ContainerFilter{
		PodSandboxId:  pod.Id,
		LabelSelector: labels,
	}
	logger.Debugw("Cri GetPid ListContainers", "incont", spew.Sdump(incont))

	respcont, err := runtimeService.ListContainers(incont)

	if err != nil {
		logger.Warnw("ListContainers error", "err", err)
		return nil, err
	}
	logger.Debugw("Cri GetPid ListContainers - got response", "respcont", spew.Sdump(respcont))

	var containers []*kubeapi.Container
	for _, cont := range respcont {
		if cont.State == kubeapi.ContainerState_CONTAINER_RUNNING {
			containers = append(containers, cont)
		}
	}
	logger.Debugw("Cri GetPid ListContainers - filtered response", "containers", spew.Sdump(containers))

	if len(containers) != 1 {
		logger.Warnw("Invalid number of containers", "containers", containers)
		return nil, errors.New("Invalid number of containers")
	}
	container := containers[0]
	containerid := container.Id

	// we check the mnt namespace cause this is the one that cannot be shared with the host...
	nstocheck := "mnt"
	// get pids
	nsinod, err := getNS(maincontext, runtimeService, nstocheck, containerid)
	if err != nil {
		logger.Warnw("getNS error", "err", err)
		return nil, err
	}

	potentialpids, err := FindPidsInNS(nsinod, nstocheck)
	if err != nil {
		logger.Warnw("FindPidsInNS error", "err", err)
		return nil, err
	}

	logger.Infow("found some pids", "potentialpids", potentialpids)
	return &platforms.ContainerInfo{Pids: potentialpids, Name: fmt.Sprintf("%s.%s", ka.Pod, ka.Namespace)}, nil
}

func getNS(origctx context.Context, cli criapi.RuntimeService, ns string, containerid string) (uint64, error) {

	cmd := []string{"ls", "-l", "/proc/self/ns/"}

	stdout, _, err := cli.ExecSync(containerid, cmd, time.Second)
	if err != nil {
		contextutils.LoggerFrom(origctx).Warnw("Error exec sync to get pid ns!", "err", err)
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
