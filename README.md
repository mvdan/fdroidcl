# fdroidcl

[![GoDoc](https://godoc.org/github.com/mvdan/fdroidcl?status.svg)](https://godoc.org/mvdan.cc/fdroidcl)
[![Build Status](https://travis-ci.org/mvdan/fdroidcl.svg?branch=master)](https://travis-ci.org/mvdan/fdroidcl)

[F-Droid](https://f-droid.org/) desktop client.

	go get -u mvdan.cc/fdroidcl/cmd/fdroidcl

While the Android client integrates with the system with regular update checks
and notifications, this is a simple command line client that talks to connected
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
	install <appid...>    Install or upgrade app
	uninstall <appid...>  Uninstall an app
	defaults              Reset to the default settings

A specific version of an app can be selected by following the appid with an
colon (:) and the version code of the app to select.

### Config

You can configure what repositories to use in the `config.json` file. On Linux,
you will likely find it at `~/.config/fdroidcl/config.json`.

You can run `fdroidcl defaults` to create the config with the default settings.

### Advantages over the Android client

* Command line interface
* Batch install/update/remove apps without root nor system privileges
* No need to install a client on the device

### What it will never do

* Run as a daemon, e.g. periodic index updates
* Act as an F-Droid server
* Swap apps with devices

### Caveats

* Index verification relies on HTTPS (not the JAR signature)
* The tool can only interact with one device at a time
* Hardware compatibility of packages is not checked
