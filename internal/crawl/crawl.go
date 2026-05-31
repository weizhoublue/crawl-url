package crawl

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"crawl-url/internal/htmlparse"
	"crawl-url/internal/output"
	"crawl-url/internal/urlnorm"
)

// Crawl runs a concurrent crawl from seed using cfg.
func Crawl(seed string, cfg Config) error {
	if err := ValidateWorkers(cfg.Workers); err != nil {
		return err
	}
	if cfg.MaxDepth != nil {
		if err := ValidateDepth(*cfg.MaxDepth); err != nil {
			return err
		}
	}
	if cfg.Workers == 0 {
		cfg.Workers = 6
	}

	prefix := urlnorm.NormalizePrefix(seed)
	seed = urlnorm.NormalizeURL(seed)
	excludes := make([]string, 0, len(cfg.ExcludePrefixes))
	for _, p := range cfg.ExcludePrefixes {
		excludes = append(excludes, urlnorm.NormalizePrefix(p))
	}
	outDir := cfg.OutputDir
	if outDir != "" {
		abs, err := filepath.Abs(outDir)
		if err != nil {
			return err
		}
		outDir = abs
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
		cfg.OutputDir = outDir
	}
	cfg.Prefix = prefix
	cfg.ExcludePrefixes = excludes

	state := newState(cfg)
	client := newHTTPClient(time.Duration(cfg.PerTimeout * float64(time.Second)))

	for i := 0; i < cfg.Workers; i++ {
		go state.worker(client)
	}

	if !state.Reserve(seed, Page, 0) {
		state.debug(fmt.Sprintf("种子 URL 未加入队列：%s", seed))
		close(state.tasks)
		return nil
	}

	limit := "不限制"
	if cfg.URLLimit != nil {
		limit = fmt.Sprintf("%d", *cfg.URLLimit)
	}
	depthLimit := "不限制"
	if cfg.MaxDepth != nil {
		depthLimit = fmt.Sprintf("%d 层（深度 0..%d）", *cfg.MaxDepth, *cfg.MaxDepth-1)
	}
	state.debug(fmt.Sprintf(
		"开始抓取：种子=%s，范围前缀=%s，工作线程=%d，URL 上限=%s，深度上限=%s",
		seed, prefix, cfg.Workers, limit, depthLimit,
	))

	state.wg.Wait()
	close(state.tasks)
	state.debug(fmt.Sprintf("抓取完成，成功 URL 数：%d", len(state.successes)))
	return nil
}

func (s *State) worker(client *http.Client) {
	for item := range s.tasks {
		s.runTask(client, item)
		s.wg.Done()
	}
}

func (s *State) runTask(client *http.Client, item task) {
	current := item.url
	s.debug(fmt.Sprintf("开始处理（%s）：%s", item.kind, current))

	var err error
	switch item.kind {
	case Page:
		err = s.processPage(client, item)
	case Media:
		err = s.processMedia(client, current)
	}
	if err != nil {
		reportFailure(current, err)
	}
}

func (s *State) processPage(client *http.Client, item task) error {
	current := item.url
	htmlText, finalRaw, err := fetchHTML(client, current)
	if err != nil {
		return err
	}
	final := urlnorm.NormalizeURL(finalRaw)
	s.noteAlias(final)

	if !urlnorm.InPrefix(final, s.cfg.Prefix) {
		return fmt.Errorf("重定向到了前缀范围外：%s", final)
	}
	if urlnorm.MatchesPrefix(final, s.cfg.ExcludePrefixes) {
		s.debug(fmt.Sprintf("重定向命中排除前缀，跳过：%s -> %s", current, final))
		return nil
	}
	if htmlText == nil {
		s.recordOutput(final)
		s.debug(fmt.Sprintf("记录非 HTML 结果：%s", final))
		return nil
	}
	if s.cfg.OutputDir != "" {
		if err := output.SaveHTML(s.cfg.OutputDir, final, *htmlText); err != nil {
			return err
		}
	}
	s.recordOutput(final)
	s.debug(fmt.Sprintf("记录结果：%s", final))

	if !s.canExpand(item.depth) {
		s.debug(fmt.Sprintf("已达最大深度，不扩展链接：%s（深度=%d）", final, item.depth))
		return nil
	}

	parsed := htmlparse.ExtractLinks(*htmlText)
	for _, raw := range parsed.Links {
		candidate := urlnorm.Join(final, raw)
		s.handleDiscovered(candidate, true, item.depth)
	}
	for _, raw := range parsed.MediaLinks {
		candidate := urlnorm.Join(final, raw)
		s.handleDiscovered(candidate, false, item.depth)
	}
	return nil
}

func (s *State) processMedia(client *http.Client, current string) error {
	finalRaw, err := accessMedia(client, current)
	if err != nil {
		return err
	}
	final := urlnorm.NormalizeURL(finalRaw)
	s.noteAlias(final)

	if !urlnorm.InPrefix(final, s.cfg.Prefix) {
		return fmt.Errorf("重定向到了前缀范围外：%s", final)
	}
	if urlnorm.MatchesPrefix(final, s.cfg.ExcludePrefixes) {
		s.debug(fmt.Sprintf("重定向命中排除前缀，跳过：%s -> %s", current, final))
		return nil
	}
	s.recordOutput(final)
	s.debug(fmt.Sprintf("记录结果：%s", final))
	return nil
}
