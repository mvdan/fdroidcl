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

func indexPath(repoName string) string {
	cache, err := appdir.Cache()
	if err != nil {
		log.Fatalf("Could not determine cache dir: %v", err)
	}
	return filepath.Join(appSubdir(cache), repoName + ".jar")
}

func updateIndex(repoName, repoURL string) error {
	p := indexPath(repoName)
	url := fmt.Sprintf("%s/%s", repoURL, "index.jar")
	if err := downloadEtag(url, p); err != nil {
		return err
	}
	return nil
}

func mustLoadIndex(repoName string) *fdroidcl.Index {
	p := indexPath(repoName)
	f, err := os.Open(p)
	if err != nil {
		log.Fatalf("Could not open index file: %v", err)
	}
	stat, err := f.Stat()
	if err != nil {
		log.Fatalf("Could not stat index file: %v", err)
	}
	index, err := fdroidcl.LoadIndexJar(f, stat.Size())
	if err != nil {
		log.Fatalf("Could not load index: %v", err)
	}
	return index
}

func startAdbIfNeeded() {
	if adb.IsServerRunning() {
		return
	}
	log.Printf("Starting ADB server...")
	if err := adb.StartServer(); err != nil {
		log.Fatalf("Could not start ADB server: %v", err)
	}
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
		err := updateIndex(repoName, *repoURL)
		if err == errNotModified {
			log.Print("Index up to date")
		} else if err != nil {
			log.Fatalf("Could not update index: %v", err)
		}
	case "list":
		index := mustLoadIndex(repoName)
		printApps(index.Apps)
	case "search":
		index := mustLoadIndex(repoName)
		apps := filterAppsSearch(index.Apps, args)
		printApps(apps)
	case "show":
		index := mustLoadIndex(repoName)
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
	case "devices":
		startAdbIfNeeded()
		devices, err := adb.Devices()
		if err != nil {
			log.Fatalf("Could not get devices: %v", err)
		}
		for _, device := range devices {
			fmt.Printf("%s - %s (%s)\n", device.Id, device.Model, device.Product)
		}
	case "installed":
		index := mustLoadIndex(repoName)
		startAdbIfNeeded()
		device := oneDevice()
		installed := mustInstalled(device)
		apps := filterAppsInstalled(index.Apps, installed)
		printApps(apps)
	default:
		log.Printf("Unrecognised command '%s'\n\n", cmd)
		flag.Usage()
		os.Exit(2)
	}
}
