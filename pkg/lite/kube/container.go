package kube

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	squashkube "github.com/solo-io/squash/pkg/platforms/kubernetes"
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
	err = syscall.Exec(fulldlv, []string{fulldlv, "attach", fmt.Sprintf("%d", pid)}, nil)
	log.WithField("err", err).Info("exec failed!")

	return errors.New("can't start dlv")
}
