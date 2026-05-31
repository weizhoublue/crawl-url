package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"crawl-url/internal/output"
)

func TestCrawlOutput(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><h1>page1</h1></body></html>`))
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

	tempDir := t.TempDir()

	_, stderr, exitCode, err := runCLI(t, []string{
		srv.URL + "/",
		"--output-dir",
		tempDir,
	})
	if err != nil {
		t.Fatalf("unexpected error running CLI: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", exitCode, stderr)
	}

	// 验证文件是否正确生成
	expectedFiles := []string{
		output.BuildOutputPath(tempDir, srv.URL+"/"),
		output.BuildOutputPath(tempDir, srv.URL+"/page1"),
	}

	for _, p := range expectedFiles {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("expected output file not found or unreadable at %q: %v", p, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("file at %q is empty", p)
		}
	}
}
