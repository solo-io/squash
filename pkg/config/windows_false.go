// +build !windows

package config

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh/terminal"
)

func (s *Squash) callLocalDebuggerCommand(dbgCmd *exec.Cmd) error {
	return ptyWrap(dbgCmd)
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
				contextutils.LoggerFrom(context.TODO()).Infof("error resizing pty: %s", err)
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
