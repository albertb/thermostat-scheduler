package program

import (
	"log"
	"thermostat-scheduler/internal/config"
	"thermostat-scheduler/internal/events"
	"time"
)

// Assemble a weekly program given a config, current time, and list of peak events.
func AssembleProgram(cfg config.Config, now time.Time, events []events.PeakEvent, verbose bool) config.WeeklyProgram {
	wp := cfg.NormalProgram
	for _, e := range events {
		startHour := hoursFromMidnight(e.Start)
		endHour := hoursFromMidnight(e.End)

		// Only consider events that end in the next 12 hours.
		if now.Before(e.End) && now.Add(12*time.Hour).After(e.End) {
			if verbose {
				log.Println("Found relevant event: ", e)
			}

			today := wp.DailyProgramOn(e.Start.Weekday())
			yesterday := wp.DailyProgramBefore(e.Start.Weekday())

			// Find the program that runs before pre-heating is meant to start.
			beforePreHeating := wp.DayEventBefore(e.Start.Weekday(),
				startHour-cfg.PeakProgram.PreHeatDuration)

			// If configured to maintain normal temp before pre-heating,
			// copy the program that runs during the peak period instead of
			// the one that runs before pre-heating.
			if cfg.PeakProgram.MaintainNormalTempBeforePreHeat {
				beforePreHeating = wp.DayEventAfter(e.Start.Weekday(), startHour)
			}

			// Find the program that runs before the end of the peak period.
			// This will be used to compute the peak temperature offsets.
			beforeEnd := wp.DayEventBefore(e.End.Weekday(),
				endHour+cfg.PeakProgram.PeakBufferDuration)

			// Pre-heating starts before the peak event, with the
			// temperature offset.
			preHeating := config.DayEvent{
				Time: startHour - cfg.PeakProgram.PreHeatDuration,
				Heat: beforeEnd.Heat + cfg.PeakProgram.PreHeatTempOffset,
				Cool: beforeEnd.Cool,
			}

			// The peak period uses the temperature offset.
			peakPeriod := config.DayEvent{
				Time: startHour - cfg.PeakProgram.PeakBufferDuration,
				Heat: beforeEnd.Heat + cfg.PeakProgram.PeakTempOffset,
				Cool: beforeEnd.Cool,
			}

			// Back to normal once the peak period is over.
			backToNormal := config.DayEvent{
				Time: endHour + cfg.PeakProgram.PeakBufferDuration,
				Heat: beforeEnd.Heat,
				Cool: beforeEnd.Cool,
			}

			// Find the program that runs after the end of the peak period.
			afterPeakPeriod := wp.DayEventAfter(e.End.Weekday(),
				endHour+cfg.PeakProgram.PeakBufferDuration)

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
	return wp
}

func hoursFromMidnight(t time.Time) time.Duration {
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return t.Sub(midnight)
}
