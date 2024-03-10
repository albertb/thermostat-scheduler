package client

import (
	"log"
	"time"
)

func AssembleProgram(config Config, now time.Time, events []DailyPeakEvents, verbose bool) WeeklyProgram {
	wp := config.NormalProgram
	for _, de := range events {
		// Reinterpret the date as midnight in the local timezone, not UTC.
		date := time.Date(de.Date.Year(), de.Date.Month(), de.Date.Day(),
			0, 0, 0, 0, time.Local)

		for _, e := range de.Events {
			start := date.Add(e.Start)
			end := date.Add(e.End)

			// Only consider events that end in the next 12 hours.
			if now.Before(end) && now.Add(12*time.Hour).After(end) {
				if verbose {
					log.Println("Found relevant event: ", e)
				}

				today := getDailyProgram(&wp, start.Weekday())
				yesterday := getPreviousDailyProgram(&wp, start.Weekday())

				// Find the program that runs before pre-heating is meant to start.
				beforePreHeating := getPreviousEvent(wp,
					e.Start-config.PeakProgram.PreHeatDuration, start.Weekday())

				// If configured to maintain normal temp before pre-heating,
				// copy the program that runs during the peak period instead of
				// the one that runs before pre-heating.
				if config.PeakProgram.MaintainNormalTempBeforePreHeat {
					beforePreHeating = getNextEvent(wp, e.Start, start.Weekday())
				}

				// Find the program that runs before the end of the peak period.
				// This will be used to compute the peak temperature offsets.
				beforeEnd := getPreviousEvent(wp,
					e.End+config.PeakProgram.PeakBufferDuration, end.Weekday())

				// Pre-heating starts before the peak event, with the
				// temperature offset.
				preHeating := DayEvent{
					Time: e.Start - config.PeakProgram.PreHeatDuration,
					Heat: beforeEnd.Heat + config.PeakProgram.PreHeatTempOffset,
					Cool: beforeEnd.Cool,
				}

				// The peak period uses the temperature offset.
				peakPeriod := DayEvent{
					Time: e.Start - config.PeakProgram.PeakBufferDuration,
					Heat: beforeEnd.Heat + config.PeakProgram.PeakTempOffset,
					Cool: beforeEnd.Cool,
				}

				// Back to normal once the peak period is over.
				backToNormal := DayEvent{
					Time: e.End + config.PeakProgram.PeakBufferDuration,
					Heat: beforeEnd.Heat,
					Cool: beforeEnd.Cool,
				}

				// Find the program that runs after the end of the peak period.
				afterPeakPeriod := getNextEvent(wp,
					e.End+config.PeakProgram.PeakBufferDuration, end.Weekday())

				// Update the weekly program with the daily programs that were
				// computed.
				yesterday.Night.Heat = beforePreHeating.Heat
				yesterday.Night.Cool = beforePreHeating.Cool
				today.Morning.Copy(preHeating)
				today.Day.Copy(peakPeriod)
				today.Evening.Copy(backToNormal)
				today.Night.Copy(afterPeakPeriod)

				// Break early -- we only handle the one peak event.
				break
			}
		}
	}
	return wp
}

func getDailyProgram(wp *WeeklyProgram, weekday time.Weekday) *DailyProgram {
	switch weekday {
	case time.Sunday:
		return &wp.Sunday
	case time.Monday:
		return &wp.Monday
	case time.Tuesday:
		return &wp.Tuesday
	case time.Wednesday:
		return &wp.Wednesday
	case time.Thursday:
		return &wp.Thursday
	case time.Friday:
		return &wp.Friday
	case time.Saturday:
		return &wp.Saturday
	}
	return nil
}

func getPreviousDailyProgram(wp *WeeklyProgram, weekday time.Weekday) *DailyProgram {
	previousWeekday := time.Weekday((int(weekday) + 7 - 1) % 7)
	return getDailyProgram(wp, previousWeekday)
}

func getNextDailyProgram(wp *WeeklyProgram, weekday time.Weekday) *DailyProgram {
	nextWeekday := time.Weekday((int(weekday) + 7 + 1) % 7)
	return getDailyProgram(wp, nextWeekday)
}

func getPreviousEvent(wp WeeklyProgram, t time.Duration, weekday time.Weekday) DayEvent {
	dp := getDailyProgram(&wp, weekday)
	if dp.Night.Time <= t {
		return dp.Night
	}
	if dp.Evening.Time <= t {
		return dp.Evening
	}
	if dp.Day.Time <= t {
		return dp.Day
	}
	if dp.Morning.Time <= t {
		return dp.Morning
	}

	// Fallback to the previous day's night program.
	dp = getPreviousDailyProgram(&wp, weekday)
	return DayEvent{
		Time: 0,
		Heat: dp.Night.Heat,
		Cool: dp.Night.Cool,
	}
}

func getNextEvent(wp WeeklyProgram, t time.Duration, weekday time.Weekday) DayEvent {
	dp := getDailyProgram(&wp, weekday)
	if dp.Morning.Time >= t {
		return dp.Morning
	}
	if dp.Day.Time >= t {
		return dp.Day
	}
	if dp.Evening.Time >= t {
		return dp.Evening
	}
	if dp.Night.Time >= t {
		return dp.Night
	}

	// Fallback to the next day's morning event.
	dp = getNextDailyProgram(&wp, weekday)
	return DayEvent{
		Time: t,
		Heat: dp.Morning.Heat,
		Cool: dp.Morning.Cool,
	}
}
