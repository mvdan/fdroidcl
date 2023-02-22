// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"mvdan.cc/fdroidcl/adb"
	"mvdan.cc/fdroidcl/fdroid"
)

var cmdInstall = &Command{
	UsageLine: "install [<appid...>]",
	Short:     "Install or upgrade apps",
	Long: `
Install or upgrade apps. When given no arguments, it reads a comma-separated
list of apps to install from standard input, like:

	packageName,versionCode,versionName
	foo.bar,120,1.2.0
`[1:],
}

var (
	installUpdates        = cmdInstall.Fset.Bool("u", false, "Upgrade all installed apps")
	installDryRun         = cmdInstall.Fset.Bool("n", false, "Only print the operations that would be done")
	installUpdatesExclude = cmdInstall.Fset.String("e", "", "Exclude apps from upgrading (comma-separated list)")
	installSkipError      = cmdInstall.Fset.Bool("s", false, "Skip to the next application if a download or install error occurs")
	installUser           = cmdInstall.Fset.String("user", "", "Install for specified user <USER_ID|current|all> (default: \"current\" for installing, and upgrading only for users who have the app installed)")
)

func init() {
	cmdInstall.Run = runInstall
}

func runInstall(args []string) error {
	if *installUpdates && len(args) > 0 {
		return fmt.Errorf("-u can only be used without arguments")
	}
	if *installUpdatesExclude != "" && !*installUpdates {
		return fmt.Errorf("-e can only be used for upgrading (i.e. -u)")
	}
	if *installUser != "" && *installUser != "all" && *installUser != "current" {
		n, err := strconv.Atoi(*installUser)
		if err != nil {
			return fmt.Errorf("-user has to be <USER_ID|current|all>")
		}
		if n < 0 {
			return fmt.Errorf("-user cannot have a negative number as USER_ID")
		}
	}
	device, err := oneDevice()
	if err != nil {
		return err
	}
	if *installUser == "current" || (*installUser == "" && !*installUpdates) {
		uid, err := device.CurrentUserId()
		if err != nil {
			return err
		}
		*installUser = strconv.Itoa(uid)
	}
	inst, err := device.Installed()
	if err != nil {
		return err
	}

	if *installUpdates {
		apps, err := loadIndexes()
		if err != nil {
			return err
		}
		apps = filterAppsUpdates(apps, inst, device)
		if *installUpdatesExclude != "" {
			excludeApps := strings.Split(*installUpdatesExclude, ",")
			installApps := make([]fdroid.App, 0)
			for _, app := range apps {
				shouldExclude := false
				for _, exclude := range excludeApps {
					if app.PackageName == exclude {
						shouldExclude = true
						break
					}
				}
				if shouldExclude {
					continue
				}
				installApps = append(installApps, app)
			}
			apps = installApps
		}
		if len(apps) == 0 {
			fmt.Fprintln(os.Stderr, "All apps up to date.")
		}
		return downloadAndDo(apps, inst, device)
	}

	if len(args) == 0 {
		// The CSV input is as follows:
		//
		// packageName,versionCode,versionName
		// foo.bar,120,1.2.0
		// ...

		r := csv.NewReader(os.Stdin)
		r.FieldsPerRecord = 3
		r.Read()
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error parsing CSV: %v", err)
			}
			// convert "foo.bar,120" into "foo.bar:120" for findApps
			args = append(args, record[0]+":"+record[1])
		}
	}

	apps, err := findApps(args)
	if err != nil {
		return err
	}
	var toInstall []fdroid.App
	for _, app := range apps {
		p, e := inst[app.PackageName]
		if !e {
			// installing an app from scratch
			toInstall = append(toInstall, app)
			continue
		}
		suggested := app.SuggestedApk(device)
		if suggested == nil {
			return fmt.Errorf("no suitable APKs found for %s", app.PackageName)
		}
		if p.VersCode >= suggested.VersCode {
			if !(*installUser == "all" && len(p.NotInstalledForUsers) > 0) { // ensure that it can't install for other user
				okSkip := *installUser == "all"
				if !okSkip {
					n, err := strconv.Atoi(*installUser)
					if err != nil {
						return err
					}
					isInstalledForUser := false
					for _, uid := range p.InstalledForUsers {
						if uid == n {
							isInstalledForUser = true
							break
						}
					}
					if isInstalledForUser {
						okSkip = true
					}
				}
				if okSkip {
					fmt.Printf("%s is up to date\n", app.PackageName)
					// app is already up to date
					continue
				}
			}
		}
		// upgrading an existing app
		toInstall = append(toInstall, app)
	}
	return downloadAndDo(toInstall, inst, device)
}

func downloadAndDo(apps []fdroid.App, installed map[string]adb.Package, device *adb.Device) error {
	type downloaded struct {
		apk  *fdroid.Apk
		app  fdroid.App
		path string
	}
	toInstall := make([]downloaded, 0)
	for _, app := range apps {
		apk := app.SuggestedApk(device)
		if apk == nil {
			return fmt.Errorf("no suitable APKs found for %s", app.PackageName)
		}
		if *installDryRun {
			fmt.Printf("install %s:%d\n", app.PackageName, apk.VersCode)
			continue
		}
		path, err := downloadApk(apk)
		if err != nil {
			if *installSkipError {
				fmt.Printf("Downloading %s failed, skipping...\n", app.PackageName)
				continue
			}
			return err
		}
		toInstall = append(toInstall, downloaded{apk: apk, app: app, path: path})
	}
	if *installDryRun {
		return nil
	}
	for _, t := range toInstall {
		var installedPkg *adb.Package = nil
		if p, e := installed[t.app.PackageName]; e {
			installedPkg = &p
		}
		if err := installApk(device, t.apk, installedPkg, t.path); err != nil {
			if *installSkipError {
				fmt.Printf("Installing %s failed, skipping...\n", t.apk.AppID)
				continue
			}
			return err
		}
	}
	return nil
}

func installApk(device *adb.Device, apk *fdroid.Apk, devicePkg *adb.Package, path string) error {
	fmt.Printf("Installing %s\n", apk.AppID)
	userId := "all"
	if *installUser != "all" {
		if *installUpdates && *installUser == "" {
			if devicePkg == nil {
				return fmt.Errorf("failed to get device package although it should be installed (please report this error)")
			}
			if len((*devicePkg).InstalledForUsers) > 0 {
				userId = strconv.Itoa((*devicePkg).InstalledForUsers[0])
			}
		} else {
			userId = *installUser
		}
	}
	if userId == "all" {
		if err := device.Install(path); err != nil {
			return fmt.Errorf("could not install %s: %v", apk.AppID, err)
		}
	} else {
		if err := device.InstallUser(path, userId); err != nil {
			return fmt.Errorf("could not install %s: %v", apk.AppID, err)
		}
	}
	return nil
}
