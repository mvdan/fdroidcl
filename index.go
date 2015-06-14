/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package fdroidcl

import (
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Index struct {
	Repo struct {
		Name        string `xml:"name,attr"`
		PubKey      string `xml:"pubkey,attr"`
		Timestamp   int    `xml:"timestamp,attr"`
		URL         string `xml:"url,attr"`
		Version     int    `xml:"version,attr"`
		MaxAge      int    `xml:"maxage,attr"`
		Description string `xml:"description"`
	} `xml:"repo"`
	Apps []App `xml:"application"`
}

type commaList []string

func (cl *commaList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}
	*cl = strings.Split(content, ",")
	return nil
}

// App is an Android application
type App struct {
	ID        string    `xml:"id"`
	Name      string    `xml:"name"`
	Summary   string    `xml:"summary"`
	Desc      string    `xml:"desc"`
	License   string    `xml:"license"`
	Categs    commaList `xml:"categories"`
	Website   string    `xml:"web"`
	Source    string    `xml:"source"`
	Tracker   string    `xml:"tracker"`
	Changelog string    `xml:"changelog"`
	Donate    string    `xml:"donate"`
	Bitcoin   string    `xml:"bitcoin"`
	Litecoin  string    `xml:"litecoin"`
	Dogecoin  string    `xml:"dogecoin"`
	FlattrID  string    `xml:"flattr"`
	Apks      []Apk     `xml:"package"`
	CVName    string    `xml:"marketversion"`
	CVCode    int       `xml:"marketvercode"`
	CurApk    *Apk
}

// Apk is an Android package
type Apk struct {
	VName   string    `xml:"version"`
	VCode   int       `xml:"versioncode"`
	Size    int64     `xml:"size"`
	MinSdk  int       `xml:"sdkver"`
	MaxSdk  int       `xml:"maxsdkver"`
	ABIs    commaList `xml:"nativecode"`
	ApkName string    `xml:"apkname"`
	SrcName string    `xml:"srcname"`
	Sig     string    `xml:"sig"`
	Added   string    `xml:"added"`
	Perms   commaList `xml:"permissions"`
	Feats   commaList `xml:"features"`
	Hash    []struct {
		Type string `xml:"type,attr"`
		Data string
	} `xml:"hash"`
}

func (app *App) calcCurApk() {
	for _, apk := range app.Apks {
		app.CurApk = &apk
		if app.CVCode >= apk.VCode {
			break
		}
	}
}

func (app *App) TextDesc(w io.Writer) {
	reader := strings.NewReader(app.Desc)
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
			if firstLine {
				firstLine = false
			} else {
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

func (app *App) prepareData() {
	app.calcCurApk()
}

type appList []App

func (al appList) Len() int           { return len(al) }
func (al appList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al appList) Less(i, j int) bool { return al[i].ID < al[j].ID }

func LoadIndexXml(r io.Reader) (*Index, error) {
	var index Index
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&index); err != nil {
		return nil, err
	}

	sort.Sort(appList(index.Apps))

	for i := range index.Apps {
		app := &index.Apps[i]
		app.prepareData()
	}
	return &index, nil
}
