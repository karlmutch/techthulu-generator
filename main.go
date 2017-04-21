package main

// This code is being used to generate test data, and prototype things in general.
// It is not generally intended to be used as is.

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mgutz/logxi/v1"
)

var (
	outputDir = flag.String("output", "./scenario", "generate test data at this location")
	logLevel  = flag.String("loglevel", "warning", "Set the desired log level")
)

// create Logger interface
var logW = log.NewLogger(log.NewConcurrentWriter(os.Stdout), "techthulu-generator")

// This module implements a module to handle communications
// with the tecthulhu device.  These devices appear to provide a WiFi
// like capability but the documentation appears to indicate a serial
// like communications protocol

type resonator struct {
	Position string `json:"position"`
	Level    int    `json:"level"`
	Health   int    `json:"health"`
	Owner    string `json:"owner"`
}

type status struct {
	Title              string      `json:"title"`
	Owner              string      `json:"owner"`
	Level              int         `json:"level"`
	Health             int         `json:"health"`
	ControllingFaction string      `json:"controllingFaction"`
	Mods               []string    `json:"mods"`
	Resonators         []resonator `json:"resonators"`
}

type portalStatus struct {
	Status status `json:"status"`
}

func main() {

	flag.Parse()

	// Wait until intialization is over before applying the log level
	logW.SetLevel(log.LevelInfo)

	switch strings.ToLower(*logLevel) {
	case "debug":
		logW.SetLevel(log.LevelDebug)
	case "info":
		logW.SetLevel(log.LevelInfo)
	case "warning", "warn":
		logW.SetLevel(log.LevelWarn)
	case "error", "err":
		logW.SetLevel(log.LevelError)
	case "fatal":
		logW.SetLevel(log.LevelFatal)
	}

	// If the directory does not exist create it

	// Ensure the directory is empty

}
