package shuba69

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

type Fetcher struct {
	client *http.Client
	cookie string
	ua     string
}

func NewFetcher(cookie string, userAgent string) *Fetcher {
	c := strings.TrimSpace(cookie)
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		ua = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	}
	return &Fetcher{
		client: &http.Client{Timeout: 30 * time.Second},
		cookie: c,
		ua:     ua,
	}
}

// CookieFromPlaywrightStorageState extracts a Cookie header string for a given targetURL
// from a Playwright storage_state JSON file.
//
// Note: This is not a bypass mechanism. It simply reuses a user-authenticated browser session.
func CookieFromPlaywrightStorageState(storageStatePath string, targetURL string) (string, error) {
	storageStatePath = strings.TrimSpace(storageStatePath)
	if storageStatePath == "" {
		return "", fmt.Errorf("storageStatePath is empty")
	}
	b, err := os.ReadFile(storageStatePath)
	if err != nil {
		return "", err
	}

	var ss struct {
		Cookies []struct {
			Name   string `json:"name"`
			Value  string `json:"value"`
			Domain string `json:"domain"`
			Path   string `json:"path"`
		} `json:"cookies"`
	}
	if err := json.Unmarshal(b, &ss); err != nil {
		return "", fmt.Errorf("parse storage_state: %w", err)
	}

	u, err := url.Parse(strings.TrimSpace(targetURL))
	if err != nil {
		return "", fmt.Errorf("parse target url: %w", err)
	}
	host := strings.ToLower(u.Host)
	path := u.Path
	if path == "" {
		path = "/"
	}

	// build cookie header from cookies matching domain and path
	var parts []string
	seen := map[string]bool{}
	for _, c := range ss.Cookies {
		name := strings.TrimSpace(c.Name)
		val := c.Value
		if name == "" || val == "" {
			continue
		}
		d := strings.ToLower(strings.TrimSpace(c.Domain))
		p := strings.TrimSpace(c.Path)
		if p == "" {
			p = "/"
		}

		// domain match: cookie domain may be ".69shuba.com" or "www.69shuba.com"
		domainOK := false
		if strings.HasPrefix(d, ".") {
			domainOK = host == strings.TrimPrefix(d, ".") || strings.HasSuffix(host, d)
		} else {
			domainOK = host == d || strings.HasSuffix(host, "."+d)
		}

		if !domainOK {
			continue
		}
		if !strings.HasPrefix(path, p) {
			continue
		}

		key := name
		if seen[key] {
			continue
		}
		seen[key] = true
		parts = append(parts, name+"="+val)
	}

	return strings.Join(parts, "; "), nil
}

func (f *Fetcher) Get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", f.ua)
	// Try to look like a normal interactive browser (avoid bot heuristics).
	// IMPORTANT: do NOT set Accept-Encoding; Go's transport will auto-negotiate gzip and transparently decode it.
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "ru,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	if strings.TrimSpace(f.cookie) != "" {
		req.Header.Set("Cookie", f.cookie)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("GET %s: status=%d body=%q", rawURL, resp.StatusCode, string(body))
	}

	reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		reader = resp.Body
	}
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(b), nil
}

