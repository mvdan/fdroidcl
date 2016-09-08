// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"log"

	"github.com/mvdan/fdroidcl"
	"github.com/mvdan/fdroidcl/adb"
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
		p, e := inst[app.ID]
		if !e {
			log.Fatalf("%s is not installed", app.ID)
		}
		suggested := app.SuggestedApk(device)
		if suggested == nil {
			log.Fatalf("No suitable APKs found for %s", app.ID)
		}
		if p.VCode >= suggested.VCode {
			log.Fatalf("%s is up to date", app.ID)
		}
	}
	downloadAndDo(apps, device, upgradeApk)
}

func upgradeApk(device *adb.Device, apk *fdroidcl.Apk, path string) {
	fmt.Printf("Upgrading %s... ", apk.AppID)
	if err := device.Upgrade(path); err != nil {
		fmt.Println()
		log.Fatalf("Could not upgrade %s: %v", apk.AppID, err)
	}
	fmt.Println("done")
}
