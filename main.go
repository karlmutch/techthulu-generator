package main

// This code is being used to generate test data, and prototype things in general.
// It is not generally intended to be used as is.

import (
	"bytes"
	"encoding/gob"
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

func (portal *status) copy() (result *status) {

	result = &status{}

	mod := &bytes.Buffer{}

	gob.NewEncoder(mod).Encode(portal)
	gob.NewDecoder(mod).Decode(result)

	return result
}

func (portal *status) fixPortal() (result *status) {

	result = &status{
		Title:              portal.Title,
		Owner:              portal.Owner,
		Level:              portal.Level,
		Health:             portal.Health,
		ControllingFaction: portal.ControllingFaction,
		Mods:               []string{},
		Resonators:         []resonator{},
	}

	for _, res := range portal.Resonators {
		if 0 == res.Health {
			continue
		}
		switch portal.ControllingFaction {
		case "1":
			res.Owner = "Morty"
		case "2":
			res.Owner = "Rick"
		}
		result.Resonators = append(result.Resonators, res)
	}

	// Check for a neutral portal before trying to
	// do the match for portal alignment etc
	if len(result.Resonators) == 0 {
		return &status{
			Title:              portal.Title,
			Owner:              "",
			Level:              0,
			Health:             0,
			ControllingFaction: "0",
			Mods:               []string{},
			Resonators:         []resonator{},
		}
	}

	result.Health = 0
	result.Level = 0
	for _, res := range result.Resonators {
		result.Health += res.Health
		result.Level += res.Level
	}

	// Integer math betwen ints is rounding down intentionally
	// based upons Ingress rules for portal leveling and health
	result.Health = result.Health / len(result.Resonators)
	result.Level = result.Level / len(result.Resonators)

	return result
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
	levels := []int{1, 1, 1, 1, 1, 1, 1, 1}

	//for _, position := range positions {
	//	portal.Status.Resonators = append(portal.Status.Resonators, resonator{Position: position, Level: 1})
	//}
	//_, second, _ := neutralToOwned(*outputDir, portal, 0)

	second := 0
	levels = []int{8, 7, 6, 6, 5, 5, 4, 4}
	for i, position := range positions {
		portal.Status.Resonators = append(portal.Status.Resonators, resonator{Position: position, Level: levels[i]})
	}
	_, second, _ = portalBuild(*outputDir, portal, second)

	writeDone(*outputDir, second)
}

func writeSlot(outputDir string, second int, portal *portalStatus) (err error) {

	dirName := path.Join(outputDir, strconv.Itoa(second), "module", "status")
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return err
	}

	output, err := json.MarshalIndent(&portalStatus{Status: portal.Status.fixPortal()}, "", "    ")
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

func neutralToOwned(scenarioDir string, template *portalStatus, offset int) (final *portalStatus, lastSecond int, err error) {

	portal := &portalStatus{
		Status: template.Status.copy(),
	}

	second := 0

	for {
		switch second {
		case 0:
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}
		case 2:
			for _, res := range template.Status.Resonators {
				portal.Status.Resonators = append(portal.Status.Resonators, resonator{Position: res.Position, Level: res.Level})
			}
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}
		case 4:
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}
		case 6, 8, 10, 12, 14, 16, 18, 20, 22, 24:
			for i, _ := range portal.Status.Resonators {
				portal.Status.Resonators[i].Health += 10
			}
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}
		case 26:

			final = &portalStatus{
				Status: portal.Status.fixPortal(),
			}
			return final, offset + second, nil
		}
		if err != nil {
			logW.Fatal(fmt.Sprintf("test data generation failed with %s", err.Error()), "error", err)
			os.Exit(-5)
		}
		second++
	}
}

func neutralToNeutralSlow(scenarioDir string, template *portalStatus, offset int) (final *portalStatus, lastSecond int, err error) {

	portal := &portalStatus{
		Status: template.Status.copy(),
	}

	second := 0

	for {
		switch second {
		case 0:
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}
		case 2:
			for _, res := range template.Status.Resonators {
				portal.Status.Resonators = append(portal.Status.Resonators, resonator{Position: res.Position, Level: res.Level})
			}
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}

		case 4:
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}

		case 6, 8, 10, 12, 14, 16, 18, 20, 22, 24:
			for i, _ := range portal.Status.Resonators {
				portal.Status.Resonators[i].Health += 10
			}
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}

		case 28, 30, 32, 34, 36, 38, 40, 42:
			for i, _ := range portal.Status.Resonators {
				portal.Status.Resonators[i].Health -= 10
			}
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}
		case 46:
			final = &portalStatus{
				Status: portal.Status.fixPortal(),
			}
			return final, offset + second, nil
		}
		if err != nil {
			logW.Fatal(fmt.Sprintf("test data generation failed with %s", err.Error()), "error", err)
			os.Exit(-5)
		}
		second++
	}
}

func portalBuild(scenarioDir string, template *portalStatus, offset int) (final *portalStatus, lastSecond int, err error) {

	portal := &portalStatus{
		Status: template.Status.copy(),
	}

	portal.Status.Resonators = []resonator{}
	level := 4
	for i, res := range template.Status.Resonators {
		// Resonators dont increase one level per position but instead
		// decreased to their maximum permitted level
		switch i {
		case 2, 4, 6, 7:
			level++
		}
		portal.Status.Resonators = append(portal.Status.Resonators, resonator{
			Position: res.Position,
			Level:    level,
			Health:   0,
		})
	}

	second := 0
	activeRes := 7

	for {
		switch second {
		// Use the follow levels of resonators at each second marked
		//   8, 7, 6, 6,  5,  5,  4,  4,
		case 0, 3, 6, 9, 12, 15, 18, 21:
			// Activate resonators starting at the right end and
			// counting down putting smaller resos at each
			// desending position
			portal.Status.Resonators[activeRes].Health = 100
			activeRes--
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}

		case 24:
			// Change faction then add resonators starting at position 0
			// with level 8 and going down as we head to the right end, so
			// prepare the new data and write the neutral portal data
			switch template.Status.ControllingFaction {
			case "1":
				portal.Status.ControllingFaction = "2"
			case "2":
				portal.Status.ControllingFaction = "1"
			}

			portal.Status.Resonators = []resonator{}
			level = 8
			for i, res := range template.Status.Resonators {
				portal.Status.Resonators = append(portal.Status.Resonators, resonator{
					Position: res.Position,
					Level:    level,
					Health:   0,
				})
				// Resonators dont increase one level per position but instead
				// decreased to their maximum permitted level
				switch i {
				case 0, 1, 3, 5:
					level--
				}
			}
			// After clearing everything make sure we write
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}
			// When we start the next test case start at res 8 and then
			// fill in the new resonators for the opposing faction, the
			// levels that each one will be will already be populated
			activeRes = 0

		case 33, 36, 39, 42, 45, 48, 51, 64:
			portal.Status.Resonators[activeRes].Health = 100
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}
			activeRes++

		case 67:
			portal.Status.Resonators = []resonator{}
			for _, res := range template.Status.Resonators {
				portal.Status.Resonators = append(portal.Status.Resonators, resonator{
					Position: res.Position,
					Level:    0,
					Health:   0,
				})
			}
			// After clearing everything make sure we write
			if err = writeSlot(scenarioDir, second+offset, portal); err != nil {
				return final, 0, err
			}

		case 76:
			final = &portalStatus{
				Status: portal.Status.fixPortal(),
			}
			return final, offset + second, nil
		}
		if err != nil {
			logW.Fatal(fmt.Sprintf("test data generation failed with %s", err.Error()), "error", err)
			os.Exit(-5)
		}
		second++
	}
}
