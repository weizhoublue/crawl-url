package htmlparse

import "testing"

func TestExtractLinks(t *testing.T) {
	html := `<html><body>
<a href="/page2">x</a>
<img src="/img/a.png">
<video poster="/vid/poster.jpg" src="/vid/a.mp4"></video>
</body></html>`
	res := ExtractLinks(html)
	if len(res.Links) != 1 || res.Links[0] != "/page2" {
		t.Fatalf("links: %#v", res.Links)
	}
	if len(res.MediaLinks) != 3 {
		t.Fatalf("media: %#v", res.MediaLinks)
	}
}
