package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCrawlHTTPErrorTolerance(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/err404", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	mux.HandleFunc("/err500", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 internal server error"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>
			<a href="/err404">err404</a>
			<a href="/err500">err500</a>
		</body></html>`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// 开启 --debug 以校验失败原因是否在 stderr 中记录
	stdout, stderr, exitCode, err := runCLI(t, []string{
		srv.URL + "/",
		"--debug",
	})
	if err != nil {
		t.Fatalf("unexpected error running CLI: %v", err)
	}

	// 单个子页面访问失败不能影响整个爬行结果，最终退出码仍应为 0
	if exitCode != 0 {
		t.Fatalf("expected exit code 0 on minor HTTP errors, got %d. stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	actualMap := make(map[string]bool)
	for _, line := range lines {
		if line != "" {
			actualMap[line] = true
		}
	}

	if !actualMap[srv.URL+"/"] {
		t.Errorf("expected seed URL in stdout")
	}

	if actualMap[srv.URL+"/err404"] || actualMap[srv.URL+"/err500"] {
		t.Errorf("failed URLs should NOT be present in stdout: %v", actualMap)
	}

	// 校验错误是否记录在 stderr 中
	if !strings.Contains(stderr, "HTTP 错误") {
		t.Errorf("expected stderr to contain error messages, got: %s", stderr)
	}
}
