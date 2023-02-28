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
		device.ABIs = getAbis(device, props)
		if len(device.ABIs) == 0 {
			return nil, fmt.Errorf("failed to get device ABIs")
		}
		api, err := AdbPropFallback(device, props, "ro.build.version.sdk")
		device.APILevel, _ = strconv.Atoi(api)
		if err != nil || device.APILevel == 0 {
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

func (d *Device) AdbProp(property string) (string, error) {
	cmd := d.AdbShell("getprop", property)
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	result := string(stdout)
	if strings.HasSuffix(result, "\r\n") {
		result = strings.TrimSuffix(result, "\r\n")
	} else if strings.HasSuffix(result, "\n") {
		result = strings.TrimSuffix(result, "\n")
	} else if strings.HasSuffix(result, "\r") {
		result = strings.TrimSuffix(result, "\r")
	}
	if result == "" {
		return "", fmt.Errorf("could not get property %s", property)
	}
	return result, nil
}

func AdbPropFallback(device *Device, props map[string]string, property string) (string, error) {
	if value, e := props[property]; e {
		return value, nil
	}
	return device.AdbProp(property)
}

func getFailureCode(r *regexp.Regexp, line string) string {
	return r.FindStringSubmatch(line)[1]
}

func getAbis(device *Device, props map[string]string) []string {
	// Android 5.0 and later specify a list of ABIs
	if abilist, err := AdbPropFallback(device, props, "ro.product.cpu.abilist"); err == nil {
		return strings.Split(abilist, ",")
	}
	// Older Android versions specify one primary ABI and optionally
	// one secondary ABI
	abi, err := AdbPropFallback(device, props, "ro.product.cpu.abi")
	if err != nil {
		return nil
	}
	if abi2, err := AdbPropFallback(device, props, "ro.product.cpu.abi2"); err == nil {
		return []string{abi, abi2}
	}
	return []string{abi}
}

var installFailureRegex = regexp.MustCompile(`^Failure \[INSTALL_(.+)\]$`)

func (d *Device) Install(path string) error {
	cmd := d.AdbCmd(append([]string{"install", "-r"}, path)...)
	output, err := cmd.CombinedOutput()
	line := getResultLine(output)
	if err == nil && line == "Success" {
		return nil
	}
	errMsg := parseError(getFailureCode(installFailureRegex, line))
	if err != nil {
		return fmt.Errorf("%v: %v", err, errMsg)
	}
	return errMsg
}

func (d *Device) InstallUser(path, user string) error {
	cmd := d.AdbCmd(append([]string{"install", "-r", "--user"}, user, path)...)
	output, err := cmd.CombinedOutput()
	line := getResultLine(output)
	if err == nil && line == "Success" {
		return nil
	}
	errMsg := parseError(getFailureCode(installFailureRegex, line))
	if err != nil {
		return fmt.Errorf("%v: %v", err, errMsg)
	}
	return errMsg
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
	line := getResultLine(output)
	if err == nil && line == "Success" {
		return nil
	}
	errMsg := parseError(getFailureCode(deleteFailureRegex, line))
	if err != nil {
		return fmt.Errorf("%v: %v", err, errMsg)
	}
	return errMsg
}

func (d *Device) UninstallUser(pkg, user string) error {
	cmd := d.AdbCmd("uninstall", "--user", user, pkg)
	output, err := cmd.CombinedOutput()
	line := getResultLine(output)
	if err == nil && line == "Success" {
		return nil
	}
	errMsg := parseError(getFailureCode(deleteFailureRegex, line))
	if err != nil {
		return fmt.Errorf("%v: %v", err, errMsg)
	}
	return errMsg
}

type Package struct {
	ID                   string
	VersCode             int
	VersName             string
	IsSystem             bool
	InstalledForUsers    []int
	NotInstalledForUsers []int
}

var (
	packageRegex           = regexp.MustCompile(`^  Package \[([^\s]+)\]`)
	verCodeRegex           = regexp.MustCompile(`^    versionCode=([0-9]+)`)
	verNameRegex           = regexp.MustCompile(`^    versionName=(.+)`)
	systemRegex            = regexp.MustCompile(`^    pkgFlags=\[.*\bSYSTEM\b.*\]`)
	installedUsersRegex    = regexp.MustCompile(`^    User (\d+): .*installed=true`)
	notInstalledUsersRegex = regexp.MustCompile(`^    User (\d+): .*installed=false`)
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
				cur.IsSystem = false
				cur.InstalledForUsers = make([]int, 0)
				cur.NotInstalledForUsers = make([]int, 0)
			}
			cur.ID = m[1]
		} else if m := verCodeRegex.FindStringSubmatch(l); m != nil {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				panic(err)
			}
			cur.VersCode = n
		} else if m := verNameRegex.FindStringSubmatch(l); m != nil {
			cur.VersName = m[1]
		} else if systemRegex.MatchString(l) {
			cur.IsSystem = true
		} else if m := installedUsersRegex.FindStringSubmatch(l); m != nil {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				panic(err)
			}
			cur.InstalledForUsers = append(cur.InstalledForUsers, n)
		} else if m := notInstalledUsersRegex.FindStringSubmatch(l); m != nil {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				panic(err)
			}
			cur.NotInstalledForUsers = append(cur.NotInstalledForUsers, n)
		}
	}
	if !first {
		packages[cur.ID] = cur
	}
	return packages, nil
}

var currentUserIdRegex = regexp.MustCompile(`^ *mUserLru: \[.*\b(\d+)\b\]`)

func (d *Device) CurrentUserId() (int, error) {
	cmd := d.AdbShell("dumpsys", "activity")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return -1, err
	}
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := currentUserIdRegex.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			panic(err)
		}
		return n, nil
	}
	return -1, fmt.Errorf("could not get current user id")
}

func AllUserIds(installed map[string]Package) map[int]struct{} {
	uidMap := make(map[int]struct{})
	for _, pkg := range installed {
		for _, uid := range pkg.InstalledForUsers {
			uidMap[uid] = struct{}{}
		}
	}
	return uidMap
}
