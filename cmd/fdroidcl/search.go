/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"strings"

	"github.com/mvdan/fdroidcl"
)

var cmdSearch = &Command{
	Name:  "search",
	Short: "Search available apps",
}

func init() {
	cmdSearch.Run = runSearch
}

func runSearch(args []string) {
	index := mustLoadIndex()
	apps := filterAppsSearch(index.Apps, args)
	printApps(apps)
}

func filterAppsSearch(apps []fdroidcl.App, terms []string) []fdroidcl.App {
	for _, term := range terms {
		term = strings.ToLower(term)
	}
	var result []fdroidcl.App
	for _, app := range apps {
		fields := []string{
			strings.ToLower(app.ID),
			strings.ToLower(app.Name),
			strings.ToLower(app.Summary),
			strings.ToLower(app.Desc),
		}
		if !appMatches(fields, terms) {
			continue
		}
		result = append(result, app)
	}
	return result
}

func appMatches(fields []string, terms []string) bool {
fieldLoop:
	for _, field := range fields {
		for _, term := range terms {
			if !strings.Contains(field, term) {
				continue fieldLoop
			}
		}
		return true
	}
	return false
}
