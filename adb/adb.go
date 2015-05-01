/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package adb

import (
	"os/exec"
)

func StartServer() error {
	cmd := exec.Command("adb", "start-server")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func DeviceList() ([]string, error) {
	cmd := exec.Command("adb", "devices")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	_ = stdout
	return nil, nil
}
