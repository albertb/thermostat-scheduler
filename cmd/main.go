package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"thermostat-scheduler/internal/app"
	"time"
)

var verbose = flag.Bool("v", false, "whether to print verbose output")
var dryRun = flag.Bool("n", false, "whether to do a dry-run and avoid making any changes to the thermostat program")

var configFile = flag.String("config", "",
	"location of the config file; default ~/.config/thermostat-scheduler/config.yaml")

var eventsCacheFile = flag.String("events-cache", "",
	"location of the peak events cache file; default ~/.cache/thermostat-scheduler/events.json")
var eventsCacheTTL = flag.Duration("events-cache-ttl", time.Hour*25,
	"how long to locally cache the peak events file for")

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

func getEventsCacheFileLocation() string {
	if len(*eventsCacheFile) > 0 {
		return *eventsCacheFile
	}
	return filepath.Join(getUserHomeDir(), ".cache", "thermostat-scheduler", "events.json")
}

func main() {
	flag.Parse()

	configFile, err := os.Open(getConfigFileLocation())
	if err != nil {
		log.Fatal(err)
	}

	eventsCacheFilename := getEventsCacheFileLocation()

	err = app.Run(configFile, eventsCacheFilename, *eventsCacheTTL, *verbose, *dryRun)
	if err != nil {
		log.Fatal(err)
	}
}
