// Copyright (c) 2018, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"mvdan.cc/fdroidcl/adb"
)

// chosenApp is the app that will be installed and uninstalled on a connected
// device. This one was chosen because it's tiny, requires no permissions, and
// should be compatible with every device.
//
// It also stores no data, so it is fine to uninstall it and the user won't lose
// any data.
const chosenApp = "org.vi_server.red_screen"

func TestCommands(t *testing.T) {
	url := config.Repos[0].URL
	client := http.Client{Timeout: 2 * time.Second}
	if _, err := client.Get(url); err != nil {
		t.Skipf("skipping since %s is unreachable: %v", url, err)
	}

	dir, err := ioutil.TempDir("", "fdroidcl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	testBasedir = dir

	mustSucceed := func(t *testing.T, want string, cmd *Command, args ...string) {
		mustRun(t, true, want, cmd, args...)
	}
	mustFail := func(t *testing.T, want string, cmd *Command, args ...string) {
		mustRun(t, false, want, cmd, args...)
	}

	t.Run("Version", func(t *testing.T) {
		mustSucceed(t, `^v`, cmdVersion)
	})

	t.Run("SearchBeforeUpdate", func(t *testing.T) {
		mustFail(t, `could not open index`, cmdSearch)
	})
	t.Run("UpdateFirst", func(t *testing.T) {
		mustSucceed(t, `done`, cmdUpdate)
	})
	t.Run("UpdateCached", func(t *testing.T) {
		mustSucceed(t, `not modified`, cmdUpdate)
	})

	t.Run("SearchNoArgs", func(t *testing.T) {
		mustSucceed(t, `F-Droid`, cmdSearch)
	})
	t.Run("SearchWithArgs", func(t *testing.T) {
		mustSucceed(t, `F-Droid`, cmdSearch, "fdroid.fdroid")
	})
	t.Run("SearchWithArgsNone", func(t *testing.T) {
		mustSucceed(t, `^$`, cmdSearch, "nomatches")
	})
	t.Run("SearchOnlyPackageNames", func(t *testing.T) {
		mustSucceed(t, `^[^ ]*$`, cmdSearch, "-q", "fdroid.fdroid")
	})

	t.Run("ShowOne", func(t *testing.T) {
		mustSucceed(t, `fdroid/fdroidclient`, cmdShow, "org.fdroid.fdroid")
	})
	t.Run("ShowMany", func(t *testing.T) {
		mustSucceed(t, `fdroid/fdroidclient.*fdroid/privileged-extension`,
			cmdShow, "org.fdroid.fdroid", "org.fdroid.fdroid.privileged")
	})

	t.Run("ListCategories", func(t *testing.T) {
		mustSucceed(t, `Development`, cmdList, "categories")
	})

	if err := startAdbIfNeeded(); err != nil {
		t.Fatal(err)
	}
	devices, err := adb.Devices()
	if err != nil {
		t.Fatal(err)
	}
	switch len(devices) {
	case 0:
		t.Log("skipping the device tests as none was found via ADB")
	case 1:
		// continue below
	default:
		t.Log("skipping the device tests as too many were found via ADB")
	}

	// try to uninstall the app first
	devices[0].Uninstall(chosenApp)
	t.Run("UninstallMissing", func(t *testing.T) {
		mustFail(t, `not installed$`, cmdUninstall, chosenApp)
	})
	t.Run("InstallVersioned", func(t *testing.T) {
		mustSucceed(t, `Installing `+regexp.QuoteMeta(chosenApp),
			cmdInstall, chosenApp+":1")
	})
	t.Run("Upgrade", func(t *testing.T) {
		mustSucceed(t, `Upgrading `+regexp.QuoteMeta(chosenApp),
			cmdUpgrade, chosenApp)
	})
	t.Run("UpgradeAlreadyInstalled", func(t *testing.T) {
		mustFail(t, `is up to date$`, cmdUpgrade, chosenApp)
	})
	t.Run("UninstallExisting", func(t *testing.T) {
		mustSucceed(t, `Uninstalling `+regexp.QuoteMeta(chosenApp),
			cmdUninstall, chosenApp)
	})
}

func mustRun(t *testing.T, success bool, wantRe string, cmd *Command, args ...string) {
	var buf bytes.Buffer
	stdout, stderr = &buf, &buf
	err := cmd.Run(args)
	out := buf.String()
	if success && err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	} else if !success && err == nil {
		t.Fatalf("expected error, got none\n%s", out)
	}
	if err != nil {
		out += err.Error()
	}
	// Let '.' match newlines, and treat the output as a single line.
	wantRe = "(?sm)" + wantRe
	if !regexp.MustCompile(wantRe).MatchString(out) {
		t.Fatalf("output does not match %#q:\n%s", wantRe, out)
	}
}
