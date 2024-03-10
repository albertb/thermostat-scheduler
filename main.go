package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"thermostat-scheduler/client"
	"time"

	"github.com/google/go-cmp/cmp"
)

var verbose = flag.Bool("verbose", false, "whether to print verbose output")
var dryRun = flag.Bool("dry-run", false, "whether to do a dry-run and avoid making any changes to the thermostat program")
var configFile = flag.String("config", "",
	"location of the config file; default ~/.config/thermostat-scheduler/config.yaml")
var eventsFile = flag.String("events", "",
	"location of the peak events file; default ~/.config/thermostat-scheduler/events.yaml")

func getUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func getConfigFileLocation() string {
	if len(*configFile) > 0 {
		return *configFile
	}
	return filepath.Join(getUserHomeDir(), ".config", "thermostat-scheduler", "config.yaml")
}

func GetEventsFileLocation() string {
	if len(*eventsFile) > 0 {
		return *eventsFile
	}
	return filepath.Join(getUserHomeDir(), ".config", "thermostat-scheduler", "events.yaml")
}

func main() {
	flag.Parse()

	config, err := client.GetConfig(getConfigFileLocation())
	if err != nil {
		log.Fatal(err)
	}

	events, err := client.GetPeakEvents(GetEventsFileLocation())
	if err != nil {
		log.Fatal(err)
	}

	// Based on the config and the list of peak events, assemble a program for
	// the current week.
	wp := client.AssembleProgram(config, time.Now(), events, *verbose)
	newStateData := wp.ToStateData()

	c := client.New()
	err = c.Login(config.Username, config.Password)
	if err != nil {
		log.Fatal(err)
	}

	d, err := c.Devices()
	if err != nil {
		log.Fatal(err)
	}
	if len(d) < 1 {
		log.Fatal("Expected one device, found none")
	}
	if len(d) > 1 {
		log.Println("Expected one device, found ", d, ". Using the first one.")
	}
	t := d[0]

	if t.StateData != newStateData {
		log.Printf("The thermostat program differs from the one that was "+
			"computed:\n%v", cmp.Diff(t.StateData, newStateData))
	} else {
		if *verbose {
			log.Println("No changes required to the thermostat program.")
		}
	}

	if *dryRun {
		log.Println("Dry-run; exiting early without any modifications.")
	} else {
		err = c.SetDeviceAttributes(t.UUID, newStateData)
		if err != nil {
			log.Fatal(err)
		}
	}
}
