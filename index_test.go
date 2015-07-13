// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package fdroidcl

import (
	"bytes"
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
		out := b.String()
		if out != c.want {
			t.Errorf("Description converting into text failed.\nInput:\n%s\nGot:\n%s\nWanted:\n%s\n",
				c.in, out, c.want)
		}
	}
}
