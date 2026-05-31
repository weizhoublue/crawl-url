package crawl

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateWorkers(t *testing.T) {
	if err := ValidateWorkers(0); err == nil {
		t.Fatal("expected error for 0 workers")
	}
	if err := ValidateWorkers(1); err != nil {
		t.Fatal(err)
	}
}

func TestReserveLimit(t *testing.T) {
	limit := 2
	cfg := Config{
		Prefix:   "https://example.com/",
		URLLimit: &limit,
		Workers:  1,
	}
	s := newState(cfg)
	go func() {
		for range s.tasks {
			s.wg.Done()
		}
	}()
	if !s.Reserve("https://example.com/a", Page, 0) {
		t.Fatal("first reserve failed")
	}
	if !s.Reserve("https://example.com/b", Page, 0) {
		t.Fatal("second reserve failed")
	}
	if s.Reserve("https://example.com/c", Page, 0) {
		t.Fatal("third reserve should fail")
	}
}

func TestValidateDepth(t *testing.T) {
	if err := ValidateDepth(0); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateDepth(1); err != nil {
		t.Fatal(err)
	}
}

func TestReserveDepth(t *testing.T) {
	max := 2
	cfg := Config{Prefix: "https://example.com/", MaxDepth: &max}
	s := newState(cfg)
	go func() {
		for range s.tasks {
			s.wg.Done()
		}
	}()
	if !s.Reserve("https://example.com/", Page, 0) {
		t.Fatal("depth 0")
	}
	if !s.Reserve("https://example.com/a", Page, 1) {
		t.Fatal("depth 1")
	}
	if s.Reserve("https://example.com/b", Page, 2) {
		t.Fatal("depth 2 should be rejected")
	}
}

func TestCrawlDepthLimit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>c</body></html>"))
	})
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><a href="/c">c</a></body></html>`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><a href="/b">b</a></body></html>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	var stdout strings.Builder
	old := printURL
	printURL = func(url string) { stdout.WriteString(url + "\n") }
	defer func() { printURL = old }()

	depth := 2
	err := Crawl(srv.URL+"/", Config{
		Prefix:     srv.URL + "/",
		MaxDepth:   &depth,
		Workers:    2,
		PerTimeout: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, "/b") {
		t.Fatalf("expected /b: %s", out)
	}
	if strings.Contains(out, "/c") {
		t.Fatalf("should not crawl /c at depth 2: %s", out)
	}

	depth = 3
	stdout.Reset()
	err = Crawl(srv.URL+"/", Config{
		Prefix:     srv.URL + "/",
		MaxDepth:   &depth,
		Workers:    2,
		PerTimeout: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "/c") {
		t.Fatalf("expected /c at depth 3: %s", stdout.String())
	}
}

func TestCrawlHTTPEndToEnd(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/data.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>leaf</body></html>"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		body := `<html><body>
<a href="/page2">p2</a>
<a href="/data.json">json</a>
</body></html>`
		_, _ = w.Write([]byte(body))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	limit := 10
	var stdout strings.Builder
	old := printURL
	printURL = func(url string) { stdout.WriteString(url + "\n") }
	defer func() { printURL = old }()

	err := Crawl(srv.URL+"/", Config{
		Prefix:     srv.URL + "/",
		URLLimit:   &limit,
		Workers:    2,
		PerTimeout: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, "/page2") {
		t.Fatalf("missing page2: %s", out)
	}
	if !strings.Contains(out, "/data.json") {
		t.Fatalf("missing json: %s", out)
	}
}
