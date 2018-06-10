// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package fdroidcl

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kr/pretty"
)

func TestTextDesc(t *testing.T) {
	for _, c := range []struct {
		in   string
		want string
	}{
		{
			"Simple description.",
			"Simple description.",
		},
		{
			"<p>Simple description.</p>",
			"Simple description.\n",
		},
		{
			"<p>Multiple</p><p>Paragraphs</p>",
			"Multiple\n\nParagraphs\n",
		},
		{
			"<p>Single, very long paragraph that is over 80 characters long and doesn't fit in a single line.</p>",
			"Single, very long paragraph that is over 80 characters long and doesn't fit in\na single line.\n",
		},
		{
			"<p>Unordered list:</p><ul><li> Item</li><li> Another item</li></ul>",
			"Unordered list:\n\n * Item\n * Another item\n",
		},
		{
			"<p>Link: <a href=\"http://foo.bar\">link title</a> text</p>",
			"Link: link title[0] text\n\n[0] http://foo.bar\n",
		},
		{
			"<p>Links: <a href=\"foo\">foo1</a> <a href=\"bar\">bar1</a></p>",
			"Links: foo1[0] bar1[1]\n\n[0] foo\n[1] bar\n",
		},
	} {
		app := App{Description: c.in}
		var b bytes.Buffer
		app.TextDesc(&b)
		got := b.String()
		if got != c.want {
			t.Fatalf("Unexpected description.\nGot:\n%s\nWant:\n%s\n",
				got, c.want)
		}
	}
}

func TestLoadIndexJSON(t *testing.T) {
	in := `
{
	"repo": {
		"name": "Foo",
		"version": 19,
		"timestamp": 1528184950000
	},
	"requests": {
		"install": [],
		"uninstall": []
	},
	"apps": [
		{
			"packageName": "foo.bar",
			"name": "Foo bar",
			"categories": ["Cat1", "Cat2"],
			"added": 1443734950000,
			"suggestedVersionName": "1.0",
			"suggestedVersionCode": "1"
		},
		{
			"packageName": "localized.app",
			"localized": {
				"en": {
					"summary": "summary in english\n"
				}
			}
		}
	],
	"packages": {
		"foo.bar": [
			{
				"versionName": "1.0",
				"versionCode": 1,
				"hash": "1e4c77d8c9fa03b3a9c42360dc55468f378bbacadeaf694daea304fe1a2750f4",
				"hashType": "sha256",
				"sig": "c0f3a6d46025bf41613c5e81781e517a",
				"signer": "573c2762a2ff87c4c1ef104b35147c8c316676e5d072ec636fc718f35df6cf22"
			}
		]
	}
}
`
	want := Index{
		Repo: Repo{
			Name:      "Foo",
			Version:   19,
			Timestamp: UnixDate{time.Unix(1528184950, 0).UTC()},
		},
		Apps: []App{
			{
				PackageName: "foo.bar",
				Name:        "Foo bar",
				Categories:  []string{"Cat1", "Cat2"},
				Added:       UnixDate{time.Unix(1443734950, 0).UTC()},
				SugVersName: "1.0",
				SugVersCode: 1,
				Apks:        []*Apk{nil},
			},
			{
				PackageName: "localized.app",
				Summary:     "summary in english",
				Localized: map[string]Localization{
					"en": {Summary: "summary in english\n"},
				},
			},
		},
		Packages: map[string][]Apk{"foo.bar": {
			{
				VersName: "1.0",
				VersCode: 1,
				Sig:      HexVal{0xc0, 0xf3, 0xa6, 0xd4, 0x60, 0x25, 0xbf, 0x41, 0x61, 0x3c, 0x5e, 0x81, 0x78, 0x1e, 0x51, 0x7a},
				Signer:   HexVal{0x57, 0x3c, 0x27, 0x62, 0xa2, 0xff, 0x87, 0xc4, 0xc1, 0xef, 0x10, 0x4b, 0x35, 0x14, 0x7c, 0x8c, 0x31, 0x66, 0x76, 0xe5, 0xd0, 0x72, 0xec, 0x63, 0x6f, 0xc7, 0x18, 0xf3, 0x5d, 0xf6, 0xcf, 0x22},
				Hash:     HexVal{0x1e, 0x4c, 0x77, 0xd8, 0xc9, 0xfa, 0x3, 0xb3, 0xa9, 0xc4, 0x23, 0x60, 0xdc, 0x55, 0x46, 0x8f, 0x37, 0x8b, 0xba, 0xca, 0xde, 0xaf, 0x69, 0x4d, 0xae, 0xa3, 0x4, 0xfe, 0x1a, 0x27, 0x50, 0xf4},
				HashType: "sha256",
			},
		}},
	}
	want.Apps[0].Apks[0] = &want.Packages["foo.bar"][0]
	r := strings.NewReader(in)
	index, err := LoadIndexJSON(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	got := *index
	for i := range want.Apps {
		app := &want.Apps[i]
		for _, apk := range app.Apks {
			apk.AppID = app.PackageName
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Unexpected index.\n%s",
			strings.Join(pretty.Diff(want, got), "\n"))
	}
}
