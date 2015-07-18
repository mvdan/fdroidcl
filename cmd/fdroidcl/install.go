// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"log"
	"path/filepath"
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
		apk := app.CurApk
		url := fmt.Sprintf("%s/%s", repoURL, apk.ApkName)
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

func apkPath(apkname string) string {
	apksDir := subdir(mustCache(), "apks")
	return filepath.Join(apksDir, apkname)
}
