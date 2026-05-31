package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCrawlBasic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><a href="/page2">page2</a><a href="https://other.com/ext">ext</a></body></html>`))
	})
	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>leaf</body></html>`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><a href="/page1">page1</a></body></html>`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/"})
	if err != nil {
		t.Fatalf("unexpected error running CLI: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", exitCode, stderr)
	}

	expected := []string{
		srv.URL + "/",
		srv.URL + "/page1",
		srv.URL + "/page2",
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	actualMap := make(map[string]bool)
	for _, line := range lines {
		if line != "" {
			actualMap[line] = true
		}
	}

	for _, exp := range expected {
		if !actualMap[exp] {
			t.Errorf("expected URL %q not found in stdout", exp)
		}
	}

	if actualMap["https://other.com/ext"] {
		t.Errorf("external URL should not be crawled")
	}

	if len(actualMap) != len(expected) {
		t.Errorf("expected %d crawled URLs, got %d: %v", len(expected), len(actualMap), actualMap)
	}
}
