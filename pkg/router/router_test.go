package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("expected non-nil router")
	}
	if r.mux == nil {
		t.Fatal("expected non-nil mux")
	}
}

func TestGET(t *testing.T) {
	r := New()
	r.GET("/hello", func(ctx *Context) (string, error) {
		return "<div>Hello World</div>", nil
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "<div>Hello World</div>" {
		t.Errorf("expected body '<div>Hello World</div>', got %q", w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/html; charset=utf-8', got %q", ct)
	}
}

func TestPOST(t *testing.T) {
	r := New()
	r.POST("/submit", func(ctx *Context) (string, error) {
		name := ctx.FormValue("name")
		return "<span>Hello " + name + "</span>", nil
	})

	body := strings.NewReader("name=World")
	req := httptest.NewRequest("POST", "/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "<span>Hello World</span>" {
		t.Errorf("expected body '<span>Hello World</span>', got %q", w.Body.String())
	}
}

func TestPUT(t *testing.T) {
	r := New()
	r.PUT("/update/{id}", func(ctx *Context) (string, error) {
		id := ctx.Param("id")
		return "<div>Updated " + id + "</div>", nil
	})

	req := httptest.NewRequest("PUT", "/update/123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Updated 123") {
		t.Errorf("expected body to contain 'Updated 123', got %q", w.Body.String())
	}
}

func TestDELETE(t *testing.T) {
	r := New()
	r.DELETE("/items/{id}", func(ctx *Context) (string, error) {
		return "", nil // Empty response for delete
	})

	req := httptest.NewRequest("DELETE", "/items/456", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestPATCH(t *testing.T) {
	r := New()
	r.PATCH("/partial/{id}", func(ctx *Context) (string, error) {
		id := ctx.Param("id")
		return "<div>Patched " + id + "</div>", nil
	})

	req := httptest.NewRequest("PATCH", "/partial/789", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestURLParams(t *testing.T) {
	r := New()
	r.GET("/users/{userID}/posts/{postID}", func(ctx *Context) (string, error) {
		userID := ctx.Param("userID")
		postID := ctx.Param("postID")
		return "<div>User: " + userID + ", Post: " + postID + "</div>", nil
	})

	req := httptest.NewRequest("GET", "/users/42/posts/99", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	expected := "<div>User: 42, Post: 99</div>"
	if w.Body.String() != expected {
		t.Errorf("expected %q, got %q", expected, w.Body.String())
	}
}

func TestQueryParams(t *testing.T) {
	r := New()
	r.GET("/search", func(ctx *Context) (string, error) {
		q := ctx.Query("q")
		page := ctx.Query("page")
		return "<div>Search: " + q + ", Page: " + page + "</div>", nil
	})

	req := httptest.NewRequest("GET", "/search?q=hello&page=2", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	expected := "<div>Search: hello, Page: 2</div>"
	if w.Body.String() != expected {
		t.Errorf("expected %q, got %q", expected, w.Body.String())
	}
}

func TestRoute(t *testing.T) {
	r := New()
	r.Route("/api", func(r *Router) {
		r.GET("/users", func(ctx *Context) (string, error) {
			return "<ul><li>User 1</li></ul>", nil
		})
		r.POST("/users", func(ctx *Context) (string, error) {
			return "<li>New User</li>", nil
		})
	})

	// Test GET /api/users
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/users: expected status 200, got %d", w.Code)
	}

	// Test POST /api/users
	req = httptest.NewRequest("POST", "/api/users", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /api/users: expected status 200, got %d", w.Code)
	}
}

func TestNotFound(t *testing.T) {
	r := New()
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<div>Custom 404</div>"))
	})

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Custom 404") {
		t.Errorf("expected custom 404 body, got %q", w.Body.String())
	}
}

func TestMiddleware(t *testing.T) {
	r := New()

	// Add custom middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Custom", "middleware-value")
			next.ServeHTTP(w, req)
		})
	})

	r.GET("/test", func(ctx *Context) (string, error) {
		return "<div>Test</div>", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Header().Get("X-Custom") != "middleware-value" {
		t.Errorf("expected X-Custom header from middleware, got %q", w.Header().Get("X-Custom"))
	}
}

func TestHTMXHeaders(t *testing.T) {
	r := New()
	r.GET("/fragment", func(ctx *Context) (string, error) {
		if ctx.IsHTMX() {
			return "<div>HTMX Request</div>", nil
		}
		return "<div>Regular Request</div>", nil
	})

	// Test HTMX request
	req := httptest.NewRequest("GET", "/fragment", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "HTMX Request") {
		t.Errorf("expected HTMX response, got %q", w.Body.String())
	}

	// Test regular request
	req = httptest.NewRequest("GET", "/fragment", nil)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Regular Request") {
		t.Errorf("expected regular response, got %q", w.Body.String())
	}
}

func TestHandlerError(t *testing.T) {
	r := New()
	r.GET("/error", func(ctx *Context) (string, error) {
		return "", http.ErrAbortHandler
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestHandle(t *testing.T) {
	r := New()
	r.Handle("/standard", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("standard handler"))
	}))

	req := httptest.NewRequest("GET", "/standard", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Body.String() != "standard handler" {
		t.Errorf("expected 'standard handler', got %q", w.Body.String())
	}
}

func TestHandleFunc(t *testing.T) {
	r := New()
	r.HandleFunc("/func", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("func handler"))
	})

	req := httptest.NewRequest("GET", "/func", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Body.String() != "func handler" {
		t.Errorf("expected 'func handler', got %q", w.Body.String())
	}
}

func TestMount(t *testing.T) {
	r := New()

	sub := http.NewServeMux()
	sub.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("mounted"))
	})

	r.Mount("/sub", sub)

	req := httptest.NewRequest("GET", "/sub/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Body.String() != "mounted" {
		t.Errorf("expected 'mounted', got %q", w.Body.String())
	}
}

func TestHandler(t *testing.T) {
	r := New()
	r.GET("/test", func(ctx *Context) (string, error) {
		return "ok", nil
	})

	handler := r.Handler()
	if handler == nil {
		t.Fatal("Handler() returned nil")
	}

	// Use the handler directly
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Body.String() != "ok" {
		t.Errorf("expected 'ok', got %q", w.Body.String())
	}
}

// BenchmarkRouter benchmarks basic routing performance
func BenchmarkRouter(b *testing.B) {
	r := New()
	r.GET("/users/{id}", func(ctx *Context) (string, error) {
		return "<div>User</div>", nil
	})

	req := httptest.NewRequest("GET", "/users/123", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// Helper function to read response body
func readBody(resp *http.Response) string {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return string(body)
}
