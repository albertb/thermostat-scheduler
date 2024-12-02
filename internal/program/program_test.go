package program

import (
	"testing"
	"thermostat-scheduler/internal/config"
	"thermostat-scheduler/internal/events"
	"time"
)

func TestAssembleProgram(t *testing.T) {
	dp := config.DailyProgram{
		Morning: config.DayEvent{Time: 7 * time.Hour, Heat: 21, Cool: 24},
		Day:     config.DayEvent{Time: 9 * time.Hour, Heat: 20, Cool: 24},
		Evening: config.DayEvent{Time: 16 * time.Hour, Heat: 21, Cool: 24},
		Night:   config.DayEvent{Time: 21 * time.Hour, Heat: 20, Cool: 25},
	}

	cfg := config.Config{
		NormalProgram: config.WeeklyProgram{
			Sunday: dp, Monday: dp, Tuesday: dp, Wednesday: dp,
			Thursday: dp, Friday: dp, Saturday: dp,
		},
		PeakProgram: config.PeakProgram{
			PreHeatDuration:    1 * time.Hour,
			PeakBufferDuration: 2 * time.Minute,
			PreHeatTempOffset:  2,
			PeakTempOffset:     -2,
		},
	}

	// Peak events are on Jan 24, 6h to 9h, and 16h to 20h.
	events := []events.PeakEvent{
		{
			Start: parseTime(t, "Wed, 24 Jan 2024 06:00:00 EST"),
			End:   parseTime(t, "Wed, 24 Jan 2024 09:00:00 EST"),
		},
		{
			Start: parseTime(t, "Wed, 24 Jan 2024 16:00:00 EST"),
			End:   parseTime(t, "Wed, 24 Jan 2024 20:00:00 EST"),
		},
	}

	// It's 4h on the morning on the peak events.
	now := parseTime(t, "Wed, 24 Jan 2024 04:00:00 EST")
	program := AssembleProgram(cfg, now, events, false)

	// Expect the current day's program to handle the morning peak period.
	expectedPeakProgram := config.DailyProgram{
		Morning: config.DayEvent{Time: 5 * time.Hour, Heat: 20 + 2, Cool: 24},
		Day:     config.DayEvent{Time: 6*time.Hour - 2*time.Minute, Heat: 20 - 2, Cool: 24},
		Evening: config.DayEvent{Time: 9*time.Hour + 2*time.Minute, Heat: 20, Cool: 24},
		Night:   config.DayEvent{Time: 16 * time.Hour, Heat: 21, Cool: 24},
	}
	if program.Wednesday != expectedPeakProgram {
		t.Errorf("want\n%v, got\n%v", expectedPeakProgram, program.Wednesday)
	}
	// Expect the previous day's night event to set now's normal temperature.
	expectedDayEvent := config.DayEvent{Time: 21 * time.Hour, Heat: 20, Cool: 25}
	if program.Tuesday.Night != expectedDayEvent {
		t.Errorf("want\n%v, got\n%v", expectedDayEvent, program.Tuesday.Night)
	}

	// Configure the program to instead maintain the usual "peak" temperature
	// before pre-heating.
	cfg.PeakProgram.MaintainNormalTempBeforePreHeat = true
	program = AssembleProgram(cfg, now, events, false)

	// Expect the same program to handle the peak period.
	if program.Wednesday != expectedPeakProgram {
		t.Errorf("at time %v, want\n%v, got\n%v", now, expectedPeakProgram, program.Wednesday)
	}
	// But expect the normal (21C) temperature before pre-heating.
	expectedBeforePreHeatingDayEvent := config.DayEvent{Time: 21 * time.Hour, Heat: 21, Cool: 24}
	if program.Tuesday.Night != expectedBeforePreHeatingDayEvent {
		t.Errorf("want\n%v, got\n%v", expectedDayEvent, program.Tuesday.Night)
	}
	// Revert back the config.
	cfg.PeakProgram.MaintainNormalTempBeforePreHeat = false

	// Four hours later, at 8h, we should still have the same program since the
	// peak period isn't over.
	now = now.Add(4 * time.Hour)
	program = AssembleProgram(cfg, now, events, false)

	// Expect the current day's program to handle the morning peak period.
	if program.Wednesday != expectedPeakProgram {
		t.Errorf("want\n%v, got\n%v", expectedPeakProgram, program.Wednesday)
	}
	// Expect the previous day's night event to set now's normal temperature.
	if program.Tuesday.Night != expectedDayEvent {
		t.Errorf("want\n%v, got\n%v", expectedDayEvent, program.Tuesday.Night)
	}

	// Two hours later at 10h, the evening peak period should now apply.
	now = now.Add(2 * time.Hour)
	program = AssembleProgram(cfg, now, events, false)

	// Expect the current day's program to handle the afteroon peak period.
	expectedPeakProgram = config.DailyProgram{
		Morning: config.DayEvent{Time: 15 * time.Hour, Heat: 21 + 2, Cool: 24},
		Day:     config.DayEvent{Time: 16*time.Hour - 2*time.Minute, Heat: 21 - 2, Cool: 24},
		Evening: config.DayEvent{Time: 20*time.Hour + 2*time.Minute, Heat: 21, Cool: 24},
		Night:   config.DayEvent{Time: 21 * time.Hour, Heat: 20, Cool: 25},
	}
	if program.Wednesday != expectedPeakProgram {
		t.Errorf("at time %v, want\n%v, got\n%v", now, expectedPeakProgram, program.Wednesday)
	}
	// Expect the previous day's night event to set the usual temperature.
	expectedDayEvent = config.DayEvent{Time: 21 * time.Hour, Heat: 20, Cool: 24}
	if program.Tuesday.Night != expectedDayEvent {
		t.Errorf("want\n%v, got\n%v", expectedDayEvent, program.Tuesday.Night)
	}

	// Twelve hours later, at 22h, we should be back on the normal program.
	now = now.Add(12 * time.Hour)
	program = AssembleProgram(cfg, now, events, false)
	if program != cfg.NormalProgram {
		t.Errorf("at time %v, want\n%v, got\n%v", now, expectedPeakProgram, program.Wednesday)
	}
}

func parseTime(t *testing.T, timeStr string) time.Time {
	parsedTime, err := time.Parse(time.RFC1123, timeStr)
	if err != nil {
		t.Fatalf("failed to parse time '%s': %v", timeStr, err)
	}
	return parsedTime
}
