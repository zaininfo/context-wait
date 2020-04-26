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
	goroutine1
	goroutine2
	goroutine11
	goroutine12
	goroutine21
	goroutine22
)

// TestWaitableContext_TypicalUsage verifies the behavior of a waitable context
func TestWaitableContext_TypicalUsage(t *testing.T) {
	ctx := context.Background()

	// a map to check the status of worker goroutines
	workSheet := sync.Map{}

	// a worker goroutine that will ignore context cancellation and report its
	// completion by marking the context as done
	importantWorker := func(ctx context.Context) {
		defer Done(ctx)

		select {
		case <-time.After(time.Second * 5):
			workSheet.Store(importantWorkerID, struct{}{})
		}
	}

	// a worker goroutine that will terminate on context cancellation
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

	// a waitable context
	ctx, wait := WithWait(ctx)

	// a cancellable (and a waitable - by virtue of parent context) context
	ctx, cancel := context.WithCancel(ctx)

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

// TestWaitableContext_Concurrency verifies the safety of a waitable context
// for concurrent usage
func TestWaitableContext_Concurrency(t *testing.T) {
	ctx := context.Background()

	// a map to check the status of goroutines
	goroutines := sync.Map{}

	// a waitable context
	ctx, wait := WithWait(ctx)
	go func(ctx context.Context) {
		defer goroutines.Store(goroutine1, struct{}{})

		// a waitable context
		ctx, wait := WithWait(ctx)
		// wait for the context
		defer wait()

		go func(ctx context.Context) {
			defer goroutines.Store(goroutine11, struct{}{})

			// mark the context as done
			Done(ctx)
		}(ctx)

		go func(ctx context.Context) {
			defer goroutines.Store(goroutine12, struct{}{})

			// mark the context as done
			Done(ctx)
		}(ctx)
	}(ctx)

	// a waitable context
	ctx, wait = WithWait(ctx)
	go func(ctx context.Context) {
		defer goroutines.Store(goroutine2, struct{}{})

		// a waitable context
		ctx, wait := WithWait(ctx)
		// wait for the context
		defer wait()

		go func(ctx context.Context) {
			defer goroutines.Store(goroutine21, struct{}{})

			// mark the context as done
			Done(ctx)
		}(ctx)

		go func(ctx context.Context) {
			defer goroutines.Store(goroutine22, struct{}{})

			// mark the context as done
			Done(ctx)
		}(ctx)
	}(ctx)

	// wait for the context
	wait()

	// give time for all goroutines to complete, this is important because
	// waitable context is marked as done in multiple goroutines, which is not
	// the recommended use case, but is tested here to ensure that there are no
	// race conditions or deadlocks
	time.Sleep(time.Second)

	if _, ok := goroutines.Load(goroutine1); !ok {
		fmt.Println("goroutine1 didn't complete")
		os.Exit(1)
	}

	if _, ok := goroutines.Load(goroutine2); !ok {
		fmt.Println("goroutine2 didn't complete")
		os.Exit(1)
	}

	if _, ok := goroutines.Load(goroutine11); !ok {
		fmt.Println("goroutine11 didn't complete")
		os.Exit(1)
	}

	if _, ok := goroutines.Load(goroutine12); !ok {
		fmt.Println("goroutine12 didn't complete")
		os.Exit(1)
	}

	if _, ok := goroutines.Load(goroutine21); !ok {
		fmt.Println("goroutine21 didn't complete")
		os.Exit(1)
	}

	if _, ok := goroutines.Load(goroutine22); !ok {
		fmt.Println("goroutine22 didn't complete")
		os.Exit(1)
	}
}
