package safemap

import "sync"

type SafeMap[T comparable, D any] struct {
	mu   sync.Mutex
	data map[T]D
}

func New[T comparable, D any]() *SafeMap[T, D] {
	return &SafeMap[T, D]{
		data: make(map[T]D),
	}
}

func (s *SafeMap[T, D]) Set(key T, value D) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *SafeMap[T, D]) Change(key T, change func(D) D) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = change(s.data[key])
}

func (s *SafeMap[T, D]) Get(key T) D {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[key]
}
