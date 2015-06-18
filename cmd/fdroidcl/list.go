/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mvdan/fdroidcl"
)

var cmdList = &Command{
	UsageLine: "list",
	Short:     "List all available apps",
}

func init() {
	cmdList.Run = runList
}

func runList(args []string) {
	index := mustLoadIndex()
	printApps(index.Apps)
}

func mustLoadIndex() *fdroidcl.Index {
	p := indexPath(repoName)
	f, err := os.Open(p)
	if err != nil {
		log.Fatalf("Could not open index file: %v", err)
	}
	stat, err := f.Stat()
	if err != nil {
		log.Fatalf("Could not stat index file: %v", err)
	}
	//pubkey, err := hex.DecodeString(repoPubkey)
	//if err != nil {
	//	log.Fatalf("Could not decode public key: %v", err)
	//}
	index, err := fdroidcl.LoadIndexJar(f, stat.Size(), nil)
	if err != nil {
		log.Fatalf("Could not load index: %v", err)
	}
	return index
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

func printApp(app fdroidcl.App, IDLen int) {
	fmt.Printf("%s%s %s %s\n", app.ID, strings.Repeat(" ", IDLen-len(app.ID)),
		app.Name, app.CurApk.VName)
	fmt.Printf("    %s\n", app.Summary)
}
