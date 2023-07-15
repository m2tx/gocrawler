package worker

import (
	"context"
	"sync"
)

type WorkerPool[T any] struct {
	NumWorkers int
	Tasks      chan T
	WorkerFunc WorkerFunc[T]
	WaitGroup  sync.WaitGroup
}

type WorkerFunc[T any] func(ctx context.Context, t T)

func NewWorkerPool[T any](numWorkers int, workerFunc WorkerFunc[T]) *WorkerPool[T] {
	return &WorkerPool[T]{
		NumWorkers: numWorkers,
		Tasks:      make(chan T),
		WorkerFunc: workerFunc,
	}
}

func (wp *WorkerPool[T]) Start(ctx context.Context) {
	for i := 0; i < wp.NumWorkers; i++ {
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool[T]) worker(ctx context.Context) {
	for {
		select {
		case t, ok := <-wp.Tasks:
			if !ok {
				return
			}
			wp.WorkerFunc(ctx, t)
			wp.WaitGroup.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool[T]) Add(t T) {
	wp.WaitGroup.Add(1)
	wp.Tasks <- t
}

func (wp *WorkerPool[T]) Wait() {
	wp.WaitGroup.Wait()
	close(wp.Tasks)
}

func (wp *WorkerPool[T]) Close() {
	close(wp.Tasks)
}
