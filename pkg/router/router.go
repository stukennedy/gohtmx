// Package router provides an HTMX-aware HTTP router built on chi.
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// FragmentHandler is a handler function that returns an HTML fragment.
// If an error is returned, an error response is automatically generated.
type FragmentHandler func(ctx *Context) (string, error)

// Router wraps chi with hypermedia-specific conventions.
type Router struct {
	mux *chi.Mux
}

// New creates a new Router with default middleware.
func New() *Router {
	r := chi.NewRouter()

	// Default middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(HXRequestMiddleware)

	return &Router{mux: r}
}

// NewWithoutMiddleware creates a Router without default middleware.
func NewWithoutMiddleware() *Router {
	return &Router{mux: chi.NewRouter()}
}

// Handler returns the underlying http.Handler for use with the adapter.
func (r *Router) Handler() http.Handler {
	return r.mux
}

// Use adds middleware to the router.
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.mux.Use(middlewares...)
}

// Fragment registers a handler that returns HTML fragments.
func (r *Router) Fragment(method, pattern string, handler FragmentHandler) {
	r.mux.Method(method, pattern, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(w, req)
		html, err := handler(ctx)
		if err != nil {
			ctx.Error(err)
			return
		}
		if !ctx.Written() {
			ctx.HTML(html)
		}
	}))
}

// GET registers a GET handler that returns HTML fragments.
func (r *Router) GET(pattern string, handler FragmentHandler) {
	r.Fragment(http.MethodGet, pattern, handler)
}

// POST registers a POST handler that returns HTML fragments.
func (r *Router) POST(pattern string, handler FragmentHandler) {
	r.Fragment(http.MethodPost, pattern, handler)
}

// PUT registers a PUT handler that returns HTML fragments.
func (r *Router) PUT(pattern string, handler FragmentHandler) {
	r.Fragment(http.MethodPut, pattern, handler)
}

// PATCH registers a PATCH handler that returns HTML fragments.
func (r *Router) PATCH(pattern string, handler FragmentHandler) {
	r.Fragment(http.MethodPatch, pattern, handler)
}

// DELETE registers a DELETE handler that returns HTML fragments.
func (r *Router) DELETE(pattern string, handler FragmentHandler) {
	r.Fragment(http.MethodDelete, pattern, handler)
}

// Handle registers a standard http.Handler.
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

// HandleFunc registers a standard http.HandlerFunc.
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
}

// Mount attaches a sub-router at the given pattern.
func (r *Router) Mount(pattern string, handler http.Handler) {
	r.mux.Mount(pattern, handler)
}

// Group creates a new route group with shared middleware.
func (r *Router) Group(fn func(r *Router)) {
	r.mux.Group(func(c chi.Router) {
		// Create sub-router that wraps the chi Router interface
		subRouter := &Router{mux: chi.NewRouter()}
		fn(subRouter)
		// Mount the sub-router's routes
		c.Mount("/", subRouter.mux)
	})
}

// Route creates a new route group at the given pattern.
func (r *Router) Route(pattern string, fn func(r *Router)) {
	r.mux.Route(pattern, func(c chi.Router) {
		subRouter := &Router{mux: c.(*chi.Mux)}
		fn(subRouter)
	})
}

// With adds inline middleware for a route.
func (r *Router) With(middlewares ...func(http.Handler) http.Handler) *Router {
	return &Router{mux: r.mux.With(middlewares...).(*chi.Mux)}
}

// NotFound registers a custom 404 handler.
func (r *Router) NotFound(handler http.HandlerFunc) {
	r.mux.NotFound(handler)
}

// MethodNotAllowed registers a custom 405 handler.
func (r *Router) MethodNotAllowed(handler http.HandlerFunc) {
	r.mux.MethodNotAllowed(handler)
}

// Static serves static files from the given filesystem.
func (r *Router) Static(pattern string, root http.FileSystem) {
	if pattern != "/" && pattern[len(pattern)-1] != '/' {
		r.mux.Get(pattern, http.RedirectHandler(pattern+"/", http.StatusMovedPermanently).ServeHTTP)
		pattern += "/"
	}
	pattern += "*"

	r.mux.Get(pattern, func(w http.ResponseWriter, req *http.Request) {
		rctx := chi.RouteContext(req.Context())
		pathPrefix := pattern[:len(pattern)-1]
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		rctx.URLParams.Add("*", req.URL.Path[len(pathPrefix):])
		fs.ServeHTTP(w, req)
	})
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
