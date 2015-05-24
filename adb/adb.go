/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package adb

import (
	"bufio"
	"os/exec"
	"regexp"
	"strings"
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
	Usb     string
	Product string
	Model   string
	Device  string
}

var deviceRegex = regexp.MustCompile(`^([^\s]+)\s+device(.*)$`)

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
		if m == nil {
			continue
		}
		device := Device{
			Id: m[1],
		}
		extras := m[2]
		for _, extra := range strings.Split(extras, " ") {
			sp := strings.SplitN(extra, ":", 2)
			if len(sp) < 2 {
				continue
			}
			switch sp[0] {
			case "usb":
				device.Usb = sp[1]
			case "product":
				device.Product = sp[1]
			case "model":
				device.Model = sp[1]
			case "device":
				device.Device = sp[1]
			}
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (d Device) AdbCmd(args ...string) *exec.Cmd {
	cmdArgs := append([]string{"-s", d.Id}, args...)
	return exec.Command("adb", cmdArgs...)
}

func (d Device) AdbShell(args ...string) *exec.Cmd {
	shellArgs := append([]string{"shell"}, args...)
	return d.AdbCmd(shellArgs...)
}

func (d Device) Install(path string) error {
	cmd := d.AdbCmd("install", path)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (d Device) Uninstall(pkg string) error {
	cmd := d.AdbCmd("uninstall", pkg)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

var packageRegex = regexp.MustCompile(`^package:([^\s]+)`)

func (d Device) Installed() ([]string, error) {
	cmd := d.AdbShell("pm", "list", "packages")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	var ids []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		m := packageRegex.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		ids = append(ids, m[1])
	}
	return ids, nil
}
