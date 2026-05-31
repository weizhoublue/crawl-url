package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// 动态拓扑 Mock 服务：构建一棵最大深度为 4 (5层) 的二叉树，总节点数为 31
// D0: / (1)
// D1: /d1-p0, /d1-p1 (2)
// D2: /d2-p0 .. /d2-p3 (4)
// D3: /d3-p0 .. /d3-p7 (8)
// D4: /d4-p0 .. /d4-p15 (16)
func setupScaleServer() *httptest.Server {
	pathRegex := regexp.MustCompile(`^/d([1-4])-p([0-9]+)$`)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if r.URL.Path == "/" {
			// 深度 0，输出指向深度 1 的链接
			_, _ = w.Write([]byte(`<html><body><a href="/d1-p0">d1-0</a><a href="/d1-p1">d1-1</a></body></html>`))
			return
		}

		matches := pathRegex.FindStringSubmatch(r.URL.Path)
		if len(matches) != 3 {
			http.NotFound(w, r)
			return
		}

		depth, _ := strconv.Atoi(matches[1])
		index, _ := strconv.Atoi(matches[2])

		if depth >= 4 {
			// 深度 4 是叶子节点，无子链接
			_, _ = w.Write([]byte(`<html><body>leaf</body></html>`))
			return
		}

		// 深度 1..3，输出指向下一层 index*2 和 index*2+1 的链接
		nextDepth := depth + 1
		leftIndex := index * 2
		rightIndex := index*2 + 1
		body := fmt.Sprintf(`<html><body>
			<a href="/d%d-p%d">left</a>
			<a href="/d%d-p%d">right</a>
		</body></html>`, nextDepth, leftIndex, nextDepth, rightIndex)
		_, _ = w.Write([]byte(body))
	})

	return httptest.NewServer(mux)
}

// 场景 1：无限制全量爬取，断言输出总数精确等于 31 个，且覆盖所有节点
func TestCrawlScaleFull(t *testing.T) {
	srv := setupScaleServer()
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	// 全量 31 个页面
	if len(actualMap) != 31 {
		t.Errorf("expected exactly 31 crawled URLs, got %d. stdout:\n%s", len(actualMap), stdout)
	}

	// 验证深度为 4 的最后一层叶子节点都在内
	for i := 0; i < 16; i++ {
		leaf := fmt.Sprintf("%s/d4-p%d", srv.URL, i)
		if !actualMap[leaf] {
			t.Errorf("leaf node %s was not crawled", leaf)
		}
	}
}

// 场景 2：设定 url-limit 限制爬取数量为 15，断言行数精确为 15
func TestCrawlScaleWithLimit(t *testing.T) {
	srv := setupScaleServer()
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/", "--url-limit", "15"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	if actualCount != 15 {
		t.Errorf("expected exactly 15 crawled URLs, got %d", actualCount)
	}
}

// 场景 3：设定 depth 限制深度为 3 (抓取深度 0, 1, 2 层，即 D0 + D1 + D2 = 7 个节点)
func TestCrawlScaleWithDepth(t *testing.T) {
	srv := setupScaleServer()
	defer srv.Close()

	stdout, stderr, exitCode, err := runCLI(t, []string{srv.URL + "/", "--depth", "3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	// D0(1) + D1(2) + D2(4) = 7
	if len(actualMap) != 7 {
		t.Errorf("expected exactly 7 crawled URLs under depth 3, got %d. stdout:\n%s", len(actualMap), stdout)
	}

	// 确认深度大于等于 3 的页面均未输出
	for k := range actualMap {
		if strings.Contains(k, "/d3-") || strings.Contains(k, "/d4-") {
			t.Errorf("URL exceeding depth 3 found: %s", k)
		}
	}
}
