package plank

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils"

	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/platforms/kubernetes"
)

const ListenHost = "127.0.0.1"

func Debug(ctx context.Context) error {
	cfg, err := GetConfig(ctx)
	if err != nil {
		return err
	}

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

	pid, err := getPid(&cfg.Attachment, info)
	if err != nil {
		return err
	}
	fmt.Println("about to serve")

	return startDebugging(cfg, pid)
}

func getPid(da *v1.DebugAttachment, info *platforms.ContainerInfo) (int, error) {
	if da.ProcessName == "" {
		return info.Pids[0], nil
	}
	reg, err := regexp.Compile(strings.ToLower(da.ProcessName))
	if err != nil {
		return 0, errors.Wrapf(err, "unable to match process name, invalid match specification")
	}
	for _, pid := range info.Pids {
		cmdLines, err := utils.GetCmdArgsByPid(pid)
		if err != nil {
			return 0, errors.Wrapf(err, "could not get command line for pid %v", pid)
		}
		preparedCmdLine := strings.ToLower(strings.Join(cmdLines, ""))
		if reg.MatchString(preparedCmdLine) {
			return pid, nil
		}
	}
	return 0, errors.Wrapf(err, "could not find a command line matching %v", da.ProcessName)
}
