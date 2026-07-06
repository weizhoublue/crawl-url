package crawl

import (
	"fmt"
	"os"
	"sync"

	"crawl-url/internal/urlnorm"
)

// Kind is the task type queued for workers.
type Kind int

const (
	Page Kind = iota
	Media
)

func (k Kind) String() string {
	switch k {
	case Page:
		return "page"
	case Media:
		return "media"
	default:
		return "unknown"
	}
}

type task struct {
	url   string
	kind  Kind
	depth int // seed = 0
}

// Config holds crawl parameters.
type Config struct {
	Prefix          string
	URLLimit        *int
	MaxDepth        *int // nil = unlimited; N allows depths 0..N-1
	Image           bool
	Video           bool
	Txt             bool
	PerTimeout      float64
	ExcludePrefixes []string
	Debug           bool
	OutputDir       string
	Workers         int
}

// State is shared crawl state protected by a mutex.
type State struct {
	cfg Config

	mu            sync.Mutex
	known         map[string]struct{}
	acceptedCount int
	successSet    map[string]struct{}
	successes     []string

	tasks chan task
	wg    sync.WaitGroup
}

func newState(cfg Config) *State {
	return &State{
		cfg:        cfg,
		known:      make(map[string]struct{}),
		successSet: make(map[string]struct{}),
		// Buffered so workers can enqueue while others fetch (avoids deadlock on unbuffered send).
		tasks:      make(chan task, 4096),
	}
}

func (s *State) debug(msg string) {
	if s.cfg.Debug {
		fmt.Fprintf(os.Stderr, "[调试] %s\n", msg)
	}
}

// Reserve enqueues url if it passes scope checks and limits.
func (s *State) Reserve(rawURL string, kind Kind, depth int) bool {
	s.mu.Lock()
	if _, ok := s.known[rawURL]; ok {
		s.debug(fmt.Sprintf("跳过已处理 URL：%s", rawURL))
		s.mu.Unlock()
		return false
	}
	if !urlnorm.InPrefix(rawURL, s.cfg.Prefix) {
		s.debug(fmt.Sprintf("跳过前缀范围外 URL：%s", rawURL))
		s.mu.Unlock()
		return false
	}
	if urlnorm.MatchesPrefix(rawURL, s.cfg.ExcludePrefixes) {
		s.debug(fmt.Sprintf("跳过命中排除前缀的 URL：%s", rawURL))
		s.mu.Unlock()
		return false
	}
	if s.cfg.MaxDepth != nil && depth >= *s.cfg.MaxDepth {
		s.debug(fmt.Sprintf("达到深度上限，跳过：%s（深度=%d）", rawURL, depth))
		s.mu.Unlock()
		return false
	}
	if s.cfg.URLLimit != nil && s.acceptedCount >= *s.cfg.URLLimit {
		s.debug(fmt.Sprintf("达到 URL 数量上限，跳过：%s", rawURL))
		s.mu.Unlock()
		return false
	}
	s.known[rawURL] = struct{}{}
	s.acceptedCount++
	s.mu.Unlock()

	s.wg.Add(1)
	s.tasks <- task{url: rawURL, kind: kind, depth: depth}
	s.debug(fmt.Sprintf("加入队列（%s，深度=%d）：%s", kind, depth, rawURL))
	return true
}

func (s *State) canExpand(depth int) bool {
	if s.cfg.MaxDepth == nil {
		return true
	}
	return depth < *s.cfg.MaxDepth-1
}

func (s *State) noteAlias(rawURL string) {
	s.mu.Lock()
	s.known[rawURL] = struct{}{}
	s.mu.Unlock()
}

func (s *State) recordOutput(rawURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.successSet[rawURL]; ok {
		return
	}
	s.successSet[rawURL] = struct{}{}
	s.successes = append(s.successes, rawURL)
	printURL(rawURL)
}

// printURL writes a successful URL to stdout (overridable in tests).
var printURL = func(url string) { fmt.Println(url) }

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

// ValidateWorkers returns an error if workers < 1.
func ValidateWorkers(workers int) error {
	if workers < 1 {
		return fmt.Errorf("--workers 必须 >= 1，当前为 %d", workers)
	}
	return nil
}

// ValidateDepth returns an error if depth < 1.
func ValidateDepth(depth int) error {
	if depth < 1 {
		return fmt.Errorf("--depth 必须 >= 1，当前为 %d", depth)
	}
	return nil
}
