//go:build !windows

package memory

import (
	"os/exec"
	"runtime"
	"syscall"
)

// PeakMemory retrieves the peak resident set size (RSS) memory usage of an
// executed command from the OS process state.
//
// Returns the memory usage in kilobytes and a boolean indicating whether the
// metric was successfully retrieved. (Note: The Windows implementation currently
// returns 0 and false as a placeholder).
func PeakMemory(cmd *exec.Cmd) (uint64, bool) {
	if cmd == nil || cmd.ProcessState == nil {
		return 0, false
	}
	usage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage)
	if !ok {
		return 0, false
	}

	if runtime.GOOS == "linux" {
		// Linux reports KB
		return uint64(usage.Maxrss), true
	}

	// macOS reports bytes
	return uint64(usage.Maxrss) / 1024, true
}
