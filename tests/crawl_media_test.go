package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupMediaServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/pic.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("fake png binary"))
	})
	mux.HandleFunc("/mov.mp4", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write([]byte("fake mp4 binary"))
	})
	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>page1</body></html>`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		body := `<html><body>
			<a href="/page1">page1</a>
			<img src="/pic.png">
			<video src="/mov.mp4"></video>
		</body></html>`
		_, _ = w.Write([]byte(body))
	})
	return httptest.NewServer(mux)
}

// 场景 1：默认不抓取媒体资源
func TestCrawlMediaDefault(t *testing.T) {
	srv := setupMediaServer()
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/"})
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

	if actualMap[srv.URL+"/pic.png"] {
		t.Errorf("image URL should not be crawled by default")
	}
	if actualMap[srv.URL+"/mov.mp4"] {
		t.Errorf("video URL should not be crawled by default")
	}
	if !actualMap[srv.URL+"/page1"] {
		t.Errorf("page1 URL should be crawled by default")
	}
}

// 场景 2：开启 --image 和 --video
func TestCrawlMediaEnabled(t *testing.T) {
	srv := setupMediaServer()
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/", "--image", "--video"})
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

	if !actualMap[srv.URL+"/pic.png"] {
		t.Errorf("image URL should be crawled with --image")
	}
	if !actualMap[srv.URL+"/mov.mp4"] {
		t.Errorf("video URL should be crawled with --video")
	}
}

// 场景 3：测试 --vedio 别名兼容性
func TestCrawlMediaAlias(t *testing.T) {
	srv := setupMediaServer()
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/", "--vedio"})
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

	if actualMap[srv.URL+"/pic.png"] {
		t.Errorf("image URL should not be crawled with --vedio only")
	}
	if !actualMap[srv.URL+"/mov.mp4"] {
		t.Errorf("video URL should be crawled with --vedio")
	}
}
