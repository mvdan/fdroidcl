/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package fdroidcl

import (
	"archive/zip"
	"errors"
	"io"
	"regexp"
)

const indexPath = "index.xml"

var (
	sigRegex = regexp.MustCompile(`^META-INF/.*\.(DSA|EC|RSA)$`)

	ErrNoIndex     = errors.New("no xml index found inside jar")
	ErrNoSigs      = errors.New("no jar signatures found")
	ErrTooManySigs = errors.New("multiple jar signatures found")
)

func verifySignature(sig io.Reader) error {
	return nil
}

func LoadIndexJar(r io.ReaderAt, size int64) (*Index, error) {
	reader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	var index io.ReadCloser
	var sig io.ReadCloser
	for _, f := range reader.File {
		if f.Name == indexPath {
			index, err = f.Open()
			if err != nil {
				return nil, err
			}
		} else if sigRegex.MatchString(f.Name) {
			if sig != nil {
				return nil, ErrTooManySigs
			}
			sig, err = f.Open()
			if err != nil {
				return nil, err
			}
		}
	}
	if index == nil {
		return nil, ErrNoIndex
	}
	defer index.Close()
	if sig == nil {
		return nil, ErrNoSigs
	}
	defer sig.Close()
	if err := verifySignature(sig); err != nil {
		return nil, err
	}
	return LoadIndexXml(index)
}
