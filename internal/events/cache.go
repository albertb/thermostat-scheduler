package events

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Cache struct {
	store store
}

func NewCache() (*Cache, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	return &Cache{
		store: &fileStore{
			path: filepath.Join(cacheDir, "thermostat-scheduler", "seen_events"),
		},
	}, nil
}

type store interface {
	Load() (map[string]struct{}, error)
	Save(id string) error
}

type fileStore struct {
	path string
}

func (s *fileStore) Load() (map[string]struct{}, error) {
	seenEvents := make(map[string]struct{})
	file, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return seenEvents, nil
		}
		return seenEvents, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		seenEvents[scanner.Text()] = struct{}{}
	}

	return seenEvents, scanner.Err()
}

func (s *fileStore) Save(id string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(id + "\n")
	return err
}

type PeakEvent struct {
	Start time.Time
	End   time.Time
}

func eventID(event PeakEvent) string {
	var b strings.Builder
	b.WriteString(event.Start.Format(time.RFC3339))
	b.WriteString(event.End.Format(time.RFC3339))
	return b.String()
}

func (c *Cache) loadSeenEvents() (map[string]struct{}, error) {
	return c.store.Load()
}

func (c *Cache) markEventAsSeen(event PeakEvent) error {
	return c.store.Save(eventID(event))
}
