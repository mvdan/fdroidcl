// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package basedir

import (
	"os"
	"os/user"
	"path/filepath"
)

// TODO: replace with os.UserCacheDir once we require Go 1.11 or later.

// Cache returns the base cache directory.
func Cache() string {
	return cache()
}

// Data returns the base data directory.
func Data() string {
	return data()
}

func firstGetenv(def string, evs ...string) string {
	for _, ev := range evs {
		if v := os.Getenv(ev); v != "" {
			return v
		}
	}
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
