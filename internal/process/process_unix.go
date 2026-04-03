// +build !windows

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

	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		if p.Debug {
			log.NewProcessLogger(p.Config.Name).Debug("Cannot get process group, killing single process")
		}
		p.Cmd.Process.Signal(syscall.SIGTERM)
	} else {
		if p.Debug {
			log.NewProcessLogger(p.Config.Name).Debug("Killing process group: -%d", pgid)
		}
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	done := make(chan error, 1)
	go func() {
		done <- p.Cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			p.Cmd.Process.Kill()
		}
		<-done
		log.LogProcessStopped(p.Config.Name, pid, "killed after timeout")
	case <-done:
	}
}

// waitForExit 使用 Wait 回收进程
func (p *Process) waitForExit() {
	// 使用 Wait 阻塞等待进程退出
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
