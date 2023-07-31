package util

import (
	"fmt"
	"os"
)

func KillProcessByPID(pid int) error {
	return SendSignalByPid(pid, os.Kill)
}

func SendSignalByPid(pid int, sig os.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %v", pid, err)
	}

	err = process.Signal(sig)
	if err != nil {
		return fmt.Errorf("failed to send signal:%v, to PID %d", sig, pid)
	}

	return nil
}
