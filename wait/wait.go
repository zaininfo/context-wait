package wait

import (
	"context"
	"sync"
)

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// waitChan is an unexported type for channel used as the blocking statement in
// the waitFunc returned by wait.WithWait.
type waitChan chan struct{}

// waitKey is the key for wait.waitChan values in Contexts. It is unexported;
// clients use wait.WithWait and wait.Done instead of using this key directly.
var waitKey key

// mutex is the mutual exclusion lock used to prevent a race condition in
// wait.Done.
var mutex sync.Mutex

// WithWait returns a new Context and a function that blocks until wait.Done is
// called with the new Context.
// If the passed Context is waitable itself, then it's returned back as it is.
func WithWait(parent context.Context) (ctx context.Context, waitFunc func()) {
	ctx = parent

	var wait waitChan
	var ok bool
	if wait, ok = getWaitChan(parent); !ok {
		wait = make(waitChan)
		ctx = context.WithValue(parent, waitKey, wait)
	}

	waitFunc = func() {
		select {
		case <-wait:
		}
	}

	return
}

// Done unblocks the waitFunc returned by wait.WithWait for the provided ctx,
// if available.
// If waitFunc has already been unblocked, then nothing happens.
func Done(ctx context.Context) {
	if wait, ok := getWaitChan(ctx); ok {
		mutex.Lock()
		defer mutex.Unlock()

		select {
		case <-wait:
		default:
			close(wait)
		}
	}

	return
}

// getWaitChan returns the wait.waitChan value from the provided Context, if
// available.
func getWaitChan(ctx context.Context) (wait waitChan, ok bool) {
	wait, ok = ctx.Value(waitKey).(waitChan)

	return
}
