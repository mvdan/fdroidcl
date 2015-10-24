// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"log"
)

var cmdUpgrade = &Command{
	UsageLine: "upgrade <appid...>",
	Short:     "Upgrade an app",
}

func init() {
	cmdUpgrade.Run = runUpgrade
}

func runUpgrade(args []string) {
	if len(args) < 1 {
		log.Fatalf("No package names given")
	}
	device := mustOneDevice()
	apps := findApps(args)
	inst := mustInstalled(device)
	for _, app := range apps {
		_, e := inst[app.ID]
		if !e {
			log.Fatalf("%s is not installed", app.ID)
		}
	}
	downloadAndInstall(apps, device)
}
