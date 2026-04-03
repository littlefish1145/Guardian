package process

import (
	"context"
	"fmt"
	"time"

	"guardian/internal/config"
	"guardian/internal/log"
	"guardian/internal/metrics"
	"guardian/internal/signal"
)

type Manager struct {
	processes map[string]*Process
	router    *signal.Router
	logEngine *log.Engine
	metrics   *metrics.Server
	ctx       context.Context
	cancel    context.CancelFunc
	debug     bool
}

func NewManager(router *signal.Router, logEngine *log.Engine, metrics *metrics.Server, debug bool) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		processes: make(map[string]*Process),
		router:    router,
		logEngine: logEngine,
		metrics:   metrics,
		ctx:       ctx,
		cancel:    cancel,
		debug:     debug,
	}
}

func (m *Manager) StartAll(configs []config.ProcessConfig) error {
	depMap := make(map[string][]string)
	for _, cfg := range configs {
		depMap[cfg.Name] = cfg.DependsOn
	}

	started := make(map[string]bool)
	for len(started) < len(configs) {
		progress := false
		for _, cfg := range configs {
			if started[cfg.Name] {
				continue
			}

			canStart := true
			for _, dep := range depMap[cfg.Name] {
				if !started[dep] {
					canStart = false
					break
				}
			}

			if canStart {
				proc := New(cfg, m.logEngine, m.metrics, m.debug)
				m.processes[cfg.Name] = proc
				m.router.Register(cfg.Name, proc)

				if err := proc.Start(); err != nil {
					log.NewProcessLogger(cfg.Name).Error("Failed to start process: %v", err)
					return err
				}
				started[cfg.Name] = true
				progress = true
			}
		}

		if !progress && len(started) < len(configs) {
			return fmt.Errorf("circular dependency detected")
		}
	}

	return nil
}

func (m *Manager) StopAll() {
	for _, proc := range m.processes {
		proc.Stop(30 * time.Second)
	}
	m.cancel()
}

func (m *Manager) GetAllProcesses() map[string]*Process {
	return m.processes
}

func (m *Manager) GetProcess(name string) (*Process, bool) {
	proc, exists := m.processes[name]
	return proc, exists
}

func (m *Manager) Monitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.reapZombies()
			for name, proc := range m.processes {
				status := proc.GetStatus()
				m.metrics.RecordState(name, string(proc.State))
				m.metrics.RecordUptime(name, time.Since(proc.LastStart).Seconds())

				if proc.FailureCount > 0 {
					m.metrics.RecordHealthCheckFailure(name)
				}

				if proc.State != proc.LastState {
					resourceInfo := log.GetProcessResourceInfo(proc.Pid)
					log.LogProcessStatus(name, string(proc.State), proc.Pid, proc.Restarts, resourceInfo)

					proc.mu.Lock()
					proc.LastState = proc.State
					proc.mu.Unlock()

					// 检查刚退出的进程是否是僵尸
					if proc.State == StateStopped && log.CheckZombieProcess(proc.Pid) {
						log.NewProcessLogger(name).Warn("Process detected as zombie during cleanup (PID: %d)", proc.Pid)
					}
				}

				// 检查运行中的进程是否是僵尸
				if proc.State == StateRunning && log.CheckZombieProcess(proc.Pid) {
					log.NewProcessLogger(name).Warn("Process is in zombie state (PID: %d)", proc.Pid)
				}

				if abandoned, ok := status["abandoned"].(bool); ok && abandoned {
					log.NewProcessLogger(name).Warn("Process has been abandoned due to repeated zombie states")
				}
			}
		}
	}
}
