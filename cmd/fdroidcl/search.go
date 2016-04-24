// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/mvdan/adb"

	"github.com/mvdan/fdroidcl"
)

var cmdSearch = &Command{
	UsageLine: "search [<regexp...>]",
	Short:     "Search available apps",
}

var (
	quiet     = cmdSearch.Flag.Bool("q", false, "Print package names only")
	installed = cmdSearch.Flag.Bool("i", false, "Filter installed apps")
	updates   = cmdSearch.Flag.Bool("u", false, "Filter apps with updates")
	category  = cmdSearch.Flag.String("c", "", "Filter apps by category")
	sortBy    = cmdSearch.Flag.String("o", "", "Sort order (added, updated)")
)

func init() {
	cmdSearch.Run = runSearch
}

func runSearch(args []string) {
	if *installed && *updates {
		fmt.Fprintf(os.Stderr, "-i is redundant if -u is specified\n")
		cmdSearch.Flag.Usage()
	}
	sfunc, err := sortFunc(*sortBy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		cmdSearch.Flag.Usage()
	}
	var device *adb.Device
	if *installed || *updates {
		device = mustOneDevice()
	}
	apps := filterAppsSearch(mustLoadIndexes(), args)
	if *installed {
		apps = filterAppsInstalled(apps, device)
	}
	if *updates {
		apps = filterAppsUpdates(apps, device)
	}
	if *category != "" {
		apps = filterAppsCategory(apps, *category)
		if apps == nil {
			fmt.Fprintf(os.Stderr, "No such category: %s\n", *category)
			cmdSearch.Flag.Usage()
		}
	}
	if sfunc != nil {
		apps = sortApps(apps, sfunc)
	}
	if *quiet {
		for _, app := range apps {
			fmt.Println(app.ID)
		}
	} else {
		printApps(apps, device)
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

func printApps(apps []fdroidcl.App, device *adb.Device) {
	maxIDLen := 0
	for _, app := range apps {
		if len(app.ID) > maxIDLen {
			maxIDLen = len(app.ID)
		}
	}
	inst := mustInstalled(device)
	for _, app := range apps {
		var pkg *adb.Package
		p, e := inst[app.ID]
		if e {
			pkg = &p
		}
		printApp(app, maxIDLen, pkg, device)
	}
}

func descVersion(app fdroidcl.App, inst *adb.Package, device *adb.Device) string {
	// With "-u" or "-i" option there must be a connected device
	if *updates || *installed {
		suggested := app.SuggestedApk(device)
		if suggested != nil && inst.VCode < suggested.VCode {
			return fmt.Sprintf("%s (%d) -> %s (%d)", inst.VName, inst.VCode,
				suggested.VName, suggested.VCode)
		}
		return fmt.Sprintf("%s (%d)", inst.VName, inst.VCode)
	}
	// Without "-u" or "-i" we only have repositories indices
	return fmt.Sprintf("%s (%d)", app.CVName, app.CVCode)
}

func printApp(app fdroidcl.App, IDLen int, inst *adb.Package, device *adb.Device) {
	fmt.Printf("%s%s %s - %s\n", app.ID, strings.Repeat(" ", IDLen-len(app.ID)),
		app.Name, descVersion(app, inst, device))
	fmt.Printf("    %s\n", app.Summary)
}

func mustInstalled(device *adb.Device) map[string]adb.Package {
	if device == nil {
		return nil
	}
	inst, err := device.Installed()
	if err != nil {
		log.Fatalf("Could not get installed packages: %v", err)
	}
	return inst
}

func filterAppsInstalled(apps []fdroidcl.App, device *adb.Device) []fdroidcl.App {
	var result []fdroidcl.App
	inst := mustInstalled(device)
	for _, app := range apps {
		if _, e := inst[app.ID]; !e {
			continue
		}
		result = append(result, app)
	}
	return result
}

func filterAppsUpdates(apps []fdroidcl.App, device *adb.Device) []fdroidcl.App {
	var result []fdroidcl.App
	inst := mustInstalled(device)
	for _, app := range apps {
		p, e := inst[app.ID]
		if !e {
			continue
		}
		suggested := app.SuggestedApk(device)
		if suggested == nil {
			continue
		}
		if p.VCode >= suggested.VCode {
			continue
		}
		result = append(result, app)
	}
	return result
}

func contains(l []string, s string) bool {
	for _, s1 := range l {
		if s1 == s {
			return true
		}
	}
	return false
}

func filterAppsCategory(apps []fdroidcl.App, categ string) []fdroidcl.App {
	var result []fdroidcl.App
	for _, app := range apps {
		if !contains(app.Categs, categ) {
			continue
		}
		result = append(result, app)
	}
	return result
}

func cmpAdded(a, b *fdroidcl.App) bool {
	return a.Added.Before(b.Added.Time)
}

func cmpUpdated(a, b *fdroidcl.App) bool {
	return a.Updated.Before(b.Updated.Time)
}

func sortFunc(sortBy string) (func(a, b *fdroidcl.App) bool, error) {
	switch sortBy {
	case "added":
		return cmpAdded, nil
	case "updated":
		return cmpUpdated, nil
	case "":
		return nil, nil
	}
	return nil, fmt.Errorf("Unknown sort order: %s", sortBy)
}

type appList struct {
	l []fdroidcl.App
	f func(a, b *fdroidcl.App) bool
}

func (al appList) Len() int           { return len(al.l) }
func (al appList) Swap(i, j int)      { al.l[i], al.l[j] = al.l[j], al.l[i] }
func (al appList) Less(i, j int) bool { return al.f(&al.l[i], &al.l[j]) }

func sortApps(apps []fdroidcl.App, f func(a, b *fdroidcl.App) bool) []fdroidcl.App {
	sort.Sort(appList{l: apps, f: f})
	return apps
}
