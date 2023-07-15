package queue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueueTimer(t *testing.T) {
	ctx := context.Background()

	queueSize := 5
	queueTimeout := 100 * time.Millisecond

	testChan := make(chan []int)

	expectedTotal := 33

	queue := NewQueueTimer[int](queueSize, queueTimeout, func(ctx context.Context, t []int) {
		testChan <- t
	})

	go func() {
		go queue.Start(ctx)

		for i := 0; i < expectedTotal; i++ {
			queue.Add(i)
		}
	}()

	var count int
	for batch := range testChan {
		length := len(batch)
		count += length

		if length == 0 {
			assert.Fail(t, "batch size equals %d", length)
		}

		if count >= expectedTotal {
			close(testChan)
		}
	}

	assert.Equal(t, expectedTotal, count)
}
