package config

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Global struct {
		LogLevel      string `yaml:"log_level"`
		MetricsPort   int    `yaml:"metrics_port"`
		APIPort       int    `yaml:"api_port"`
		LogDir        string `yaml:"log_dir"`
		MaxLogSizeMB  int    `yaml:"max_log_size_mb"`
		MaxLogBackups int    `yaml:"max_log_backups"`
		Debug         bool   `yaml:"debug"`
	} `yaml:"global"`
	Processes []ProcessConfig `yaml:"processes"`
}

type ProcessConfig struct {
	Name          string         `yaml:"name"`
	Command       []string       `yaml:"command"`
	WorkingDir    string         `yaml:"working_dir"`
	HealthCheck   HealthCheck    `yaml:"health_check"`
	RestartPolicy RestartPolicy  `yaml:"restart_policy"`
	Resources     ResourceLimits `yaml:"resources"`
	Logging       LoggingConfig  `yaml:"logging"`
	DependsOn     []string       `yaml:"depends_on"`
}

type HealthCheck struct {
	Type             string        `yaml:"type"`
	Endpoint         string        `yaml:"endpoint"`
	Interval         time.Duration `yaml:"interval"`
	Timeout          time.Duration `yaml:"timeout"`
	FailureThreshold int           `yaml:"failure_threshold"`
}

type RestartPolicy struct {
	Policy             string        `yaml:"policy"`
	MaxRestarts        int           `yaml:"max_restarts"`
	BaseDelay          time.Duration `yaml:"base_delay"`
	MaxDelay           time.Duration `yaml:"max_delay"`
	ZombieMaxRestarts  int           `yaml:"zombie_max_restarts"`
	ZombieCheckEnabled bool          `yaml:"zombie_check_enabled"`
}

type ResourceLimits struct {
	MemoryLimit string `yaml:"memory_limit"`
	CPUQuota    string `yaml:"cpu_quota"`
	NoFile      int    `yaml:"nofile"`
}

type LoggingConfig struct {
	Stdout     bool   `yaml:"stdout"`
	Stderr     bool   `yaml:"stderr"`
	File       string `yaml:"file"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

var current atomic.Pointer[Config]

func Get() *Config { return current.Load() }

func Watch(ctx context.Context, path string) error {
	if _, err := load(path); err != nil {
		return err
	}

	w, _ := fsnotify.NewWatcher()
	defer w.Close()
	w.Add(filepath.Dir(path))

	var t *time.Timer
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-w.Events:
			if filepath.Base(e.Name) == filepath.Base(path) && (e.Has(fsnotify.Write) || e.Has(fsnotify.Create)) {
				if t != nil {
					t.Stop()
				}
				t = time.AfterFunc(300*time.Millisecond, func() {
					if _, err := load(path); err != nil {
					}
				})
			}
		}
	}
}

func load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	setDefaults(cfg)
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Global.LogLevel == "" {
		cfg.Global.LogLevel = "info"
	}
	if cfg.Global.MetricsPort == 0 {
		cfg.Global.MetricsPort = 9090
	}
	if cfg.Global.APIPort == 0 {
		cfg.Global.APIPort = 8080
	}
	if cfg.Global.LogDir == "" {
		cfg.Global.LogDir = "/var/log/guardian"
	}
	if cfg.Global.MaxLogSizeMB == 0 {
		cfg.Global.MaxLogSizeMB = 100
	}
	if cfg.Global.MaxLogBackups == 0 {
		cfg.Global.MaxLogBackups = 5
	}

	for i := range cfg.Processes {
		if cfg.Processes[i].HealthCheck.Interval == 0 {
			cfg.Processes[i].HealthCheck.Interval = 10 * time.Second
		}
		if cfg.Processes[i].HealthCheck.Timeout == 0 {
			cfg.Processes[i].HealthCheck.Timeout = 3 * time.Second
		}
		if cfg.Processes[i].HealthCheck.FailureThreshold == 0 {
			cfg.Processes[i].HealthCheck.FailureThreshold = 3
		}
		if cfg.Processes[i].RestartPolicy.Policy == "" {
			cfg.Processes[i].RestartPolicy.Policy = "on-failure"
		}
		if cfg.Processes[i].RestartPolicy.MaxRestarts == 0 {
			cfg.Processes[i].RestartPolicy.MaxRestarts = 5
		}
		if cfg.Processes[i].RestartPolicy.ZombieMaxRestarts == 0 {
			cfg.Processes[i].RestartPolicy.ZombieMaxRestarts = 3
		}
		if !cfg.Processes[i].RestartPolicy.ZombieCheckEnabled {
			cfg.Processes[i].RestartPolicy.ZombieCheckEnabled = true
		}
		if cfg.Processes[i].Logging.MaxSizeMB == 0 {
			cfg.Processes[i].Logging.MaxSizeMB = cfg.Global.MaxLogSizeMB
		}
		if cfg.Processes[i].Logging.MaxBackups == 0 {
			cfg.Processes[i].Logging.MaxBackups = cfg.Global.MaxLogBackups
		}
	}
}
