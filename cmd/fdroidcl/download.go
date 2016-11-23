// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
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

func runDownload(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no package names given")
	}
	apps, err := findApps(args)
	if err != nil {
		return err
	}
	device, err := maybeOneDevice()
	if err != nil {
		return err
	}
	for _, app := range apps {
		apk := app.SuggestedApk(device)
		if apk == nil {
			return fmt.Errorf("no suggested APK found for %s", app.ID)
		}
		path, err := downloadApk(apk)
		if err != nil {
			return err
		}
		fmt.Printf("APK available in %s\n", path)
	}
	return nil
}

func downloadApk(apk *fdroidcl.Apk) (string, error) {
	url := apk.URL()
	path := apkPath(apk.ApkName)
	if err := downloadEtag(url, path, apk.Hash.Data); err == errNotModified {
	} else if err != nil {
		return "", fmt.Errorf("could not download %s: %v", apk.AppID, err)
	}
	return path, nil
}

func apkPath(apkname string) string {
	apksDir := subdir(mustCache(), "apks")
	return filepath.Join(apksDir, apkname)
}
