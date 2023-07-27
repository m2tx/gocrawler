package worker

import (
	"context"
	"sync"
)

type WorkerPool[T any] struct {
	nWorkers   int
	queue      chan T
	workerFunc WorkerFunc[T]
	waitGroup  sync.WaitGroup
}

type WorkerFunc[T any] func(ctx context.Context, t T)

func NewWorkerPool[T any](nWorkers int, workerFunc WorkerFunc[T]) *WorkerPool[T] {
	return &WorkerPool[T]{
		nWorkers:   nWorkers,
		queue:      make(chan T),
		workerFunc: workerFunc,
	}
}

func (wp *WorkerPool[T]) Start(ctx context.Context) {
	for i := 0; i < wp.nWorkers; i++ {
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool[T]) worker(ctx context.Context) {
	for {
		select {
		case t, ok := <-wp.queue:
			if !ok {
				return
			}
			wp.workerFunc(ctx, t)
			wp.waitGroup.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool[T]) Add(t T) {
	wp.waitGroup.Add(1)
	wp.queue <- t
}

func (wp *WorkerPool[T]) Wait() {
	wp.waitGroup.Wait()
}

func (wp *WorkerPool[T]) Close() {
	close(wp.queue)
}
