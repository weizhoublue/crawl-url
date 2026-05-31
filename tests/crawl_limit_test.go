package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCrawlLimit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>a</body></html>`))
	})
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>b</body></html>`))
	})
	mux.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>c</body></html>`))
	})
	mux.HandleFunc("/d", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>d</body></html>`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		body := `<html><body>
			<a href="/a">a</a>
			<a href="/b">b</a>
			<a href="/c">c</a>
			<a href="/d">d</a>
		</body></html>`
		_, _ = w.Write([]byte(body))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/", "--url-limit", "2"})
	if err != nil {
		t.Fatalf("unexpected error running CLI: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	actualCount := 0
	for _, line := range lines {
		if line != "" {
			actualCount++
		}
	}

	if actualCount != 2 {
		t.Errorf("expected exactly 2 crawled URLs, got %d. stdout: %s", actualCount, stdout)
	}

	// 确认包含种子 URL
	if !strings.Contains(stdout, srv.URL+"/") {
		t.Errorf("stdout should contain seed URL: %s", stdout)
	}
}
