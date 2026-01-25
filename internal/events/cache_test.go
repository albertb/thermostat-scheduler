package events

import (
	"testing"
	"time"
)

type inMemoryStore struct {
	events map[string]struct{}
}

func (s *inMemoryStore) Load() (map[string]struct{}, error) {
	return s.events, nil
}

func (s *inMemoryStore) Save(id string) error {
	s.events[id] = struct{}{}
	return nil
}

func NewInMemoryCache() *Cache {
	return &Cache{
		store: &inMemoryStore{
			events: make(map[string]struct{}),
		},
	}
}

func TestEventCache(t *testing.T) {
	cache := NewInMemoryCache()

	event := PeakEvent{
		Start: time.Now(),
		End:   time.Now().Add(1 * time.Hour),
	}

	// 1. Test that the cache is empty
	seenEvents, err := cache.loadSeenEvents()
	if err != nil {
		t.Fatalf("failed to load seen events: %v", err)
	}
	if len(seenEvents) != 0 {
		t.Errorf("expected 0 seen events, got %d", len(seenEvents))
	}

	// 2. Test that we can mark an event as seen
	if err := cache.markEventAsSeen(event); err != nil {
		t.Fatalf("failed to mark event as seen: %v", err)
	}

	// 3. Test that the event is now in the cache
	seenEvents, err = cache.loadSeenEvents()
	if err != nil {
		t.Fatalf("failed to load seen events: %v", err)
	}
	if len(seenEvents) != 1 {
		t.Errorf("expected 1 seen event, got %d", len(seenEvents))
	}
	if _, seen := seenEvents[eventID(event)]; !seen {
		t.Errorf("event not found in cache")
	}
}
