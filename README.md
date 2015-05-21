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

 * Multi-repo support
 * Interaction with a device via ADB:
   - Probably relying on the installed command `adb`
   - Fetch device details (Android version, architecture, ...)
   - Fetch installed applications
   - Should use a `sync` command if always fetching the data above is slow
   - Install, update and remove applications
 * Interaction with multiple devices at once via ADB:
   - Transfer apps and their data from one device to another
 * Apk caching

### Advantages over the Android client

 * Faster to use command line interface
 * Ability to batch install/update/remove without root nor system privileges
 * Handle multiple Android devices

### Android client features this will never have

 * "Update available" notifications
 * Run on Android with a user interface
 * Swap apps over WiFi or Bluetooth and local repos
