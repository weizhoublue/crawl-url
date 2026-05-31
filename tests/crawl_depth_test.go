package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCrawlDepth(t *testing.T) {
	mux := http.NewServeMux()
	// /d2 (D2) -> /d3 (D3)
	mux.HandleFunc("/d2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><a href="/d3">d3</a></body></html>`))
	})
	// /d1 (D1) -> /d2 (D2)
	mux.HandleFunc("/d1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><a href="/d2">d2</a></body></html>`))
	})
	// / (D0) -> /d1 (D1)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><a href="/d1">d1</a></body></html>`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// --depth 2 表示抓取深度 0 和 1（即仅抓取种子和直接子链接）
	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/", "--depth", "2"})
	if err != nil {
		t.Fatalf("unexpected error running CLI: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	actualMap := make(map[string]bool)
	for _, line := range lines {
		if line != "" {
			actualMap[line] = true
		}
	}

	expected := []string{
		srv.URL + "/",
		srv.URL + "/d1",
	}

	for _, exp := range expected {
		if !actualMap[exp] {
			t.Errorf("expected URL %q not found in stdout", exp)
		}
	}

	if actualMap[srv.URL+"/d2"] {
		t.Errorf("URL %s (depth 2) should NOT be crawled or output when --depth is 2", srv.URL+"/d2")
	}

	if len(actualMap) != len(expected) {
		t.Errorf("expected exactly %d URLs, got %d: %v", len(expected), len(actualMap), actualMap)
	}
}
