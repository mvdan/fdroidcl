# fdroidcl

F-Droid desktop client.

This is **not** a replacement for the Android client. While the Android client
integrates with the system with regular update checks and notifications, this
is a command line client that talks to connected devices via ADB.

For simplicity, it tries to follow the `apt-get`/`apt-cache` commands where it
makes sense such as `update`, `show`, `install` and `remove`.

### Current features

 * Single repo support
 * Update the index
 * List all apps
 * Search by keywords
 * Show details of an app

### Missing features

 * Proper index update checking via ETag
 * Multi-repo support
 * Interaction with a device via ADB:
   - Probably relying on the installed command `adb`
   - Fetch device details (Android version, architecture, ...)
   - Fetch installed applications
   - Should use a `sync` command if always fetching the data above is slow
   - Install, update and remove applications
 * Apk caching

### Advantages over the Android client

 * Faster to use command line interface
 * Ability to batch install/update/remove without root nor system privileges
 * Interact between multiple Android devices:
   - Transfer apps and their data from one device to another

### Android client features this will never have

 * "Update available" notifications
 * Swap apps over WiFi or Bluetooth and local repos
