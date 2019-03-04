// +build !windows

package config

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kr/pty"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"

	squashv1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/debuggers/local"
)

func (s *Squash) connectUser(da *squashv1.DebugAttachment, remoteDbgPort int) error {
	if s.Machine {
		return nil
	}
	debugger := local.GetParticularDebugger(s.Debugger)
	kubectlCmd := debugger.GetRemoteConnectionCmd(
		da.PlankName,
		s.SquashNamespace,
		s.Pod,
		s.Namespace,
		s.LocalPort,
		remoteDbgPort,
	)
	// Starting port forward in background.
	if err := kubectlCmd.Start(); err != nil {
		// s.printError(createdPodName)
		return err
	}
	// kill the kubectl port-forward process on exit to free the port
	// this defer must be called after Start() initializes Process
	defer kubectlCmd.Process.Kill()

	// Delaying to allow port forwarding to complete.
	time.Sleep(5 * time.Second)
	if os.Getenv("DEBUG_SELF") != "" {
		fmt.Println("FOR DEBUGGING SQUASH'S DEBUGGER CONTAINER:")
		fmt.Println("TODO")
		// s.printError(createdPod)
	}

	dbgCmd := debugger.GetDebugCmd(s.LocalPort)
	if err := ptyWrap(dbgCmd); err != nil {
		// s.printError(createdPodName)
		return err
	}
	return nil
}

func ptyWrap(c *exec.Cmd) error {

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)

	return nil
}
