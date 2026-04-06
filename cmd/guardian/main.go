package main

import (
	"context"
	"flag"
	"os"
	"time"

	"guardian/internal/api"
	"guardian/internal/config"
	"guardian/internal/log"
	"guardian/internal/metrics"
	"guardian/internal/process"
	"guardian/internal/signal"
)

func main() {
	configPath := flag.String("config", "guardian.yaml", "Path to configuration file")
	flag.Parse()

	sysLogger := log.NewSystemLogger()
	sysLogger.Info("Guardian process manager starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go config.Watch(ctx, *configPath)

	cfg := config.Get()
	// if err != nil {
	// 	sysLogger.Error("Failed to load config: %v", err)
	// 	os.Exit(1)
	// }

	logEngine, err := log.NewEngine(cfg.Global.LogDir, cfg.Global.MaxLogSizeMB, cfg.Global.MaxLogBackups, true)
	if err != nil {
		sysLogger.Error("Failed to initialize log engine: %v", err)
		os.Exit(1)
	}

	sysLogger.Info("Log engine initialized, directory: %s", cfg.Global.LogDir)

	metricsServer := metrics.NewServer(cfg.Global.MetricsPort)
	metricsServer.Start()
	sysLogger.Info("Metrics server started on port %d", cfg.Global.MetricsPort)

	router := signal.NewRouter()

	procMgr := process.NewManager(router, logEngine, metricsServer, cfg.Global.Debug)

	apiServer := api.NewServer(cfg.Global.APIPort, procMgr)
	if err := apiServer.Start(); err != nil {
		sysLogger.Error("Failed to start API server: %v", err)
		os.Exit(1)
	}
	sysLogger.Info("API server started on port %d", cfg.Global.APIPort)

	if cfg.Global.Debug {
		sysLogger.Info("Debug mode enabled")
	}

	router.SetShutdownHandlers(
		func() error {
			sysLogger.Info("Stopping all processes...")
			procMgr.StopAll()
			return nil
		},
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return metricsServer.Stop(ctx)
		},
		func() error {
			sysLogger.Info("Closing log engine...")
			return logEngine.Close()
		},
	)

	router.Start()

	sysLogger.Info("Starting %d processes...", len(cfg.Processes))
	if err := procMgr.StartAll(cfg.Processes); err != nil {
		sysLogger.Error("Failed to start processes: %v", err)
		os.Exit(1)
	}

	sysLogger.Info("All processes started successfully, entering monitor mode")
	go procMgr.Monitor()

	<-router.Done()
	sysLogger.Info("Guardian shutdown complete")
}
