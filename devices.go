// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"

	"mvdan.cc/fdroidcl/adb"
)

var cmdDevices = &Command{
	UsageLine: "devices",
	Short:     "List connected devices",
}

func init() {
	cmdDevices.Run = runDevices
}

func runDevices(args []string) error {
	if err := startAdbIfNeeded(); err != nil {
		return err
	}
	devices, err := adb.Devices()
	if err != nil {
		return fmt.Errorf("could not get devices: %v", err)
	}
	for _, device := range devices {
		fmt.Printf("%s - %s (%s)\n", device.ID, device.Model, device.Product)
	}
	return nil
}

func startAdbIfNeeded() error {
	if adb.IsServerRunning() {
		return nil
	}
	if err := adb.StartServer(); err != nil {
		return fmt.Errorf("could not start ADB server: %v", err)
	}
	return nil
}

func maybeOneDevice() (*adb.Device, error) {
	if err := startAdbIfNeeded(); err != nil {
		return nil, err
	}
	devices, err := adb.Devices()
	if err != nil {
		return nil, fmt.Errorf("could not get devices: %v", err)
	}
	if len(devices) < 1 {
		return nil, nil
	}
	envAndroidSerial := os.Getenv("ANDROID_SERIAL")
	if len(envAndroidSerial) > 0 {
		for _, device := range devices {
			if device.ID == envAndroidSerial {
				return device, nil
			}
		}
		return nil, fmt.Errorf("ANDROID_SERIAL set, but no device with this serial found")
	}
	if len(devices) > 1 {
		return nil, fmt.Errorf("at most one connected device can be used, if ANDROID_SERIAL was not set")
	}
	return devices[0], nil
}

func oneDevice() (*adb.Device, error) {
	device, err := maybeOneDevice()
	if err == nil && device == nil {
		err = fmt.Errorf("a connected device is needed")
	}
	return device, err
}
