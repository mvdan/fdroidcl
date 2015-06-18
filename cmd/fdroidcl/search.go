/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/mvdan/fdroidcl"
	"github.com/mvdan/fdroidcl/adb"
)

var cmdSearch = &Command{
	UsageLine: "search <regexp...>",
	Short:     "Search available apps",
}

var (
	quiet     = cmdSearch.Flag.Bool("q", false, "Print package names only")
	installed = cmdSearch.Flag.Bool("i", false, "Filter installed packages")
)

func init() {
	cmdSearch.Run = runSearch
}

func runSearch(args []string) {
	index := mustLoadIndex()
	apps := filterAppsSearch(index.Apps, args)
	if *installed {
		device := oneDevice()
		installed := mustInstalled(device)
		apps = filterAppsInstalled(apps, installed)
	}
	if *quiet {
		for _, app := range apps {
			fmt.Println(app.ID)
		}
	} else {
		printApps(apps)
	}
}

func filterAppsSearch(apps []fdroidcl.App, terms []string) []fdroidcl.App {
	regexes := make([]*regexp.Regexp, len(terms))
	for i, term := range terms {
		regexes[i] = regexp.MustCompile(term)
	}
	var result []fdroidcl.App
	for _, app := range apps {
		fields := []string{
			strings.ToLower(app.ID),
			strings.ToLower(app.Name),
			strings.ToLower(app.Summary),
			strings.ToLower(app.Desc),
		}
		if !appMatches(fields, regexes) {
			continue
		}
		result = append(result, app)
	}
	return result
}

func appMatches(fields []string, regexes []*regexp.Regexp) bool {
fieldLoop:
	for _, field := range fields {
		for _, regex := range regexes {
			if !regex.MatchString(field) {
				continue fieldLoop
			}
		}
		return true
	}
	return false
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

func mustInstalled(device adb.Device) []adb.Package {
	installed, err := device.Installed()
	if err != nil {
		log.Fatalf("Could not get installed packages: %v", err)
	}
	return installed
}

func filterAppsInstalled(apps []fdroidcl.App, installed []adb.Package) []fdroidcl.App {
	instMap := make(map[string]struct{}, len(installed))
	for _, p := range installed {
		instMap[p.ID] = struct{}{}
	}
	var result []fdroidcl.App
	for _, app := range apps {
		if _, e := instMap[app.ID]; !e {
			continue
		}
		result = append(result, app)
	}
	return result
}
