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

var cmdShow = &Command{
	Name:  "show",
	Short: "Show detailed info of an app",
}

func init() {
	cmdShow.Run = runShow
}

func runShow(args []string) {
	index := mustLoadIndex()
	found := make(map[string]*fdroidcl.App, len(args))
	for _, appID := range args {
		found[appID] = nil
	}
	for i := range index.Apps {
		app := &index.Apps[i]
		_, e := found[app.ID]
		if !e {
			continue
		}
		found[app.ID] = app
	}
	for i, appID := range args {
		app, _ := found[appID]
		if app == nil {
			log.Fatalf("Could not find app with ID '%s'", appID)
		}
		if i > 0 {
			fmt.Printf("\n--\n\n")
		}
		printAppDetailed(*app)
	}
}

func printAppDetailed(app fdroidcl.App) {
	p := func(title string, format string, args ...interface{}) {
		if format == "" {
			fmt.Println(title)
		} else {
			fmt.Printf("%s %s\n", title, fmt.Sprintf(format, args...))
		}
	}
	p("Package          :", "%s", app.ID)
	p("Name             :", "%s", app.Name)
	p("Summary          :", "%s", app.Summary)
	p("Current Version  :", "%s (%d)", app.CurApk.VName, app.CurApk.VCode)
	p("Upstream Version :", "%s (%d)", app.CVName, app.CVCode)
	p("License          :", "%s", app.License)
	if app.Categs != nil {
		p("Categories       :", "%s", strings.Join(app.Categs, ", "))
	}
	if app.Website != "" {
		p("Website          :", "%s", app.Website)
	}
	if app.Source != "" {
		p("Source           :", "%s", app.Source)
	}
	if app.Tracker != "" {
		p("Tracker          :", "%s", app.Tracker)
	}
	if app.Changelog != "" {
		p("Changelog        :", "%s", app.Changelog)
	}
	if app.Donate != "" {
		p("Donate           :", "%s", app.Donate)
	}
	if app.Bitcoin != "" {
		p("Bitcoin          :", "bitcoin:%s", app.Bitcoin)
	}
	if app.Litecoin != "" {
		p("Litecoin         :", "litecoin:%s", app.Litecoin)
	}
	if app.Dogecoin != "" {
		p("Dogecoin         :", "dogecoin:%s", app.Dogecoin)
	}
	if app.FlattrID != "" {
		p("Flattr           :", "https://flattr.com/thing/%s", app.FlattrID)
	}
	fmt.Println()
	p("Description :", "")
	fmt.Println()
	app.TextDesc(os.Stdout)
	fmt.Println()
	p("Available Versions :", "")
	for _, apk := range app.Apks {
		fmt.Println()
		p("    Name   :", "%s (%d)", apk.VName, apk.VCode)
		p("    Size   :", "%d", apk.Size)
		p("    MinSdk :", "%d", apk.MinSdk)
		if apk.MaxSdk > 0 {
			p("    MaxSdk :", "%d", apk.MaxSdk)
		}
		if apk.ABIs != nil {
			p("    ABIs   :", "%s", strings.Join(apk.ABIs, ", "))
		}
	}
}
