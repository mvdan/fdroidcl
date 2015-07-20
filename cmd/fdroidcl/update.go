// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/mvdan/fdroidcl"
)

var cmdUpdate = &Command{
	UsageLine: "update",
	Short:     "Update the index",
}

func init() {
	cmdUpdate.Run = runUpdate
}

func runUpdate(args []string) {
	for _, r := range config.Repos {
		if !r.Enabled {
			continue
		}
		if err := r.updateIndex(); err != nil {
			log.Fatalf("Could not update index: %v", err)
		}
	}
}

func (r *repo) updateIndex() error {
	url := fmt.Sprintf("%s/%s", r.URL, "index.jar")
	if err := downloadEtag(url, indexPath(r.ID), nil); err != nil {
		return err
	}
	return nil
}

func (r *repo) loadIndex() (*fdroidcl.Index, error) {
	p := indexPath(r.ID)
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("could not open index: %v", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat index: %v", err)
	}
	//pubkey, err := hex.DecodeString(repoPubkey)
	//if err != nil {
	//	return nil, fmt.Errorf("could not decode public key: %v", err)
	//}
	index, err := fdroidcl.LoadIndexJar(f, stat.Size(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not load index: %v", err)
	}
	return index, nil
}

func respEtag(resp *http.Response) string {
	etags, e := resp.Header["Etag"]
	if !e || len(etags) == 0 {
		return ""
	}
	return etags[0]
}

func downloadEtag(url, path string, sum []byte) error {
	fmt.Printf("Downloading %s... ", url)
	defer fmt.Println()
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
		fmt.Printf("not modified")
		return nil
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if sum == nil {
		_, err := io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
	} else {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		got := sha256.Sum256(data)
		if !bytes.Equal(sum, got[:]) {
			return errors.New("sha256 mismatch")
		}
		if _, err := f.Write(data); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(etagPath, []byte(respEtag(resp)), 0644); err != nil {
		return err
	}
	fmt.Printf("done")
	return nil
}

func indexPath(name string) string {
	return filepath.Join(mustConfig(), name+".jar")
}

func mustLoadIndexes() []fdroidcl.App {
	m := make(map[string]*fdroidcl.App)
	for _, r := range config.Repos {
		if !r.Enabled {
			continue
		}
		index, err := r.loadIndex()
		if err != nil {
			log.Fatalf("Error while loading %s: %v", r.ID, err)
		}
		for i := range index.Apps {
			app := index.Apps[i]
			orig, e := m[app.ID]
			if !e {
				m[app.ID] = &app
				continue
			}
			apks := append(orig.Apks, app.Apks...)
			// We use a stable sort so that repository order
			// (priority) is preserved amongst apks with the same
			// vercode on apps
			sort.Stable(fdroidcl.ApkList(apks))
			m[app.ID].Apks = apks
		}
	}
	apps := make([]fdroidcl.App, 0, len(m))
	for _, a := range m {
		apps = append(apps, *a)
	}
	sort.Sort(fdroidcl.AppList(apps))
	return apps
}
