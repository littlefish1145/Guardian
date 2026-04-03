// +build !windows

package process

import (
	"syscall"

	"guardian/internal/log"
)

func (m *Manager) reapZombies() {
	for {
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
		if err != nil || pid <= 0 {
			break
		}

		processName := "SYSTEM"
		for name, proc := range m.processes {
			if proc.Pid == pid {
				processName = name
				break
			}
		}
		log.NewProcessLogger(processName).Info("Detected and reaped zombie process (PID: %d)", pid)
	}
}
