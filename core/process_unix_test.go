//go:build !windows

package core

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestProcessGroupCancellation(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("POSIX shell unavailable")
	}
	root := t.TempDir()
	record := filepath.Join(root, "child")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan processResult, 1)
	go func() {
		done <- execute(ctx, root, shell, []string{"-c", `sleep 30 & child=$!; printf '%s' "$child" > "$1"; wait`, "eitri-test", record}, "")
	}()
	var pid int
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(record)
		if err == nil {
			pid, _ = strconv.Atoi(string(data))
			if pid > 0 {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if pid == 0 {
		t.Fatal("child did not start")
	}
	defer syscall.Kill(pid, syscall.SIGKILL)
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("process tree cancellation stalled")
	}
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err == syscall.ESRCH {
			return
		}
		// Linux test containers may not reap orphan zombies. A zombie is terminated.
		if stat, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat")); err == nil && strings.Contains(string(stat), ") Z ") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("child process survived cancellation")
}
