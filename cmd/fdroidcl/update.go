/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"fmt"
	"log"
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
