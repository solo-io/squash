package kube

import (
	"context"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	sqOpts "github.com/solo-io/squash/pkg/options"
	"go.uber.org/zap"
)

func startServer(ctx context.Context, cfg Config, pid int) error {
	// we proxy so we can exit the debugger when disconnection occours

	dbgInfo := debuggerServer[cfg.Debugger]
	if dbgInfo == nil {
		return errors.New("unknown debugger")
	}

	cmd, err := startDebuggerServer(ctx, pid, dbgInfo)
	if err != nil {
		return err
	}

	errchan := make(chan error, 1)
	reporterr := func(err error) {
		select {
		case errchan <- err:
		default:
		}
	}
	go func() {
		reporterr(cmd.Wait())
	}()

	conn, err := startLocalServer()
	if err != nil {
		return err
	}
	defer conn.Close()

	// connect to debug server
	conn2, err := net.Dial("tcp", ListenHost+":"+sqOpts.DebuggerPort)
	if err != nil {
		return err
	}

	go func() {
		_, err := io.Copy(conn2, conn)
		reporterr(err)
	}()
	go func() {
		// if the client ends the session - no error
		io.Copy(conn, conn2)
		reporterr(nil)
	}()

	return <-errchan
}

func startDebuggerServer(ctx context.Context, pid int, dbgInfo *DebuggerInfo) (*exec.Cmd, error) {
	// TODO: use squash's interfaces for a debug server
	cmd := exec.Command("dlv", dbgInfo.CmdlineGen(pid)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugw("dlv command", "cmd", cmd, "args", cmd.Args)

	err := cmd.Start()
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to start dlv")

		return nil, err
	}

	return cmd, nil
}

func startLocalServer() (net.Conn, error) {
	l, err := net.Listen("tcp", ListenHost+":"+OutPort)
	if err != nil {
		return nil, err
	}
	defer l.Close()
	conn, err := l.Accept()
	return conn, err

}
