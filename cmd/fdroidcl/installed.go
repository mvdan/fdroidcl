/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"log"

	"github.com/mvdan/fdroidcl"
	"github.com/mvdan/fdroidcl/adb"
)

var cmdInstalled = &Command{
	UsageLine: "installed",
	Short:     "List installed apps",
}

func init() {
	cmdInstalled.Run = runInstalled
}

func runInstalled(args []string) {
	index := mustLoadIndex()
	startAdbIfNeeded()
	device := oneDevice()
	installed := mustInstalled(device)
	apps := filterAppsInstalled(index.Apps, installed)
	printApps(apps)
}

func mustInstalled(device adb.Device) []string {
	installed, err := device.Installed()
	if err != nil {
		log.Fatalf("Could not get installed packages: %v", err)
	}
	return installed
}

func filterAppsInstalled(apps []fdroidcl.App, installed []string) []fdroidcl.App {
	instMap := make(map[string]struct{}, len(installed))
	for _, id := range installed {
		instMap[id] = struct{}{}
	}
	var result []fdroidcl.App
	for _, app := range apps {
		if _, e := instMap[app.ID]; !e {
			continue
		}
		result = append(result, app)
	}
	return result
}
