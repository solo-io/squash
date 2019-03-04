// +build windows

package config

import (
	"fmt"
	"os/exec"

	"github.com/solo-io/squash/pkg/debuggers/local"
)

// since dlv is proxied we are using a pseudoterminal to correctly handle
// control characters. However, the pseudoterminal we are using does not have
// a windows implementation. For now, skip connecting for windows users.
func (s *Squash) callLocalDebuggerCommand(dbgCmd *exec.Cmd) error {
	debugger := local.GetParticularDebugger(s.Debugger)
	if warning := debugger.WindowsSupportWarning(); warning != "" {
		return fmt.Errorf(warning)
	}
	return dbgCmd.Run()
}
