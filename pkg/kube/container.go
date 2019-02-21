package kube

import (
	"context"
	"fmt"

	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/platforms/kubernetes"
)

const OutPort = "1236"
const ListenHost = "127.0.0.1"

func Debug(ctx context.Context) error {
	cfg := GetConfig()

	var err error
	var containerProcess platforms.ContainerProcess

	containerProcess, err = kubernetes.NewContainerProcess()
	if err != nil {
		containerProcess, err = kubernetes.NewCRIContainerProcessAlphaV1()
		if err != nil {
			return err
		}
	}

	info, err := containerProcess.GetContainerInfo(ctx, &cfg.Attachment)
	if err != nil {
		return err
	}

	pid := info.Pids[0]
	fmt.Println("about to serve")

	return startDebugging(cfg, pid)
}
