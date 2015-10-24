// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"log"

	"github.com/mvdan/fdroidcl"
	"github.com/mvdan/fdroidcl/adb"
)

var cmdInstall = &Command{
	UsageLine: "install <appid...>",
	Short:     "Install an app",
}

func init() {
	cmdInstall.Run = runInstall
}

func runInstall(args []string) {
	if len(args) < 1 {
		log.Fatalf("No package names given")
	}
	device := mustOneDevice()
	apps := findApps(args)
	inst := mustInstalled(device)
	for _, app := range apps {
		if _, e := inst[app.ID]; e {
			log.Fatalf("%s is already installed", app.ID)
		}
	}
	downloadAndDo(apps, device, installApk)
}

func downloadAndDo(apps []*fdroidcl.App, device *adb.Device, doApk func(*adb.Device, *fdroidcl.Apk, string)) {
	type downloaded struct {
		apk  *fdroidcl.Apk
		path string
	}
	toInstall := make([]downloaded, len(apps))
	for i, app := range apps {
		apk := app.CurApk()
		if apk == nil {
			log.Fatalf("No current apk found for %s", app.ID)
		}
		path := downloadApk(apk)
		toInstall[i] = downloaded{apk: apk, path: path}
	}
	for _, t := range toInstall {
		doApk(device, t.apk, t.path)
	}
}

func installApk(device *adb.Device, apk *fdroidcl.Apk, path string) {
	fmt.Printf("Installing %s... ", apk.App.ID)
	if err := device.Install(path); err != nil {
		fmt.Println()
		log.Fatalf("Could not install %s: %v", apk.App.ID, err)
	}
	fmt.Println("done")
}

func upgradeApk(device *adb.Device, apk *fdroidcl.Apk, path string) {
	fmt.Printf("Upgrading %s... ", apk.App.ID)
	if err := device.Upgrade(path); err != nil {
		fmt.Println()
		log.Fatalf("Could not upgrade %s: %v", apk.App.ID, err)
	}
	fmt.Println("done")
}
