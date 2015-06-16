# fdroidcl

[F-Droid](https://f-droid.org/) desktop client.

This is **not** a replacement for the [Android client](https://gitlab.com/fdroid/fdroidclient).
While the Android client integrates with the system with regular update checks
and notifications, this is a command line client that talks to connected
devices via [ADB](https://developer.android.com/tools/help/adb.html).

### Commands

	update             Update the index
	list               List all available apps
	search <term...>   Search available apps
	show <app...>      Show detailed info about an app
	devices            List connected devices
	installed          List installed apps

### Missing commands

	updates            List apps to update
	install <app...>   Install an app
	upgrade <app...>   Upgrade an app
	remove <app...>    Remove an app

### Missing features

 * Index verification via jar signature
   - Cannot be currently done since MD5WithRSA is unimplemented
 * Multi-repo support
 * Interaction with multiple devices at once

### Advantages over the Android client

 * Command line interface
 * Batch install/update/remove apps without root nor system privileges
 * Handle multiple Android devices

### What it will never do

 * Run as a daemon, e.g. periodic index updates
 * Graphical user interface
 * Act as an F-Droid server
 * Swap apps with devices running the Android client
