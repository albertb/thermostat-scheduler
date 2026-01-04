package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"thermostat-scheduler/internal/app"
)

var verbose = flag.Bool("v", false, "whether to print verbose output")
var dryRun = flag.Bool("n", false, "whether to do a dry-run and avoid making any changes to the thermostat program")

var configFile = flag.String("config", "",
	"location of the config file; default ~/.config/thermostat-scheduler/config.yaml")

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

func main() {
	flag.Parse()

	configFile, err := os.Open(getConfigFileLocation())
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(configFile, *verbose, *dryRun)
	if err != nil {
		log.Fatal(err)
	}
}
