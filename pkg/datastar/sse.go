// Package datastar provides Server-Sent Events (SSE) support for Datastar integration.
// It wraps the datastar-go SDK with convenience methods for the Irgo framework.
package datastar

import (
	"bytes"
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
)

// SSE wraps the datastar ServerSentEventGenerator with additional convenience methods.
type SSE struct {
	*datastar.ServerSentEventGenerator
}

// NewSSE creates a new SSE writer for streaming responses to the client.
// The connection stays alive until the context is canceled or the handler returns.
func NewSSE(w http.ResponseWriter, r *http.Request, opts ...datastar.SSEOption) *SSE {
	return &SSE{
		ServerSentEventGenerator: datastar.NewSSE(w, r, opts...),
	}
}

// PatchTempl renders a templ component and patches it into the DOM.
// This is a convenience method that wraps PatchElementTempl.
func (s *SSE) PatchTempl(c templ.Component, opts ...datastar.PatchElementOption) error {
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// PatchTemplByID renders a templ component and patches it into a specific element by ID.
func (s *SSE) PatchTemplByID(id string, c templ.Component, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithSelectorID(id))
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// PatchHTML patches raw HTML into the DOM.
func (s *SSE) PatchHTML(html string, opts ...datastar.PatchElementOption) error {
	return s.ServerSentEventGenerator.PatchElements(html, opts...)
}

// PatchHTMLByID patches raw HTML into a specific element by ID.
func (s *SSE) PatchHTMLByID(id string, html string, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithSelectorID(id))
	return s.ServerSentEventGenerator.PatchElements(html, opts...)
}

// AppendTempl appends a templ component inside an element.
func (s *SSE) AppendTempl(c templ.Component, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithModeAppend())
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// AppendTemplByID appends a templ component inside a specific element by ID.
func (s *SSE) AppendTemplByID(id string, c templ.Component, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithSelectorID(id), datastar.WithModeAppend())
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// PrependTempl prepends a templ component inside an element.
func (s *SSE) PrependTempl(c templ.Component, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithModePrepend())
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// PrependTemplByID prepends a templ component inside a specific element by ID.
func (s *SSE) PrependTemplByID(id string, c templ.Component, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithSelectorID(id), datastar.WithModePrepend())
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// RemoveByID removes an element by its ID.
func (s *SSE) RemoveByID(id string) error {
	return s.ServerSentEventGenerator.RemoveElementByID(id)
}

// Remove removes elements matching a CSS selector.
func (s *SSE) Remove(selector string, opts ...datastar.PatchElementOption) error {
	return s.ServerSentEventGenerator.RemoveElement(selector, opts...)
}

// PatchSignals updates client-side signals with the provided data.
// The signals parameter should be a struct or map that will be marshaled to JSON.
func (s *SSE) PatchSignals(signals any, opts ...datastar.PatchSignalsOption) error {
	return s.ServerSentEventGenerator.MarshalAndPatchSignals(signals, opts...)
}

// PatchSignalsIfMissing updates signals only if they don't already exist on the client.
func (s *SSE) PatchSignalsIfMissing(signals any, opts ...datastar.PatchSignalsOption) error {
	return s.ServerSentEventGenerator.MarshalAndPatchSignalsIfMissing(signals, opts...)
}

// Redirect navigates the client to a new URL.
func (s *SSE) Redirect(url string, opts ...datastar.ExecuteScriptOption) error {
	return s.ServerSentEventGenerator.Redirect(url, opts...)
}

// ExecuteScript executes JavaScript on the client.
func (s *SSE) ExecuteScript(script string, opts ...datastar.ExecuteScriptOption) error {
	return s.ServerSentEventGenerator.ExecuteScript(script, opts...)
}

// ConsoleLog logs a message to the browser console.
func (s *SSE) ConsoleLog(msg string, opts ...datastar.ExecuteScriptOption) error {
	return s.ServerSentEventGenerator.ConsoleLog(msg, opts...)
}

// ConsoleError logs an error to the browser console.
func (s *SSE) ConsoleError(err error, opts ...datastar.ExecuteScriptOption) error {
	return s.ServerSentEventGenerator.ConsoleError(err, opts...)
}

// DispatchEvent dispatches a custom event on the client.
func (s *SSE) DispatchEvent(eventName string, detail any, opts ...datastar.DispatchCustomEventOption) error {
	return s.ServerSentEventGenerator.DispatchCustomEvent(eventName, detail, opts...)
}

// Context returns the request context, which can be used to detect client disconnection.
func (s *SSE) Context() context.Context {
	return s.ServerSentEventGenerator.Context()
}

// IsClosed returns true if the client has disconnected.
func (s *SSE) IsClosed() bool {
	return s.ServerSentEventGenerator.IsClosed()
}

// ReadSignals extracts Datastar signals from an HTTP request and unmarshals them.
// For GET requests, signals are read from URL query parameters.
// For other methods, signals are read from the JSON-encoded request body.
func ReadSignals(r *http.Request, signals any) error {
	return datastar.ReadSignals(r, signals)
}

// RenderTempl renders a templ component to a string.
// This is useful for testing or when you need the HTML string directly.
func RenderTempl(c templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Re-export commonly used options for convenience
var (
	// Element patch modes
	WithModeOuter   = datastar.WithModeOuter
	WithModeInner   = datastar.WithModeInner
	WithModeRemove  = datastar.WithModeRemove
	WithModeReplace = datastar.WithModeReplace
	WithModePrepend = datastar.WithModePrepend
	WithModeAppend  = datastar.WithModeAppend
	WithModeBefore  = datastar.WithModeBefore
	WithModeAfter   = datastar.WithModeAfter

	// Selectors
	WithSelector    = datastar.WithSelector
	WithSelectorID  = datastar.WithSelectorID
	WithSelectorf   = datastar.WithSelectorf

	// View transitions
	WithViewTransitions    = datastar.WithViewTransitions
	WithoutViewTransitions = datastar.WithoutViewTransitions
)

// Sugar functions for generating action attributes in templates
var (
	GetSSE    = datastar.GetSSE
	PostSSE   = datastar.PostSSE
	PutSSE    = datastar.PutSSE
	PatchSSE  = datastar.PatchSSE
	DeleteSSE = datastar.DeleteSSE
)
