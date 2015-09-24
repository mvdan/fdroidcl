// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"log"
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
	paths := make([]string, len(apps))
	for i, app := range apps {
		apk := app.CurApk()
		if apk == nil {
			log.Fatalf("No current apk found for %s", app.ID)
		}
		url := apk.URL()
		path := apkPath(apk.ApkName)
		if err := downloadEtag(url, path, apk.Hash.Data); err != nil {
			log.Fatalf("Could not download '%s': %v", app.ID, err)
		}
		paths[i] = path
	}
	for i, app := range apps {
		path := paths[i]
		fmt.Printf("Installing %s... ", app.ID)
		if err := device.Install(path); err != nil {
			fmt.Println()
			log.Fatalf("Could not install '%s': %v", app.ID, err)
		}
		fmt.Println("done")
	}
}
