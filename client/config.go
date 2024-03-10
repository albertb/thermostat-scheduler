package client

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// A single event in a program.
type DayEvent struct {
	Time time.Duration // The time this event starts, as a duration from midnight, e.g., "14h"
	Heat int           // The temperature to heat to.
	Cool int           // The temperature to cool to.
}

func (d *DayEvent) Copy(other DayEvent) {
	d.Time = other.Time
	d.Heat = other.Heat
	d.Cool = other.Cool
}

// The program for a single day.
type DailyProgram struct {
	Morning DayEvent
	Day     DayEvent
	Evening DayEvent
	Night   DayEvent
}

// The program for a whole week, one program per day.
type WeeklyProgram struct {
	Sunday    DailyProgram
	Monday    DailyProgram
	Tuesday   DailyProgram
	Wednesday DailyProgram
	Thursday  DailyProgram
	Friday    DailyProgram
	Saturday  DailyProgram
}

type PeakProgram struct {
	// How long to pre-heat before a peak event, e.g., "1h"
	// Careful not to overlap the previous event with an overly long
	// pre-heating.
	PreHeatDuration time.Duration `yaml:"pre_heat_duration"`

	// How long before and after the peak event to keep the peak temperature to
	// account for potential clock drift, e.g., "2m"
	//
	// TODO: It seems as though the thermostat always rounds down to the nearest
	// 10 minute, e.g, 18m->10m, so this doesn't work as expected.
	PeakBufferDuration time.Duration `yaml:"peak_buffer"`

	// How much to change the temperature by during pre-heating, relative to
	// the normal program temperature at the start of the peak event, e.g., 2
	PreHeatTempOffset int `yaml:"pre_heat_temp_offset"`

	// How much to change the temperature during the peak event, relative to the
	// normal program temperature at the start of the peak event, e.g., -1
	PeakTempOffset int `yaml:"peak_temp_offset"`

	// Whether to maintain the normal temperature of the peak period before the
	// pre-heat period. This makes it so we don't need to heat as much during
	// pre-heating.
	MaintainNormalTempBeforePreHeat bool `yaml:"maintain_normal_temp_before_preheat"`
}

type Config struct {
	Username string
	Password string

	// The normal every day program.
	NormalProgram WeeklyProgram `yaml:"normal_program"`

	// How to modify the program during peak events.
	PeakProgram PeakProgram `yaml:"peak_program"`
}

func GetConfig(path string) (Config, error) {
	c := Config{}

	file, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal([]byte(file), &c)
	if err != nil {
		return c, err
	}

	return validate(c)
}

func validate(c Config) (Config, error) {
	if len(c.Username) < 1 || len(c.Password) < 1 {
		return c, errors.New("username and password are required")
	}
	err := validateWeeklyProgram(c.NormalProgram)
	if err != nil {
		return c, err
	}
	err = validatePeakProgram(c.PeakProgram)
	if err != nil {
		return c, err
	}
	return c, nil
}

func validateWeeklyProgram(p WeeklyProgram) error {
	for _, dp := range []DailyProgram{p.Sunday, p.Monday, p.Tuesday, p.Wednesday, p.Thursday, p.Friday, p.Saturday} {
		err := validateDailyProgram(dp)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateDailyProgram(p DailyProgram) error {
	// First, validate that each day event is valid.
	for _, e := range []DayEvent{p.Morning, p.Day, p.Evening, p.Night} {
		err := validateDayEvent(e)
		if err != nil {
			return err
		}
	}

	// Then, make sure the start times are in order.
	last := -1 * time.Second
	for _, t := range []time.Duration{p.Morning.Time, p.Day.Time, p.Evening.Time, p.Night.Time} {
		if t < last {
			return fmt.Errorf("daily program times aren't in order, %v !< %v", last, t)
		}
		last = t
	}
	return nil
}

func validateDayEvent(e DayEvent) error {
	if e.Time < 0 || e.Time > 24*time.Hour {
		return fmt.Errorf("event time must be between 0s and 24h, got: %v", e)
	}
	// Losely enforce that the temperature are within acceptable ranges.
	if e.Heat < 0 || e.Heat > 50 {
		return fmt.Errorf("event heat must be between 0C and 50C, got: %v", e)
	}
	if e.Cool < 0 || e.Cool > 50 {
		return fmt.Errorf("event cool must be between 0C and 50C, got: %v", e)
	}
	return nil
}

func validatePeakProgram(p PeakProgram) error {
	// Loosely enforce that the peak program values are within acceptable ranges.
	if p.PreHeatDuration < 0 || p.PreHeatDuration > 2*time.Hour {
		return fmt.Errorf("pre-heat duration should be between 0s and 2h, got %v", p.PreHeatDuration)
	}
	if p.PeakBufferDuration < 0 || p.PeakBufferDuration > 10*time.Minute {
		return fmt.Errorf("peak buffer duration should be between 0s and 10m, got %v", p.PeakBufferDuration)
	}
	if p.PreHeatTempOffset < 0 || p.PreHeatTempOffset > 10 {
		return fmt.Errorf("pre-heat temp offset should be between 0C and 10C, got %v", p.PreHeatTempOffset)
	}
	if p.PeakTempOffset < -10 || p.PeakTempOffset > 0 {
		return fmt.Errorf("peak temp offset should be between -10C and 0C, got %v", p.PeakTempOffset)
	}
	return nil
}

func (wp WeeklyProgram) ToStateData() StateData {
	return StateData{
		Program1: wp.Monday.ToProgramString(),
		Program2: wp.Tuesday.ToProgramString(),
		Program3: wp.Wednesday.ToProgramString(),
		Program4: wp.Thursday.ToProgramString(),
		Program5: wp.Friday.ToProgramString(),
		Program6: wp.Saturday.ToProgramString(),
		Program7: wp.Sunday.ToProgramString(),
	}
}

func (dp DailyProgram) ToProgramString() string {
	return dp.Morning.ToProgramHeatStringPart() +
		dp.Day.ToProgramHeatStringPart() +
		dp.Evening.ToProgramHeatStringPart() +
		dp.Night.ToProgramHeatStringPart() +
		dp.Morning.ToProgramCoolStringPart() +
		dp.Day.ToProgramCoolStringPart() +
		dp.Evening.ToProgramCoolStringPart() +
		dp.Night.ToProgramCoolStringPart()
}

func (de DayEvent) ToProgramHeatStringPart() string {
	start := time.Time{}.Add(de.Time)
	return fmt.Sprintf("%02d", start.Hour()) +
		fmt.Sprintf("%02d", start.Minute()) +
		fmt.Sprintf("%03d", int(de.Heat*10))
}

func (de DayEvent) ToProgramCoolStringPart() string {
	start := time.Time{}.Add(de.Time)
	return fmt.Sprintf("%02d", start.Hour()) +
		fmt.Sprintf("%02d", start.Minute()) +
		fmt.Sprintf("%03d", int(de.Cool*10))
}
