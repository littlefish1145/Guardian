package log

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelDebug LogLevel = "DEBUG"
)

type ProcessLogger struct {
	processName string
}

func NewProcessLogger(processName string) *ProcessLogger {
	return &ProcessLogger{
		processName: processName,
	}
}

func (l *ProcessLogger) log(level LogLevel, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] [%s] [%s] %s\n", timestamp, level, l.processName, message)
}

func (l *ProcessLogger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *ProcessLogger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *ProcessLogger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

func (l *ProcessLogger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

type SystemLogger struct{}

func NewSystemLogger() *SystemLogger {
	return &SystemLogger{}
}

func (l *SystemLogger) log(level LogLevel, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] [%s] [SYSTEM] %s\n", timestamp, level, message)
}

func (l *SystemLogger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *SystemLogger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *SystemLogger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

func (l *SystemLogger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

type ResourceInfo struct {
	CPUUsage    float64
	MemoryMB    float64
	GoroutineCount int
}

func GetSystemResourceInfo() ResourceInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ResourceInfo{
		CPUUsage:       0,
		MemoryMB:       float64(m.Sys) / 1024 / 1024,
		GoroutineCount: runtime.NumGoroutine(),
	}
}

func GetProcessResourceInfo(pid int) ResourceInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ResourceInfo{
		CPUUsage:       0,
		MemoryMB:       float64(m.Sys) / 1024 / 1024,
		GoroutineCount: runtime.NumGoroutine(),
	}
}

func FormatResourceInfo(info ResourceInfo) string {
	return fmt.Sprintf("Memory: %.2fMB, Goroutines: %d", info.MemoryMB, info.GoroutineCount)
}

func LogProcessStatus(processName string, state string, pid int, restarts int, resourceInfo ResourceInfo) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	resourceStr := FormatResourceInfo(resourceInfo)
	fmt.Printf("[%s] [INFO] [%s] State: %s, PID: %d, Restarts: %d, %s\n",
		timestamp, processName, state, pid, restarts, resourceStr)
}

func LogZombieDetected(processName string, pid int, zombieRestarts int, maxRestarts int) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Printf("[%s] [WARN] [%s] Zombie process detected! PID: %d, Zombie restarts: %d/%d\n",
		timestamp, processName, pid, zombieRestarts, maxRestarts)
}

func LogAbandonRestart(processName string, pid int, reason string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Printf("[%s] [ERROR] [%s] Abandoning restart for PID: %d, Reason: %s\n",
		timestamp, processName, pid, reason)
}

func LogProcessStarted(processName string, pid int, resourceInfo ResourceInfo) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	resourceStr := FormatResourceInfo(resourceInfo)
	fmt.Printf("[%s] [INFO] [%s] Process started successfully, PID: %d, %s\n",
		timestamp, processName, pid, resourceStr)
}

func LogProcessStopped(processName string, pid int, reason string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Printf("[%s] [INFO] [%s] Process stopped, PID: %d, Reason: %s\n",
		timestamp, processName, pid, reason)
}

func LogHealthCheckFailed(processName string, pid int, failureCount int, threshold int) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	if failureCount > threshold {
		fmt.Printf("[%s] [WARN] [%s] Health check failed (exceeded), PID: %d, Failures: %d (threshold: %d)\n",
			timestamp, processName, pid, failureCount, threshold)
	} else {
		fmt.Printf("[%s] [WARN] [%s] Health check failed, PID: %d, Failures: %d/%d\n",
			timestamp, processName, pid, failureCount, threshold)
	}
}

func LogHealthCheckFailedDebug(processName string, pid int, failureCount int, threshold int, reason string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	if failureCount > threshold {
		fmt.Printf("[%s] [WARN] [%s] Health check failed (exceeded), PID: %d, Failures: %d (threshold: %d), Reason: %s\n",
			timestamp, processName, pid, failureCount, threshold, reason)
	} else {
		fmt.Printf("[%s] [WARN] [%s] Health check failed, PID: %d, Failures: %d/%d, Reason: %s\n",
			timestamp, processName, pid, failureCount, threshold, reason)
	}
}

func isZombie(pid int) bool {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return false
	}

	var (
		pidField    int
		comm        string
		state       string
		rest        string
	)

	_, err = fmt.Sscanf(string(data), "%d %s %s %s", &pidField, &comm, &state, &rest)
	if err != nil {
		return false
	}

	return state == "Z"
}

func CheckZombieProcess(pid int) bool {
	return isZombie(pid)
}
