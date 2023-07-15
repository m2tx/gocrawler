package queue

import (
	"context"
	"time"
)

type QueueTimer[T any] struct {
	Queue       chan T
	TriggerFunc TriggerFunc[[]T]
	Size        int
	Timeout     time.Duration
}

type TriggerFunc[T any] func(ctx context.Context, t T)

func NewQueueTimer[T any](size int, timeout time.Duration, triggerFunc TriggerFunc[[]T]) *QueueTimer[T] {
	return &QueueTimer[T]{
		Queue:       make(chan T),
		TriggerFunc: triggerFunc,
		Size:        size,
		Timeout:     timeout,
	}
}

func (qt *QueueTimer[T]) Start(ctx context.Context) {
	queue := make([]T, 0)

	for {
		select {
		case t, ok := <-qt.Queue:
			if !ok {
				if len(queue) > 0 {
					qt.TriggerFunc(ctx, queue)
					queue = nil
				}

				return
			}

			queue = append(queue, t)
			if len(queue) == qt.Size {
				qt.TriggerFunc(ctx, queue)
				queue = make([]T, 0)
			}
		case <-time.After(qt.Timeout):
			if len(queue) > 0 {
				qt.TriggerFunc(ctx, queue)
				queue = make([]T, 0)
			}
		case <-ctx.Done():
			close(qt.Queue)
			if len(queue) > 0 {
				qt.TriggerFunc(ctx, queue)
				queue = nil
			}
			return
		}
	}
}

func (qt *QueueTimer[T]) Add(t T) {
	qt.Queue <- t
}
