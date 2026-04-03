package signal

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Router struct {
	mu        sync.RWMutex
	processes map[string]interface{}
	pgid      int
	handlers  []func() error
	done      chan struct{}
}

func NewRouter() *Router {
	return &Router{
		processes: make(map[string]interface{}),
		pgid:      os.Getpid(),
		done:      make(chan struct{}),
	}
}

func (r *Router) Register(name string, proc interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processes[name] = proc
}

func (r *Router) SetShutdownHandlers(handlers ...func() error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = handlers
}

func (r *Router) Start() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	go func() {
		for sig := range sigs {
			switch sig {
			case os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT:
				fmt.Printf("Received %v, gracefully stopping processes...\n", sig)
				r.shutdownAll()
				close(r.done)
				os.Exit(0)
			default:
				fmt.Printf("Received signal: %v\n", sig)
			}
		}
	}()
}

func (r *Router) shutdownAll() {
	r.mu.RLock()
	procs := make([]interface{}, 0, len(r.processes))
	for _, p := range r.processes {
		procs = append(procs, p)
	}
	handlers := append([]func() error{}, r.handlers...)
	r.mu.RUnlock()

	var wg sync.WaitGroup
	for _, p := range procs {
		wg.Add(1)
		go func(proc interface{}) {
			defer wg.Done()
			if stopper, ok := proc.(interface{ Stop(time.Duration) error }); ok {
				stopper.Stop(30 * time.Second)
			}
		}(p)
	}
	wg.Wait()

	for _, h := range handlers {
		h()
	}
}

func (r *Router) GetPgid() int {
	return r.pgid
}

func (r *Router) Done() <-chan struct{} {
	return r.done
}
