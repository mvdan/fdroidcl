/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mvdan/appdir"
)

var cmdUpdate = &Command{
	UsageLine: "update",
	Short:     "Update the index",
}

func init() {
	cmdUpdate.Run = runUpdate
}

func runUpdate(args []string) {
	if err := updateIndex(); err != nil {
		log.Fatalf("Could not update index: %v", err)
	}
}

func updateIndex() error {
	url := fmt.Sprintf("%s/%s", repoURL, "index.jar")
	if err := downloadEtag(url, indexPath(repoName)); err != nil {
		return err
	}
	return nil
}

func respEtag(resp *http.Response) string {
	etags, e := resp.Header["Etag"]
	if !e || len(etags) == 0 {
		return ""
	}
	return etags[0]
}

func downloadEtag(url, path string) error {
	fmt.Printf("Downloading %s...", url)
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
		fmt.Println(" not modified")
		return nil
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	err2 := ioutil.WriteFile(etagPath, []byte(respEtag(resp)), 0644)
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	fmt.Println(" done")
	return nil
}

func indexPath(name string) string {
	cache, err := appdir.Cache()
	if err != nil {
		log.Fatalf("Could not determine cache dir: %v", err)
	}
	return filepath.Join(appSubdir(cache), repoName+".jar")
}

func appSubdir(appdir string) string {
	p := filepath.Join(appdir, "fdroidcl")
	if err := os.MkdirAll(p, 0755); err != nil {
		log.Fatalf("Could not create app dir: %v", err)
	}
	return p
}
