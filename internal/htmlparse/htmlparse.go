package htmlparse

import (
	"strings"

	"golang.org/x/net/html"
)

// Result holds link and media references extracted from HTML.
type Result struct {
	Links      []string
	MediaLinks []string
}

// ExtractLinks parses HTML and collects href, src, and poster attributes.
func ExtractLinks(htmlText string) Result {
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return Result{}
	}
	var res Result
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			tag := n.Data
			for _, a := range n.Attr {
				if a.Val == "" {
					continue
				}
				switch {
				case tag == "a" && a.Key == "href":
					res.Links = append(res.Links, a.Val)
				case a.Key == "src" || a.Key == "poster":
					res.MediaLinks = append(res.MediaLinks, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return res
}
