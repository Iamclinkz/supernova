//go:build windows
// +build windows

package processor_plugin_shell

import (
	"os/exec"
)

func setCmdAttr(cmd *exec.Cmd, config map[string]string) error {
	return nil
}
