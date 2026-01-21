package router

import (
	"net/http"
	"strings"
)

// SecretValidationMiddleware validates the X-Irgo-Secret header on requests.
// Requests without a valid secret receive a 403 Forbidden response.
//
// The following requests bypass validation:
//   - GET, HEAD, OPTIONS requests (safe methods that can't mutate state)
//   - Paths matching excludePaths prefixes (e.g., "/static/")
//
// This allows the webview to load the initial page and static assets,
// while protecting state-changing operations (POST, PUT, DELETE, PATCH).
func SecretValidationMiddleware(secret string, excludePaths []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Safe methods bypass secret validation
			// GET/HEAD can't mutate state, OPTIONS is for CORS preflight
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check if path is excluded
			for _, prefix := range excludePaths {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Validate secret header for state-changing requests
			headerSecret := r.Header.Get("X-Irgo-Secret")
			if headerSecret != secret {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// StrictOriginMiddleware validates the Origin header exactly.
// For non-safe methods (not GET, HEAD, OPTIONS), the Origin must exactly match
// one of the allowed origins. This prevents DNS rebinding and CSRF attacks.
// Uses exact string matching only - no suffix/prefix patterns.
func StrictOriginMiddleware(allowedOrigins ...string) func(http.Handler) http.Handler {
	// Build a set for O(1) lookup
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Safe methods don't require origin validation
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get("Origin")

			// If no Origin header, this might be a same-origin request
			// from the webview or a non-browser client.
			// We allow it since the secret header provides authentication.
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Exact match only - no endsWith or contains
			if _, ok := originSet[origin]; !ok {
				http.Error(w, "Forbidden: invalid origin", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WebSocketSecretMiddleware validates the secret for WebSocket upgrade requests.
// Since the WebSocket API doesn't support custom headers, the secret is passed
// as a query parameter: ?secret=xxx
func WebSocketSecretMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only check WebSocket upgrade requests
			if !isWebSocketUpgrade(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Validate secret from query parameter
			querySecret := r.URL.Query().Get("secret")
			if querySecret != secret {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade.
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}
