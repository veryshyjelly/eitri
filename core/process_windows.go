package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Start suspended so the program cannot spawn children before it belongs to our
// job. Closing the job terminates remaining descendants on every exit path.
func runProcess(ctx context.Context, cmd *exec.Cmd) error {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return fmt.Errorf("create process job: %w", err)
	}
	defer windows.CloseHandle(job)
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return fmt.Errorf("set process job limits: %w", err)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED}
	cmd.Cancel = func() error {
		_ = windows.TerminateJobObject(job, 1)
		return cmd.Process.Kill() // also covers cancellation before job assignment
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	abort := func(err error) error { _ = cmd.Process.Kill(); _ = cmd.Wait(); return err }
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		return abort(fmt.Errorf("open process job member: %w", err))
	}
	err = windows.AssignProcessToJobObject(job, process)
	windows.CloseHandle(process)
	if err != nil {
		return abort(fmt.Errorf("assign process job: %w", err))
	}
	if err := ctx.Err(); err != nil {
		return abort(err)
	}
	if err := resumeProcess(uint32(cmd.Process.Pid)); err != nil {
		return abort(err)
	}
	return cmd.Wait()
}

func resumeProcess(pid uint32) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return fmt.Errorf("find suspended thread: %w", err)
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ThreadEntry32{Size: uint32(unsafe.Sizeof(windows.ThreadEntry32{}))}
	for err = windows.Thread32First(snapshot, &entry); err == nil; err = windows.Thread32Next(snapshot, &entry) {
		if entry.OwnerProcessID != pid {
			continue
		}
		thread, err := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
		if err != nil {
			return fmt.Errorf("open suspended thread: %w", err)
		}
		_, err = windows.ResumeThread(thread)
		windows.CloseHandle(thread)
		return err
	}
	if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return fmt.Errorf("suspended process thread not found")
	}
	return err
}
