package util

import (
	"fmt"
	"os"
)

func KillProcessByPID(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %v", pid, err)
	}

	err = process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %v", pid, err)
	}

	return nil
}
