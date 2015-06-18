/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mvdan/fdroidcl"
)

var cmdSearch = &Command{
	UsageLine: "search <regexp...>",
	Short:     "Search available apps",
}

var (
	quiet = cmdSearch.Flag.Bool("q", false, "Show package name only")
)

func init() {
	cmdSearch.Run = runSearch
}

func runSearch(args []string) {
	index := mustLoadIndex()
	apps := filterAppsSearch(index.Apps, args)
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
