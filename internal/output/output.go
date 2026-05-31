package output

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// BuildOutputPath mirrors Python build_output_path.
func BuildOutputPath(outputDir, rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return filepath.Join(outputDir, "invalid.html")
	}
	host := url.PathEscape(strings.ToLower(u.Host))
	rawPath := u.Path
	if rawPath == "" {
		rawPath = "/"
	}
	if strings.HasSuffix(rawPath, "/") {
		rawPath += "index"
	}
	segments := strings.Split(rawPath, "/")
	var pathSegments []string
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		escaped := url.PathEscape(seg)
		if escaped == "" {
			escaped = "_"
		}
		pathSegments = append(pathSegments, escaped)
	}
	if len(pathSegments) == 0 {
		pathSegments = []string{"index"}
	}
	filename := pathSegments[len(pathSegments)-1]
	sum := sha256.Sum256([]byte(rawURL))
	digest := hex.EncodeToString(sum[:])[:12]
	if u.RawQuery != "" {
		filename += "__query"
	}
	filename = fmt.Sprintf("%s__%s.html", filename, digest)
	parts := append([]string{outputDir, host}, pathSegments[:len(pathSegments)-1]...)
	parts = append(parts, filename)
	return filepath.Join(parts...)
}

// SaveHTML writes htmlText to the path derived from url.
func SaveHTML(outputDir, rawURL, htmlText string) error {
	p := BuildOutputPath(outputDir, rawURL)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(htmlText), 0o644)
}
