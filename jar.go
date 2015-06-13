/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package fdroidcl

import (
	"archive/zip"
	"io"
)

func LoadIndexJar(r io.ReaderAt, size int64) (*Index, error) {
	reader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	var rc io.ReadCloser
	for _, f := range reader.File {
		if f.Name != "index.xml" {
			continue
		}
		rc, err = f.Open()
		if err != nil {
			return nil, err
		}
		break
	}
	defer rc.Close()
	return LoadIndexXml(rc)
}
