package client

import (
	"testing"
	"time"
)

func TestAssembleProgram(t *testing.T) {
	dp := DailyProgram{
		Morning: DayEvent{Time: 7 * time.Hour, Heat: 21, Cool: 24},
		Day:     DayEvent{Time: 9 * time.Hour, Heat: 20, Cool: 24},
		Evening: DayEvent{Time: 16 * time.Hour, Heat: 21, Cool: 24},
		Night:   DayEvent{Time: 21 * time.Hour, Heat: 20, Cool: 25},
	}

	config := Config{
		NormalProgram: WeeklyProgram{
			Sunday: dp, Monday: dp, Tuesday: dp, Wednesday: dp,
			Thursday: dp, Friday: dp, Saturday: dp,
		},
		PeakProgram: PeakProgram{
			PreHeatDuration:    1 * time.Hour,
			PeakBufferDuration: 2 * time.Minute,
			PreHeatTempOffset:  2,
			PeakTempOffset:     -2,
		},
	}

	// Peak events are on Jan 24, 6h to 9h, and 16h to 20h.
	events := []DailyPeakEvents{
		{Date: time.Date(2024, 1, 24, 0, 0, 0, 0, time.UTC),
			Events: []PeakEvent{
				{Start: 6 * time.Hour, End: 9 * time.Hour},
				{Start: 16 * time.Hour, End: 20 * time.Hour}}},
	}

	// It's 4h on the morning on the peak events.
	now, _ := time.Parse(time.RFC1123, "Wed, 24 Jan 2024 04:00:00 EST")
	program := AssembleProgram(config, now, events, false)

	// Expect the current day's program to handle the morning peak period.
	expectedPeakProgram := DailyProgram{
		Morning: DayEvent{Time: 5 * time.Hour, Heat: 20 + 2, Cool: 24},
		Day:     DayEvent{Time: 6*time.Hour - 2*time.Minute, Heat: 20 - 2, Cool: 24},
		Evening: DayEvent{Time: 9*time.Hour + 2*time.Minute, Heat: 20, Cool: 24},
		Night:   DayEvent{Time: 16 * time.Hour, Heat: 21, Cool: 24},
	}
	if program.Wednesday != expectedPeakProgram {
		t.Errorf("want\n%v, got\n%v", expectedPeakProgram, program.Wednesday)
	}
	// Expect the previous day's night event to set now's normal temperature.
	expectedDayEvent := DayEvent{Time: 21 * time.Hour, Heat: 20, Cool: 25}
	if program.Tuesday.Night != expectedDayEvent {
		t.Errorf("want\n%v, got\n%v", expectedDayEvent, program.Tuesday.Night)
	}

	// Configure the program to instead maintain the usual "peak" temperature
	// before pre-heating.
	config.PeakProgram.MaintainNormalTempBeforePreHeat = true
	program = AssembleProgram(config, now, events, false)

	// Expect the same program to handle the peak period.
	if program.Wednesday != expectedPeakProgram {
		t.Errorf("at time %v, want\n%v, got\n%v", now, expectedPeakProgram, program.Wednesday)
	}
	// But expect the normal (21C) temperature before pre-heating.
	expectedBeforePreHeatingDayEvent := DayEvent{Time: 21 * time.Hour, Heat: 21, Cool: 24}
	if program.Tuesday.Night != expectedBeforePreHeatingDayEvent {
		t.Errorf("want\n%v, got\n%v", expectedDayEvent, program.Tuesday.Night)
	}
	// Revert back the config.
	config.PeakProgram.MaintainNormalTempBeforePreHeat = false

	// Four hours later, at 8h, we should still have the same program since the
	// peak period isn't over.
	now = now.Add(4 * time.Hour)
	program = AssembleProgram(config, now, events, false)

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
	program = AssembleProgram(config, now, events, false)

	// Expect the current day's program to handle the afteroon peak period.
	expectedPeakProgram = DailyProgram{
		Morning: DayEvent{Time: 15 * time.Hour, Heat: 21 + 2, Cool: 24},
		Day:     DayEvent{Time: 16*time.Hour - 2*time.Minute, Heat: 21 - 2, Cool: 24},
		Evening: DayEvent{Time: 20*time.Hour + 2*time.Minute, Heat: 21, Cool: 24},
		Night:   DayEvent{Time: 21 * time.Hour, Heat: 20, Cool: 25},
	}
	if program.Wednesday != expectedPeakProgram {
		t.Errorf("at time %v, want\n%v, got\n%v", now, expectedPeakProgram, program.Wednesday)
	}
	// Expect the previous day's night event to set the usual temperature.
	expectedDayEvent = DayEvent{Time: 21 * time.Hour, Heat: 20, Cool: 24}
	if program.Tuesday.Night != expectedDayEvent {
		t.Errorf("want\n%v, got\n%v", expectedDayEvent, program.Tuesday.Night)
	}

	// Twelve hours later, at 22h, we should be back on the normal program.
	now = now.Add(12 * time.Hour)
	program = AssembleProgram(config, now, events, false)
	if program != config.NormalProgram {
		t.Errorf("at time %v, want\n%v, got\n%v", now, expectedPeakProgram, program.Wednesday)
	}
}
