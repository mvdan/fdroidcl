// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package basedir

import (
	"os"
	"os/user"
	"path/filepath"
)

// TODO: replace with https://github.com/golang/go/issues/29960 if accepted.

// Data returns the base data directory.
func Data() string {
	return dataDir
}

func firstGetenv(def string, evs ...string) string {
	for _, ev := range evs {
		if v := os.Getenv(ev); v != "" {
			return v
		}
	}
	// TODO: replace with os.UserHomeDir once we require Go 1.12 or later.
	home := os.Getenv("HOME")
	if home == "" {
		curUser, err := user.Current()
		if err != nil {
			return ""
		}
		home = curUser.HomeDir
	}
	return filepath.Join(home, def)
}
