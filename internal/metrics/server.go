package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	mu               sync.RWMutex
	port             int
	processRestarts  *prometheus.CounterVec
	processState     *prometheus.GaugeVec
	processHealthy   *prometheus.GaugeVec
	processUptime    *prometheus.GaugeVec
	healthCheckFails *prometheus.CounterVec
	server           *http.Server
}

func NewServer(port int) *Server {
	s := &Server{
		port: port,
		processRestarts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "guardian_process_restarts_total",
				Help: "Total number of process restarts",
			},
			[]string{"process"},
		),
		processState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "guardian_process_state",
				Help: "Process state (0=stopped, 1=running, 2=failed)",
			},
			[]string{"process"},
		),
		processHealthy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "guardian_process_healthy",
				Help: "Process health check status",
			},
			[]string{"process"},
		),
		processUptime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "guardian_process_uptime_seconds",
				Help: "Process uptime in seconds",
			},
			[]string{"process"},
		),
		healthCheckFails: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "guardian_health_check_failures_total",
				Help: "Total health check failures",
			},
			[]string{"process"},
		),
	}

	prometheus.MustRegister(s.processRestarts)
	prometheus.MustRegister(s.processState)
	prometheus.MustRegister(s.processHealthy)
	prometheus.MustRegister(s.processUptime)
	prometheus.MustRegister(s.healthCheckFails)

	return s
}

func (s *Server) RecordRestart(procName string) {
	s.processRestarts.WithLabelValues(procName).Inc()
}

func (s *Server) RecordState(procName string, state string) {
	stateVal := 0.0
	switch state {
	case "running":
		stateVal = 1.0
	case "failed":
		stateVal = 2.0
	}
	s.processState.WithLabelValues(procName).Set(stateVal)
}

func (s *Server) RecordHealthy(procName string, healthy bool) {
	val := 0.0
	if healthy {
		val = 1.0
	}
	s.processHealthy.WithLabelValues(procName).Set(val)
}

func (s *Server) RecordUptime(procName string, seconds float64) {
	s.processUptime.WithLabelValues(procName).Set(seconds)
}

func (s *Server) RecordHealthCheckFailure(procName string) {
	s.healthCheckFails.WithLabelValues(procName).Inc()
}

func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	addr := fmt.Sprintf(":%d", s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		fmt.Printf("Starting metrics server on %s\n", s.server.Addr)
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
