package app

import (
	"errors"
	"io"
	"log"
	"thermostat-scheduler/internal/client"
	"thermostat-scheduler/internal/config"
	"thermostat-scheduler/internal/events"
	"thermostat-scheduler/internal/program"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Run(configReader io.Reader, eventsCacheFilename string, eventsCacheTTL time.Duration, verbose, dryRun bool) error {
	cfg, err := config.ReadConfig(configReader)
	if err != nil {
		return err
	}

	events, err := events.GetPeakEvents(eventsCacheFilename, eventsCacheTTL)
	if err != nil {
		return err
	}

	// Based on the config and the list of peak events, assemble a program for
	// the current week.
	wp := program.AssembleProgram(cfg, time.Now(), events, verbose)
	newStateData := wp.ToStateData()

	apiClient := client.New()
	err = apiClient.Login(cfg.Username, cfg.Password)
	if err != nil {
		log.Fatal(err)
	}

	devices, err := apiClient.Devices()
	if err != nil {
		return err
	}

	if len(devices) < 1 {
		return errors.New("expected one device, found none")
	}

	if len(devices) > 1 {
		log.Println("Expected exactly one device, but found ", devices, ". Using the first one.")
	}
	device := devices[0]

	if device.StateData == newStateData {
		if verbose {
			log.Println("No changes required to the thermostat program.")
		}
		return nil
	}

	log.Printf("The thermostat program differs from the one that was computed:\n%v", cmp.Diff(device.StateData, newStateData))
	if dryRun {
		log.Println("Dry-run; exiting early without any modifications.")
		return nil
	}

	_, err = apiClient.SetDeviceAttributes(device.UUID, newStateData)
	if err != nil {
		return err
	}
	return nil
}
