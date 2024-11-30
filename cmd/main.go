package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"thermostat-scheduler/internal/app"
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

func getEventsFileLocation() string {
	if len(*eventsFile) > 0 {
		return *eventsFile
	}
	return filepath.Join(getUserHomeDir(), ".config", "thermostat-scheduler", "events.yaml")
}

func main() {
	flag.Parse()

	configFile, err := os.Open(getConfigFileLocation())
	if err != nil {
		log.Fatal(err)
	}

	eventsFile, err := os.Open(getEventsFileLocation())
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(configFile, eventsFile, *verbose, *dryRun)
	if err != nil {
		log.Fatal(err)
	}
}
