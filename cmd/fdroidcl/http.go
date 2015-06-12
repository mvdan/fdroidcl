/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var errNotModified = errors.New("etag matches, file was not modified")

func respEtag(resp *http.Response) string {
	etags, e := resp.Header["Etag"]
	if !e || len(etags) == 0 {
		return ""
	}
	return etags[0]
}

func downloadEtag(url, path string) error {
	log.Printf("Downloading %s", url)
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
		return errNotModified
	}
	jar, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, jar, 0644)
	etag := respEtag(resp)
	err2 := ioutil.WriteFile(etagPath, []byte(etag), 0644)
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	return nil
}
