package client

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// Keeps track of the start and end time of an event as durations from
// midnight. |start| must be less than |end|, and both values must be between
// 0s and 24h.
//
// E.g., { start: "16h", end: "20h" } for a event from 16h to 20h.
type PeakEvent struct {
	Start time.Duration `yaml:"start"`
	End   time.Duration `yaml:"end"`
}

// Keeps track of peak events during a single day.
type DailyPeakEvents struct {
	Date   time.Time   `yaml:"date"`
	Events []PeakEvent `yaml:"events"`
}

func GetPeakEvents(path string) ([]DailyPeakEvents, error) {
	e := []DailyPeakEvents{}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return e, err
	}

	err = yaml.Unmarshal([]byte(file), &e)
	if err != nil {
		return e, err
	}

	err = validateDailyPeakEvents(e)
	if err != nil {
		return e, err
	}

	return e, nil
}

func validateDailyPeakEvents(events []DailyPeakEvents) error {
	lastDate := time.Time{}
	for _, dpe := range events {
		if !dpe.Date.After(lastDate) && !dpe.Date.IsZero() {
			return fmt.Errorf("peak event dates must be in order and cannot repeat")
		}
		lastDate = dpe.Date

		lastEnd := -1 * time.Second
		for _, pe := range dpe.Events {
			if pe.Start < 0 || pe.Start > 24*time.Hour {
				return fmt.Errorf("peak event start time must be between 0s and 24h, got %v", pe.Start)
			}
			if pe.End < 0 || pe.End > 24*time.Hour {
				return fmt.Errorf("peak event end time must be between 0s and 24h, got %v", pe.End)
			}
			if pe.Start >= pe.End {
				return fmt.Errorf("peak event start time must be strictly before end time, got %v vs %v", pe.Start, pe.End)
			}
			if pe.Start < lastEnd {
				return fmt.Errorf("peak events must be in order and cannot overlap, %v vs %v", lastEnd, pe)
			}
			lastEnd = pe.End
		}
	}
	return nil
}
