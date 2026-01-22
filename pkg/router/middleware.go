package router

import (
	"context"
	"net/http"
)

// contextKey is used for context values.
type contextKey string

const (
	// DatastarRequestKey is the context key for Datastar SSE request detection.
	DatastarRequestKey contextKey = "datastar-request"
)

// DatastarRequestMiddleware detects and tags Datastar SSE requests.
// Datastar sends Accept: text/event-stream for SSE requests.
func DatastarRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isDatastar := r.Header.Get("Accept") == "text/event-stream"
		ctx := context.WithValue(r.Context(), DatastarRequestKey, isDatastar)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsDatastarRequest returns true if the request is a Datastar SSE request.
func IsDatastarRequest(r *http.Request) bool {
	if v, ok := r.Context().Value(DatastarRequestKey).(bool); ok {
		return v
	}
	return r.Header.Get("Accept") == "text/event-stream"
}

// LayoutWrapper wraps fragment responses in a full page layout
// when the request is not from Datastar (direct browser navigation).
type LayoutWrapper struct {
	Layout func(content string) string
}

// Wrap returns middleware that wraps non-Datastar responses in a layout.
func (l *LayoutWrapper) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsDatastarRequest(r) || l.Layout == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Capture the response
		rec := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(rec, r)

		// If it's HTML and not an error, wrap in layout
		contentType := rec.Header().Get("Content-Type")
		if rec.statusCode < 400 && isHTML(contentType) {
			wrapped := l.Layout(string(rec.body))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(rec.statusCode)
			w.Write([]byte(wrapped))
		} else {
			// Pass through as-is
			w.WriteHeader(rec.statusCode)
			w.Write(rec.body)
		}
	})
}

// responseRecorder captures the response for post-processing.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}

func isHTML(contentType string) bool {
	return contentType == "" ||
		contentType == "text/html" ||
		contentType == "text/html; charset=utf-8"
}

// NoCacheMiddleware sets headers to prevent caching.
func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds CORS headers for development.
func CORSMiddleware(allowedOrigins ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireDatastar returns 400 if the request is not a Datastar SSE request.
func RequireDatastar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsDatastarRequest(r) {
			http.Error(w, "Datastar SSE request required", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
