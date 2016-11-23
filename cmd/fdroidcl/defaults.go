// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"encoding/json"
	"fmt"
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
		log.Fatal("No arguments allowed")
	}
	if err := writeConfig(&config); err != nil {
		log.Fatal(err)
	}
}

func writeConfig(c *userConfig) error {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return fmt.Errorf("cannot encode config: %v", err)
	}
	f, err := os.Create(configPath())
	if err != nil {
		return fmt.Errorf("cannot create config file: %v", err)
	}
	_, err = f.Write(b)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
