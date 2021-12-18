// This extension adds replacer renderer to change urls.

package srcreplacer

import (
	"bytes"
	"fmt"
	"log"

	"golang.org/x/net/html"
)

func ReplaceHTML(rawHTML []byte, linkReplacer, imgReplacer func(string) string) ([]byte, error) {
	getAttr := func(attrs []html.Attribute, key string) string {
		for _, attr := range attrs {
			if attr.Key == key {
				return attr.Val
			}
		}
		return ""
	}

	setAttr := func(attrs []html.Attribute, key, val string) {
		for i, attr := range attrs {
			if attr.Key == key {
				attrs[i].Val = val
			}
		}
	}

	reader := bytes.NewReader(rawHTML)
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("scrreplacer.replaceHTML: %w", err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				href := getAttr(n.Attr, "href")
				href = linkReplacer(href)
				setAttr(n.Attr, "href", href)
			case "img":
				src := getAttr(n.Attr, "src")
				src = imgReplacer(src)
				setAttr(n.Attr, "src", src)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		return nil, fmt.Errorf("scrreplacer.replaceHTML: %w", err)
	}

	log.Println(string(buf.Bytes()))

	return buf.Bytes(), nil
}
