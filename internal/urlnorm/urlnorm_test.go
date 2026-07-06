package urlnorm

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"HTTPS://Example.COM/docs/", "https://example.com/docs/"},
		{"https://example.com/a/../b", "https://example.com/b"},
		{"https://example.com/foo?x=1#frag", "https://example.com/foo?x=1"},
	}
	for _, tc := range tests {
		if got := NormalizeURL(tc.in); got != tc.want {
			t.Errorf("NormalizeURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizePrefix(t *testing.T) {
	got := NormalizePrefix("https://example.com/docs/page")
	want := "https://example.com/docs/page/"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = NormalizePrefix("https://example.com/docs/")
	want = "https://example.com/docs/"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestInPrefix(t *testing.T) {
	prefix := "https://example.com/docs/"
	if !InPrefix("https://example.com/docs/a.html", prefix) {
		t.Fatal("expected in prefix")
	}
	if !InPrefix("https://example.com/docs", prefix) {
		t.Fatal("expected directory URL without trailing slash in prefix")
	}
	if InPrefix("https://other.com/docs/a.html", prefix) {
		t.Fatal("expected out of prefix")
	}
}

func TestHasExtension(t *testing.T) {
	if !HasExtension("https://x.com/a.PNG", ImageExtensions) {
		t.Fatal("expected image extension")
	}
}

func TestHasExtensionText(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://x.com/README.md", true},
		{"https://x.com/README.MD", true}, // case-insensitive
		{"https://x.com/config.yaml", true},
		{"https://x.com/config.yml", true},
		{"https://x.com/data.json", true},
		{"https://x.com/data.xml", true},
		{"https://x.com/notes.txt", true},
		{"https://x.com/index.mdx", true},
		{"https://x.com/report.csv", true},
		{"https://x.com/docs.rst", true},
		{"https://x.com/pyproject.toml", true},
		{"https://x.com/page.html", false},
		{"https://x.com/image.png", false},
		{"https://x.com/video.mp4", false},
	}
	for _, tc := range cases {
		got := HasExtension(tc.url, TextExtensions)
		if got != tc.want {
			t.Errorf("HasExtension(%q, TextExtensions) = %v, want %v", tc.url, got, tc.want)
		}
	}
}
