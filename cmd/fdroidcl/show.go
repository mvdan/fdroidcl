// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"strconv"
	"strings"

	"mvdan.cc/fdroidcl"
)

var cmdShow = &Command{
	UsageLine: "show <appid...>",
	Short:     "Show detailed info about an app",
}

func init() {
	cmdShow.Run = runShow
}

func runShow(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no package names given")
	}
	apps, err := findApps(args)
	if err != nil {
		return err
	}
	for i, app := range apps {
		if i > 0 {
			fmt.Fprintf(stdout, "\n--\n\n")
		}
		printAppDetailed(*app)
	}
	return nil
}

func appsMap(apps []fdroidcl.App) map[string]*fdroidcl.App {
	m := make(map[string]*fdroidcl.App, len(apps))
	for i := range apps {
		app := &apps[i]
		m[app.PackageName] = app
	}
	return m
}

func findApps(ids []string) ([]*fdroidcl.App, error) {
	apps, err := loadIndexes()
	if err != nil {
		return nil, err
	}
	byId := appsMap(apps)
	result := make([]*fdroidcl.App, len(ids))
	for i, id := range ids {
		var vcode = -1
		j := strings.Index(id, ":")
		if j > -1 {
			var err error
			vcode, err = strconv.Atoi(id[j+1:])
			if err != nil {
				return nil, fmt.Errorf("could not parse version code from '%s'", id)
			}
			id = id[:j]
		}

		app, e := byId[id]
		if !e {
			return nil, fmt.Errorf("could not find app with ID '%s'", id)
		}

		if vcode > -1 {
			found := false
			for _, apk := range app.Apks {
				if apk.VersCode == vcode {
					app.Apks = []*fdroidcl.Apk{apk}
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("could not find version %d for app with ID '%s'", vcode, id)
			}
		}
		result[i] = app
	}
	return result, nil
}

func printAppDetailed(app fdroidcl.App) {
	p := func(title string, format string, args ...interface{}) {
		if format == "" {
			fmt.Fprintln(stdout, title)
		} else {
			fmt.Fprintf(stdout, "%s %s\n", title, fmt.Sprintf(format, args...))
		}
	}
	p("Package          :", "%s", app.PackageName)
	p("Name             :", "%s", app.Name)
	p("Summary          :", "%s", app.Summary)
	p("Added            :", "%s", app.Added.String())
	p("Last Updated     :", "%s", app.Updated.String())
	p("Version          :", "%s (%d)", app.SugVersName, app.SugVersCode)
	p("License          :", "%s", app.License)
	if app.Categories != nil {
		p("Categories       :", "%s", strings.Join(app.Categories, ", "))
	}
	if app.Website != "" {
		p("Website          :", "%s", app.Website)
	}
	if app.SourceCode != "" {
		p("Source Code      :", "%s", app.SourceCode)
	}
	if app.IssueTracker != "" {
		p("Issue Tracker    :", "%s", app.IssueTracker)
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
	if app.FlattrID != "" {
		p("Flattr           :", "https://flattr.com/thing/%s", app.FlattrID)
	}
	fmt.Fprintln(stdout)
	p("Description :", "")
	fmt.Fprintln(stdout)
	app.TextDesc(stdout)
	fmt.Fprintln(stdout)
	p("Available Versions :", "")
	for _, apk := range app.Apks {
		fmt.Fprintln(stdout)
		p("    Version :", "%s (%d)", apk.VersName, apk.VersCode)
		p("    Size    :", "%d", apk.Size)
		p("    MinSdk  :", "%d", apk.MinSdk)
		if apk.MaxSdk > 0 {
			p("    MaxSdk  :", "%d", apk.MaxSdk)
		}
		if apk.ABIs != nil {
			p("    ABIs    :", "%s", strings.Join(apk.ABIs, ", "))
		}
		if apk.Perms != nil {
			p("    Perms   :", "%s", strings.Join(apk.Perms, ", "))
		}
	}
}
