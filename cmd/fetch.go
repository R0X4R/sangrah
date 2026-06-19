package cmd

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var httpClient *http.Client

var (
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	defaultAccept    = "application/javascript, */*"
)

// InitClient sets up the shared HTTP client with connection pooling, HTTP/2
// support, and per-host connection limits based on concurrency.
func InitClient(concurrency int) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: concurrency * 2,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	httpClient = &http.Client{
		Transport: transport,
		Timeout:   120 * time.Second,
	}
}

// FetchJS downloads JavaScript from a URL with automatic HTTPS prefixing
// and exponential-backoff retries (capped at 10s). Returns the raw bytes
// or an error if all retries are exhausted.
func FetchJS(rawURL string, opts *Options) ([]byte, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	var lastErr error
	for attempt := 0; attempt <= opts.Retries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 10*time.Second {
				backoff = 10 * time.Second
			}
			time.Sleep(backoff)
		}

		data, err := doFetch(rawURL, opts)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("after %d retries: %w", opts.Retries, lastErr)
}

// doFetch performs a single HTTP GET with custom headers, cookies, and a
// browser-like User-Agent. Returns an error on non-200 status or read failure.
func doFetch(rawURL string, opts *Options) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", defaultAccept)

	for _, h := range opts.Headers {
		if idx := strings.Index(h, ":"); idx > 0 {
			req.Header.Set(strings.TrimSpace(h[:idx]), strings.TrimSpace(h[idx+1:]))
		}
	}

	if opts.Cookie != "" {
		req.Header.Set("Cookie", opts.Cookie)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	return data, nil
}
