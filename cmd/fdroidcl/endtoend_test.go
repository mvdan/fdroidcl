// Copyright (c) 2018, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

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
}

func mustRun(t *testing.T, success bool, wantRe, name string, args ...string) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if success && err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	} else if !success && err == nil {
		t.Fatalf("expected error, got none\n%s", out)
	}
	if !regexp.MustCompile(wantRe).Match(out) {
		t.Fatalf("output does not match %#q:\n%s", wantRe, out)
	}
}
