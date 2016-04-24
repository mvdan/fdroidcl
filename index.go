// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package fdroidcl

import (
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mvdan/adb"
)

type Index struct {
	Repo Repo  `xml:"repo"`
	Apps []App `xml:"application"`
}

type Repo struct {
	Name        string `xml:"name,attr"`
	PubKey      string `xml:"pubkey,attr"`
	Timestamp   int    `xml:"timestamp,attr"`
	URL         string `xml:"url,attr"`
	Version     int    `xml:"version,attr"`
	MaxAge      int    `xml:"maxage,attr"`
	Description string `xml:"description"`
}

// App is an Android application
type App struct {
	ID        string    `xml:"id"`
	Name      string    `xml:"name"`
	Summary   string    `xml:"summary"`
	Added     DateVal   `xml:"added"`
	Updated   DateVal   `xml:"lastupdated"`
	Icon      string    `xml:"icon"`
	Desc      string    `xml:"desc"`
	License   string    `xml:"license"`
	Categs    CommaList `xml:"categories"`
	Website   string    `xml:"web"`
	Source    string    `xml:"source"`
	Tracker   string    `xml:"tracker"`
	Changelog string    `xml:"changelog"`
	Donate    string    `xml:"donate"`
	Bitcoin   string    `xml:"bitcoin"`
	Litecoin  string    `xml:"litecoin"`
	FlattrID  string    `xml:"flattr"`
	Apks      []Apk     `xml:"package"`
	CVName    string    `xml:"marketversion"`
	CVCode    int       `xml:"marketvercode"`
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
	return fmt.Sprintf("%s/%s/%s", a.Apks[0].Repo.URL,
		getIconsDir(density), a.Icon)
}

func (a *App) IconURL() string {
	return a.IconURLForDensity(UnknownDensity)
}

func (a *App) TextDesc(w io.Writer) {
	reader := strings.NewReader(a.Desc)
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
			case "p":
				fmt.Fprintln(w)
			case "ul":
				fmt.Fprintln(w)
			case "ol":
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
	VName   string    `xml:"version"`
	VCode   int       `xml:"versioncode"`
	Size    int64     `xml:"size"`
	MinSdk  int       `xml:"sdkver"`
	MaxSdk  int       `xml:"maxsdkver"`
	ABIs    CommaList `xml:"nativecode"`
	ApkName string    `xml:"apkname"`
	SrcName string    `xml:"srcname"`
	Sig     HexVal    `xml:"sig"`
	Added   DateVal   `xml:"added"`
	Perms   CommaList `xml:"permissions"`
	Feats   CommaList `xml:"features"`
	Hash    HexHash   `xml:"hash"`

	App  *App  `xml:"-"`
	Repo *Repo `xml:"-"`
}

func (a *Apk) URL() string {
	return fmt.Sprintf("%s/%s", a.Repo.URL, a.ApkName)
}

func (a *Apk) SrcURL() string {
	return fmt.Sprintf("%s/%s", a.Repo.URL, a.SrcName)
}

func (apk *Apk) IsCompatibleABI(ABIs []string) bool {
	if len(apk.ABIs) == 0 {
		return true // APK does not contain native code
	}
	for i := range apk.ABIs {
		for j := range ABIs {
			if apk.ABIs[i] == ABIs[j] {
				return true
			}
		}
	}
	return false
}

func (apk *Apk) IsCompatibleAPILevel(sdk int) bool {
	return sdk >= apk.MinSdk && (apk.MaxSdk == 0 || sdk <= apk.MaxSdk)
}

func (apk *Apk) IsCompatible(device *adb.Device) bool {
	return apk.IsCompatibleABI(device.ABIs) &&
		apk.IsCompatibleAPILevel(device.APILevel)
}

type AppList []App

func (al AppList) Len() int           { return len(al) }
func (al AppList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al AppList) Less(i, j int) bool { return al[i].ID < al[j].ID }

type ApkList []Apk

func (al ApkList) Len() int           { return len(al) }
func (al ApkList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al ApkList) Less(i, j int) bool { return al[i].VCode > al[j].VCode }

func LoadIndexXML(r io.Reader) (*Index, error) {
	var index Index
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&index); err != nil {
		return nil, err
	}

	sort.Sort(AppList(index.Apps))

	for i := range index.Apps {
		app := &index.Apps[i]
		sort.Sort(ApkList(app.Apks))
		for j := range app.Apks {
			apk := &app.Apks[j]
			apk.App = app
			apk.Repo = &index.Repo
		}
	}
	return &index, nil
}

func (a *App) ApksByVName(vname string) []Apk {
	var apks []Apk
	for i := range a.Apks {
		if vname == a.Apks[i].VName {
			apks = append(apks, a.Apks[i])
		}
	}
	return apks
}

func (a *App) SuggestedVName() string {
	for i := range a.Apks {
		apk := &a.Apks[i]
		if a.CVCode >= apk.VCode {
			return apk.VName
		}
	}
	return ""
}

func (a *App) SuggestedApks() []Apk {
	// No APKs => nothing to suggest
	if len(a.Apks) == 0 {
		return nil
	}

	// First, try to follow CV
	apks := a.ApksByVName(a.SuggestedVName())
	if len(apks) > 0 {
		return apks
	}

	// When CV is missing current version code or it's invalid (no APKs
	// match it), use heuristic: find all APKs having the same version
	// string as the APK with the greatest version code
	return a.ApksByVName(a.Apks[0].VName)
}

func (a *App) SuggestedApk(device *adb.Device) *Apk {
	for i := range a.Apks {
		apk := &a.Apks[i]
		if a.CVCode >= apk.VCode && apk.IsCompatible(device) {
			return apk
		}
	}
	for i := range a.Apks {
		apk := &a.Apks[i]
		if apk.IsCompatible(device) {
			return apk
		}
	}
	return nil
}
