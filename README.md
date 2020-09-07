### Waitable context
![GitHub Workflow Status](https://github.com/zaininfo/context-wait/workflows/CI/badge.svg)

The `wait` module introduces two semantics for contexts to support waiting for a goroutine, which complements the existing functionality to cancel a goroutine.

Due to the parent-child hierarchy of contexts, a waitable context will be unblocked if any one of the child contexts marks it as completed.

Therefore, the best use-case for a waitable context is when there is one essential goroutine and one or more non-essential goroutines.

#### Functions

- `func WithWait(parent context.Context) (ctx context.Context, waitFunc func())`

This function returns a new context and a `wait` function that can be used to wait for completion of the new context. In case, the parent context is already a waitable context, this function returns it as it is.

- `func Done(ctx context.Context)`

`Done` function marks the passed context as completed. This function is idempotent for the same context.

#### Usage

```go
package main

import (
	"context"
	"time"

	wContext "github.com/zaininfo/context-wait/wait"
)

func main() {
	ctx, wait := wContext.WithWait(context.Background())

	go someFunc(ctx)

	// wait for context completion
	wait()
}

func someFunc(ctx context.Context) {
	time.Sleep(time.Minute)

	// mark context as completed
	wContext.Done(ctx)
}
```

The test file contains a more realistic use-case i.e. waiting for an important goroutine while cancelling unimportant ones.
