// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package fdroidcl

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"mvdan.cc/fdroidcl/adb"
)

type Index struct {
	Repo     Repo             `json:"repo"`
	Apps     []App            `json:"apps"`
	Packages map[string][]Apk `json:"packages"`
}

type Repo struct {
	Name        string   `json:"name"`
	Timestamp   UnixDate `json:"timestamp"`
	Address     string   `json:"address"`
	Icon        string   `json:"icon"`
	Version     int      `json:"version"`
	MaxAge      int      `json:"maxage"`
	Description string   `json:"description"`
}

// App is an Android application
type App struct {
	PackageName  string   `json:"packageName"`
	Name         string   `json:"name"`
	Summary      string   `json:"summary"`
	Added        UnixDate `json:"added"`
	Updated      UnixDate `json:"lastUpdated"`
	Icon         string   `json:"icon"`
	Description  string   `json:"description"`
	License      string   `json:"license"`
	Categories   []string `json:"categories"`
	Website      string   `json:"webSite"`
	SourceCode   string   `json:"sourceCode"`
	IssueTracker string   `json:"issueTracker"`
	Changelog    string   `json:"changelog"`
	Donate       string   `json:"donate"`
	Bitcoin      string   `json:"bitcoin"`
	Litecoin     string   `json:"litecoin"`
	FlattrID     string   `json:"flattr"`
	SugVersName  string   `json:"suggestedVersionName"`
	SugVersCode  int      `json:"suggestedVersionCode,string"`

	Localized map[string]Localization `json:"localized"`

	Apks []*Apk `json:"-"`
}

type Localization struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type IconDensity uint

const (
	UnknownDensity IconDensity = 0
	LowDensity     IconDensity = 120
	MediumDensity  IconDensity = 160
	HighDensity    IconDensity = 240
	XHighDensity   IconDensity = 320
	XXHighDensity  IconDensity = 480
	XXXHighDensity IconDensity = 640
)

func getIconsDir(density IconDensity) string {
	if density == UnknownDensity {
		return "icons"
	}
	for _, d := range [...]IconDensity{
		XXXHighDensity,
		XXHighDensity,
		XHighDensity,
		HighDensity,
		MediumDensity,
	} {
		if density >= d {
			return fmt.Sprintf("icons-%d", d)
		}
	}
	return fmt.Sprintf("icons-%d", LowDensity)
}

func (a *App) IconURLForDensity(density IconDensity) string {
	if len(a.Apks) == 0 {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", a.Apks[0].RepoURL,
		getIconsDir(density), a.Icon)
}

func (a *App) IconURL() string {
	return a.IconURLForDensity(UnknownDensity)
}

func (a *App) TextDesc(w io.Writer) {
	reader := strings.NewReader(a.Description)
	decoder := xml.NewDecoder(reader)
	firstParagraph := true
	linePrefix := ""
	colsUsed := 0
	var links []string
	linked := false
	for {
		token, err := decoder.Token()
		if err == io.EOF || token == nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				if firstParagraph {
					firstParagraph = false
				} else {
					fmt.Fprintln(w)
				}
				linePrefix = ""
				colsUsed = 0
			case "li":
				fmt.Fprint(w, "\n *")
				linePrefix = "   "
				colsUsed = 0
			case "a":
				for _, attr := range t.Attr {
					if attr.Name.Local == "href" {
						links = append(links, attr.Value)
						linked = true
						break
					}
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p", "ul", "ol":
				fmt.Fprintln(w)
			}
		case xml.CharData:
			left := string(t)
			if linked {
				left += fmt.Sprintf("[%d]", len(links)-1)
				linked = false
			}
			limit := 80 - len(linePrefix) - colsUsed
			firstLine := true
			for len(left) > limit {
				last := 0
				for i, c := range left {
					if i >= limit {
						break
					}
					if c == ' ' {
						last = i
					}
				}
				if firstLine {
					firstLine = false
					limit += colsUsed
				} else {
					fmt.Fprint(w, linePrefix)
				}
				fmt.Fprintln(w, left[:last])
				left = left[last+1:]
				colsUsed = 0
			}
			if !firstLine {
				fmt.Fprint(w, linePrefix)
			}
			fmt.Fprint(w, left)
			colsUsed += len(left)
		}
	}
	if len(links) > 0 {
		fmt.Fprintln(w)
		for i, link := range links {
			fmt.Fprintf(w, "[%d] %s\n", i, link)
		}
	}
}

// Apk is an Android package
type Apk struct {
	VersName string   `json:"versionName"`
	VersCode int      `json:"versionCode"`
	Size     int64    `json:"size"`
	MinSdk   int      `json:"sdkver"`
	MaxSdk   int      `json:"maxsdkver"`
	ABIs     []string `json:"nativecode"`
	ApkName  string   `json:"apkname"`
	SrcName  string   `json:"srcname"`
	Sig      HexVal   `json:"sig"`
	Signer   HexVal   `json:"signer"`
	Added    UnixDate `json:"added"`
	Perms    []string `json:"permissions"`
	Feats    []string `json:"features"`
	Hash     HexVal   `json:"hash"`
	HashType string   `json:"hashType"`

	AppID   string `json:"-"`
	RepoURL string `json:"-"`
}

func (a *Apk) URL() string {
	return fmt.Sprintf("%s/%s", a.RepoURL, a.ApkName)
}

func (a *Apk) SrcURL() string {
	return fmt.Sprintf("%s/%s", a.RepoURL, a.SrcName)
}

func (a *Apk) IsCompatibleABI(ABIs []string) bool {
	if len(a.ABIs) == 0 {
		return true // APK does not contain native code
	}
	for _, apkABI := range a.ABIs {
		for _, abi := range ABIs {
			if apkABI == abi {
				return true
			}
		}
	}
	return false
}

func (a *Apk) IsCompatibleAPILevel(sdk int) bool {
	return sdk >= a.MinSdk && (a.MaxSdk == 0 || sdk <= a.MaxSdk)
}

func (a *Apk) IsCompatible(device *adb.Device) bool {
	if device == nil {
		return true
	}
	return a.IsCompatibleABI(device.ABIs) &&
		a.IsCompatibleAPILevel(device.APILevel)
}

type AppList []App

func (al AppList) Len() int           { return len(al) }
func (al AppList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al AppList) Less(i, j int) bool { return al[i].PackageName < al[j].PackageName }

type ApkList []Apk

func (al ApkList) Len() int           { return len(al) }
func (al ApkList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al ApkList) Less(i, j int) bool { return al[i].VersCode > al[j].VersCode }

func LoadIndexJSON(r io.Reader) (*Index, error) {
	var index Index
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&index); err != nil {
		return nil, err
	}

	sort.Sort(AppList(index.Apps))

	for i := range index.Apps {
		app := &index.Apps[i]
		english, enOK := app.Localized["en"]
		if !enOK {
			english, enOK = app.Localized["en-US"]
		}
		if app.Summary == "" && enOK {
			app.Summary = english.Summary
		}
		if app.Description == "" && enOK {
			app.Description = english.Description
		}
		app.Summary = strings.TrimSpace(app.Summary)
		sort.Sort(ApkList(index.Packages[app.PackageName]))
		for i := range index.Packages[app.PackageName] {
			apk := &index.Packages[app.PackageName][i]
			apk.AppID = app.PackageName
			apk.RepoURL = index.Repo.Address
			app.Apks = append(app.Apks, apk)
		}
	}
	return &index, nil
}

func (a *App) SuggestedApk(device *adb.Device) *Apk {
	for _, apk := range a.Apks {
		if a.SugVersCode >= apk.VersCode && apk.IsCompatible(device) {
			return apk
		}
	}
	// fall back to the first compatible apk
	for _, apk := range a.Apks {
		if apk.IsCompatible(device) {
			return apk
		}
	}
	return nil
}
