# fdroidcl

[![GoDoc](https://godoc.org/github.com/mvdan/fdroidcl?status.svg)](https://godoc.org/mvdan.cc/fdroidcl)

[F-Droid](https://f-droid.org/) desktop client. Requires Go 1.18 or later.

	go get mvdan.cc/fdroidcl

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

Show all available updates, and install them:

	fdroidcl search -u
	fdroidcl install -u

Unofficial packages are available on: [Debian](https://packages.debian.org/buster/fdroidcl) and [Ubuntu](https://packages.ubuntu.com/eoan/fdroidcl).

### Commands

	update                Update the index
	search [<regexp...>]  Search available apps
	show <appid...>       Show detailed info about apps
	install [<appid...>]  Install or upgrade apps
	uninstall <appid...>  Uninstall an app
	download <appid...>   Download an app
	devices               List connected devices
	list (categories)     List all known values of a kind
	defaults              Reset to the default settings
	version               Print version information
	clean                 Clean index and/or cache
	repo                  Manage repositories


An appid is just an app's unique package name. A specific version of an app can
be selected by following the appid with a colon and the version code. The
'search' and 'show' commands can be used to find these strings. For example:

	$ fdroidcl search redreader
	$ fdroidcl show org.quantumbadger.redreader
	$ fdroidcl install org.quantumbadger.redreader:85

### Config

You can configure what repositories to use in the `config.json` file. On Linux,
you will likely find it at `~/.config/fdroidcl/config.json`.

You can run `fdroidcl defaults` to create the config with the default settings.

#### *new: you can manage the repositories now directly via cli*

```
usage: fdroidcl repo

List, add, remove, enable or disable repositories.
When a repository is added, it is enabled by default.

List repositories:

        $ fdroidcl repo

Modify repositories:

        $ fdroidcl repo add <NAME> <URL>
        $ fdroidcl repo remove <NAME>
        $ fdroidcl repo enable <NAME>
        $ fdroidcl repo disable <NAME>
```

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

### FAQ

* What's the point of a desktop client?

This client works with Android devices connected via ADB; it does not install
apps on the host machine.

* Why not just use the f-droid.org website to download APKs?

That's always an option. However, an F-Droid client supports multiple
repositories, searching for apps, filtering by compatibility with your device,
showing available updates, et cetera.
