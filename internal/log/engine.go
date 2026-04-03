package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Entry struct {
	Timestamp string `json:"ts"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Stream    string `json:"stream"`
	Message   string `json:"message"`
}

type Engine struct {
	mu       sync.Mutex
	file     *os.File
	maxSize  int64
	maxBkups int
	logDir   string
}

func NewEngine(logDir string, maxSizeMB int, maxBackups int, compress bool) (*Engine, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	return &Engine{
		logDir:   logDir,
		maxSize:  int64(maxSizeMB) * 1024 * 1024,
		maxBkups: maxBackups,
	}, nil
}

func (e *Engine) openFile(name string) error {
	path := filepath.Join(e.logDir, name+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	e.file = f
	return nil
}

func (e *Engine) rotate(name string) error {
	if e.file != nil {
		e.file.Close()
	}

	logPath := filepath.Join(e.logDir, name+".log")
	stat, err := os.Stat(logPath)
	if err != nil || stat.Size() < e.maxSize {
		return e.openFile(name)
	}

	for i := e.maxBkups - 1; i > 0; i-- {
		oldPath := filepath.Join(e.logDir, fmt.Sprintf("%s.%d.log", name, i))
		newPath := filepath.Join(e.logDir, fmt.Sprintf("%s.%d.log", name, i+1))
		os.Rename(oldPath, newPath)
	}

	newPath := filepath.Join(e.logDir, fmt.Sprintf("%s.1.log", name))
	os.Rename(logPath, newPath)

	return e.openFile(name)
}

func (e *Engine) Capture(ctx context.Context, procName, stream string, reader io.Reader) {
	e.mu.Lock()
	if e.file == nil {
		e.openFile(procName)
	}
	e.mu.Unlock()

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := reader.Read(buf)
		if n > 0 {
			entry := Entry{
				Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
				Level:     "info",
				Source:    procName,
				Stream:    stream,
				Message:   string(buf[:n]),
			}
			data, _ := json.Marshal(entry)

			e.mu.Lock()
			e.file.Write(data)
			e.file.WriteString("\n")
			e.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.file == nil {
		return nil
	}

	err := e.file.Close()
	e.file = nil
	return err
}
