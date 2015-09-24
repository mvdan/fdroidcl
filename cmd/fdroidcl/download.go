// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/mvdan/fdroidcl"
)

var cmdDownload = &Command{
	UsageLine: "download <appid...>",
	Short:     "Download an app",
}

func init() {
	cmdDownload.Run = runDownload
}

func runDownload(args []string) {
	if len(args) < 1 {
		log.Fatalf("No package names given")
	}
	apps := findApps(args)
	for _, app := range apps {
		apk := app.CurApk()
		if apk == nil {
			log.Fatalf("No current apk found for %s", app.ID)
		}
		path := downloadApk(apk)
		fmt.Printf("APK available in %s\n", path)
	}
}

func downloadApk(apk *fdroidcl.Apk) string {
	url := apk.URL()
	path := apkPath(apk.ApkName)
	if err := downloadEtag(url, path, apk.Hash.Data); err != nil {
		log.Fatalf("Could not download '%s': %v", apk.App.ID, err)
	}
	return path
}

func apkPath(apkname string) string {
	apksDir := subdir(mustCache(), "apks")
	return filepath.Join(apksDir, apkname)
}
