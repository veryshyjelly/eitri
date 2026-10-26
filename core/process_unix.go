//go:build !windows

package core

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func runProcess(_ context.Context, cmd *exec.Cmd) error {
	defer cleanupProcess(cmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return cmd.Run()
}

func cleanupProcess(cmd *exec.Cmd) {
	// Also terminate descendants left behind after an otherwise successful parent.
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
