package background

import "sync"

type backgroundServerStatus struct {
	mu         sync.RWMutex
	isStarting bool
	isStarted  bool
	isStopping bool
}

func newBackgroundServerStatus() *backgroundServerStatus {
	return &backgroundServerStatus{
		mu:         sync.RWMutex{},
		isStarting: false,
		isStarted:  false,
		isStopping: false,
	}
}

func (s *backgroundServerStatus) getIsStarting() bool {
	s.mu.RLock()
	ret := s.isStarting
	s.mu.RUnlock()
	return ret
}

func (s *backgroundServerStatus) setIsStarting(val bool) {
	s.mu.Lock()
	s.isStarting = val
	s.mu.Unlock()
}

func (s *backgroundServerStatus) getIsStarted() bool {
	s.mu.RLock()
	ret := s.isStarted
	s.mu.RUnlock()
	return ret
}

func (s *backgroundServerStatus) setIsStarted(val bool) {
	s.mu.Lock()
	s.isStarted = val
	s.mu.Unlock()
}

func (s *backgroundServerStatus) getIsStopping() bool {
	s.mu.RLock()
	ret := s.isStopping
	s.mu.RUnlock()
	return ret
}

func (s *backgroundServerStatus) setIsStopping(val bool) {
	s.mu.Lock()
	s.isStopping = val
	s.mu.Unlock()
}
