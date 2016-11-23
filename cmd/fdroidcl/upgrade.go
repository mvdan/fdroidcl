// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"

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

func runUpgrade(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no package names given")
	}
	apps, err := findApps(args)
	if err != nil {
		return err
	}
	device, err := oneDevice()
	if err != nil {
		return err
	}
	inst, err := device.Installed()
	if err != nil {
		return err
	}
	for _, app := range apps {
		p, e := inst[app.ID]
		if !e {
			return fmt.Errorf("%s is not installed", app.ID)
		}
		suggested := app.SuggestedApk(device)
		if suggested == nil {
			return fmt.Errorf("no suitable APKs found for %s", app.ID)
		}
		if p.VCode >= suggested.VCode {
			return fmt.Errorf("%s is up to date", app.ID)
		}
	}
	return downloadAndDo(apps, device, upgradeApk)
}

func upgradeApk(device *adb.Device, apk *fdroidcl.Apk, path string) error {
	fmt.Printf("Upgrading %s\n", apk.AppID)
	if err := device.Upgrade(path); err != nil {
		return fmt.Errorf("could not upgrade %s: %v", apk.AppID, err)
	}
	return nil
}
