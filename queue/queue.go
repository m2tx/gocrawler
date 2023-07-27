package queue

import (
	"context"
	"time"
)

type QueueTimer[T any] struct {
	queue       chan T
	triggerFunc TriggerFunc[[]T]
	size        int
	timeout     time.Duration
}

type TriggerFunc[T any] func(ctx context.Context, t T)

func NewQueueTimer[T any](size int, timeout time.Duration, triggerFunc TriggerFunc[[]T]) *QueueTimer[T] {
	return &QueueTimer[T]{
		queue:       make(chan T),
		triggerFunc: triggerFunc,
		size:        size,
		timeout:     timeout,
	}
}

func (qt *QueueTimer[T]) Start(ctx context.Context) {
	queue := make([]T, 0)

	for {
		select {
		case t, ok := <-qt.queue:
			if !ok {
				if len(queue) > 0 {
					qt.triggerFunc(ctx, queue)
					queue = nil
				}

				return
			}

			queue = append(queue, t)
			if len(queue) == qt.size {
				qt.triggerFunc(ctx, queue)
				queue = make([]T, 0)
			}
		case <-time.After(qt.timeout):
			if len(queue) > 0 {
				qt.triggerFunc(ctx, queue)
				queue = make([]T, 0)
			}
		case <-ctx.Done():
			qt.Close()
			if len(queue) > 0 {
				qt.triggerFunc(ctx, queue)
				queue = nil
			}
			return
		}
	}
}

func (qt *QueueTimer[T]) Add(t T) {
	qt.queue <- t
}

func (qt *QueueTimer[T]) Close() {
	close(qt.queue)
}
