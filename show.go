// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"mvdan.cc/fdroidcl/fdroid"
)

var cmdShow = &Command{
	UsageLine: "show <appid...>",
	Short:     "Show detailed info about apps",
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
			fmt.Printf("\n--\n\n")
		}
		printAppDetailed(app)
	}
	return nil
}

func appsMap(apps []fdroid.App) map[string]*fdroid.App {
	m := make(map[string]*fdroid.App, len(apps))
	for i := range apps {
		app := &apps[i]
		m[app.PackageName] = app
	}
	return m
}

func findApps(ids []string) ([]fdroid.App, error) {
	apps, err := loadIndexes()
	if err != nil {
		return nil, err
	}
	byId := appsMap(apps)
	result := make([]fdroid.App, len(ids))
	for i, id := range ids {
		vcode := -1
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
					app.Apks = []*fdroid.Apk{apk}
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("could not find version %d for app with ID '%s'", vcode, id)
			}
		}
		result[i] = *app
	}
	return result, nil
}

func printAppDetailed(app fdroid.App) {
	fmt.Printf("Package          : %s\n", app.PackageName)
	fmt.Printf("Name             : %s\n", app.Name)
	fmt.Printf("Summary          : %s\n", app.Summary)
	fmt.Printf("Added            : %s\n", app.Added.String())
	fmt.Printf("Last Updated     : %s\n", app.Updated.String())
	fmt.Printf("Version          : %s (%d)\n", app.SugVersName, app.SugVersCode)
	fmt.Printf("License          : %s\n", app.License)
	if app.Categories != nil {
		fmt.Printf("Categories       : %s\n", strings.Join(app.Categories, ", "))
	}
	if app.Website != "" {
		fmt.Printf("Website          : %s\n", app.Website)
	}
	if app.SourceCode != "" {
		fmt.Printf("Source Code      : %s\n", app.SourceCode)
	}
	if app.IssueTracker != "" {
		fmt.Printf("Issue Tracker    : %s\n", app.IssueTracker)
	}
	if app.Changelog != "" {
		fmt.Printf("Changelog        : %s\n", app.Changelog)
	}
	if app.Donate != "" {
		fmt.Printf("Donate           : %s\n", app.Donate)
	}
	if app.Bitcoin != "" {
		fmt.Printf("Bitcoin          : bitcoin:%s\n", app.Bitcoin)
	}
	if app.Litecoin != "" {
		fmt.Printf("Litecoin         : litecoin:%s\n", app.Litecoin)
	}
	if app.FlattrID != "" {
		fmt.Printf("Flattr           : https://flattr.com/thing/%s\n", app.FlattrID)
	}
	fmt.Println()
	fmt.Println("Description :")
	fmt.Println()
	app.TextDesc(os.Stdout)
	fmt.Println()
	fmt.Println("Available Versions :")
	for _, apk := range app.Apks {
		fmt.Println()
		fmt.Printf("    Version : %s (%d)\n", apk.VersName, apk.VersCode)
		fmt.Printf("    Size    : %d\n", apk.Size)
		fmt.Printf("    MinSdk  : %d\n", apk.MinSdk)
		if apk.MaxSdk > 0 {
			fmt.Printf("    MaxSdk  : %d\n", apk.MaxSdk)
		}
		if apk.ABIs != nil {
			fmt.Printf("    ABIs    : %s\n", strings.Join(apk.ABIs, ", "))
		}
		if apk.Perms != nil {
			fmt.Printf("    Perms   : %s\n", strings.Join(apk.Perms, ", "))
		}
	}
}
