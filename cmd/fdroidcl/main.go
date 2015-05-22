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

func filterAppsInstalled(apps *map[string]fdroidcl.App, installed []string) {
	instMap := make(map[string]struct{}, len(installed))
	for _, id := range installed {
		instMap[id] = struct{}{}
	}
	for appID := range *apps {
		if _, e := instMap[appID]; !e {
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

func printApp(app fdroidcl.App, IDLen int) {
	fmt.Printf("%s%s %s %s\n", app.ID, strings.Repeat(" ", IDLen-len(app.ID)),
		app.Name, app.CurApk.VName)
	fmt.Printf("    %s\n", app.Summary)
}

func printApps(apps map[string]fdroidcl.App) {
	maxIDLen := 0
	for appID := range apps {
		if len(appID) > maxIDLen {
			maxIDLen = len(appID)
		}
	}
	for _, app := range sortedApps(apps) {
		printApp(app, maxIDLen)
	}
}

var repoURL = flag.String("r", "https://f-droid.org/repo", "repository address")

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: fdroidcl [-h] [-r <repo address>] <command> [<args>]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Available commands:")
		fmt.Fprintln(os.Stderr, "   update             Update the index")
		fmt.Fprintln(os.Stderr, "   list               List all available apps")
		fmt.Fprintln(os.Stderr, "   search <term...>   Search available apps")
		fmt.Fprintln(os.Stderr, "   show <appid...>    Show detailed info of an app")
		fmt.Fprintln(os.Stderr, "   devices            List connected devices")
	}
}

func mustLoadApps(repoName string) map[string]fdroidcl.App {
	apps, err := fdroidcl.LoadApps(repoName)
	if err != nil {
		log.Fatalf("Could not load apps: %v", err)
	}
	return apps
}

func mustInstalled(device adb.Device) []string {
	installed, err := device.Installed()
	if err != nil {
		log.Fatalf("Could not get installed packages: %v", err)
	}
	return installed
}

func oneDevice() adb.Device {
	devices, err := adb.Devices()
	if err != nil {
		log.Fatalf("Could not get devices: %v", err)
	}
	if len(devices) == 0 {
		log.Fatalf("No devices found")
	}
	if len(devices) > 1 {
		log.Fatalf("Too many devices found")
	}
	return devices[0]
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
		printApps(apps)
	case "search":
		apps := mustLoadApps(repoName)
		filterAppsSearch(&apps, args)
		printApps(apps)
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
			log.Fatalf("Could not get devices: %v", err)
		}
		for _, device := range devices {
			fmt.Printf("%s - %s (%s)\n", device.Id, device.Model, device.Product)
		}
	case "installed":
		apps := mustLoadApps(repoName)
		device := oneDevice()
		installed := mustInstalled(device)
		filterAppsInstalled(&apps, installed)
		printApps(apps)
	default:
		log.Printf("Unrecognised command '%s'\n\n", cmd)
		flag.Usage()
		os.Exit(2)
	}
}
