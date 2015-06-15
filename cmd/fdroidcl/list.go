/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"github.com/mvdan/fdroidcl"
)

var cmdList = &Command{
	Name:  "list",
	Short: "List all available apps",
}

func init() {
	cmdList.Run = runList
}

func runList(args []string) {
	index := mustLoadIndex()
	printApps(index.Apps)
}

func printApps(apps []fdroidcl.App) {
	maxIDLen := 0
	for _, app := range apps {
		if len(app.ID) > maxIDLen {
			maxIDLen = len(app.ID)
		}
	}
	for _, app := range apps {
		printApp(app, maxIDLen)
	}
}
