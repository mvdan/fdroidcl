// Copyright (c) 2018, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// chosenApp is the app that will be installed and uninstalled on a connected
// device. This one was chosen because it's tiny, requires no permissions, and
// should be compatible with every device.
//
// It also stores no data, so it is fine to uninstall it and the user won't lose
// any data.
const chosenApp = "org.vi_server.red_screen"

func TestEndToEnd(t *testing.T) {
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

	// Build fdroidcl in the temporary directory.
	fdroidcl := filepath.Join(dir, "fdroidcl")
	if out, err := exec.Command("go", "build",
		"-ldflags=-X main.testBasedir="+dir,
		"-o", fdroidcl).CombinedOutput(); err != nil {
		t.Fatalf("%s", out)
	}

	mustSucceed := func(t *testing.T, want string, args ...string) {
		mustRun(t, true, want, fdroidcl, args...)
	}
	mustFail := func(t *testing.T, want string, args ...string) {
		mustRun(t, false, want, fdroidcl, args...)
	}

	t.Run("Help", func(t *testing.T) {
		mustFail(t, `Usage: fdroidcl`, "-h")
	})
	t.Run("UnknownCommand", func(t *testing.T) {
		mustFail(t, `Unrecognised command`, "unknown")
	})
	t.Run("Version", func(t *testing.T) {
		mustSucceed(t, `^v`, "version")
	})

	t.Run("SearchBeforeUpdate", func(t *testing.T) {
		mustFail(t, `could not open index`, "search")
	})
	t.Run("UpdateFirst", func(t *testing.T) {
		mustSucceed(t, `done`, "update")
	})
	t.Run("UpdateCached", func(t *testing.T) {
		mustSucceed(t, `not modified`, "update")
	})

	t.Run("SearchNoArgs", func(t *testing.T) {
		mustSucceed(t, `F-Droid`, "search")
	})
	t.Run("SearchWithArgs", func(t *testing.T) {
		mustSucceed(t, `F-Droid`, "search", "fdroid.fdroid")
	})
	t.Run("SearchWithArgsNone", func(t *testing.T) {
		mustSucceed(t, `^$`, "search", "nomatches")
	})
	t.Run("SearchOnlyPackageNames", func(t *testing.T) {
		mustSucceed(t, `^[^ ]*$`, "search", "-q", "fdroid.fdroid")
	})

	t.Run("ShowOne", func(t *testing.T) {
		mustSucceed(t, `fdroid/fdroidclient`, "show", "org.fdroid.fdroid")
	})
	t.Run("ShowMany", func(t *testing.T) {
		mustSucceed(t, `fdroid/fdroidclient.*fdroid/privileged-extension`,
			"show", "org.fdroid.fdroid", "org.fdroid.fdroid.privileged")
	})

	t.Run("ListCategories", func(t *testing.T) {
		mustSucceed(t, `Development`, "list", "categories")
	})

	out, err := exec.Command(fdroidcl, "devices").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	switch bytes.Count(out, []byte("\n")) {
	case 0:
		t.Log("skipping the device tests as none was found via ADB")
	case 1:
		// continue below
	default:
		t.Log("skipping the device tests as too many were found via ADB")
	}

	// try to uninstall the app first
	exec.Command(fdroidcl, "uninstall", chosenApp).Run()
	t.Run("UninstallMissing", func(t *testing.T) {
		mustFail(t, `not installed$`, "uninstall", chosenApp)
	})
	t.Run("InstallVersioned", func(t *testing.T) {
		mustSucceed(t, `Installing `+regexp.QuoteMeta(chosenApp),
			"install", chosenApp+":1")
	})
	t.Run("Upgrade", func(t *testing.T) {
		mustSucceed(t, `Upgrading `+regexp.QuoteMeta(chosenApp),
			"upgrade", chosenApp)
	})
	t.Run("UpgradeAlreadyInstalled", func(t *testing.T) {
		mustFail(t, `is up to date$`, "upgrade", chosenApp)
	})
	t.Run("UninstallExisting", func(t *testing.T) {
		mustSucceed(t, `Uninstalling `+regexp.QuoteMeta(chosenApp),
			"uninstall", chosenApp)
	})
}

func mustRun(t *testing.T, success bool, wantRe, name string, args ...string) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if success && err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	} else if !success && err == nil {
		t.Fatalf("expected error, got none\n%s", out)
	}
	// Let '.' match newlines, and treat the output as a single line.
	wantRe = "(?sm)" + wantRe
	if !regexp.MustCompile(wantRe).Match(out) {
		t.Fatalf("output does not match %#q:\n%s", wantRe, out)
	}
}
