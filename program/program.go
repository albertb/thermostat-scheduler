package program

import (
	"log"
	"thermostat-scheduler/config"
	"thermostat-scheduler/event"
	"time"
)

// Assemble a weekly program given a config, current time, and list of peak events.
func AssembleProgram(cfg config.Config, now time.Time, events []event.DailyPeakEvents, verbose bool) config.WeeklyProgram {
	wp := cfg.NormalProgram
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

				today := wp.DailyProgramOn(start.Weekday())
				yesterday := wp.DailyProgramBefore(start.Weekday())

				// Find the program that runs before pre-heating is meant to start.
				beforePreHeating := wp.DayEventBefore(start.Weekday(),
					e.Start-cfg.PeakProgram.PreHeatDuration)

				// If configured to maintain normal temp before pre-heating,
				// copy the program that runs during the peak period instead of
				// the one that runs before pre-heating.
				if cfg.PeakProgram.MaintainNormalTempBeforePreHeat {
					beforePreHeating = wp.DayEventAfter(start.Weekday(), e.Start)
				}

				// Find the program that runs before the end of the peak period.
				// This will be used to compute the peak temperature offsets.
				beforeEnd := wp.DayEventBefore(end.Weekday(),
					e.End+cfg.PeakProgram.PeakBufferDuration)

				// Pre-heating starts before the peak event, with the
				// temperature offset.
				preHeating := config.DayEvent{
					Time: e.Start - cfg.PeakProgram.PreHeatDuration,
					Heat: beforeEnd.Heat + cfg.PeakProgram.PreHeatTempOffset,
					Cool: beforeEnd.Cool,
				}

				// The peak period uses the temperature offset.
				peakPeriod := config.DayEvent{
					Time: e.Start - cfg.PeakProgram.PeakBufferDuration,
					Heat: beforeEnd.Heat + cfg.PeakProgram.PeakTempOffset,
					Cool: beforeEnd.Cool,
				}

				// Back to normal once the peak period is over.
				backToNormal := config.DayEvent{
					Time: e.End + cfg.PeakProgram.PeakBufferDuration,
					Heat: beforeEnd.Heat,
					Cool: beforeEnd.Cool,
				}

				// Find the program that runs after the end of the peak period.
				afterPeakPeriod := wp.DayEventAfter(end.Weekday(),
					e.End+cfg.PeakProgram.PeakBufferDuration)

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
