// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"errors"
	"fmt"
	"log"
)

var cmdUninstall = &Command{
	UsageLine: "uninstall <appid...>",
	Short:     "Uninstall an app",
}

func init() {
	cmdUninstall.Run = runUninstall
}

func runUninstall(args []string) {
	if len(args) < 1 {
		log.Fatalf("No package names given")
	}
	device := mustOneDevice()
	inst := mustInstalled(device)
	for _, id := range args {
		var err error
		fmt.Printf("Uninstalling %s... ", id)
		if _, installed := inst[id]; installed {
			err = device.Uninstall(id)
		} else {
			err = errors.New("not installed")
		}
		if err != nil {
			fmt.Println()
			log.Fatalf("Could not uninstall %s: %v", id, err)
		}
		fmt.Println("done")
	}
}
