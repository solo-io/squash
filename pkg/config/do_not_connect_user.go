// +build windows

package config

import (
	"fmt"

	squashv1 "github.com/solo-io/squash/pkg/api/v1"
)

// since dlv is proxied we are using a pseudoterminal to correctly handle
// control characters. However, the pseudoterminal we are using does not have
// a windows implementation. For now, skip connecting for windows users.
func (s *Squash) connectUser(da *squashv1.DebugAttachment, remoteDbgPort int) error {
	if s.Machine {
		return nil
	}
	fmt.Println("Interactive command line is not currently available on Windows, please use the vscode extension or include the --machine flag.")
	return nil
}
