# fdroidcl

F-Droid desktop client.

Not yet a replacement for the Android client, since many features are still
missing.

For simplicity, it tries to follow the `apt-get`/`apt-cache` commands where it
makes sense such as `update`, `show`, `install` and `remove`.

### Current features

 * Single repo support
 * Update the index
 * List all apps
 * Show details of an app

### Missing features

 * Searching
 * Multi-repo support
 * Interaction with a device
   - Probably via the command `adb`
   - Fetch device details (Android version, architecture, ...)
   - Fetch installed applications
   - Should use a `sync` command if always fetching the data above is slow
   - Install, update and remove applications

