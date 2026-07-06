package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"crawl-url/internal/crawl"
)

var errHelp = fmt.Errorf("help")
var errVersion = fmt.Errorf("version")

const version = "v0.2.0"

type cliOptions struct {
	seed string
	cfg  crawl.Config
}

func parseCLI(args []string) (cliOptions, error) {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cliOptions{}, errHelp
		}
		if arg == "--version" || arg == "-v" {
			return cliOptions{}, errVersion
		}
	}

	seed, flagArgs := splitSeedAndFlags(args)
	if seed == "" {
		return cliOptions{}, fmt.Errorf("缺少种子 URL")
	}

	opts := cliOptions{seed: seed}
	var urlLimit int
	var hasLimit bool
	var maxDepth int
	var hasDepth bool

	for i := 0; i < len(flagArgs); i++ {
		arg := flagArgs[i]
		if arg == "-h" || arg == "--help" {
			return cliOptions{}, errHelp
		}
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			return cliOptions{}, fmt.Errorf("请使用 -- 长选项（例如 --url-limit），不支持: %s", arg)
		}
		if !strings.HasPrefix(arg, "--") {
			return cliOptions{}, fmt.Errorf("无法识别的参数: %s", arg)
		}

		name, value, hasValue := splitLongFlag(arg)
		if !hasValue && flagTakesValue(name) {
			i++
			if i >= len(flagArgs) {
				return cliOptions{}, fmt.Errorf("选项 --%s 需要参数", name)
			}
			value = flagArgs[i]
			hasValue = true
		}

		switch name {
		case "url-limit":
			if !hasValue {
				return cliOptions{}, fmt.Errorf("选项 --url-limit 需要参数")
			}
			n, err := strconv.Atoi(value)
			if err != nil {
				return cliOptions{}, fmt.Errorf("--url-limit 需要整数: %w", err)
			}
			urlLimit = n
			hasLimit = true
		case "exclude-prefix":
			if !hasValue {
				return cliOptions{}, fmt.Errorf("选项 --exclude-prefix 需要参数")
			}
			opts.cfg.ExcludePrefixes = append(opts.cfg.ExcludePrefixes, value)
		case "image":
			if hasValue {
				return cliOptions{}, fmt.Errorf("选项 --image 不接受参数")
			}
			opts.cfg.Image = true
		case "video", "vedio":
			if hasValue {
				return cliOptions{}, fmt.Errorf("选项 --%s 不接受参数", name)
			}
			opts.cfg.Video = true
		case "per-timeout":
			if !hasValue {
				return cliOptions{}, fmt.Errorf("选项 --per-timeout 需要参数")
			}
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return cliOptions{}, fmt.Errorf("--per-timeout 需要数字: %w", err)
			}
			opts.cfg.PerTimeout = f
		case "depth":
			if !hasValue {
				return cliOptions{}, fmt.Errorf("选项 --depth 需要参数")
			}
			n, err := strconv.Atoi(value)
			if err != nil {
				return cliOptions{}, fmt.Errorf("--depth 需要整数: %w", err)
			}
			maxDepth = n
			hasDepth = true
		case "workers":
			if !hasValue {
				return cliOptions{}, fmt.Errorf("选项 --workers 需要参数")
			}
			n, err := strconv.Atoi(value)
			if err != nil {
				return cliOptions{}, fmt.Errorf("--workers 需要整数: %w", err)
			}
			opts.cfg.Workers = n
		case "debug":
			if hasValue {
				return cliOptions{}, fmt.Errorf("选项 --debug 不接受参数")
			}
			opts.cfg.Debug = true
		case "txt":
			if hasValue {
				return cliOptions{}, fmt.Errorf("选项 --txt 不接受参数")
			}
			opts.cfg.Txt = true
		case "output-dir":
			if !hasValue {
				return cliOptions{}, fmt.Errorf("选项 --output-dir 需要参数")
			}
			opts.cfg.OutputDir = value
		default:
			return cliOptions{}, fmt.Errorf("未知选项: --%s", name)
		}
	}

	if hasLimit {
		opts.cfg.URLLimit = &urlLimit
	}
	if hasDepth {
		opts.cfg.MaxDepth = &maxDepth
	}
	if opts.cfg.Workers == 0 {
		opts.cfg.Workers = 6
	}
	if opts.cfg.PerTimeout == 0 {
		opts.cfg.PerTimeout = 10
	}
	return opts, nil
}

func splitLongFlag(arg string) (name, value string, hasValue bool) {
	body := strings.TrimPrefix(arg, "--")
	if idx := strings.IndexByte(body, '='); idx >= 0 {
		return body[:idx], body[idx+1:], true
	}
	return body, "", false
}

func flagTakesValue(name string) bool {
	switch name {
	case "url-limit", "exclude-prefix", "per-timeout", "workers", "output-dir", "depth":
		return true
	default:
		return false
	}
}

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
