/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/mvdan/fdroidcl"
	"github.com/mvdan/fdroidcl/adb"
)

func appMatches(fields []string, terms []string) bool {
	for _, field := range fields {
		for _, term := range terms {
			if !strings.Contains(field, term) {
				goto next
			}
		}
		return true
	next:
	}
	return false
}

func filterAppsSearch(apps *map[string]fdroidcl.App, terms []string) {
	for _, term := range terms {
		term = strings.ToLower(term)
	}
	for appID, app := range *apps {
		fields := []string{
			strings.ToLower(app.ID),
			strings.ToLower(app.Name),
			strings.ToLower(app.Summary),
			strings.ToLower(app.Desc),
		}
		if !appMatches(fields, terms) {
			delete(*apps, appID)
		}
	}
}

type appList []fdroidcl.App

func (al appList) Len() int           { return len(al) }
func (al appList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al appList) Less(i, j int) bool { return al[i].ID < al[j].ID }

func sortedApps(apps map[string]fdroidcl.App) []fdroidcl.App {
	list := make(appList, 0, len(apps))
	for appID := range apps {
		list = append(list, apps[appID])
	}
	sort.Sort(list)
	return list
}

var repoURL = flag.String("r", "https://f-droid.org/repo", "repository address")

func init() {
	flag.Usage = func() {
		p := func(args ...interface{}) {
			fmt.Fprintln(os.Stderr, args...)
		}
		p("Usage: fdroidcl [-h] [-r <repo address>] <command> [<args>]")
		p()
		p("Available commands:")
		p("   update             Update the index")
		p("   list               List all available apps")
		p("   search <term...>   Search available apps")
		p("   show <appid...>    Show detailed info of an app")
		p("   devices            List connected devices")
	}
}

func mustLoadApps(repoName string) map[string]fdroidcl.App {
	apps, err := fdroidcl.LoadApps(repoName)
	if err != nil {
		log.Fatalf("Could not load apps: %v", err)
	}
	return apps
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	cmd := flag.Args()[0]
	args := flag.Args()[1:]

	repoName := "index"

	switch cmd {
	case "update":
		err := fdroidcl.UpdateIndex(repoName, *repoURL)
		if err == fdroidcl.ErrNotModified {
			log.Print("Index up to date")
		} else if err != nil {
			log.Fatalf("Could not update index: %v", err)
		}
	case "list":
		apps := mustLoadApps(repoName)
		for _, app := range sortedApps(apps) {
			app.WriteShort(os.Stdout)
		}
	case "search":
		apps := mustLoadApps(repoName)
		filterAppsSearch(&apps, args)
		for _, app := range sortedApps(apps) {
			app.WriteShort(os.Stdout)
		}
	case "show":
		apps := mustLoadApps(repoName)
		for _, appID := range args {
			app, e := apps[appID]
			if !e {
				log.Fatalf("Could not find app with ID '%s'", appID)
			}
			app.WriteDetailed(os.Stdout)
		}
	case "devices":
		devices, err := adb.Devices()
		if err != nil {
			log.Fatalf("Could not list devices: %v", err)
		}
		for _, device := range devices {
			fmt.Println(device)
		}
	default:
		log.Printf("Unrecognised command '%s'\n\n", cmd)
		flag.Usage()
		os.Exit(2)
	}
}
