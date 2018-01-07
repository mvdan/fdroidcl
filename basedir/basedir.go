// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package basedir

import (
	"os"
	"os/user"
	"path/filepath"
)

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
	home, err := homeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, def)
}

func homeDir() (string, error) {
	curUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return curUser.HomeDir, nil
}
