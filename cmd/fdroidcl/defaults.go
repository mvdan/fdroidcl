// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"encoding/json"
	"log"
	"os"
)

var cmdDefaults = &Command{
	UsageLine: "defaults",
	Short:     "Reset to the default settings",
}

func init() {
	cmdDefaults.Run = runDefaults
}

func runDefaults(args []string) {
	if len(args) > 0 {
		log.Fatalf("No arguments allowed")
	}
	writeConfig(&config)
}

func writeConfig(c *userConfig) {
	f, err := os.Create(configPath())
	if err != nil {
		log.Fatalf("Error when creating config file: %v", err)
	}
	defer f.Close()
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		log.Fatalf("Error when encoding config file: %v", err)
	}
	if _, err := f.Write(b); err != nil {
		log.Fatalf("Error when writing config file: %v", err)
	}
}
