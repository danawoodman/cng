package internal

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func forkProcess(prog string, args []string) *exec.Cmd {
	cmd := exec.Command(prog, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	return cmd
}

func killProcess(proc *os.Process) error {
	kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(proc.Pid))
	err := kill.Run()
	if err != nil {
		return err
	}
	return nil
}
