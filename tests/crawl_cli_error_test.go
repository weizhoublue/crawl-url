package tests

import (
	"strings"
	"testing"
)

func TestCrawlCLIErrorMissingSeed(t *testing.T) {
	_, stderr, exitCode, err := runCLI(t, []string{})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if exitCode != 2 {
		t.Errorf("expected exit code 2 for missing seed, got %d", exitCode)
	}
	if !strings.Contains(stderr, "缺少种子 URL") {
		t.Errorf("expected error message to contain '缺少种子 URL', got: %s", stderr)
	}
}

func TestCrawlCLIErrorShortFlag(t *testing.T) {
	_, stderr, exitCode, err := runCLI(t, []string{"http://localhost", "-workers", "4"})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if exitCode != 2 {
		t.Errorf("expected exit code 2 for short flag, got %d", exitCode)
	}
	if !strings.Contains(stderr, "请使用 -- 长选项") {
		t.Errorf("expected error message to contain '请使用 -- 长选项', got: %s", stderr)
	}
}

func TestCrawlCLIErrorUnknownFlag(t *testing.T) {
	_, stderr, exitCode, err := runCLI(t, []string{"http://localhost", "--foo"})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if exitCode != 2 {
		t.Errorf("expected exit code 2 for unknown flag, got %d", exitCode)
	}
	if !strings.Contains(stderr, "未知选项: --foo") {
		t.Errorf("expected error message to contain '未知选项: --foo', got: %s", stderr)
	}
}

func TestCrawlCLIErrorInvalidWorkers(t *testing.T) {
	_, stderr, exitCode, err := runCLI(t, []string{"http://localhost", "--workers", "-1"})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for invalid workers count, got %d", exitCode)
	}
	if !strings.Contains(stderr, "workers 必须 >= 1") {
		t.Errorf("expected error message to contain 'workers 必须 >= 1', got: %s", stderr)
	}
}

func TestCrawlCLIErrorInvalidDepth(t *testing.T) {
	_, stderr, exitCode, err := runCLI(t, []string{"http://localhost", "--depth", "0"})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for invalid depth, got %d", exitCode)
	}
	if !strings.Contains(stderr, "depth 必须 >= 1") {
		t.Errorf("expected error message to contain 'depth 必须 >= 1', got: %s", stderr)
	}
}
