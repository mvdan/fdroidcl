/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mvdan/appdir"

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

func filterAppsInstalled(apps []fdroidcl.App, installed []string) []fdroidcl.App {
	instMap := make(map[string]struct{}, len(installed))
	for _, id := range installed {
		instMap[id] = struct{}{}
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

func printApp(app fdroidcl.App, IDLen int) {
	fmt.Printf("%s%s %s %s\n", app.ID, strings.Repeat(" ", IDLen-len(app.ID)),
		app.Name, app.CurApk.VName)
	fmt.Printf("    %s\n", app.Summary)
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

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: fdroidcl [-h] <command> [<args>]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Available commands:")
		fmt.Fprintln(os.Stderr, "   update             Update the index")
		fmt.Fprintln(os.Stderr, "   list               List all available apps")
		fmt.Fprintln(os.Stderr, "   search <term...>   Search available apps")
		fmt.Fprintln(os.Stderr, "   show <appid...>    Show detailed info of an app")
		fmt.Fprintln(os.Stderr, "   devices            List connected devices")
		fmt.Fprintln(os.Stderr, "   installed          List installed apps")
	}
}

func appSubdir(appdir string) string {
	p := filepath.Join(appdir, "fdroidcl")
	if err := os.MkdirAll(p, 0755); err != nil {
		log.Fatalf("Could not create app dir: %v", err)
	}
	return p
}

func indexPath(name string) string {
	cache, err := appdir.Cache()
	if err != nil {
		log.Fatalf("Could not determine cache dir: %v", err)
	}
	return filepath.Join(appSubdir(cache), repoName+".jar")
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

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(args []string)

	// Name of the command.
	Name string

	// Short is the short description.
	Short string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

// Commands lists the available commands.
var commands = []*Command{
	cmdUpdate,
	cmdList,
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	for _, cmd := range commands {
		if cmd.Name != args[0] {
			continue
		}
		cmd.Flag.Parse(args[1:])
		args = cmd.Flag.Args()
		cmd.Run(args)
		os.Exit(0)
	}

	switch args[0] {
	case "search":
		args = args[1:]
		index := mustLoadIndex()
		apps := filterAppsSearch(index.Apps, args)
		printApps(apps)
	case "show":
		args = args[1:]
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
	case "installed":
		index := mustLoadIndex()
		startAdbIfNeeded()
		device := oneDevice()
		installed := mustInstalled(device)
		apps := filterAppsInstalled(index.Apps, installed)
		printApps(apps)
	default:
		log.Printf("Unrecognised command '%s'\n\n", args[0])
		flag.Usage()
		os.Exit(2)
	}
}
