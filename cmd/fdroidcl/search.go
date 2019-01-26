// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"mvdan.cc/fdroidcl"
	"mvdan.cc/fdroidcl/adb"
)

var cmdSearch = &Command{
	UsageLine: "search [<regexp...>]",
	Short:     "Search available apps",
}

var (
	searchQuiet     = cmdSearch.Fset.Bool("q", false, "Print package names only")
	searchInstalled = cmdSearch.Fset.Bool("i", false, "Filter installed apps")
	searchUpdates   = cmdSearch.Fset.Bool("u", false, "Filter apps with updates")
	searchDays      = cmdSearch.Fset.Int("d", 0, "Select apps last updated in the last <n> days; a negative value drops them instead")
	searchCategory  = cmdSearch.Fset.String("c", "", "Filter apps by category")
	searchSortBy    = cmdSearch.Fset.String("o", "", "Sort order (added, updated)")
)

func init() {
	cmdSearch.Run = runSearch
}

func runSearch(args []string) error {
	if *searchInstalled && *searchUpdates {
		return fmt.Errorf("-i is redundant if -u is specified")
	}
	sfunc, err := sortFunc(*searchSortBy)
	if err != nil {
		return err
	}
	apps, err := loadIndexes()
	if err != nil {
		return err
	}
	if len(apps) > 0 && *searchCategory != "" {
		apps = filterAppsCategory(apps, *searchCategory)
		if apps == nil {
			return fmt.Errorf("no such category: %s", *searchCategory)
		}
	}
	if len(apps) > 0 && len(args) > 0 {
		apps = filterAppsSearch(apps, args)
	}
	var device *adb.Device
	var inst map[string]adb.Package
	if *searchInstalled || *searchUpdates {
		if device, err = oneDevice(); err != nil {
			return err
		}
		if inst, err = device.Installed(); err != nil {
			return err
		}
	}
	if len(apps) > 0 && *searchInstalled {
		apps = filterAppsInstalled(apps, inst)
	}
	if len(apps) > 0 && *searchUpdates {
		apps = filterAppsUpdates(apps, inst, device)
	}
	if len(apps) > 0 && *searchDays != 0 {
		apps = filterAppsLastUpdated(apps, *searchDays)
	}
	if sfunc != nil {
		apps = sortApps(apps, sfunc)
	}
	if *searchQuiet {
		for _, app := range apps {
			fmt.Fprintln(stdout, app.PackageName)
		}
	} else {
		printApps(apps, inst, device)
	}
	return nil
}

func filterAppsSearch(apps []fdroidcl.App, terms []string) []fdroidcl.App {
	regexes := make([]*regexp.Regexp, len(terms))
	for i, term := range terms {
		regexes[i] = regexp.MustCompile(term)
	}
	var result []fdroidcl.App
	for _, app := range apps {
		fields := []string{
			strings.ToLower(app.PackageName),
			strings.ToLower(app.Name),
			strings.ToLower(app.Summary),
			strings.ToLower(app.Description),
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

func printApps(apps []fdroidcl.App, inst map[string]adb.Package, device *adb.Device) {
	maxIDLen := 0
	for _, app := range apps {
		if len(app.PackageName) > maxIDLen {
			maxIDLen = len(app.PackageName)
		}
	}
	for _, app := range apps {
		var pkg *adb.Package
		p, e := inst[app.PackageName]
		if e {
			pkg = &p
		}
		printApp(app, maxIDLen, pkg, device)
	}
}

func descVersion(app fdroidcl.App, inst *adb.Package, device *adb.Device) string {
	if inst != nil {
		suggested := app.SuggestedApk(device)
		if suggested != nil && inst.VersCode < suggested.VersCode {
			return fmt.Sprintf("%s (%d) -> %s (%d)", inst.VersName, inst.VersCode,
				suggested.VersName, suggested.VersCode)
		}
		return fmt.Sprintf("%s (%d)", inst.VersName, inst.VersCode)
	}
	return fmt.Sprintf("%s (%d)", app.SugVersName, app.SugVersCode)
}

func printApp(app fdroidcl.App, IDLen int, inst *adb.Package, device *adb.Device) {
	fmt.Fprintf(stdout, "%s%s %s - %s\n", app.PackageName, strings.Repeat(" ", IDLen-len(app.PackageName)),
		app.Name, descVersion(app, inst, device))
	fmt.Fprintf(stdout, "    %s\n", app.Summary)
}

func filterAppsInstalled(apps []fdroidcl.App, inst map[string]adb.Package) []fdroidcl.App {
	var result []fdroidcl.App
	for _, app := range apps {
		if _, e := inst[app.PackageName]; !e {
			continue
		}
		result = append(result, app)
	}
	return result
}

func filterAppsUpdates(apps []fdroidcl.App, inst map[string]adb.Package, device *adb.Device) []fdroidcl.App {
	var result []fdroidcl.App
	for _, app := range apps {
		p, e := inst[app.PackageName]
		if !e {
			continue
		}
		suggested := app.SuggestedApk(device)
		if suggested == nil {
			continue
		}
		if p.VersCode >= suggested.VersCode {
			continue
		}
		result = append(result, app)
	}
	return result
}

func filterAppsLastUpdated(apps []fdroidcl.App, days int) []fdroidcl.App {
	var result []fdroidcl.App
	newer := true
	if days < 0 {
		days = -days
		newer = false
	}
	date := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1-days)
	for _, app := range apps {
		if app.Updated.Before(date) == newer {
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
		if !contains(app.Categories, categ) {
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
	return nil, fmt.Errorf("unknown sort order: %s", sortBy)
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
