package main

import "sync"

// ConcurrentSlice is a thread safe integer store
type ConcurrentSlice struct {
	sync.RWMutex
	items []int64
}

// Add item to slice
func (s *ConcurrentSlice) Add(e int64) {
	s.Lock()
	defer s.Unlock()

	s.items = append(s.items, e)
}

// Remove item from slice
func (s *ConcurrentSlice) Remove(e int64) {
	s.Lock()
	defer s.Unlock()

	i := s.Index(e)
	s.items = append(s.items[0:i], s.items[i+1:]...)
}

// Index of current item or -1 if not found
func (s *ConcurrentSlice) Index(e int64) int {
	for i, el := range s.items {
		if el == e {
			return i
		}
	}
	return -1
}
