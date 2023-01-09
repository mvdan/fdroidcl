// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var cmdClean = &Command{
	UsageLine: "clean",
	Short:     "Clean index and/or cache",
	Long: `
Clean index and/or cache.
The index basically contains information about available APKs to download.
The cache is where all downloaded APKs are cached.
Usage:

	$ fdroidcl clean
	$ fdroidcl clean index
	$ fdroidcl clean cache
`[1:],
}

func init() {
	cmdClean.Run = runClean
}

func runClean(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("wrong amount of arguments")
	}
	if len(args) == 1 && (args[0] != "index" && args[0] != "cache") {
		return fmt.Errorf("pass either index or cache as parameter, or no parameter at all")
	}
	if len(args) == 0 || args[0] == "index" {
		err := cleanIndex()
		if err != nil {
			return err
		}
	}
	if len(args) == 0 || args[0] == "cache" {
		err := cleanCache()
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanIndex() error {
	cachePath := filepath.Join(mustCache(), "cache-gob")
	err := removeFile(cachePath)
	if err != nil {
		return err
	}
	err = removeGlob(mustData() + "/*.jar")
	if err != nil {
		return err
	}
	err = removeGlob(mustData() + "/*.jar-etag")
	if err != nil {
		return err
	}
	return nil
}

func cleanCache() error {
	apksDir := subdir(mustCache(), "apks")
	err := os.RemoveAll(apksDir)
	if err != nil {
		return err
	}
	return nil
}

func removeFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		e, ok := err.(*os.PathError)
		if ok && e.Err == syscall.ENOENT {
			// The file didn't exist, ignore
		} else {
			return err
		}
	}
	return nil
}

func removeGlob(glob string) error {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	for _, value := range matches {
		err := removeFile(value)
		if err != nil {
			return err
		}
	}
	return nil
}
