package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/utils/processwatcher"

	log "github.com/Sirupsen/logrus"

	"google.golang.org/grpc"
	kubeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

const criRuntime = "/var/run/cri.sock"

type CRIContainer2Pid struct {
}

func NewContainer2Pid() platforms.Container2Pid {
	return &CRIContainer2Pid{}
}

func getDialer(a string, t time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", a, t)
}
func (c *CRIContainer2Pid) GetPid(maincontext context.Context, attachmentname string) (int, error) {
	log.WithField("attachmentname", attachmentname).Debug("Cri GetPid called")
	parts := strings.SplitN(attachmentname, ":", 2)
	if len(parts) != 2 {
		return 0, errors.New("bad container name format")
	}
	podname := parts[0]
	containername := parts[1]

	// contact the local CRI and get the container

	cc, err := grpc.Dial(criRuntime, grpc.WithInsecure(), grpc.WithDialer(getDialer))
	runtimeService := kubeapi.NewRuntimeServiceClient(cc)

	labels := make(map[string]string)
	labels["io.kubernetes.pod.name"] = podname
	inpod := &kubeapi.ListPodSandboxRequest{
		Filter: &kubeapi.PodSandboxFilter{
			LabelSelector: labels,
		},
	}

	log.WithField("inpod", spew.Sdump(inpod)).Debug("Cri GetPid ListPodSandbox")

	ctx, cancel := context.WithTimeout(maincontext, time.Second)
	resp, err := runtimeService.ListPodSandbox(ctx, inpod)
	cancel()
	if err != nil {
		log.WithField("err", err).Warn("ListPodSandbox error")
		return 0, err
	}
	if len(resp.Items) != 1 {
		log.WithField("items", spew.Sdump(resp.Items)).Warn("Invalid number of pods")
		return 0, errors.New("Invalid number of pods")
	}
	pod := resp.Items[0]

	labels = make(map[string]string)
	labels["io.kubernetes.container.name"] = containername
	incont := &kubeapi.ListContainersRequest{
		Filter: &kubeapi.ContainerFilter{
			PodSandboxId:  pod.Id,
			LabelSelector: labels,
		},
	}
	log.WithField("incont", spew.Sdump(incont)).Debug("Cri GetPid ListContainers")

	ctx, cancel = context.WithTimeout(maincontext, time.Second)
	respcont, err := runtimeService.ListContainers(ctx, incont)
	cancel()

	if err != nil {
		log.WithField("err", err).Warn("ListContainers error")
		return 0, err
	}
	log.WithField("respcont", spew.Sdump(respcont)).Debug("Cri GetPid ListContainers - got response")

	var containers []*kubeapi.Container
	for _, cont := range respcont.Containers {
		if cont.State == kubeapi.ContainerState_CONTAINER_RUNNING {
			containers = append(containers, cont)
		}
	}
	log.WithField("containers", spew.Sdump(containers)).Debug("Cri GetPid ListContainers - filtered response")

	if len(containers) != 1 {
		log.WithField("containers", containers).Warn("Invalid number of containers")
		return 0, errors.New("Invalid number of containers")
	}
	container := containers[0]
	containerid := container.Id

	// we check the mnt namespace cause this is the one that cannot be shared with the host...
	nstocheck := "mnt"
	// get pids
	nsinod, err := getNS(maincontext, runtimeService, nstocheck, containerid)
	if err != nil {
		log.WithField("err", err).Warn("getNS error")
		return 0, err
	}

	potentialpids, err := FindPidsInNS(nsinod, nstocheck)
	if err != nil {
		log.WithField("err", err).Warn("FindPidsInNS error")
		return 0, err
	}

	log.WithField("potentialpids", potentialpids).Info("found some pids")
	pid, err := FindFirstProcess(potentialpids)
	if err != nil {
		log.WithField("err", err).Warn("FindFirstProcess error")
		return 0, err
	}
	return pid, nil
}

func FindFirstProcess(pids []int) (int, error) {
	minpid := 0
	var mintime *time.Time
	for _, pid := range pids {
		p := filepath.Join("/proc", fmt.Sprintf("%d", pid), "exe")
		n, err := os.Stat(p)
		if err != nil {
			continue
		}
		t := n.ModTime()
		if (mintime == nil) || t.Before(*mintime) {
			mintime = &t
			minpid = pid
		}
	}

	if minpid == 0 {
		return 0, errors.New("no process found")
	}
	return minpid, nil
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

func getNS(origctx context.Context, cli kubeapi.RuntimeServiceClient, ns string, containerid string) (uint64, error) {

	req := &kubeapi.ExecSyncRequest{
		ContainerId: containerid,
		Cmd:         []string{"ls", "-l", "/proc/self/ns/"},
		Timeout:     1,
	}

	ctx, cancel := context.WithTimeout(origctx, time.Second)
	result, err := cli.ExecSync(ctx, req)
	cancel()
	if err != nil {
		log.WithField("err", err).Warn("Error exec sync to get pid ns!")
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
