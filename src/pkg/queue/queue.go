package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/casapps/casci/src/pkg/builds"
)

// BuildQueue manages the queue of builds to be executed
type BuildQueue struct {
	builds    []*builds.Build
	mu        sync.RWMutex
	buildChan chan *builds.Build
	stopChan  chan struct{}
	workers   int
	buildSvc  *builds.Service
	executor  Executor
}

// Executor interface for build execution
type Executor interface {
	Execute(ctx context.Context, build *builds.Build) error
}

// NewBuildQueue creates a new build queue
func NewBuildQueue(buildSvc *builds.Service, executor Executor, workers int) *BuildQueue {
	if workers <= 0 {
		workers = 5 // Default to 5 concurrent builds
	}

	return &BuildQueue{
		builds:    make([]*builds.Build, 0),
		buildChan: make(chan *builds.Build, 100),
		stopChan:  make(chan struct{}),
		workers:   workers,
		buildSvc:  buildSvc,
		executor:  executor,
	}
}

// Start starts the queue workers
func (q *BuildQueue) Start(ctx context.Context) {
	log.Printf("Starting build queue with %d workers", q.workers)

	// Start worker goroutines
	for i := 0; i < q.workers; i++ {
		go q.worker(ctx, i)
	}

	// Start queue monitor
	go q.monitor(ctx)
}

// Stop stops the queue
func (q *BuildQueue) Stop() {
	close(q.stopChan)
}

// Enqueue adds a build to the queue
func (q *BuildQueue) Enqueue(build *builds.Build) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Add to queue list
	q.builds = append(q.builds, build)

	// Send to channel for workers
	select {
	case q.buildChan <- build:
		log.Printf("Build #%d for project %d enqueued", build.BuildNumber, build.ProjectID)
		return nil
	default:
		return fmt.Errorf("build queue is full")
	}
}

// worker processes builds from the queue
func (q *BuildQueue) worker(ctx context.Context, id int) {
	log.Printf("Build queue worker %d started", id)

	for {
		select {
		case <-q.stopChan:
			log.Printf("Build queue worker %d stopped", id)
			return
		case <-ctx.Done():
			log.Printf("Build queue worker %d context canceled", id)
			return
		case build := <-q.buildChan:
			q.processBuild(ctx, id, build)
		}
	}
}

// processBuild processes a single build
func (q *BuildQueue) processBuild(ctx context.Context, workerID int, build *builds.Build) {
	log.Printf("Worker %d processing build #%d for project %d", workerID, build.BuildNumber, build.ProjectID)

	// Remove from queue list
	q.removeFromQueue(build.ID)

	// Mark build as started
	if err := q.buildSvc.Start(ctx, build.ID); err != nil {
		log.Printf("Failed to mark build as started: %v", err)
		return
	}

	// Execute the build
	err := q.executor.Execute(ctx, build)

	// Update build status
	if err != nil {
		log.Printf("Build #%d failed: %v", build.BuildNumber, err)
		if err := q.buildSvc.Complete(ctx, build.ID, builds.StatusFailed); err != nil {
			log.Printf("Failed to mark build as failed: %v", err)
		}
	} else {
		log.Printf("Build #%d completed successfully", build.BuildNumber)
		if err := q.buildSvc.Complete(ctx, build.ID, builds.StatusSuccess); err != nil {
			log.Printf("Failed to mark build as complete: %v", err)
		}
	}
}

// removeFromQueue removes a build from the queue list
func (q *BuildQueue) removeFromQueue(buildID int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, b := range q.builds {
		if b.ID == buildID {
			q.builds = append(q.builds[:i], q.builds[i+1:]...)
			return
		}
	}
}

// monitor periodically checks for queued builds in the database
func (q *BuildQueue) monitor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.checkQueuedBuilds(ctx)
		}
	}
}

// checkQueuedBuilds checks for builds in queued status and adds them to the queue
func (q *BuildQueue) checkQueuedBuilds(ctx context.Context) {
	queuedBuilds, err := q.buildSvc.ListQueued(ctx)
	if err != nil {
		log.Printf("Failed to list queued builds: %v", err)
		return
	}

	for _, build := range queuedBuilds {
		// Check if already in queue
		if !q.isInQueue(build.ID) {
			if err := q.Enqueue(build); err != nil {
				log.Printf("Failed to enqueue build #%d: %v", build.BuildNumber, err)
			}
		}
	}
}

// isInQueue checks if a build is already in the queue
func (q *BuildQueue) isInQueue(buildID int) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, b := range q.builds {
		if b.ID == buildID {
			return true
		}
	}
	return false
}

// GetQueueLength returns the current queue length
func (q *BuildQueue) GetQueueLength() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.builds)
}

// GetQueuedBuilds returns all builds currently in the queue
func (q *BuildQueue) GetQueuedBuilds() []*builds.Build {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*builds.Build, len(q.builds))
	copy(result, q.builds)
	return result
}
