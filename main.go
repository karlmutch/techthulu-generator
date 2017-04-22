package main

// This code is being used to generate test data, and prototype things in general.
// It is not generally intended to be used as is.

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mgutz/logxi/v1"
)

var (
	outputDir = flag.String("output", "./scenario", "generate test data at this location")
	logLevel  = flag.String("loglevel", "warning", "Set the desired log level")

	// create Logger interface
	logW = log.NewLogger(log.NewConcurrentWriter(os.Stdout), "techthulu-generator")
)

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
	Status *status `json:"status"`
}

func (portal *status) fixPortal() {

	// Check for a neutral portal before trying to
	// do the match for portal alignment etc
	if len(portal.Resonators) == 0 {
		portal.Health = 0
		portal.Level = 0
		portal.Owner = ""
		portal.Mods = []string{}
		return
	}

	hlth := 0
	levels := 0
	for _, res := range portal.Resonators {
		hlth += res.Health
		levels += res.Level
	}
	// Integer math betwen ints is roundign down intentionally
	// based upons Ingress rules for portal leveling and health
	portal.Health = hlth / len(portal.Resonators)
	portal.Level = levels / len(portal.Resonators)
}

func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return true, nil
		}
		return false, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func setLogLevel() {
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
}

func main() {

	flag.Parse()

	setLogLevel()

	// If the directory does not exist create it
	exists, err := dirExists(*outputDir)
	if err != nil {
		logW.Fatal(err.Error(), "error", err)
		os.Exit(-2)
	}
	if !exists {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			logW.Fatal(fmt.Sprintf("could not create %s due to %s", *outputDir, err.Error()), "error", err)
			os.Exit(-1)
		}
	}

	// Ensure the directory is empty
	files, err := ioutil.ReadDir(*outputDir)
	if err != nil {
		logW.Fatal(err.Error(), "error", err)
		os.Exit(-3)
	}
	if len(files) != 0 {
		logW.Fatal(fmt.Sprintf("Your scenario directory has files already in it"))
		os.Exit(-4)
	}

	portal := &portalStatus{
		Status: &status{
			Title:              "Camp Navarro",
			Health:             0,
			Level:              0,
			Owner:              "",
			ControllingFaction: "1",
			Mods:               []string{},
			Resonators:         []resonator{},
		},
	}

	positions := []string{"E", "NE", "N", "NW", "W", "SW", "S", "SE"}

	// Start at second 0 with nothing then go and step once every 2 seconds
	// using a switch for each step along the way generating changes and writting
	// then until we are done
	second := 0

	func() {
		for {
			switch second {
			case 0:
				err = writeSlot(*outputDir, 0, portal)
			case 2:
				for _, position := range positions {
					portal.Status.Resonators = append(portal.Status.Resonators, resonator{Position: position})
				}
				err = writeSlot(*outputDir, 0, portal)
			case 4:
				err = writeSlot(*outputDir, 0, portal)
			case 6, 8, 10, 12, 14, 16, 18, 20, 22, 24:
				for i, _ := range portal.Status.Resonators {
					portal.Status.Resonators[i].Health += 10
					portal.Status.Resonators[i].Level = 1
					portal.Status.Resonators[i].Owner = "Morty"
				}
				err = writeSlot(*outputDir, second, portal)
			case 26:
				return
			}
			if err != nil {
				logW.Fatal(fmt.Sprintf("test data generation failed with %s", err.Error()), "error", err)
				os.Exit(-5)
			}
			second++
		}
	}()
	writeDone(*outputDir, second)
}

func writeSlot(outputDir string, second int, portal *portalStatus) (err error) {

	portal.Status.fixPortal()

	dirName := path.Join(outputDir, strconv.Itoa(second), "module", "status")
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return err
	}

	output, err := json.MarshalIndent(portal, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(dirName, "json"), output, 0755)
}

func writeDone(outputDir string, second int) (err error) {
	dirName := path.Join(outputDir, strconv.Itoa(second))
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(dirName, "finish"), []byte{}, 0755)
}
