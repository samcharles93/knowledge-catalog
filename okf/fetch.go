package okf

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"golang.org/x/net/html/charset"
)

const (
	fetchUserAgent   = "okf-reference-agent/0.1 (+https://github.com/samcharles93/knowledge-catalog)"
	fetchTimeout     = 10 * time.Second
	maxMarkdownBytes = 40 * 1024
)

var (
	titleRE = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	hrefRE  = regexp.MustCompile(`(?i)href\s*=\s*["']([^"'\s]+)["']`)
)

// Page is a fetched web page reduced to its title, markdown body, and the
// absolute, fragment-stripped links found on it.
type Page struct {
	URL      string
	Title    string
	Markdown string
	Links    []string
}

// FetchPage retrieves url, requires an HTML response, and converts it to a
// Page with extracted title, markdown body, and outbound links.
func FetchPage(rawURL string) (Page, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return Page{}, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	req.Header.Set("User-Agent", fetchUserAgent)
	req.Header.Set("Accept", "text/html,*/*;q=0.5")

	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return Page{}, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "html") {
		label := contentType
		if label == "" {
			label = "unknown"
		}
		return Page{}, fmt.Errorf("fetch %s: non-HTML content-type: %s", rawURL, label)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Page{}, fmt.Errorf("fetch %s: %w", rawURL, err)
	}

	finalURL := rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	html := decodeHTML(bodyBytes, contentType)
	title := extractTitle(html)
	links := extractPageLinks(html, finalURL)

	converter := md.NewConverter("", true, &md.Options{HeadingStyle: "atx"})
	markdown, err := converter.ConvertString(html)
	if err != nil {
		markdown = ""
	}
	markdown = truncateMarkdown(strings.TrimSpace(markdown), maxMarkdownBytes)

	return Page{URL: finalURL, Title: title, Markdown: markdown, Links: links}, nil
}

// decodeHTML normalizes body bytes to UTF-8 using the response's declared
// (or sniffed) charset, self-healing to the raw bytes if detection fails.
func decodeHTML(body []byte, contentType string) string {
	reader, err := charset.NewReader(bytes.NewReader(body), contentType)
	if err != nil {
		return string(body)
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return string(body)
	}
	return string(decoded)
}

func extractTitle(html string) string {
	m := titleRE.FindStringSubmatch(html)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(strings.Join(strings.Fields(m[1]), " "))
}

// extractPageLinks resolves every href on the page to an absolute URL
// relative to baseURL, drops non-http(s) schemes (mailto:, javascript:,
// ...), strips fragments, and de-duplicates while preserving order.
func extractPageLinks(html, baseURL string) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	var out []string
	seen := map[string]bool{}
	for _, m := range hrefRE.FindAllStringSubmatch(html, -1) {
		href := strings.TrimSpace(m[1])
		if href == "" {
			continue
		}
		parsed, err := url.Parse(href)
		if err != nil {
			continue
		}
		if scheme := strings.ToLower(parsed.Scheme); scheme != "" && scheme != "http" && scheme != "https" {
			continue
		}

		absolute := base.ResolveReference(parsed)
		absolute.Fragment = ""
		absStr := absolute.String()
		if seen[absStr] {
			continue
		}
		seen[absStr] = true
		out = append(out, absStr)
	}
	return out
}

func truncateMarkdown(text string, maxBytes int) string {
	if len(text) <= maxBytes {
		return text
	}
	truncated := text[:maxBytes]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated + "\n\n[...truncated...]"
}
