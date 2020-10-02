package faggot

import "sync"

// concurrentSlice is a thread safe integer store
type concurrentSlice struct {
	sync.RWMutex
	items []int64
}

// Add item to slice
func (s *concurrentSlice) Add(e int64) {
	s.Lock()
	defer s.Unlock()

	s.items = append(s.items, e)
}

// Remove item from slice
func (s *concurrentSlice) Remove(e int64) {
	s.Lock()
	defer s.Unlock()

	i := s.index(e)
	s.items = append(s.items[0:i], s.items[i+1:]...)
}

// Index of current item or -1 if not found
func (s *concurrentSlice) Index(e int64) int {
	s.Lock()
	defer s.Unlock()

	return s.index(e)
}

func (s *concurrentSlice) index(e int64) int {
	for i, el := range s.items {
		if el == e {
			return i
		}
	}
	return -1
}
