package fanqie

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

type Fetcher struct {
	client *http.Client
	cookie string
	ua     string
}

func NewFetcher() *Fetcher {
	return NewFetcherWithOptions("", "")
}

// NewFetcherWithOptions allows overriding Cookie/UA. If empty, generates its own.
func NewFetcherWithOptions(cookie string, userAgent string) *Fetcher {
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		ua = randomUA()
	}
	c := strings.TrimSpace(cookie)
	if c == "" {
		// fanqie doesn't require a specific cookie for public pages, but we can set a stable random cookie
		// to look like a normal browser session.
		c = "sessionid=" + randHex(16) + "; csrftoken=" + randHex(16)
	}

	return &Fetcher{
		client: &http.Client{Timeout: 25 * time.Second},
		cookie: c,
		ua:     ua,
	}
}

func (f *Fetcher) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", f.ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,ru;q=0.7")
	req.Header.Set("Cookie", f.cookie)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("GET %s: status=%d body=%q", url, resp.StatusCode, string(body))
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

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "deadbeef"
	}
	return hex.EncodeToString(b)
}

func randomUA() string {
	uas := []string{
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	}
	idx := 0
	if n, err := randInt(len(uas)); err == nil {
		idx = n
	}
	return uas[idx]
}

func randInt(max int) (int, error) {
	if max <= 1 {
		return 0, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

