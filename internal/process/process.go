package process

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"guardian/internal/config"
	"guardian/internal/log"
	"guardian/internal/metrics"
)

type State string

const (
	StateStopped   State = "stopped"
	StateStarting  State = "starting"
	StateRunning   State = "running"
	StateStopping  State = "stopping"
	StateFailed    State = "failed"
	StateReclaimed State = "reclaimed"
)

type Process struct {
	mu             sync.RWMutex
	Config         config.ProcessConfig
	State          State
	LastState      State
	Cmd            *exec.Cmd
	Pid            int
	Restarts       int
	LastStart      time.Time
	FailureCount   int
	HealthyCount   int
	ZombieRestarts int
	Abandoned      bool
	Debug          bool
	logEngine      *log.Engine
	metricsServer  *metrics.Server
	cancelCtx      context.CancelFunc
	ctx            context.Context
}

func New(cfg config.ProcessConfig, logEngine *log.Engine, metricsServer *metrics.Server, debug bool) *Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &Process{
		Config:        cfg,
		State:         StateStopped,
		LastState:     StateStopped,
		Debug:         debug,
		logEngine:     logEngine,
		metricsServer: metricsServer,
		ctx:           ctx,
		cancelCtx:     cancel,
	}
}

func (p *Process) resetContext() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelCtx != nil {
		p.cancelCtx()
	}
	p.ctx, p.cancelCtx = context.WithCancel(context.Background())
}

func (p *Process) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.State == StateRunning {
		return fmt.Errorf("process %s already running", p.Config.Name)
	}

	if p.Abandoned {
		return fmt.Errorf("process %s has been abandoned due to repeated zombie states", p.Config.Name)
	}

	p.State = StateStarting
	p.Cmd = exec.CommandContext(p.ctx, p.Config.Command[0], p.Config.Command[1:]...)
	p.Cmd.Dir = p.Config.WorkingDir

	if p.Config.Logging.Stdout {
		stdout, _ := p.Cmd.StdoutPipe()
		go p.logEngine.Capture(p.ctx, p.Config.Name, "stdout", stdout)
	}
	if p.Config.Logging.Stderr {
		stderr, _ := p.Cmd.StderrPipe()
		go p.logEngine.Capture(p.ctx, p.Config.Name, "stderr", stderr)
	}

	if err := p.Cmd.Start(); err != nil {
		p.LastState = p.State
		p.State = StateStopped
		log.NewProcessLogger(p.Config.Name).Error("Failed to start process: %v", err)
		return err
	}

	p.Pid = p.Cmd.Process.Pid
	p.State = StateRunning
	p.LastStart = time.Now()
	p.FailureCount = 0

	resourceInfo := log.GetProcessResourceInfo(p.Pid)
	log.LogProcessStarted(p.Config.Name, p.Pid, resourceInfo)

	go p.healthCheckLoop()
	go p.waitForExit()

	return nil
}

func (p *Process) Stop(timeout time.Duration) error {
	p.mu.Lock()
	if p.State != StateRunning {
		p.mu.Unlock()
		return nil
	}
	p.State = StateStopping
	pid := p.Pid
	p.mu.Unlock()

	log.LogProcessStopped(p.Config.Name, pid, "graceful shutdown initiated")

	p.killProcessGroup(pid, timeout)

	log.LogProcessStopped(p.Config.Name, pid, "stopped successfully")
	return nil
}

func (p *Process) healthCheckLoop() {
	if p.Config.HealthCheck.Type == "" {
		return
	}

	ticker := time.NewTicker(p.Config.HealthCheck.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.mu.RLock()
			if p.State != StateRunning {
				p.mu.RUnlock()
				return
			}
			p.mu.RUnlock()

			healthy, reason := p.checkHealthWithReason()
			if healthy {
				p.mu.Lock()
				p.HealthyCount++
				p.FailureCount = 0
				p.mu.Unlock()
			} else {
				p.mu.Lock()
				p.FailureCount++
				currentFailures := p.FailureCount
				threshold := p.Config.HealthCheck.FailureThreshold
				debug := p.Debug
				p.mu.Unlock()

				if debug {
					log.LogHealthCheckFailedDebug(p.Config.Name, p.Pid, currentFailures, threshold, reason)
				} else {
					log.LogHealthCheckFailed(p.Config.Name, p.Pid, currentFailures, threshold)
				}

				if threshold != -1 && currentFailures > threshold {
					p.TryRestart()
					return
				}
			}
		}
	}
}

func (p *Process) checkHealthWithReason() (bool, string) {
	if p.Config.HealthCheck.Type == "http" {
		client := &http.Client{Timeout: p.Config.HealthCheck.Timeout}
		resp, err := client.Get(p.Config.HealthCheck.Endpoint)
		if err != nil {
			return false, fmt.Sprintf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return false, fmt.Sprintf("HTTP status code: %d", resp.StatusCode)
		}
		return true, ""
	}
	return true, ""
}

func (p *Process) checkHealth() bool {
	healthy, _ := p.checkHealthWithReason()
	return healthy
}

func (p *Process) cleanup() {
	if p.Cmd != nil && p.Cmd.Process != nil {
		p.Cmd.Process.Release()
	}

	if p.Debug {
		log.NewProcessLogger(p.Config.Name).Debug("Process resources cleaned up")
	}
}

func (p *Process) TryRestart() error {
	p.mu.RLock()
	abandoned := p.Abandoned
	currentRestarts := p.Restarts
	maxRestarts := p.Config.RestartPolicy.MaxRestarts
	p.mu.RUnlock()

	if abandoned {
		log.NewProcessLogger(p.Config.Name).Error("Process has been abandoned due to repeated zombie states")
		return fmt.Errorf("process %s has been abandoned", p.Config.Name)
	}

	if currentRestarts >= maxRestarts {
		p.mu.Lock()
		p.Abandoned = true
		p.LastState = p.State
		p.State = StateStopped
		p.mu.Unlock()

		log.NewProcessLogger(p.Config.Name).Warn("Stopping process: exceeded max restarts (%d/%d)", currentRestarts, maxRestarts)
		p.Stop(5 * time.Second)
		return fmt.Errorf("process %s stopped: exceeded max restarts", p.Config.Name)
	}

	if p.Config.RestartPolicy.ZombieCheckEnabled && log.CheckZombieProcess(p.Pid) {
		p.mu.Lock()
		p.ZombieRestarts++
		currentZombieRestarts := p.ZombieRestarts
		maxZombieRestarts := p.Config.RestartPolicy.ZombieMaxRestarts
		p.mu.Unlock()

		log.LogZombieDetected(p.Config.Name, p.Pid, currentZombieRestarts, maxZombieRestarts)

		if currentZombieRestarts >= maxZombieRestarts {
			p.mu.Lock()
			p.Abandoned = true
			p.LastState = p.State
			p.State = StateReclaimed
			p.mu.Unlock()

			log.LogAbandonRestart(p.Config.Name, p.Pid,
				fmt.Sprintf("exceeded max zombie restarts (%d/%d)", currentZombieRestarts, maxZombieRestarts))

			p.Stop(5 * time.Second)
			return fmt.Errorf("process %s abandoned due to repeated zombie states", p.Config.Name)
		}
	}

	log.NewProcessLogger(p.Config.Name).Info("Restarting process (attempt %d/%d)", currentRestarts+1, maxRestarts)

	p.mu.Lock()
	p.cancelCtx()
	p.ctx, p.cancelCtx = context.WithCancel(context.Background())
	p.mu.Unlock()

	p.Stop(5 * time.Second)
	time.Sleep(time.Second)

	err := p.Start()
	if err != nil {
		log.NewProcessLogger(p.Config.Name).Error("Failed to restart: %v", err)
	}
	return err
}

func (p *Process) Restart() error {
	p.mu.RLock()
	abandoned := p.Abandoned
	p.mu.RUnlock()

	if abandoned {
		log.LogAbandonRestart(p.Config.Name, p.Pid, "process has been abandoned due to repeated zombie states")
		return fmt.Errorf("process %s has been abandoned", p.Config.Name)
	}

	if p.Config.RestartPolicy.ZombieCheckEnabled && log.CheckZombieProcess(p.Pid) {
		p.mu.Lock()
		p.ZombieRestarts++
		currentZombieRestarts := p.ZombieRestarts
		maxZombieRestarts := p.Config.RestartPolicy.ZombieMaxRestarts
		p.mu.Unlock()

		log.LogZombieDetected(p.Config.Name, p.Pid, currentZombieRestarts, maxZombieRestarts)

		if currentZombieRestarts >= maxZombieRestarts {
			p.mu.Lock()
			p.Abandoned = true
			p.LastState = p.State
			p.State = StateReclaimed
			p.mu.Unlock()

			log.LogAbandonRestart(p.Config.Name, p.Pid,
				fmt.Sprintf("exceeded max zombie restarts (%d/%d)", currentZombieRestarts, maxZombieRestarts))

			p.Stop(5 * time.Second)
			return fmt.Errorf("process %s abandoned due to repeated zombie states", p.Config.Name)
		}
	}

	p.Stop(5 * time.Second)
	time.Sleep(time.Second)
	return p.Start()
}

func (p *Process) GetStatus() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"name":            p.Config.Name,
		"state":           p.State,
		"pid":             p.Pid,
		"restarts":        p.Restarts,
		"last_start":      p.LastStart,
		"failure_count":   p.FailureCount,
		"healthy_count":   p.HealthyCount,
		"zombie_restarts": p.ZombieRestarts,
		"abandoned":       p.Abandoned,
	}
}
