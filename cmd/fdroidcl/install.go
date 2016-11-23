// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"

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

func runInstall(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no package names given")
	}
	device, err := oneDevice()
	if err != nil {
		return err
	}
	apps, err := findApps(args)
	if err != nil {
		return err
	}
	inst, err := device.Installed()
	if err != nil {
		return err
	}
	for _, app := range apps {
		if _, e := inst[app.ID]; e {
			return fmt.Errorf("%s is already installed", app.ID)
		}
	}
	return downloadAndDo(apps, device, installApk)
}

func downloadAndDo(apps []*fdroidcl.App, device *adb.Device, doApk func(*adb.Device, *fdroidcl.Apk, string) error) error {
	type downloaded struct {
		apk  *fdroidcl.Apk
		path string
	}
	toInstall := make([]downloaded, len(apps))
	for i, app := range apps {
		apk := app.SuggestedApk(device)
		if apk == nil {
			return fmt.Errorf("no suitable APKs found for %s", app.ID)
		}
		path, err := downloadApk(apk)
		if err != nil {
			return err
		}
		toInstall[i] = downloaded{apk: apk, path: path}
	}
	for _, t := range toInstall {
		if err := doApk(device, t.apk, t.path); err != nil {
			return err
		}
	}
	return nil
}

func installApk(device *adb.Device, apk *fdroidcl.Apk, path string) error {
	fmt.Printf("Installing %s\n", apk.AppID)
	if err := device.Install(path); err != nil {
		return fmt.Errorf("could not install %s: %v", apk.AppID, err)
	}
	return nil
}
