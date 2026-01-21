package events

import (
	"bufio"
	"os"
	"path/filepath"
)

var getCacheDir = func() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "thermostat-scheduler"), nil
}

func getCacheFile() (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "seen_events"), nil
}

func loadSeenEvents() (map[string]bool, error) {
	seenEvents := make(map[string]bool)
	cacheFile, err := getCacheFile()
	if err != nil {
		return seenEvents, err
	}

	file, err := os.Open(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return seenEvents, nil
		}
		return seenEvents, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		seenEvents[scanner.Text()] = true
	}

	return seenEvents, scanner.Err()
}

func eventID(event PeakEvent) string {
	return event.Start.String() + event.End.String()
}

func markEventAsSeen(event PeakEvent) error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	cacheFile, err := getCacheFile()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(cacheFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(eventID(event) + "\n")
	return err
}
