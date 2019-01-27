// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"errors"
	"fmt"
)

var cmdUninstall = &Command{
	UsageLine: "uninstall <appid...>",
	Short:     "Uninstall an app",
}

func init() {
	cmdUninstall.Run = runUninstall
}

func runUninstall(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no package names given")
	}
	device, err := oneDevice()
	if err != nil {
		return err
	}
	inst, err := device.Installed()
	if err != nil {
		return err
	}
	for _, id := range args {
		var err error
		fmt.Fprintf(stdout, "Uninstalling %s\n", id)
		if _, installed := inst[id]; installed {
			err = device.Uninstall(id)
		} else {
			err = errors.New("not installed")
		}
		if err != nil {
			return fmt.Errorf("could not uninstall %s: %v", id, err)
		}
	}
	return nil
}
