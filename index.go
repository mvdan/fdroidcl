/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	ID       string    `xml:"id"`
	Name     string    `xml:"name"`
	Summary  string    `xml:"summary"`
	Desc     string    `xml:"desc"`
	License  string    `xml:"license"`
	Categs   CommaList `xml:"categories"`
	Website  string    `xml:"web"`
	Source   string    `xml:"source"`
	Tracker  string    `xml:"tracker"`
	Donate   string    `xml:"donate"`
	Bitcoin  string    `xml:"bitcoin"`
	Litecoin string    `xml:"litecoin"`
	Dogecoin string    `xml:"dogecoin"`
	FlattrID string    `xml:"flattr"`
	Apks     []Apk     `xml:"package"`
	CVName   string    `xml:"marketversion"`
	CVCode   uint      `xml:"marketvercode"`
	CurApk   *Apk
}

// Apk is an Android package
type Apk struct {
	VName  string    `xml:"version"`
	VCode  uint      `xml:"versioncode"`
	Size   int       `xml:"size"`
	MinSdk int       `xml:"sdkver"`
	MaxSdk int       `xml:"maxsdkver"`
	ABIs   CommaList `xml:"nativecode"`
}

func (app *App) calcCurApk() {
	for _, apk := range app.Apks {
		app.CurApk = &apk
		if app.CVCode >= apk.VCode {
			break
		}
	}
}

func (app *App) writeTextDesc(w io.Writer) {
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
	app.writeTextDesc(w)
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

var ErrNotModified = errors.New("etag matches, file was not modified")

func downloadEtag(url, path string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	etagPath := path + "-etag"
	if _, err := os.Stat(path); err == nil {
		etag, _ := ioutil.ReadFile(etagPath)
		req.Header.Add("If-None-Match", string(etag))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified {
		return ErrNotModified
	}
	jar, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, jar, 0644)
	err2 := ioutil.WriteFile(etagPath, []byte(resp.Header["Etag"][0]), 0644)
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	return nil
}

const indexName = "index.jar"

func updateIndex() {
	url := fmt.Sprintf("%s/%s", *repoURL, indexName)
	log.Printf("Downloading %s", url)
	err := downloadEtag(url, indexName)
	if err == ErrNotModified {
		log.Printf("Index is already up to date")
	} else if err != nil {
		log.Fatalf("Could not update index: %s", err)
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
