# fdroidcl

[F-Droid](https://f-droid.org/) desktop client.

This is **not** a replacement for the [Android client](https://gitlab.com/fdroid/fdroidclient).
While the Android client integrates with the system with regular update checks
and notifications, this is a command line client that talks to connected
devices via [ADB](https://developer.android.com/tools/help/adb.html).

For simplicity, it tries to follow the `apt-get`/`apt-cache` commands where it
makes sense such as `update`, `show`, `install` and `remove`.

### Current features

 * Single repo support
 * Update the index
 * List all apps
 * Search by keywords
 * Show details of an app

### Missing features

 * Index verification via jar signature
 * Apk verification via checksum
 * Multi-repo support
 * Interaction with a device via ADB:
   - Fetch device details (Android version, architecture, ...)
   - Should use a `sync` command if always fetching the data above is slow
   - Install, update and remove applications
 * Interaction with multiple devices at once via ADB:
   - Transfer apps and their data from one device to another
 * Apk caching

### Advantages over the Android client

 * Command line interface
 * Batch install/update/remove apps without root nor system privileges
 * Handle multiple Android devices

### What it will never do

 * Run as a daemon, e.g. periodic index updates
 * Graphical user interface
 * Act as an F-Droid server
 * Swap apps with devices running the Android client
