# Design: --txt and --version CLI flags

Date: 2026-07-06

## Summary

Add two new CLI options to `crawl-url`:

1. `--txt` — opt-in flag to capture text file URLs (`.md`, `.txt`, `.yaml`, etc.) during crawling. Off by default.
2. `--version` / `-v` — print the current version string and exit.

## Background

The crawler already supports `--image` and `--video` flags that opt-in to capturing media URLs matching known extensions. The `--txt` flag follows the same pattern for text/document file extensions. When `--txt` is off, text file URLs are silently dropped (consistent with image/video behavior). When on, they are verified via GET and recorded in output.

## Scope

Files changed:

- `internal/urlnorm/urlnorm.go` — add `TextExtensions`
- `internal/crawl/state.go` — add `Txt bool` to `Config`; extend `handleDiscovered`
- `cmd/crawl-url/flags.go` — parse `--txt`, `--version`, `-v`; add `version` constant; update usage
- `cmd/crawl-url/main.go` — handle `errVersion` sentinel
- `cmd/crawl-url/flags_test.go` — tests for new flags
- `internal/crawl/crawl_test.go` — tests for `handleDiscovered` with text extensions

## Design

### 1. TextExtensions (urlnorm)

Add a new exported variable alongside `ImageExtensions` and `VideoExtensions`:

```go
TextExtensions = extSet(
    ".md", ".mdx", ".txt", ".yaml", ".yml",
    ".json", ".xml", ".csv", ".rst", ".toml",
)
```

### 2. Config.Txt and handleDiscovered (crawl/state)

Add field to `Config`:

```go
Txt bool
```

Extend `handleDiscovered` after the Video block, before the final `s.Reserve(candidate, Page, childDepth)`:

```go
if urlnorm.HasExtension(candidate, urlnorm.TextExtensions) {
    if s.cfg.Txt {
        s.Reserve(candidate, Media, childDepth)
    }
    return
}
```

Behavior:
- `--txt` off → text-extension URLs are dropped silently (same as `--image` off for images)
- `--txt` on → queued as `Media`; processed by `accessMedia` (GET, discard body, record URL on success)

### 3. CLI (flags.go / main.go)

**Version constant:**

```go
const version = "v0.2.0"
```

**New sentinel:**

```go
var errVersion = fmt.Errorf("version")
```

**parseCLI** — detect `--version` / `-v` at the same early pass as `--help`:

```go
if arg == "--version" || arg == "-v" {
    return cliOptions{}, errVersion
}
```

**main.go** — handle the new sentinel after `errHelp`:

```go
if err == errVersion {
    fmt.Println(version)
    os.Exit(0)
}
```

**flagTakesValue** — no change needed (`--txt` is a boolean switch).

**parseCLI switch** — new case:

```go
case "txt":
    if hasValue {
        return cliOptions{}, fmt.Errorf("选项 --txt 不接受参数")
    }
    opts.cfg.Txt = true
```

**printUsage** — add two lines:

```
  --txt                     抓取同前缀文本文件 URL（.md .txt .yaml .yml .json 等）
  --version, -v             打印版本号
```

### 4. Error handling

- `--txt=value` → error "选项 --txt 不接受参数"
- `--version=value` / `-v=value` → version detection happens at the early scan (before flag parsing), so no `=value` form is possible at that stage; if it somehow reaches the switch, treat as unknown option (acceptable edge case)

### 5. Tests

**flags_test.go:**
- `--txt` sets `cfg.Txt = true`
- `--txt=foo` returns error
- `--version` returns `errVersion`
- `-v` returns `errVersion`

**crawl_test.go:**
- `handleDiscovered` with a `.md` URL and `cfg.Txt = false` → not queued
- `handleDiscovered` with a `.md` URL and `cfg.Txt = true` → queued as Media
- Same for `.yaml`, `.json` (spot-check two other extensions)

## Non-goals

- Custom extension list via `--txt-ext` (YAGNI)
- Parsing links out of text file content
- Dynamic version from git tags or build flags (hardcoded constant is sufficient for now)
