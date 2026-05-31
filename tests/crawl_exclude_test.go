package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCrawlExclude(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/allowed", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>allowed page</body></html>`))
	})
	mux.HandleFunc("/static/img.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>image page</body></html>`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>
			<a href="/allowed">allowed</a>
			<a href="/static/img.png">exclude me</a>
		</body></html>`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{
		srv.URL + "/",
		"--exclude-prefix",
		srv.URL + "/static/",
	})
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

	if !actualMap[srv.URL+"/allowed"] {
		t.Errorf("expected allowed URL to be crawled")
	}

	if actualMap[srv.URL+"/static/img.png"] {
		t.Errorf("URL with excluded prefix was crawled: %s", srv.URL+"/static/img.png")
	}
}
