// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
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
	for _, id := range args {
		fmt.Printf("Uninstalling %s... ", id)
		if err := device.Uninstall(id); err != nil {
			fmt.Println()
			log.Fatalf("Could not uninstall '%s': %v", id, err)
		}
		fmt.Println("done")
	}
}
