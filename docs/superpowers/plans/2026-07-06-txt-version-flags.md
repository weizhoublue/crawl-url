# --txt and --version Flags Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `--txt` opt-in flag to capture text-file URLs and `--version`/`-v` flag to print the version string.

**Architecture:** Follows the existing `--image`/`--video` pattern exactly: new `TextExtensions` set in `urlnorm`, `Txt bool` in `Config`, early-exit guard in `handleDiscovered`, and flag parsing in `flags.go`/`main.go`. Version is a hardcoded constant with an `errVersion` sentinel mirroring the existing `errHelp` pattern.

**Tech Stack:** Go 1.21+, standard library only.

## Global Constraints

- All new flag names use `--` long-option format (existing convention).
- Boolean flags must reject `--flag=value` form with an error message.
- New extensions are all lowercase; `HasExtension` already lowercases the URL path ext before lookup.
- Version string: `v0.2.0` (hardcoded constant `version` in `flags.go`).
- No new dependencies.

---

## Environment Note

**Pre-existing build failures:** `internal/crawl` and `cmd/crawl-url` packages currently fail to build due to a `golang.org/x/text` module proxy issue on this machine (git `--end-of-options` not supported). This is unrelated to our changes. Test commands per task reflect this:

- `go test ./internal/urlnorm/...` → runs fine
- `go test ./internal/crawl/...` → build fails (pre-existing); verify by inspecting compilation errors — if the ONLY errors are the pre-existing charset errors, the logic is correct
- `go test ./cmd/crawl-url/...` → same pre-existing build failure applies

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `internal/urlnorm/urlnorm.go` | Modify | Add `TextExtensions` variable |
| `internal/urlnorm/urlnorm_test.go` | Modify | Add `HasExtension` test for text extensions |
| `internal/crawl/state.go` | Modify | Add `Txt bool` to `Config`; extend `handleDiscovered` |
| `internal/crawl/crawl_test.go` | Modify | Add unit tests for `handleDiscovered` with text extensions |
| `cmd/crawl-url/flags.go` | Modify | Add `version` const, `errVersion`, `--txt` case, `--version`/`-v` early scan, update `printUsage` |
| `cmd/crawl-url/main.go` | Modify | Handle `errVersion` sentinel |
| `cmd/crawl-url/flags_test.go` | Modify | Tests for `--txt` and `--version`/`-v` |

---

## Task 1: TextExtensions in urlnorm

**Files:**
- Modify: `internal/urlnorm/urlnorm.go`
- Modify: `internal/urlnorm/urlnorm_test.go`

**Interfaces:**
- Produces: `TextExtensions map[string]struct{}` — used by Task 2's `handleDiscovered`

- [ ] **Step 1: Write the failing test**

Add to `internal/urlnorm/urlnorm_test.go` after the existing `TestHasExtension`:

```go
func TestHasExtensionText(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://x.com/README.md", true},
		{"https://x.com/README.MD", true}, // case-insensitive
		{"https://x.com/config.yaml", true},
		{"https://x.com/config.yml", true},
		{"https://x.com/data.json", true},
		{"https://x.com/data.xml", true},
		{"https://x.com/notes.txt", true},
		{"https://x.com/index.mdx", true},
		{"https://x.com/report.csv", true},
		{"https://x.com/docs.rst", true},
		{"https://x.com/pyproject.toml", true},
		{"https://x.com/page.html", false},
		{"https://x.com/image.png", false},
		{"https://x.com/video.mp4", false},
	}
	for _, tc := range cases {
		got := HasExtension(tc.url, TextExtensions)
		if got != tc.want {
			t.Errorf("HasExtension(%q, TextExtensions) = %v, want %v", tc.url, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/urlnorm/... -run TestHasExtensionText -v
```

Expected: `FAIL` — `undefined: TextExtensions`

- [ ] **Step 3: Add TextExtensions to urlnorm.go**

In `internal/urlnorm/urlnorm.go`, extend the `var` block (after `VideoExtensions`):

```go
var (
	ImageExtensions = extSet(
		".apng", ".avif", ".gif", ".jpeg", ".jpg", ".png", ".svg", ".webp",
	)
	VideoExtensions = extSet(
		".avi", ".m4v", ".mkv", ".mov", ".mp4", ".mpeg", ".mpg", ".ogv", ".webm",
	)
	TextExtensions = extSet(
		".md", ".mdx", ".txt", ".yaml", ".yml",
		".json", ".xml", ".csv", ".rst", ".toml",
	)
)
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/urlnorm/... -v
```

Expected: all tests PASS (including the new `TestHasExtensionText`)

- [ ] **Step 5: Commit**

```bash
git add internal/urlnorm/urlnorm.go internal/urlnorm/urlnorm_test.go
git commit -m "feat(urlnorm): add TextExtensions for text file URLs"
```

---

## Task 2: Config.Txt and handleDiscovered

**Files:**
- Modify: `internal/crawl/state.go`
- Modify: `internal/crawl/crawl_test.go`

**Interfaces:**
- Consumes: `urlnorm.TextExtensions` (from Task 1)
- Produces: `Config.Txt bool` — consumed by Task 3's CLI parsing

- [ ] **Step 1: Write the failing tests**

Add to `internal/crawl/crawl_test.go`:

```go
func TestHandleDiscoveredTxtOff(t *testing.T) {
	cfg := Config{Prefix: "https://example.com/", Workers: 1}
	s := newState(cfg)
	// drain tasks so channel never blocks (buffer=4096, but be safe)
	go func() {
		for range s.tasks {
			s.wg.Done()
		}
	}()

	s.handleDiscovered("https://example.com/README.md", true, 0)
	s.handleDiscovered("https://example.com/config.yaml", true, 0)
	s.handleDiscovered("https://example.com/data.json", true, 0)

	s.mu.Lock()
	count := s.acceptedCount
	s.mu.Unlock()

	if count != 0 {
		t.Fatalf("--txt off: expected 0 accepted, got %d", count)
	}
}

func TestHandleDiscoveredTxtOn(t *testing.T) {
	cfg := Config{Prefix: "https://example.com/", Txt: true, Workers: 1}
	s := newState(cfg)
	go func() {
		for range s.tasks {
			s.wg.Done()
		}
	}()

	s.handleDiscovered("https://example.com/README.md", true, 0)
	s.handleDiscovered("https://example.com/config.yaml", true, 0)
	s.handleDiscovered("https://example.com/data.json", true, 0)

	s.mu.Lock()
	count := s.acceptedCount
	s.mu.Unlock()

	if count != 3 {
		t.Fatalf("--txt on: expected 3 accepted, got %d", count)
	}
}
```

- [ ] **Step 2: Attempt build to verify it fails**

```bash
go build ./internal/crawl/...
```

Expected: build error — `cfg.Txt undefined (type Config has no field or method Txt)` (plus the pre-existing charset error; ignore that)

- [ ] **Step 3: Add Txt bool to Config**

In `internal/crawl/state.go`, extend the `Config` struct (after `Video bool`):

```go
type Config struct {
	Prefix          string
	URLLimit        *int
	MaxDepth        *int
	Image           bool
	Video           bool
	Txt             bool
	PerTimeout      float64
	ExcludePrefixes []string
	Debug           bool
	OutputDir       string
	Workers         int
}
```

- [ ] **Step 4: Extend handleDiscovered**

In `internal/crawl/state.go`, in the `handleDiscovered` method, add the text extension block **after** the Video block and **before** the final `s.Reserve(candidate, Page, childDepth)`:

```go
func (s *State) handleDiscovered(candidate string, crawlPage bool, parentDepth int) {
	childDepth := parentDepth + 1
	if !urlnorm.InPrefix(candidate, s.cfg.Prefix) {
		return
	}
	if urlnorm.HasExtension(candidate, urlnorm.ImageExtensions) {
		if s.cfg.Image {
			s.Reserve(candidate, Media, childDepth)
		}
		return
	}
	if urlnorm.HasExtension(candidate, urlnorm.VideoExtensions) {
		if s.cfg.Video {
			s.Reserve(candidate, Media, childDepth)
		}
		return
	}
	if urlnorm.HasExtension(candidate, urlnorm.TextExtensions) {
		if s.cfg.Txt {
			s.Reserve(candidate, Media, childDepth)
		}
		return
	}
	if !crawlPage {
		return
	}
	s.Reserve(candidate, Page, childDepth)
}
```

- [ ] **Step 5: Attempt build to confirm only pre-existing errors remain**

```bash
go build ./internal/crawl/... 2>&1 | grep -v "golang.org/x/text"
```

Expected: no output (only charset errors are suppressed; our code compiles cleanly)

- [ ] **Step 6: Commit**

```bash
git add internal/crawl/state.go internal/crawl/crawl_test.go
git commit -m "feat(crawl): add --txt support to Config and handleDiscovered"
```

---

## Task 3: CLI --txt and --version flags

**Files:**
- Modify: `cmd/crawl-url/flags.go`
- Modify: `cmd/crawl-url/main.go`
- Modify: `cmd/crawl-url/flags_test.go`

**Interfaces:**
- Consumes: `Config.Txt bool` (from Task 2), `crawl.Config` struct
- Produces: `--txt`, `--version`, `-v` CLI options

- [ ] **Step 1: Write the failing tests**

Add to `cmd/crawl-url/flags_test.go`:

```go
func TestParseCLITxt(t *testing.T) {
	opts, err := parseCLI([]string{"https://example.com/docs", "--txt"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.cfg.Txt {
		t.Fatal("expected cfg.Txt = true")
	}
}

func TestParseCLITxtRejectsValue(t *testing.T) {
	_, err := parseCLI([]string{"https://example.com/docs", "--txt=foo"})
	if err == nil {
		t.Fatal("expected error for --txt=foo")
	}
}

func TestParseCLIVersion(t *testing.T) {
	_, err := parseCLI([]string{"--version"})
	if err != errVersion {
		t.Fatalf("expected errVersion, got %v", err)
	}
}

func TestParseCLIVersionShort(t *testing.T) {
	_, err := parseCLI([]string{"-v"})
	if err != errVersion {
		t.Fatalf("expected errVersion for -v, got %v", err)
	}
}

func TestParseCLIVersionBeforeSeed(t *testing.T) {
	_, err := parseCLI([]string{"https://example.com/docs", "--version"})
	if err != errVersion {
		t.Fatalf("expected errVersion when --version appears after seed, got %v", err)
	}
}
```

- [ ] **Step 2: Attempt build to confirm it fails**

```bash
go build ./cmd/crawl-url/... 2>&1 | grep -v "golang.org/x/text"
```

Expected: `errVersion undefined` (plus pre-existing charset errors; ignore those)

- [ ] **Step 3: Add version constant and errVersion to flags.go**

In `cmd/crawl-url/flags.go`, after `var errHelp = fmt.Errorf("help")`:

```go
var errHelp = fmt.Errorf("help")
var errVersion = fmt.Errorf("version")

const version = "v0.2.0"
```

- [ ] **Step 4: Add --version/-v to early scan in parseCLI**

In `cmd/crawl-url/flags.go`, extend the early scan loop (the first `for _, arg := range args` that currently only checks for `--help`/`-h`):

```go
for _, arg := range args {
	if arg == "--help" || arg == "-h" {
		return cliOptions{}, errHelp
	}
	if arg == "--version" || arg == "-v" {
		return cliOptions{}, errVersion
	}
}
```

- [ ] **Step 5: Add --txt case to parseCLI switch**

In `cmd/crawl-url/flags.go`, in the `switch name` block, add after the `"debug"` case:

```go
case "txt":
	if hasValue {
		return cliOptions{}, fmt.Errorf("选项 --txt 不接受参数")
	}
	opts.cfg.Txt = true
```

- [ ] **Step 6: Update printUsage**

In `cmd/crawl-url/flags.go`, in `printUsage`, add two lines after the `--video` / `--vedio` lines:

```go
fmt.Fprintf(w, "  --txt                     抓取同前缀文本文件 URL（.md .txt .yaml .yml .json 等）\n")
```

And after `--output-dir`:

```go
fmt.Fprintf(w, "  --version, -v             打印版本号\n")
```

The full updated `printUsage` body:

```go
func printUsage() {
	w := os.Stderr
	fmt.Fprintf(w, "用法: crawl-url <url> [选项]\n\n")
	fmt.Fprintf(w, "从种子 URL 开始并发抓取同一文档前缀下的页面，将成功 URL 输出到标准输出。\n\n")
	fmt.Fprintf(w, "选项:\n")
	fmt.Fprintf(w, "  --url-limit <n>           最多入队 URL 数量（默认不限制）\n")
	fmt.Fprintf(w, "  --depth <n>               最大层数，种子为深度 0（默认不限制；1=仅种子）\n")
	fmt.Fprintf(w, "  --exclude-prefix <前缀>   排除 URL 前缀（可重复）\n")
	fmt.Fprintf(w, "  --image                   抓取同前缀图片 URL\n")
	fmt.Fprintf(w, "  --video                   抓取同前缀视频 URL\n")
	fmt.Fprintf(w, "  --vedio                   --video 的兼容别名\n")
	fmt.Fprintf(w, "  --txt                     抓取同前缀文本文件 URL（.md .txt .yaml .yml .json 等）\n")
	fmt.Fprintf(w, "  --per-timeout <秒>        单请求超时（默认 10）\n")
	fmt.Fprintf(w, "  --workers <n>             并发 worker 数（默认 6）\n")
	fmt.Fprintf(w, "  --debug                   调试日志到 stderr\n")
	fmt.Fprintf(w, "  --output-dir <目录>       保存 HTML 页面\n")
	fmt.Fprintf(w, "  --version, -v             打印版本号\n")
	fmt.Fprintf(w, "  --help, -h                显示此帮助\n")
	fmt.Fprintf(w, "\n示例:\n")
	fmt.Fprintf(w, "  crawl-url https://clerk.com/docs --url-limit 20 --workers 8\n")
}
```

- [ ] **Step 7: Handle errVersion in main.go**

In `cmd/crawl-url/main.go`, after the `if err == errHelp` block:

```go
func main() {
	opts, err := parseCLI(os.Args[1:])
	if err == errHelp {
		printUsage()
		os.Exit(0)
	}
	if err == errVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n\n", err)
		printUsage()
		os.Exit(2)
	}
	// ... rest unchanged
```

- [ ] **Step 8: Attempt build to confirm only pre-existing errors remain**

```bash
go build ./cmd/crawl-url/... 2>&1 | grep -v "golang.org/x/text"
```

Expected: no output

- [ ] **Step 9: Spot-check version flag manually (build binary first)**

```bash
go build -o /tmp/crawl-url-test ./cmd/crawl-url/ 2>/dev/null || echo "build failed (charset issue, expected)"
```

If the build succeeds despite the environment issue:
```bash
/tmp/crawl-url-test --version
/tmp/crawl-url-test -v
```
Expected output: `v0.2.0`

- [ ] **Step 10: Commit**

```bash
git add cmd/crawl-url/flags.go cmd/crawl-url/main.go cmd/crawl-url/flags_test.go
git commit -m "feat(cli): add --txt and --version/-v flags"
```
