/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package fdroidcl

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

func indexPath(repoName string) string {
	return repoName + ".jar"
}

func UpdateIndex(repoName, repoURL string) error {
	path := indexPath(repoName)
	url := fmt.Sprintf("%s/%s", repoURL, path)
	if err := downloadEtag(url, path); err != nil {
		return err
	}
	return nil
}

func LoadRepo(repoName string) (*Repo, error) {
	path := indexPath(repoName)
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buf := new(bytes.Buffer)

	for _, f := range r.File {
		if f.Name != "index.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		if _, err = io.Copy(buf, rc); err != nil {
			return nil, err
		}
		rc.Close()
		break
	}

	var repo Repo
	if err := xml.Unmarshal(buf.Bytes(), &repo); err != nil {
		return nil, err
	}

	for i := range repo.Apps {
		app := &repo.Apps[i]
		app.prepareData()
	}
	return &repo, nil
}
