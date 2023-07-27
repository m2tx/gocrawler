package safemap

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeMap(t *testing.T) {
	var waitGroup sync.WaitGroup
	key := "number"

	m := New[string, int]()

	waitGroup.Add(1)
	go func() {
		for i := 0; i < 1000; i++ {
			waitGroup.Add(1)
			go func() {
				m.Change(key, func(d int) int {
					return d + 1
				})
				waitGroup.Done()
			}()
		}
		waitGroup.Done()
	}()

	waitGroup.Wait()

	assert.Equal(t, 1000, m.Get(key))
}
