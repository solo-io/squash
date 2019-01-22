package kube

import (
	"context"
	"fmt"

	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/platforms/kubernetes"
)

type DebuggerInfo struct {
	CmdlineGen func(int) []string
}

var debuggers map[string]*DebuggerInfo
var debuggerServer map[string]*DebuggerInfo

const DebuggerPort = "1235"
const OutPort = "1236"
const ListenHost = "127.0.0.1"

func GetDebugServerAddress() string {
	return fmt.Sprintf("%v:%v", ListenHost, OutPort)
}

func init() {
	debuggers = make(map[string]*DebuggerInfo)
	debuggers["dlv"] = &DebuggerInfo{CmdlineGen: func(pid int) []string {
		return []string{"attach", fmt.Sprintf("%d", pid)}
	}}

	debuggers["gdb"] = &DebuggerInfo{CmdlineGen: func(pid int) []string {
		return []string{"-p", fmt.Sprintf("%d", pid)}
	}}
	debuggerServer = make(map[string]*DebuggerInfo)
	debuggerServer["dlv"] = &DebuggerInfo{CmdlineGen: func(pid int) []string {
		return []string{"attach", fmt.Sprintf("%d", pid), "--listen=127.0.0.1:" + DebuggerPort, "--headless", "--log", "--api-version=2"}
	}}
}

func Debug() error {
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

	info, err := containerProcess.GetContainerInfo(context.TODO(), &cfg.Attachment)
	if err != nil {
		return err
	}

	pid := info.Pids[0]

	return startServer(cfg, pid)
}
