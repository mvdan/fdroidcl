// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"errors"
	"fmt"
	"strconv"
)

var cmdUninstall = &Command{
	UsageLine: "uninstall <appid...>",
	Short:     "Uninstall an app",
}

var (
	uninstallUser = cmdUninstall.Fset.String("user", "all", "Uninstall for specified user <USER_ID|current|all>")
)

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
	if *uninstallUser == "current" {
		uid, err := device.CurrentUserId()
		if err != nil {
			return err
		}
		*uninstallUser = strconv.Itoa(uid)
	}
	inst, err := device.Installed()
	if err != nil {
		return err
	}
	for _, id := range args {
		var err error
		fmt.Printf("Uninstalling %s\n", id)
		if _, installed := inst[id]; installed {
			if *uninstallUser == "all" {
				err = device.Uninstall(id)
			} else {
				err = device.UninstallUser(id, *uninstallUser)
			}
		} else {
			err = errors.New("not installed")
		}
		if err != nil {
			return fmt.Errorf("could not uninstall %s: %v", id, err)
		}
	}
	return nil
}
