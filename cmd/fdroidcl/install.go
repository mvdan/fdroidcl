/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/mvdan/appdir"
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
	apksDir := apksCacheDir()
	for i, app := range apps {
		apk := app.CurApk
		url := fmt.Sprintf("%s/%s", repoURL, apk.ApkName)
		path := filepath.Join(apksDir, apk.ApkName)
		if err := downloadEtag(url, path, apk.Hash.Data); err != nil {
			log.Fatalf("Could not download '%s': %v", app.ID, err)
		}
		paths[i] = path
	}
	for i, app := range apps {
		path := paths[i]
		if err := device.Install(path); err != nil {
			log.Fatalf("Could not install '%s': %v", app.ID, err)
		}
	}
}

func apksCacheDir() string {
	cache, err := appdir.Cache()
	if err != nil {
		log.Fatalf("Could not determine cache dir: %v", err)
	}
	return appSubdir(cache, "apks")
}

func apkPath(apkname string) string {
	cache, err := appdir.Cache()
	if err != nil {
		log.Fatalf("Could not determine cache dir: %v", err)
	}
	return filepath.Join(appSubdir(cache, "apks"), apkname)
}
