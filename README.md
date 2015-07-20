# fdroidcl

[F-Droid](https://f-droid.org/) desktop client.

This is **not** a replacement for the [Android client](https://gitlab.com/fdroid/fdroidclient).
While the Android client integrates with the system with regular update checks
and notifications, this is a command line client that talks to connected
devices via [ADB](https://developer.android.com/tools/help/adb.html).

This software is **unfinished** and **unstable**. Since it lacks JAR
verification signature, it also relies solely on HTTPS for security. Use at
your own risk.

### Commands

	update                Update the index
	search <regexp...>    Search available apps
	show <appid...>       Show detailed info about an app
	devices               List connected devices
	download <appid...>   Download an app
	install <appid...>    Install an app
	uninstall <appid...>  Uninstall an app
	defaults              Reset to the default settings

### Missing commands

	upgrade <appid...>    Upgrade an app

### Missing features

 * Index verification via jar signature
   - Cannot be currently done since MD5WithRSA is unimplemented
 * Interaction with multiple devices at once

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
