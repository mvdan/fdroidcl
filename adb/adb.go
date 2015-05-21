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

type Device struct {
	Id      string
	Product string
	Model   string
	Device  string
}

var deviceRegex = regexp.MustCompile(`^([^\s]+)\s+device\s+product:([^\s]+)\s+model:([^\s]+)\s+device:([^\s]+)`)

func Devices() ([]Device, error) {
	cmd := exec.Command("adb", "devices", "-l")
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
		device := Device{
			Id:      m[1],
			Product: m[2],
			Model:   m[3],
			Device:  m[4],
		}
		if device.Id == "" {
			continue
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (d Device) AdbCommand(args ...string) *exec.Cmd {
	cmdArgs := append([]string{"-s", d.Id}, args...)
	return exec.Command("adb", cmdArgs...)
}

func (d Device) Install(path string) error {
	cmd := d.AdbCommand("install", path)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (d Device) Uninstall(pkg string) error {
	cmd := d.AdbCommand("uninstall", pkg)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
