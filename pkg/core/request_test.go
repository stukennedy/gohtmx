package core

import (
	"testing"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest("GET", "/test?foo=bar")

	if req.Method != "GET" {
		t.Errorf("expected Method GET, got %s", req.Method)
	}
	if req.URL != "/test?foo=bar" {
		t.Errorf("expected URL /test?foo=bar, got %s", req.URL)
	}
	if req.Headers != "{}" {
		t.Errorf("expected empty headers {}, got %s", req.Headers)
	}
}

func TestRequestHeaders(t *testing.T) {
	req := NewRequest("POST", "/api")

	// Test SetHeader
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("X-Custom", "value")

	// Test GetHeader
	if ct := req.GetHeader("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
	if custom := req.GetHeader("X-Custom"); custom != "value" {
		t.Errorf("expected X-Custom value, got %s", custom)
	}

	// Test GetHeaders
	headers := req.GetHeaders()
	if len(headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(headers))
	}
}

func TestRequestPath(t *testing.T) {
	tests := []struct {
		url      string
		wantPath string
		wantQuery string
	}{
		{"/test", "/test", ""},
		{"/test?foo=bar", "/test", "foo=bar"},
		{"/api/users?page=1&limit=10", "/api/users", "page=1&limit=10"},
		{"/?q=search", "/", "q=search"},
	}

	for _, tt := range tests {
		req := NewRequest("GET", tt.url)
		if path := req.Path(); path != tt.wantPath {
			t.Errorf("Path(%q) = %q, want %q", tt.url, path, tt.wantPath)
		}
		if query := req.Query(); query != tt.wantQuery {
			t.Errorf("Query(%q) = %q, want %q", tt.url, query, tt.wantQuery)
		}
	}
}

func TestRequestQueryValue(t *testing.T) {
	req := NewRequest("GET", "/search?q=hello&page=2&limit=10")

	if v := req.QueryValue("q"); v != "hello" {
		t.Errorf("QueryValue(q) = %q, want %q", v, "hello")
	}
	if v := req.QueryValue("page"); v != "2" {
		t.Errorf("QueryValue(page) = %q, want %q", v, "2")
	}
	if v := req.QueryValue("missing"); v != "" {
		t.Errorf("QueryValue(missing) = %q, want empty", v)
	}
}

func TestRequestIsHTMX(t *testing.T) {
	req := NewRequest("GET", "/")

	if req.IsHTMX() {
		t.Error("expected IsHTMX() = false for new request")
	}

	req.SetHeader("HX-Request", "true")
	if !req.IsHTMX() {
		t.Error("expected IsHTMX() = true after setting HX-Request header")
	}
}

func TestRequestBody(t *testing.T) {
	req := NewRequest("POST", "/api")
	req.Body = []byte(`{"name": "test"}`)

	if s := req.BodyString(); s != `{"name": "test"}` {
		t.Errorf("BodyString() = %q, want %q", s, `{"name": "test"}`)
	}
}
