// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package fdroidcl

import (
	"encoding/hex"
	"strings"
	"time"
)

type CommaList []string

func (cl *CommaList) String() string {
	return strings.Join(*cl, ",")
}

func (cl *CommaList) UnmarshalText(text []byte) error {
	*cl = strings.Split(string(text), ",")
	return nil
}

type HexHash struct {
	Type string `xml:"type,attr"`
	Data HexVal `xml:",chardata"`
}

type HexVal []byte

func (hv *HexVal) String() string {
	return hex.EncodeToString(*hv)
}

func (hv *HexVal) UnmarshalText(text []byte) error {
	b, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	*hv = b
	return nil
}

type DateVal struct {
	time.Time
}

const dateFormat = "2006-01-02"

func (dv *DateVal) String() string {
	return dv.Format(dateFormat)
}

func (dv *DateVal) UnmarshalText(text []byte) error {
	t, err := time.Parse(dateFormat, string(text))
	if err != nil {
		return err
	}
	*dv = DateVal{t}
	return nil
}
