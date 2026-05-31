package output

import (
	"strings"
	"testing"
)

func TestBuildOutputPath(t *testing.T) {
	p := BuildOutputPath("/tmp/out", "https://Example.COM/docs/page?q=1")
	if !strings.Contains(p, "example.com") {
		t.Fatalf("path missing host: %s", p)
	}
	if !strings.HasSuffix(p, ".html") {
		t.Fatalf("expected .html suffix: %s", p)
	}
	if !strings.Contains(p, "__") {
		t.Fatalf("expected digest in filename: %s", p)
	}
}
