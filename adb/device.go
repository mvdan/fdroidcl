// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package adb

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Device struct {
	ID       string
	Usb      string
	Product  string
	Model    string
	Device   string
	ABIs     []string
	APILevel int
}

var deviceRegex = regexp.MustCompile(`^([^\s]+)\s+device(.*)$`)

func Devices() ([]*Device, error) {
	cmd := exec.Command("adb", "devices", "-l")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	var devices []*Device
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := deviceRegex.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		device := &Device{
			ID: m[1],
		}
		for _, extra := range strings.Split(m[2], " ") {
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

		props, err := device.AdbProps()
		if err != nil {
			return nil, err
		}
		device.ABIs = getAbis(props)
		if len(device.ABIs) == 0 {
			return nil, fmt.Errorf("failed to get device ABIs")
		}
		device.APILevel, _ = strconv.Atoi(props["ro.build.version.sdk"])
		if device.APILevel == 0 {
			return nil, fmt.Errorf("failed to get device API level")
		}

		devices = append(devices, device)
	}
	return devices, nil
}

func (d *Device) AdbCmd(args ...string) *exec.Cmd {
	cmdArgs := append([]string{"-s", d.ID}, args...)
	return exec.Command("adb", cmdArgs...)
}

func (d *Device) AdbShell(args ...string) *exec.Cmd {
	shellArgs := append([]string{"shell"}, args...)
	return d.AdbCmd(shellArgs...)
}

var propLineRegex = regexp.MustCompile(`^\[(.*)\]: \[(.*)\]$`)

func (d *Device) AdbProps() (map[string]string, error) {
	cmd := d.AdbShell("getprop")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	props := make(map[string]string)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := propLineRegex.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		key, val := m[1], m[2]
		props[key] = val
	}
	return props, nil
}

func getFailureCode(r *regexp.Regexp, line string) string {
	return r.FindStringSubmatch(line)[1]
}

func getAbis(props map[string]string) []string {
	// Android 5.0 and later specify a list of ABIs
	if abilist, e := props["ro.product.cpu.abilist"]; e {
		return strings.Split(abilist, ",")
	}
	// Older Android versions specify one primary ABI and optionally
	// one secondary ABI
	abi, e := props["ro.product.cpu.abi"]
	if !e {
		return nil
	}
	if abi2, e := props["ro.product.cpu.abi2"]; e {
		return []string{abi, abi2}
	}
	return []string{abi}
}

var installFailureRegex = regexp.MustCompile(`^Failure \[INSTALL_(.+)\]$`)

func withOpts(cmd string, opts []string, args ...string) []string {
	v := append([]string{cmd}, opts...)
	return append(v, args...)
}

func (d *Device) install(opts []string, path string) error {
	cmd := d.AdbCmd(withOpts("install", opts, path)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	line := getResultLine(output)
	if line == "Success" {
		return nil
	}
	return parseError(getFailureCode(installFailureRegex, line))
}

func (d *Device) Install(path string) error {
	return d.install(nil, path)
}

func (d *Device) Upgrade(path string) error {
	return d.install([]string{"-r"}, path)
}

func getResultLine(output []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "Success") {
			return l
		}
		failure := strings.Index(l, "Failure")
		if failure >= 0 {
			return l[failure:]
		}
	}
	return ""
}

var deleteFailureRegex = regexp.MustCompile(`^Failure \[DELETE_(.+)\]$`)

func (d *Device) Uninstall(pkg string) error {
	cmd := d.AdbCmd("uninstall", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	line := getResultLine(output)
	if line == "Success" {
		return nil
	}
	return parseError(getFailureCode(deleteFailureRegex, line))
}

type Package struct {
	ID    string
	VCode int
	VName string
}

var (
	packageRegex = regexp.MustCompile(`^  Package \[([^\s]+)\]`)
	verCodeRegex = regexp.MustCompile(`^    versionCode=([0-9]+)`)
	verNameRegex = regexp.MustCompile(`^    versionName=(.+)`)
)

func (d *Device) Installed() (map[string]Package, error) {
	cmd := d.AdbShell("dumpsys", "package", "packages")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	packages := make(map[string]Package)
	scanner := bufio.NewScanner(stdout)
	var cur Package
	first := true
	for scanner.Scan() {
		l := scanner.Text()
		if m := packageRegex.FindStringSubmatch(l); m != nil {
			if first {
				first = false
			} else {
				packages[cur.ID] = cur
				cur = Package{}
			}
			cur.ID = m[1]
		} else if m := verCodeRegex.FindStringSubmatch(l); m != nil {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				panic(err)
			}
			cur.VCode = n
		} else if m := verNameRegex.FindStringSubmatch(l); m != nil {
			cur.VName = m[1]
		}
	}
	if !first {
		packages[cur.ID] = cur
	}
	return packages, nil
}
