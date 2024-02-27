//go:build !windows

package internal

import (
	"os"
	"os/exec"
	"syscall"
)

func forkProcess(prog string, args []string) *exec.Cmd {
	cmd := exec.Command(prog, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd
}

func killProcess(proc *os.Process) error {
	pgid, err := syscall.Getpgid(proc.Pid)
	if err == nil {
		err = syscall.Kill(-pgid, syscall.SIGINT)
		if err != nil {
			return err
		}
	}
	return nil
}
