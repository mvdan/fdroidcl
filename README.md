# fdroidcl

[![GoDoc](https://godoc.org/github.com/mvdan/fdroidcl?status.svg)](https://godoc.org/github.com/mvdan/fdroidcl)
[![Build Status](https://travis-ci.org/mvdan/fdroidcl.svg?branch=master)](https://travis-ci.org/mvdan/fdroidcl)

[F-Droid](https://f-droid.org/) desktop client.

	go get -u github.com/mvdan/fdroidcl/cmd/fdroidcl

This is **not** a replacement for the [Android client](https://gitlab.com/fdroid/fdroidclient).

While the Android client integrates with the system with regular update checks
and notifications, this is a command line client that talks to connected
devices via [ADB](https://developer.android.com/tools/help/adb.html).

### Quickstart

Download the index:

	fdroidcl update

Show all available apps:

	fdroidcl search

Install an app:

	fdroidcl install org.adaway

### Commands

	update                Update the index
	search <regexp...>    Search available apps
	show <appid...>       Show detailed info about an app
	devices               List connected devices
	download <appid...>   Download an app
	install <appid...>    Install an app
	upgrade <appid...>    Upgrade an app
	uninstall <appid...>  Uninstall an app
	defaults              Reset to the default settings

### Config

You can configure the repositories to use in the `config.json` file,
located in `fdroidcl`'s config directory. This will be
`~/.config/fdroidcl/config.json` on Linux.

You can run `fdroidcl defaults` to create the config with the default
settings.

### Missing features

 * Index verification via jar signature - currently relies on HTTPS
 * Interaction with multiple devices at once
 * Hardware features filtering

### Advantages over the Android client

 * Command line interface
 * Batch install/update/remove apps without root nor system privileges
 * Handle multiple Android devices
 * No need to install a client on the device

### What it will never do

 * Run as a daemon, e.g. periodic index updates
 * Graphical user interface
 * Act as an F-Droid server
 * Swap apps with devices running the Android client
