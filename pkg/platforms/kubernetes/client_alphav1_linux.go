package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/davecgh/go-spew/spew"

	squashv1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/utils/processwatcher"

	"google.golang.org/grpc"

	kubeapi "github.com/solo-io/squash/pkg/platforms/kubernetes/alphav1"
	k8models "github.com/solo-io/squash/pkg/platforms/kubernetes/models"
)

type CRIContainerProcessAlphaV1 struct{}

func NewCRIContainerProcessAlphaV1() (*CRIContainerProcessAlphaV1, error) {
	// test that we have access to the runtime service
	cc, err := grpc.Dial(CriRuntime, grpc.WithInsecure(), grpc.WithDialer(getDialer))
	if err != nil {
		return nil, err
	}
	runtimeService := kubeapi.NewRuntimeServiceClient(cc)

	in := &kubeapi.StatusRequest{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = runtimeService.Status(ctx, in)
	cancel()
	if err != nil {
		return nil, err
	}
	return &CRIContainerProcessAlphaV1{}, nil
}

func getDialer(a string, t time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", a, t)
}

func (c *CRIContainerProcessAlphaV1) GetContainerInfo(maincontext context.Context, attachment *squashv1.DebugAttachment) (*platforms.ContainerInfo, error) {

	fmt.Println("v1")
	fmt.Println(attachment)
	logger := contextutils.LoggerFrom(context.TODO())
	logger.Debugw("Cri GetPid called", "attachment", attachment)

	ka, err := k8models.DebugAttachmentToKubeAttachment(attachment)

	if err != nil {
		return nil, errors.New("bad attachment format")
	}

	// contact the local CRI and get the container

	cc, err := grpc.Dial(CriRuntime, grpc.WithInsecure(), grpc.WithDialer(getDialer))
	runtimeService := kubeapi.NewRuntimeServiceClient(cc)

	labels := make(map[string]string)
	labels["io.kubernetes.pod.name"] = ka.Pod
	labels["io.kubernetes.pod.namespace"] = ka.Namespace
	st := kubeapi.PodSandboxStateValue{State: kubeapi.PodSandboxState_SANDBOX_READY}
	inpod := &kubeapi.ListPodSandboxRequest{
		Filter: &kubeapi.PodSandboxFilter{
			LabelSelector: labels,
			State:         &st,
		},
	}

	logger.Debugw("Cri GetPid ListPodSandbox", "inpod", spew.Sdump(inpod))

	ctx, cancel := context.WithTimeout(maincontext, time.Second)
	resp, err := runtimeService.ListPodSandbox(ctx, inpod)
	cancel()
	if err != nil {
		logger.Warnw("ListPodSandbox error", "err", err)
		return nil, err
	}
	if len(resp.Items) != 1 {
		logger.Warnw("Invalid number of pods", "items", spew.Sdump(resp.Items))
		return nil, errors.New("Invalid number of pods")
	}
	pod := resp.Items[0]

	labels = make(map[string]string)
	labels["io.kubernetes.container.name"] = ka.Container
	incont := &kubeapi.ListContainersRequest{
		Filter: &kubeapi.ContainerFilter{
			PodSandboxId:  pod.Id,
			LabelSelector: labels,
		},
	}
	logger.Debugw("Cri GetPid ListContainers", "incont", spew.Sdump(incont))

	ctx, cancel = context.WithTimeout(maincontext, time.Second)
	respcont, err := runtimeService.ListContainers(ctx, incont)
	cancel()

	if err != nil {
		logger.Warnw("ListContainers error", "err", err)
		return nil, err
	}
	logger.Debugw("Cri GetPid ListContainers - got response", "respcont", spew.Sdump(respcont))

	var containers []*kubeapi.Container
	for _, cont := range respcont.Containers {
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
	nsinod, err := getNSAlphav1(maincontext, runtimeService, nstocheck, containerid)
	if err != nil {
		logger.Warnw("getNSAlphav1 error", "err", err)
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

func FindPidsInNS(inod uint64, ns string) ([]int, error) {
	var res []int
	files, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}

		p := filepath.Join("/proc", f.Name(), "ns", ns)
		if inod2, err := processwatcher.PathToInode(p); err != nil {
			continue
		} else if inod == inod2 {
			res = append(res, pid)
		}
	}

	return res, nil
}

func getNSAlphav1(origctx context.Context, cli kubeapi.RuntimeServiceClient, ns string, containerid string) (uint64, error) {

	req := &kubeapi.ExecSyncRequest{
		ContainerId: containerid,
		Cmd:         []string{"ls", "-l", "/proc/self/ns/"},
		Timeout:     1,
	}

	ctx, cancel := context.WithTimeout(origctx, time.Second)
	result, err := cli.ExecSync(ctx, req)
	cancel()
	if err != nil {
		contextutils.LoggerFrom(origctx).Warnw("Error exec sync to get pid ns!", "err", err)
		return 0, err
	}
	/* output looks like:
	lrwxrwxrwx 1 root root 0 Jul 28 16:39 /proc/1/ns/pid -> pid:[4026532605]
	...
	*/
	output := result.Stdout
	regex := regexp.MustCompile(ns + `:\[(\d+)\]`)
	matches := regex.FindStringSubmatch(string(output))
	if len(matches) != 2 {
		return 0, errors.New("mnt namespace not found")
	}

	inod, err := strconv.ParseInt(matches[1], 10, 0)
	return uint64(inod), err
}
