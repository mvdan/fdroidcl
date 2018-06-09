// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"

	"mvdan.cc/fdroidcl"
	"mvdan.cc/fdroidcl/adb"
)

var cmdInstall = &Command{
	UsageLine: "install <appid...>",
	Short:     "Install or upgrade an app",
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
	var toInstall []*fdroidcl.App
	for _, app := range apps {
		p, e := inst[app.PackageName]
		if !e {
			// installing an app from scratch
			toInstall = append(toInstall, app)
			continue
		}
		suggested := app.SuggestedApk(device)
		if suggested == nil {
			return fmt.Errorf("no suitable APKs found for %s", app.PackageName)
		}
		if p.VersCode >= suggested.VersCode {
			fmt.Fprintf(stdout, "%s is up to date\n", app.PackageName)
			// app is already up to date
			continue
		}
		// upgrading an existing app
		toInstall = append(toInstall, app)
	}
	return downloadAndDo(toInstall, device)
}

func downloadAndDo(apps []*fdroidcl.App, device *adb.Device) error {
	type downloaded struct {
		apk  *fdroidcl.Apk
		path string
	}
	toInstall := make([]downloaded, len(apps))
	for i, app := range apps {
		apk := app.SuggestedApk(device)
		if apk == nil {
			return fmt.Errorf("no suitable APKs found for %s", app.PackageName)
		}
		path, err := downloadApk(apk)
		if err != nil {
			return err
		}
		toInstall[i] = downloaded{apk: apk, path: path}
	}
	for _, t := range toInstall {
		if err := installApk(device, t.apk, t.path); err != nil {
			return err
		}
	}
	return nil
}

func installApk(device *adb.Device, apk *fdroidcl.Apk, path string) error {
	fmt.Fprintf(stdout, "Installing %s\n", apk.AppID)
	if err := device.Install(path); err != nil {
		return fmt.Errorf("could not install %s: %v", apk.AppID, err)
	}
	return nil
}
