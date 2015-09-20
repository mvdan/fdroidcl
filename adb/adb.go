// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package adb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	port = 5037
)

var (
	// Common install and uninstall errors
	ErrInternalError  = errors.New("internal error")
	ErrUserRestricted = errors.New("user restricted")
	ErrAborted        = errors.New("aborted")

	// Install errors
	ErrAlreadyExists            = errors.New("already exists")
	ErrInvalidApk               = errors.New("invalid apk")
	ErrInvalidURI               = errors.New("invalid uri")
	ErrInsufficientStorage      = errors.New("insufficient storage")
	ErrDuplicatePackage         = errors.New("duplicate package")
	ErrNoSharedUser             = errors.New("no shared user")
	ErrUpdateIncompatible       = errors.New("update incompatible")
	ErrSharedUserIncompatible   = errors.New("shared user incompatible")
	ErrMissingSharedLibrary     = errors.New("missing shared library")
	ErrReplaceCouldntDelete     = errors.New("replace couldn't delete")
	ErrDexopt                   = errors.New("dexopt")
	ErrOlderSdk                 = errors.New("older sdk")
	ErrConflictingProvider      = errors.New("conflicting provider")
	ErrNewerSdk                 = errors.New("newer sdk")
	ErrTestOnly                 = errors.New("test only")
	ErrCPUAbiIncompatible       = errors.New("cpu abi incompatible")
	ErrMissingFeature           = errors.New("missing feature")
	ErrContainerError           = errors.New("combiner error")
	ErrInvalidInstallLocation   = errors.New("invalid install location")
	ErrMediaUnavailable         = errors.New("media unavailable")
	ErrVerificationTimeout      = errors.New("verification timeout")
	ErrVerificationFailure      = errors.New("verification failure")
	ErrPackageChanged           = errors.New("package changed")
	ErrUIDChanged               = errors.New("uid changed")
	ErrVersionDowngrade         = errors.New("version downgrade")
	ErrNotApk                   = errors.New("not apk")
	ErrBadManifest              = errors.New("bad manifest")
	ErrUnexpectedException      = errors.New("unexpected exception")
	ErrNoCertificates           = errors.New("no certificates")
	ErrInconsistentCertificates = errors.New("inconsistent certificates")
	ErrCertificateEncoding      = errors.New("certificate encoding")
	ErrBadPackageName           = errors.New("bad package name")
	ErrBadSharedUserID          = errors.New("bad shared user id")
	ErrManifestMalformed        = errors.New("manifest malformed")
	ErrManifestEmpty            = errors.New("manifest empty")
	ErrDuplicatePermission      = errors.New("duplicate permission")
	ErrNoMatchingAbis           = errors.New("no matching abis")

	// Uninstall errors
	ErrDevicePolicyManager = errors.New("device policy manager")
	ErrOwnerBlocked        = errors.New("owner blocked")
)

func parseError(s string) error {
	switch s {
	case "FAILED_ALREADY_EXISTS":
		return ErrAlreadyExists
	case "FAILED_INVALID_APK":
		return ErrInvalidApk
	case "FAILED_INVALID_URI":
		return ErrInvalidURI
	case "FAILED_INSUFFICIENT_STORAGE":
		return ErrInsufficientStorage
	case "FAILED_DUPLICATE_PACKAGE":
		return ErrDuplicatePackage
	case "FAILED_NO_SHARED_USER":
		return ErrNoSharedUser
	case "FAILED_UPDATE_INCOMPATIBLE":
		return ErrUpdateIncompatible
	case "FAILED_SHARED_USER_INCOMPATIBLE":
		return ErrSharedUserIncompatible
	case "FAILED_MISSING_SHARED_LIBRARY":
		return ErrMissingSharedLibrary
	case "FAILED_REPLACE_COULDNT_DELETE":
		return ErrReplaceCouldntDelete
	case "FAILED_DEXOPT":
		return ErrDexopt
	case "FAILED_OLDER_SDK":
		return ErrOlderSdk
	case "FAILED_CONFLICTING_PROVIDER":
		return ErrConflictingProvider
	case "FAILED_NEWER_SDK":
		return ErrNewerSdk
	case "FAILED_TEST_ONLY":
		return ErrTestOnly
	case "FAILED_CPU_ABI_INCOMPATIBLE":
		return ErrCPUAbiIncompatible
	case "FAILED_MISSING_FEATURE":
		return ErrMissingFeature
	case "FAILED_CONTAINER_ERROR":
		return ErrContainerError
	case "FAILED_INVALID_INSTALL_LOCATION":
		return ErrInvalidInstallLocation
	case "FAILED_MEDIA_UNAVAILABLE":
		return ErrMediaUnavailable
	case "FAILED_VERIFICATION_TIMEOUT":
		return ErrVerificationTimeout
	case "FAILED_VERIFICATION_FAILURE":
		return ErrVerificationFailure
	case "FAILED_PACKAGE_CHANGED":
		return ErrPackageChanged
	case "FAILED_UID_CHANGED":
		return ErrUIDChanged
	case "FAILED_VERSION_DOWNGRADE":
		return ErrVersionDowngrade
	case "PARSE_FAILED_NOT_APK":
		return ErrNotApk
	case "PARSE_FAILED_BAD_MANIFEST":
		return ErrBadManifest
	case "PARSE_FAILED_UNEXPECTED_EXCEPTION":
		return ErrUnexpectedException
	case "PARSE_FAILED_NO_CERTIFICATES":
		return ErrNoCertificates
	case "PARSE_FAILED_INCONSISTENT_CERTIFICATES":
		return ErrInconsistentCertificates
	case "PARSE_FAILED_CERTIFICATE_ENCODING":
		return ErrCertificateEncoding
	case "PARSE_FAILED_BAD_PACKAGE_NAME":
		return ErrBadPackageName
	case "PARSE_FAILED_BAD_SHARED_USER_ID":
		return ErrBadSharedUserID
	case "PARSE_FAILED_MANIFEST_MALFORMED":
		return ErrManifestMalformed
	case "PARSE_FAILED_MANIFEST_EMPTY":
		return ErrManifestEmpty
	case "FAILED_INTERNAL_ERROR":
		return ErrInternalError
	case "FAILED_USER_RESTRICTED":
		return ErrUserRestricted
	case "FAILED_DUPLICATE_PERMISSION":
		return ErrDuplicatePermission
	case "FAILED_NO_MATCHING_ABIS":
		return ErrNoMatchingAbis
	case "FAILED_ABORTED":
		return ErrAborted
	}
	return fmt.Errorf("unknown error: %s", s)
}

func IsServerRunning() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func StartServer() error {
	return exec.Command("adb", "start-server").Run()
}

type Device struct {
	ID      string
	Usb     string
	Product string
	Model   string
	Device  string
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
		line := scanner.Text()
		m := deviceRegex.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		device := &Device{
			ID: m[1],
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

func (d *Device) AdbCmd(args ...string) *exec.Cmd {
	cmdArgs := append([]string{"-s", d.ID}, args...)
	return exec.Command("adb", cmdArgs...)
}

func (d *Device) AdbShell(args ...string) *exec.Cmd {
	shellArgs := append([]string{"shell"}, args...)
	return d.AdbCmd(shellArgs...)
}

func getFailureCode(r *regexp.Regexp, line string) string {
	return r.FindStringSubmatch(line)[1]
}

var installFailureRegex = regexp.MustCompile(`^Failure \[INSTALL_(.+)\]$`)

func (d *Device) Install(path string) error {
	cmd := d.AdbCmd("install", path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	line := getResultLine(stdout)
	if line == "Success" {
		return nil
	}
	return parseError(getFailureCode(installFailureRegex, line))
}

func getResultLine(out io.ReadCloser) string {
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "Failure") || strings.HasPrefix(l, "Success") {
			return l
		}
	}
	return ""
}

var deleteFailureRegex = regexp.MustCompile(`^Failure \[DELETE_(.+)\]$`)

func (d *Device) Uninstall(pkg string) error {
	cmd := d.AdbCmd("uninstall", pkg)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	line := getResultLine(stdout)
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
