/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

// Repo is an F-Droid repository holding apps and apks
type Repo struct {
	Apps []App `xml:"application"`
}

type CommaList []string

func (cl *CommaList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}
	*cl = strings.Split(content, ",")
	return nil
}

// App is an Android application
type App struct {
	ID      string    `xml:"id"`
	Name    string    `xml:"name"`
	Summary string    `xml:"summary"`
	Desc    string    `xml:"desc"`
	License string    `xml:"license"`
	Categs  CommaList `xml:"categories"`
	CVName  string    `xml:"marketversion"`
	CVCode  uint      `xml:"marketvercode"`
	Website string    `xml:"web"`
	Source  string    `xml:"source"`
	Tracker string    `xml:"tracker"`
	Apks    []Apk     `xml:"package"`
	CurApk  *Apk
}

// Apk is an Android package
type Apk struct {
	VName  string    `xml:"version"`
	VCode  uint      `xml:"versioncode"`
	Size   int       `xml:"size"`
	MinSdk int       `xml:"sdkver"`
	ABIs   CommaList `xml:"nativecode"`
}

func (app *App) prepareData() {
	for _, apk := range app.Apks {
		app.CurApk = &apk
		if app.CVCode >= apk.VCode {
			break
		}
	}
}

func (app *App) writeShort(w io.Writer) {
	fmt.Fprintf(w, "%s | %s %s\n", app.ID, app.Name, app.CurApk.VName)
	fmt.Fprintf(w, "    %s\n", app.Summary)
}

func (app *App) writeDetailed(w io.Writer) {
	p := func(title string, format string, args ...interface{}) {
		if format == "" {
			fmt.Fprintln(w, title)
		} else {
			fmt.Fprintf(w, "%s %s\n", title, fmt.Sprintf(format, args...))
		}
	}
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
	// p("Description     :", "%s", app.Desc) // TODO: parse html, 80 column wrapping
	fmt.Println()
	p("Available Versions :", "")
	for _, apk := range app.Apks {
		fmt.Println()
		p("    Name     :", "%s (%d)", apk.VName, apk.VCode)
		p("    Size     :", "%d", apk.Size)
		p("    MinSdk   :", "%d", apk.MinSdk)
		if apk.ABIs != nil {
			p("    ABIs     :", "%s", strings.Join(apk.ABIs, ", "))
		}
	}
}

const indexName = "index.jar"

func updateIndex() {
	url := fmt.Sprintf("%s/%s", *repoURL, indexName)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch '%s': %s", url, err)
	}
	defer resp.Body.Close()
	out, err := os.Create(indexName)
	if err != nil {
		log.Fatalf("Failed to create file '%s': %s", indexName, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Fatal(err)
	}
}

func loadApps() map[string]App {
	r, err := zip.OpenReader(indexName)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	buf := new(bytes.Buffer)

	for _, f := range r.File {
		if f.Name != "index.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		if _, err = io.Copy(buf, rc); err != nil {
			log.Fatal(err)
		}
		rc.Close()
		break
	}

	var repo Repo
	if err := xml.Unmarshal(buf.Bytes(), &repo); err != nil {
		log.Fatalf("Could not read xml: %s", err)
	}
	apps := make(map[string]App)

	for i := range repo.Apps {
		app := repo.Apps[i]
		app.prepareData()
		apps[app.ID] = app
	}
	return apps
}

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

func filterAppsSearch(apps *map[string]App, terms []string) {
	for _, term := range terms {
		term = strings.ToLower(term)
	}
	for appID, app := range *apps {
		fields := []string{
			strings.ToLower(app.ID),
			strings.ToLower(app.Name),
			strings.ToLower(app.Summary),
			// strings.ToLower(app.Desc), // TODO remove html
		}
		if !appMatches(fields, terms) {
			delete(*apps, appID)
		}
	}
}

type appList []App

func (al appList) Len() int           { return len(al) }
func (al appList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al appList) Less(i, j int) bool { return al[i].ID < al[j].ID }

func sortedApps(apps map[string]App) []App {
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
		p("   update           Update the index")
		p("   list             List all available apps")
		p("   search <term...> Search available apps")
		p("   show <appid...>   Show detailed info of an app")
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	cmd := flag.Args()[0]
	args := flag.Args()[1:]

	switch cmd {
	case "update":
		updateIndex()
	case "list":
		apps := loadApps()
		for _, app := range sortedApps(apps) {
			app.writeShort(os.Stdout)
		}
	case "search":
		apps := loadApps()
		filterAppsSearch(&apps, args)
		for _, app := range sortedApps(apps) {
			app.writeShort(os.Stdout)
		}
	case "show":
		apps := loadApps()
		for _, appID := range args {
			app, e := apps[appID]
			if !e {
				log.Fatalf("Could not find app with ID '%s'", appID)
			}
			app.writeDetailed(os.Stdout)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unrecognised command '%s'\n\n", cmd)
		flag.Usage()
		os.Exit(2)
	}
}
