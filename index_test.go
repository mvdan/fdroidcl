// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package fdroidcl

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
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
		app := App{Desc: c.in}
		var b bytes.Buffer
		app.TextDesc(&b)
		got := b.String()
		if got != c.want {
			t.Fatalf("Unexpected description.\nGot:\n%s\nWant:\n%s\n",
				got, c.want)
		}
	}
}

func TestLoadIndexXML(t *testing.T) {
	tests := []struct {
		in   string
		want Index
	}{
		{
			`<fdroid>
			<repo name="Foo" version="14"/>
			<application>
				<id>foo.bar</id>
				<name>Foo bar</name>
				<package>
					<version>1.0</version>
					<versioncode>1</versioncode>
				</package>
			</application>
			</fdroid>`,
			Index{
				Repo: Repo{
					Name: "Foo",
					Version: 14,
				},
				Apps: []App{
					{
						ID:   "foo.bar",
						Name: "Foo bar",
						Apks: []Apk{
							{
								VName: "1.0",
								VCode: 1,
							},
						},
					},
				},
			},
		},
	}
	for _, c := range tests {
		r := strings.NewReader(c.in)
		index, err := LoadIndexXML(r)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		got := *index
		for i := range c.want.Apps {
			app := &c.want.Apps[i]
			for i := range app.Apks {
				apk := &app.Apks[i]
				apk.Repo = &got.Repo
				apk.App = app
			}
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("Unexpected index.\nGot:\n%v\nWant:\n%v\n",
				got, c.want)
		}
	}
}
