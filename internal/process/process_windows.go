// +build windows

package process

import (
	"syscall"
	"time"

	"guardian/internal/log"
)

func (p *Process) killProcessGroup(pid int, timeout time.Duration) {
	if pid <= 0 {
		return
	}

	if p.Debug {
		log.NewProcessLogger(p.Config.Name).Debug("Killing process on Windows: PID %d", pid)
	}

	p.Cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		done <- p.Cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		p.Cmd.Process.Kill()
		<-done
		log.LogProcessStopped(p.Config.Name, pid, "killed after timeout")
	case <-done:
	}
}

// waitForExit Windows 版本使用 Wait 回收进程
func (p *Process) waitForExit() {
	p.Cmd.Wait()

	p.mu.Lock()
	p.LastState = p.State
	if p.State != StateStopping {
		p.State = StateStopped
		p.Restarts++
	}
	p.mu.Unlock()

	p.cleanup()
}
