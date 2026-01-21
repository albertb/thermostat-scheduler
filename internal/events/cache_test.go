package events

import (
	"os"
	"testing"
	"time"
)

func TestEventCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// monkey patch the cache dir functions
	oldCacheDirFn := getCacheDir
	getCacheDir = func() (string, error) { return tmpDir, nil }
	defer func() { getCacheDir = oldCacheDirFn }()

	event := PeakEvent{
		Start: time.Now(),
		End:   time.Now().Add(1 * time.Hour),
	}

	// 1. Test that the cache is empty
	seenEvents, err := loadSeenEvents()
	if err != nil {
		t.Fatalf("failed to load seen events: %v", err)
	}
	if len(seenEvents) != 0 {
		t.Errorf("expected 0 seen events, got %d", len(seenEvents))
	}

	// 2. Test that we can mark an event as seen
	if err := markEventAsSeen(event); err != nil {
		t.Fatalf("failed to mark event as seen: %v", err)
	}

	// 3. Test that the event is now in the cache
	seenEvents, err = loadSeenEvents()
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
