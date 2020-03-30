package wait

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	importantWorkerID = iota
	optionalWorkerID
)

// TestWaitableContext verifies the behavior of a waitable context
func TestWaitableContext(t *testing.T) {
	ctx := context.Background()

	// a map to check the status of worker goroutines
	workSheet := sync.Map{}

	// a worker goroutine that will ignore context cancellation
	// and report its completion by marking the context as done
	importantWorker := func(ctx context.Context) {
		defer Done(ctx)

		select {
		case <-time.After(time.Second * 5):
			workSheet.Store(importantWorkerID, struct{}{})
		}
	}

	// a worker goroutine that will terminate on context
	// cancellation
	optionalWorker := func(ctx context.Context) {
		select {
		case <-time.After(time.Second * 10):
			workSheet.Store(optionalWorkerID, struct{}{})
		case <-ctx.Done():
		}
	}

	startWorkers := func(ctx context.Context) {
		go importantWorker(ctx)
		go optionalWorker(ctx)
	}

	ctx, cancel := context.WithCancel(ctx)

	// a waitable (and cancellable, by virtue of parent context)
	// context
	ctx, wait := WithWait(ctx)

	// 1) start workers
	startWorkers(ctx)

	// 2) timeout optional worker
	go func() {
		select {
		case <-time.After(time.Second):
			cancel()
		}
	}()

	// 3) wait for important worker
	wait()

	if _, ok := workSheet.Load(importantWorkerID); !ok {
		fmt.Println("important worker was not waited for")
		os.Exit(1)
	}

	if _, ok := workSheet.Load(optionalWorkerID); ok {
		fmt.Println("optional worker was waited for")
		os.Exit(1)
	}
}
