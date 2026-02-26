package hasher

import (
	"errors"
	"runtime"
	"sync"
)

var ErrQueueFull = errors.New("hasher: job queue is full")

// Job is a closed interface — only types in this package can implement it.
type Job interface {
	execute()
}

// Dispatcher manages a fixed pool of worker goroutines that process hash jobs.
type Dispatcher struct {
	jobs chan Job
	wg   sync.WaitGroup
}

// NewDispatcher creates a Dispatcher with a buffered job channel sized at 2 * NumCPU.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		jobs: make(chan Job, 2*runtime.NumCPU()),
	}
}

// Start launches runtime.NumCPU() worker goroutines.
func (d *Dispatcher) Start() {
	n := runtime.NumCPU()
	d.wg.Add(n)
	for range n {
		go d.worker()
	}
}

// Stop closes the job channel and waits for all workers to drain.
func (d *Dispatcher) Stop() {
	close(d.jobs)
	d.wg.Wait()
}

// Submit enqueues a job. Returns ErrQueueFull if the buffer is at capacity.
func (d *Dispatcher) Submit(job Job) error {
	select {
	case d.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for job := range d.jobs {
		job.execute()
	}
}
