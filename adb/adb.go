/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package adb

import (
	"bufio"
	"os/exec"
	"regexp"
)

func StartServer() error {
	cmd := exec.Command("adb", "start-server")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

type Device string

var deviceRegex = regexp.MustCompile(`^(.+)\tdevice`)

func Devices() ([]Device, error) {
	cmd := exec.Command("adb", "devices")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	var devices []Device
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		m := deviceRegex.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		device := Device(m[1])
		if device == "" {
			continue
		}
		devices = append(devices, device)
	}
	return devices, nil
}
