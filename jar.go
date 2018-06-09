// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package fdroidcl

import (
	"archive/zip"
	"errors"
	"io"
)

const indexPath = "index-v1.json"

var ErrNoIndex = errors.New("no json index found inside jar")

func LoadIndexJar(r io.ReaderAt, size int64, pubkey []byte) (*Index, error) {
	reader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	var index io.ReadCloser
	for _, f := range reader.File {
		if f.Name == indexPath {
			index, err = f.Open()
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if index == nil {
		return nil, ErrNoIndex
	}
	defer index.Close()
	return LoadIndexJSON(index)
}
