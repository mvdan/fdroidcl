/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package adb

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

const (
	port = 5037
)

func IsServerRunning() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

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

type Package struct {
	ID    string
	VName string
	VCode int

}

var packageRegex = regexp.MustCompile(`Package \[([^\s]+)\]`)

func (d Device) Installed() ([]Package, error) {
	cmd := d.AdbShell("dumpsys", "package", "packages")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	var packages []Package
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		m := packageRegex.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		p := Package{
			ID: m[1],
		}
		packages = append(packages, p)
	}
	return packages, nil
}
