package scanner

import (
	"os"
	"sync"

	"github.com/Sho2010/dup-finder/internal/models"
)

// ScanJob represents a file to be scanned
type ScanJob struct {
	Path      string      // File path
	Directory string      // Root directory
	Info      os.FileInfo // File info
}

// ScanResult represents the result of scanning a file
type ScanResult struct {
	FileInfo models.FileInfo
	Error    error
}

// WorkerPool manages parallel file processing
type WorkerPool struct {
	numWorkers int
	jobs       chan ScanJob
	results    chan ScanResult
	wg         sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan ScanJob, numWorkers*2),
		results:    make(chan ScanResult, numWorkers*10),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// worker processes jobs from the jobs channel
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for job := range wp.jobs {
		fileInfo := models.FileInfo{
			Path:      job.Path,
			Directory: job.Directory,
			Size:      job.Info.Size(),
			ModTime:   job.Info.ModTime(),
		}
		wp.results <- ScanResult{
			FileInfo: fileInfo,
			Error:    nil,
		}
	}
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job ScanJob) {
	wp.jobs <- job
}

// Close closes the jobs channel and waits for workers to finish
func (wp *WorkerPool) Close() {
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
}

// Results returns the results channel
func (wp *WorkerPool) Results() <-chan ScanResult {
	return wp.results
}
