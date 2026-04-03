package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"guardian/internal/process"
)

type Server struct {
	port    int
	manager *process.Manager
	server  *http.Server
}

func NewServer(port int, manager *process.Manager) *Server {
	return &Server{
		port:    port,
		manager: manager,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/processes", s.handleProcesses)
	mux.HandleFunc("/api/process/", s.handleProcess)
	mux.HandleFunc("/api/health", s.handleHealth)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleProcesses(w http.ResponseWriter, r *http.Request) {
	processes := s.manager.GetAllProcesses()
	
	response := make([]ProcessInfo, 0, len(processes))
	for name, proc := range processes {
		status := proc.GetStatus()
		info := ProcessInfo{
			Name:           name,
			State:          getString(status, "state"),
			PID:            getInt(status, "pid"),
			Restarts:       getInt(status, "restarts"),
			MemoryMB:       getProcessMemory(getInt(status, "pid")),
			LastStart:      getTime(status, "last_start"),
			FailureCount:   getInt(status, "failure_count"),
			HealthyCount:   getInt(status, "healthy_count"),
			ZombieRestarts: getInt(status, "zombie_restarts"),
			Abandoned:      getBool(status, "abandoned"),
		}
		response = append(response, info)
	}

	w.Header().Set("Content-Type", "application/json")
	data, _ := json.MarshalIndent(response, "", "  ")
	w.Write(data)
}

func (s *Server) handleProcess(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/api/process/"):]
	if name == "" {
		http.Error(w, "Process name required", http.StatusBadRequest)
		return
	}

	proc, exists := s.manager.GetProcess(name)
	if !exists {
		http.Error(w, "Process not found", http.StatusNotFound)
		return
	}

	status := proc.GetStatus()
	info := ProcessInfo{
		Name:           name,
		State:          getString(status, "state"),
		PID:            getInt(status, "pid"),
		Restarts:       getInt(status, "restarts"),
		MemoryMB:       getProcessMemory(getInt(status, "pid")),
		LastStart:      getTime(status, "last_start"),
		FailureCount:   getInt(status, "failure_count"),
		HealthyCount:   getInt(status, "healthy_count"),
		ZombieRestarts: getInt(status, "zombie_restarts"),
		Abandoned:      getBool(status, "abandoned"),
	}

	w.Header().Set("Content-Type", "application/json")
	data, _ := json.MarshalIndent(info, "", "  ")
	w.Write(data)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

type ProcessInfo struct {
	Name           string    `json:"name"`
	State          string    `json:"state"`
	PID            int       `json:"pid"`
	Restarts       int       `json:"restarts"`
	MemoryMB       float64   `json:"memory_mb"`
	LastStart      time.Time `json:"last_start"`
	FailureCount   int       `json:"failure_count"`
	HealthyCount   int       `json:"healthy_count"`
	ZombieRestarts int       `json:"zombie_restarts"`
	Abandoned      bool      `json:"abandoned"`
}

func getProcessMemory(pid int) float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Sys) / 1024 / 1024
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case process.State:
			return string(v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getTime(m map[string]interface{}, key string) time.Time {
	if val, ok := m[key]; ok {
		if t, ok := val.(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}
