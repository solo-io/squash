package kube

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	squashkube "github.com/solo-io/squash/pkg/platforms/kubernetes"
)

type DebuggerInfo struct {
	CmdlineGen func(int) []string
}

var debuggers map[string]*DebuggerInfo

func init() {
	debuggers = make(map[string]*DebuggerInfo)
	debuggers["dlv"] = &DebuggerInfo{CmdlineGen: func(pid int) []string {
		return []string{"attach", fmt.Sprintf("%d", pid)}
	}}

	debuggers["gdb"] = &DebuggerInfo{CmdlineGen: func(pid int) []string {
		return []string{"-p", fmt.Sprintf("%d", pid)}
	}}
}

func Debug() error {
	cfg := GetConfig()

	dbgInfo := debuggers[cfg.Debugger]
	if dbgInfo == nil {
		return errors.New("unknown debugger")
	}

	containerProcess := squashkube.NewContainerProcess()
	info, err := containerProcess.GetContainerInfoKube(nil, &cfg.Attachment)
	if err != nil {
		return err
	}

	pid := info.Pids[0]

	// exec into dlv
	log.WithField("pid", pid).Info("attaching with " + cfg.Debugger)
	fullpath, err := exec.LookPath(cfg.Debugger)
	if err != nil {
		return err
	}
	err = syscall.Exec(fullpath, append([]string{fullpath}, dbgInfo.CmdlineGen(pid)...), nil)
	log.WithField("err", err).Info("exec failed!")

	return errors.New("can't start " + cfg.Debugger)
}
