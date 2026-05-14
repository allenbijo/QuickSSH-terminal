//go:build !windows

package ssh

import (
	"os/exec"
	"syscall"
)

func hideWindow(_ *exec.Cmd) {}

func killGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(syscall.SIGTERM)
}
