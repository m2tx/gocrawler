package worker

import (
	"context"
	"sync"
	"testing"
)

func TestWorkerPool(t *testing.T) {
	ctx := context.Background()

	expectedTotal := 1000
	var waitGroup sync.WaitGroup

	pool := NewWorkerPool[int](10, func(ctx context.Context, t int) {
		waitGroup.Done()
	})
	pool.Start(ctx)

	for i := 0; i < expectedTotal; i++ {
		waitGroup.Add(1)
		pool.Add(i)
	}

	pool.Wait()

	waitGroup.Wait()
}
