package parserclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New() *Client {
	base := strings.TrimSpace(os.Getenv("PARSER_SERVICE_URL"))
	if base == "" {
		// docker-compose service name
		base = "http://parser-service:8000"
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		// Parsing can take a while (Playwright navigations). Do not kill the request
		// just because headers aren't sent quickly.
		ResponseHeaderTimeout: 0,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &Client{
		baseURL: strings.TrimRight(base, "/"),
		// IMPORTANT: do not set a short client timeout here.
		// Parser-service may take a long time for big books (hundreds/thousands of chapters).
		// We rely on the caller's ctx timeout/cancellation for safety.
		http: &http.Client{Timeout: 0, Transport: transport},
	}
}

type ParseRequest struct {
	URL              string `json:"url"`
	Site             string `json:"site,omitempty"` // "101kks" | "69shuba"
	ChaptersLimit    int    `json:"chapters_limit,omitempty"`
	UserAgent        string `json:"user_agent,omitempty"`
	Referer          string `json:"referer,omitempty"`
	CookieHeader     string `json:"cookie_header,omitempty"`
	StorageStatePath string `json:"storage_state_path,omitempty"` // path in parser-service container, e.g. /data/101kks_storage.json
	NavigationTimeoutMS int  `json:"navigation_timeout_ms,omitempty"` // Playwright navigation timeout override (ms)
	Humanize         bool   `json:"humanize,omitempty"`
	Locale           string `json:"locale,omitempty"`
	TimezoneID       string `json:"timezone_id,omitempty"`
	ViewportWidth    int    `json:"viewport_width,omitempty"`
	ViewportHeight   int    `json:"viewport_height,omitempty"`
	HumanDelayMSMin  int    `json:"human_delay_ms_min,omitempty"`
	HumanDelayMSMax  int    `json:"human_delay_ms_max,omitempty"`
	CloudflareWaitMS int    `json:"cloudflare_wait_ms,omitempty"`
}

type ChapterRef struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Number *int   `json:"number,omitempty"`
}

type Book struct {
	Title      string       `json:"title"`
	CoverURL   string       `json:"cover_url"`
	Description string      `json:"description"`
	Author     string       `json:"author"`
	Category   string       `json:"category"`
	Tags       []string     `json:"tags"`
	CatalogURL string       `json:"catalog_url"`
	Chapters   []ChapterRef `json:"chapters"`
}

type Chapter struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ParseResponse struct {
	Site     string                 `json:"site"`
	Book     Book                   `json:"book"`
	Chapters []Chapter               `json:"chapters"`
	Debug    map[string]interface{} `json:"debug"`
}

func (c *Client) Parse(ctx context.Context, req ParseRequest) (*ParseResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/parse", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("parser-service: status=%d body=%q", resp.StatusCode, string(b))
	}

	var out ParseResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

