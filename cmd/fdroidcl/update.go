/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mvdan/appdir"
)

var cmdUpdate = &Command{
	Name:  "update",
	Short: "Update the index",
}

func init() {
	cmdUpdate.Run = runUpdate
}

func runUpdate(args []string) {
	if err := updateIndex(); err != nil {
		log.Fatalf("Could not update index: %v", err)
	}
}

func updateIndex() error {
	url := fmt.Sprintf("%s/%s", repoURL, "index.jar")
	if err := downloadEtag(url, indexPath(repoName)); err != nil {
		return err
	}
	return nil
}

func indexPath(name string) string {
	cache, err := appdir.Cache()
	if err != nil {
		log.Fatalf("Could not determine cache dir: %v", err)
	}
	return filepath.Join(appSubdir(cache), repoName+".jar")
}

func appSubdir(appdir string) string {
	p := filepath.Join(appdir, "fdroidcl")
	if err := os.MkdirAll(p, 0755); err != nil {
		log.Fatalf("Could not create app dir: %v", err)
	}
	return p
}
