package wait

import "context"

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// waitKey is the key for wait.waitChan values in Contexts. It is
// unexported; clients use wait.WithWait and wait.Done
// instead of using this key directly.
var waitKey key

// waitChan is the channel that serves as the blocking statement
// in the function returned by wait.WithWait.
// It is unexported; clients use the function returned by
// wait.WithWait and wait.Done instead of using this channel
// directly.
type waitChan chan struct{}

// WithWait returns a new Context and a function that blocks until
// wait.Done is called with the new Context.
func WithWait(parent context.Context) (ctx context.Context, waitFunc func()) {
	wait := make(waitChan)

	ctx = context.WithValue(parent, waitKey, wait)
	waitFunc = func() {
		select {
		case <-wait:
		}
	}

	return
}

// Done unblocks the function returned by wait.WithWait for the
// provided ctx, if available.
func Done(ctx context.Context) {
	if wait, ok := ctx.Value(waitKey).(waitChan); ok {
		close(wait)
	}

	return
}
