// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/schollz/progressbar/v3"
	"mvdan.cc/fdroidcl/fdroid"
)

var cmdUpdate = &Command{
	UsageLine: "update",
	Short:     "Update the index",
}

func init() {
	cmdUpdate.Run = runUpdate
}

func runUpdate(args []string) error {
	anyModified := false
	for _, r := range config.Repos {
		if !r.Enabled {
			continue
		}
		if err := r.updateIndex(); err == errNotModified {
		} else if err != nil {
			return fmt.Errorf("could not update index: %v", err)
		} else {
			anyModified = true
		}
	}
	if anyModified {
		cachePath := filepath.Join(mustCache(), "cache-gob")
		os.Remove(cachePath)
	}
	return nil
}

const jarFile = "index-v1.jar"

func (r *repo) updateIndex() error {
	url := fmt.Sprintf("%s/%s", r.URL, jarFile)
	return downloadEtag(url, indexPath(r.ID), nil)
}

func (r *repo) loadIndex() (*fdroid.Index, error) {
	p := indexPath(r.ID)
	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("index does not exist; try 'fdroidcl update'")
	} else if err != nil {
		return nil, fmt.Errorf("could not open index: %v", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat index: %v", err)
	}
	return fdroid.LoadIndexJar(f, stat.Size(), nil)
}

func respEtag(resp *http.Response) string {
	etags, e := resp.Header["Etag"]
	if !e || len(etags) == 0 {
		return ""
	}
	return etags[0]
}

var errNotModified = fmt.Errorf("not modified")

var httpClient = &http.Client{}

func downloadEtag(url, target_path string, sum []byte) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	etagPath := target_path + "-etag"
	if _, err := os.Stat(target_path); err == nil {
		etag, _ := os.ReadFile(etagPath)
		req.Header.Add("If-None-Match", string(etag))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	filename := path.Base(url)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s download failed: %d %s",
			filename, resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	if resp.StatusCode == http.StatusNotModified {
		fmt.Printf("%s not modified\n", filename)
		return errNotModified
	}
	f, err := os.OpenFile(target_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription(fmt.Sprintf("%-50s", filename)),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionShowBytes(true),
		progressbar.OptionThrottle(50*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stdout, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionUseANSICodes(runtime.GOOS != "windows"),
		progressbar.OptionFullWidth(),
	)
	if sum == nil {
		_, err := io.Copy(io.MultiWriter(f, bar), resp.Body)
		if err != nil {
			return err
		}
	} else {
		hash := sha256.New()
		_, err := io.Copy(io.MultiWriter(f, bar, hash), resp.Body)
		if err != nil {
			return err
		}
		got := hash.Sum(nil)
		if !bytes.Equal(sum, got[:]) {
			return fmt.Errorf("%s sha256 mismatch", filename)
		}
	}
	if err := os.WriteFile(etagPath, []byte(respEtag(resp)), 0o644); err != nil {
		return err
	}
	return nil
}

func indexPath(name string) string {
	return filepath.Join(mustData(), name+".jar")
}

const cacheVersion = 2

type cache struct {
	Version int
	Apps    []fdroid.App
}

type apkPtrList []*fdroid.Apk

func (al apkPtrList) Len() int           { return len(al) }
func (al apkPtrList) Swap(i, j int)      { al[i], al[j] = al[j], al[i] }
func (al apkPtrList) Less(i, j int) bool { return al[i].VersCode > al[j].VersCode }

func loadIndexes() ([]fdroid.App, error) {
	cachePath := filepath.Join(mustCache(), "cache-gob")
	if f, err := os.Open(cachePath); err == nil {
		defer f.Close()
		var c cache
		if err := gob.NewDecoder(f).Decode(&c); err == nil && c.Version == cacheVersion {
			return c.Apps, nil
		}
	}
	m := make(map[string]*fdroid.App)
	for _, r := range config.Repos {
		if !r.Enabled {
			continue
		}
		index, err := r.loadIndex()
		if err != nil {
			return nil, fmt.Errorf("error while loading %s: %v", r.ID, err)
		}
		for i := range index.Apps {
			app := index.Apps[i]
			orig, e := m[app.PackageName]
			if !e {
				m[app.PackageName] = &app
				continue
			}
			apks := append(orig.Apks, app.Apks...)
			// We use a stable sort so that repository order
			// (priority) is preserved amongst apks with the same
			// vercode on apps
			sort.Stable(apkPtrList(apks))
			m[app.PackageName].Apks = apks
		}
	}
	apps := make([]fdroid.App, 0, len(m))
	for _, a := range m {
		apps = append(apps, *a)
	}
	sort.Sort(fdroid.AppList(apps))
	if f, err := os.Create(cachePath); err == nil {
		defer f.Close()
		gob.NewEncoder(f).Encode(cache{
			Version: cacheVersion,
			Apps:    apps,
		})
	}
	return apps, nil
}
