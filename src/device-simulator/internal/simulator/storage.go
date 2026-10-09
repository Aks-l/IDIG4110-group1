package simulator

import (
	"fmt"
	"sync"
	"time"
)

type Storage interface {
	SaveReading(Reading) error
	SaveStateChange(StateChange) error
	CurrentState(string) (map[string]any, bool)
	Latest(string) (Reading, bool)
	History(string, time.Time, time.Time) []Reading
}

type StateChange struct {
	DeviceID string
	State    map[string]any
	Time     time.Time
}

type MemoryStorage struct {
	mu      sync.RWMutex
	latest  map[string]Reading
	history map[string][]Reading
	states  map[string]map[string]any
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{latest: map[string]Reading{}, history: map[string][]Reading{}, states: map[string]map[string]any{}}
}

func (s *MemoryStorage) SaveReading(reading Reading) error {
	if reading.DeviceID == "" || reading.Property == "" || reading.Time.IsZero() {
		return fmt.Errorf("reading requires device_id, property and time")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := reading.DeviceID + "/" + reading.Property
	s.latest[key] = reading
	s.history[key] = append(s.history[key], reading)
	return nil
}

func (s *MemoryStorage) SaveStateChange(change StateChange) error {
	if change.DeviceID == "" || change.Time.IsZero() {
		return fmt.Errorf("state change requires device_id and time")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := make(map[string]any, len(change.State))
	for key, value := range change.State {
		state[key] = value
	}
	s.states[change.DeviceID] = state
	return nil
}

func (s *MemoryStorage) CurrentState(deviceID string) (map[string]any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.states[deviceID]
	if !ok {
		return nil, false
	}
	copy := make(map[string]any, len(state))
	for key, value := range state {
		copy[key] = value
	}
	return copy, true
}

func (s *MemoryStorage) Latest(key string) (Reading, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	reading, ok := s.latest[key]
	return reading, ok
}

func (s *MemoryStorage) History(key string, from, to time.Time) []Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Reading
	for _, reading := range s.history[key] {
		if !reading.Time.Before(from) && !reading.Time.After(to) {
			result = append(result, reading)
		}
	}
	return result
}
