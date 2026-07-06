package urlnorm

import (
	"net/url"
	"path"
	"strings"
)

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

func extSet(exts ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		m[e] = struct{}{}
	}
	return m
}

// NormalizeURL unifies scheme, host, and path for deduplication (fragment dropped).
func NormalizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""

	p := u.Path
	if p == "" {
		p = "/"
	}
	trailingSlash := strings.HasSuffix(p, "/")
	p = path.Clean(p)
	if p == "." {
		p = "/"
	}
	if trailingSlash && !strings.HasSuffix(p, "/") {
		p += "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u.Path = p
	return u.String()
}

// NormalizePrefix returns the crawl scope prefix derived from a URL.
func NormalizePrefix(raw string) string {
	normalized := NormalizeURL(raw)
	u, err := url.Parse(normalized)
	if err != nil {
		return normalized
	}
	p := u.Path
	if p == "" {
		p = "/"
	}
	last := p
	if i := strings.LastIndex(p, "/"); i >= 0 {
		last = p[i+1:]
	}
	if !strings.HasSuffix(p, "/") && !strings.Contains(last, ".") {
		p += "/"
	}
	u.Path = p
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// InPrefix reports whether url is under prefix (same scheme and host).
func InPrefix(rawURL, prefix string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	p, err := url.Parse(prefix)
	if err != nil {
		return false
	}
	if strings.ToLower(u.Scheme) != strings.ToLower(p.Scheme) {
		return false
	}
	if strings.ToLower(u.Host) != strings.ToLower(p.Host) {
		return false
	}
	up, pp := u.Path, p.Path
	if strings.HasSuffix(pp, "/") {
		if strings.HasPrefix(up, pp) {
			return true
		}
		// Prefix /docs/ must also match the directory URL /docs (no trailing slash).
		if base := strings.TrimSuffix(pp, "/"); base != "" && up == base {
			return true
		}
		return false
	}
	return up == pp || strings.HasPrefix(up, pp+"/")
}

// MatchesPrefix reports whether url matches any exclude prefix.
func MatchesPrefix(rawURL string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if InPrefix(rawURL, prefix) {
			return true
		}
	}
	return false
}

// HasExtension reports whether the URL path ends with one of the extensions.
func HasExtension(rawURL string, exts map[string]struct{}) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	ext := strings.ToLower(path.Ext(u.Path))
	_, ok := exts[ext]
	return ok
}

// Join resolves ref against base and normalizes.
func Join(base, ref string) string {
	refURL, err := url.Parse(ref)
	if err != nil {
		return NormalizeURL(ref)
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return NormalizeURL(ref)
	}
	return NormalizeURL(baseURL.ResolveReference(refURL).String())
}
