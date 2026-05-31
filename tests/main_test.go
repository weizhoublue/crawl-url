package tests

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var binPath string

func TestMain(m *testing.M) {
	flag.Parse()

	// 1. 创建临时目录存放编译后的 CLI 二进制
	tmpDir, err := os.MkdirTemp("", "crawl-url-test-*")
	if err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath = filepath.Join(tmpDir, "crawl-url")

	// 2. 编译项目到临时目录
	cmd := exec.Command("go", "build", "-o", binPath, "../cmd/crawl-url")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to build crawl-url CLI: %v", err)
	}

	// 3. 执行测试套件
	code := m.Run()
	os.Exit(code)
}

// runCLI 运行编译好的二进制，并捕获输出与退出码
func runCLI(t *testing.T, args []string) (string, string, int, error) {
	t.Helper()
	cmd := exec.Command(binPath, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return stdout, stderr, -1, err
		}
	}

	return stdout, stderr, exitCode, nil
}
