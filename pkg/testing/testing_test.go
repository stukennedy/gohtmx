package testing

import (
	"net/http"
	"testing"
)

func newTestHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<h1>Welcome</h1>"))
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			name := r.FormValue("name")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("<div id=\"user\">" + name + "</div>"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<ul><li>User 1</li></ul>"))
	})

	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") == "text/event-stream" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Write([]byte("event: datastar-patch-elements\ndata: fragments <div>SSE Response</div>\n\n"))
			return
		}
		w.Write([]byte("<div>Regular Response</div>"))
	})

	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	})

	mux.HandleFunc("/notfound", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return mux
}

func TestClientGet(t *testing.T) {
	client := NewClient(newTestHandler())

	resp := client.Get("/")
	resp.AssertOK(t)
	resp.AssertContains(t, "Welcome")
	resp.AssertHTML(t)
}

func TestClientPostForm(t *testing.T) {
	client := NewClient(newTestHandler())

	resp := client.PostForm("/users", map[string]string{"name": "John"})
	resp.AssertCreated(t)
	resp.AssertContains(t, "John")
}

func TestClientDatastar(t *testing.T) {
	client := NewClient(newTestHandler())

	// Regular request
	resp := client.Get("/sse")
	resp.AssertContains(t, "Regular Response")

	// Datastar SSE request
	resp = client.Datastar().Get("/sse")
	resp.AssertSSE(t)
	resp.AssertSSEEvent(t, "datastar-patch-elements")
	resp.AssertSSEContains(t, "SSE Response")
}

func TestClientDelete(t *testing.T) {
	client := NewClient(newTestHandler())

	resp := client.Delete("/delete")
	resp.AssertNoContent(t)
}

func TestResponseAssertions(t *testing.T) {
	client := NewClient(newTestHandler())

	// Test NotFound assertion
	resp := client.Get("/notfound")
	resp.AssertNotFound(t)

	// Test BadRequest assertion
	resp = client.Get("/error")
	resp.AssertBadRequest(t)

	// Test Redirect assertion
	resp = client.Get("/redirect")
	resp.AssertRedirect(t)

	// Test JSON assertion
	resp = client.Get("/json")
	resp.AssertJSON(t)
}

func TestResponseContainsAll(t *testing.T) {
	client := NewClient(newTestHandler())

	resp := client.Get("/users")

	if !resp.ContainsAll("ul", "li", "User 1") {
		t.Error("expected body to contain all elements")
	}

	resp.AssertContainsAll(t, "ul", "li", "User 1")
}

func TestHTMLAssertions(t *testing.T) {
	client := NewClient(newTestHandler())

	resp := client.PostForm("/users", map[string]string{"name": "Test"})
	resp.HTML(t).ContainsID("user")
	resp.HTML(t).ContainsElement("div", `id="user"`)
}

func TestRequestBuilder(t *testing.T) {
	handler := newTestHandler()

	resp := NewRequest("POST", "/users").
		WithFormBody(map[string]string{"name": "Builder Test"}).
		Execute(handler)

	resp.AssertCreated(t)
	resp.AssertContains(t, "Builder Test")
}

func TestRequestBuilderDatastar(t *testing.T) {
	handler := newTestHandler()

	resp := NewRequest("GET", "/sse").
		AsDatastar().
		Execute(handler)

	resp.AssertSSE(t)
	resp.AssertSSEContains(t, "SSE Response")
}

func TestWithHeader(t *testing.T) {
	client := NewClient(newTestHandler())

	// Add custom header
	customClient := client.WithHeader("X-Custom", "value")
	_ = customClient.Get("/")

	// Original client should be unchanged
	if _, ok := client.headers["X-Custom"]; ok {
		t.Error("original client should not have custom header")
	}
}

func TestMockRenderer(t *testing.T) {
	mock := &MockRenderer{}

	// Render something
	_, _ = mock.Render(nil)
	_, _ = mock.Render(nil)

	mock.AssertRenderedCount(t, 2)

	mock.Reset()
	mock.AssertRenderedCount(t, 0)
}

func TestBodyReader(t *testing.T) {
	reader := BodyReader("test content")
	if reader == nil {
		t.Error("expected non-nil reader")
	}
}
