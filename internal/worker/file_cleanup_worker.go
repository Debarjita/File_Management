package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"file-sharing-platform/internal/service"
)

// FileCleanupWorker is a worker that cleans up expired files
type FileCleanupWorker struct {
	fileService  *service.FileService
	interval     time.Duration
	batchSize    int
	stopChan     chan struct{}
	wg           sync.WaitGroup
	isRunning    bool
	runningMutex sync.Mutex
}

// NewFileCleanupWorker creates a new file cleanup worker
func NewFileCleanupWorker(fileService *service.FileService, interval time.Duration, batchSize int) *FileCleanupWorker {
	return &FileCleanupWorker{
		fileService: fileService,
		interval:    interval,
		batchSize:   batchSize,
		stopChan:    make(chan struct{}),
	}
}

// Start starts the worker
func (w *FileCleanupWorker) Start() {
	w.runningMutex.Lock()
	defer w.runningMutex.Unlock()

	if w.isRunning {
		return
	}

	w.isRunning = true
	w.wg.Add(1)

	go w.run()

	log.Println("File cleanup worker started")
}

// Stop stops the worker
func (w *FileCleanupWorker) Stop() {
	w.runningMutex.Lock()
	defer w.runningMutex.Unlock()

	if !w.isRunning {
		return
	}

	close(w.stopChan)
	w.wg.Wait()
	w.isRunning = false

	log.Println("File cleanup worker stopped")
}

// run runs the worker
func (w *FileCleanupWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run once on startup
	w.cleanupFiles()

	for {
		select {
		case <-ticker.C:
			w.cleanupFiles()
		case <-w.stopChan:
			return
		}
	}
}

// cleanupFiles cleans up expired files
func (w *FileCleanupWorker) cleanupFiles() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	count, err := w.fileService.CleanupExpiredFiles(ctx, w.batchSize)
	if err != nil {
		log.Printf("Error cleaning up expired files: %v", err)
		return
	}

	if count > 0 {
		log.Printf("Cleaned up %d expired files", count)
	}
}
