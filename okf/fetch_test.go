package okf

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchPageExtractsTitleLinksAndMarkdown(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title>  Hello   World </title></head>` +
			`<body><h1>Hi</h1><a href="/other">Other</a><a href="https://example.com/x#frag">Ext</a></body></html>`))
	}))
	defer srv.Close()

	page, err := FetchPage(srv.URL)
	if err != nil {
		t.Fatalf("FetchPage() error = %v", err)
	}
	if page.Title != "Hello World" {
		t.Errorf("Title = %q, want %q", page.Title, "Hello World")
	}
	if !strings.Contains(page.Markdown, "Hi") {
		t.Errorf("Markdown = %q, want to contain heading text", page.Markdown)
	}
	wantLink := srv.URL + "/other"
	found := false
	for _, l := range page.Links {
		if l == wantLink {
			found = true
		}
		if strings.Contains(l, "#frag") {
			t.Errorf("link %q should have fragment stripped", l)
		}
	}
	if !found {
		t.Errorf("Links = %v, want to contain %q", page.Links, wantLink)
	}
}

func TestFetchPageRejectsNonHTMLContentType(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	if _, err := FetchPage(srv.URL); err == nil {
		t.Fatal("FetchPage() error = nil, want non-HTML error")
	}
}

func TestFetchPageWrapsTransportErrors(t *testing.T) {
	t.Parallel()

	if _, err := FetchPage("http://127.0.0.1:1"); err == nil {
		t.Fatal("FetchPage() error = nil, want connection error")
	}
}
