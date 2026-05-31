package crawl

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html/charset"
)

const userAgent = "docs-crawler/1.0"

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
}

func fetchHTML(client *http.Client, rawURL string) (htmlText *string, finalURL string, err error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", &httpError{code: resp.StatusCode}
	}

	final := resp.Request.URL.String()
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		mediaType = resp.Header.Get("Content-Type")
	}
	if !strings.Contains(strings.ToLower(mediaType), "html") {
		return nil, final, nil
	}

	reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		reader = resp.Body
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}
	text := string(body)
	return &text, final, nil
}

func accessMedia(client *http.Client, rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return "", &httpError{code: resp.StatusCode}
	}
	return resp.Request.URL.String(), nil
}

type httpError struct {
	code int
}

func (e *httpError) Error() string {
	return fmt.Sprintf("http status %d", e.code)
}

func describeFailure(err error) string {
	var he *httpError
	if errors.As(err, &he) {
		return fmt.Sprintf("HTTP 错误，状态码：%d", he.code)
	}
	var uerr *url.Error
	if errors.As(err, &uerr) {
		if uerr.Timeout() || errors.Is(uerr.Err, os.ErrDeadlineExceeded) {
			return "请求超时"
		}
		var nerr net.Error
		if errors.As(uerr.Err, &nerr) && nerr.Timeout() {
			return "请求超时"
		}
		if errors.Is(uerr.Err, syscall.ETIMEDOUT) {
			return "请求超时"
		}
		return "网络请求失败"
	}
	if isUnicodeError(err) {
		return "响应内容解码失败"
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return "系统调用失败"
	}
	if errors.Is(err, os.ErrInvalid) || errors.Is(err, os.ErrPermission) {
		return "系统调用失败"
	}
	if _, ok := err.(*os.PathError); ok {
		return "系统调用失败"
	}
	if errors.Is(err, os.ErrNotExist) {
		return "系统调用失败"
	}
	if isOSError(err) {
		return "系统调用失败"
	}
	return err.Error()
}

func isUnicodeError(err error) bool {
	// UTF-8 decode issues surface as string replacement in Go; treat explicit types.
	type unicodeError interface {
		error
	}
	return strings.Contains(err.Error(), "invalid UTF-8") ||
		strings.Contains(err.Error(), "encoding")
}

func isOSError(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	return errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, os.ErrClosed)
}

func reportFailure(rawURL string, err error) {
	fmt.Fprintf(os.Stderr, "[失败] %s\t%s\n", rawURL, describeFailure(err))
}
