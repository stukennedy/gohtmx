package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Context provides request data and response helpers for fragment handlers.
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	written  bool
}

// NewContext creates a new Context from the standard http types.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
	}
}

// Param returns a URL path parameter extracted by chi router.
func (c *Context) Param(key string) string {
	return chi.URLParam(c.Request, key)
}

// Query returns a query string parameter.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryDefault returns a query parameter or default value if not present.
func (c *Context) QueryDefault(key, defaultValue string) string {
	v := c.Query(key)
	if v == "" {
		return defaultValue
	}
	return v
}

// FormValue returns a form field value (works for POST form data).
func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

// IsHTMX returns true if this is an HTMX request.
func (c *Context) IsHTMX() bool {
	return c.Request.Header.Get("HX-Request") == "true"
}

// HXTarget returns the target element ID from HX-Target header.
func (c *Context) HXTarget() string {
	return c.Request.Header.Get("HX-Target")
}

// HXTrigger returns the triggering element ID from HX-Trigger header.
func (c *Context) HXTrigger() string {
	return c.Request.Header.Get("HX-Trigger")
}

// HXTriggerName returns the trigger name from HX-Trigger-Name header.
func (c *Context) HXTriggerName() string {
	return c.Request.Header.Get("HX-Trigger-Name")
}

// HXCurrentURL returns the current URL from HX-Current-URL header.
func (c *Context) HXCurrentURL() string {
	return c.Request.Header.Get("HX-Current-URL")
}

// HXPrompt returns the user's prompt response from HX-Prompt header.
func (c *Context) HXPrompt() string {
	return c.Request.Header.Get("HX-Prompt")
}

// Header returns a request header value.
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header.
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// HTML writes an HTML response with 200 status.
func (c *Context) HTML(html string) {
	c.HTMLStatus(http.StatusOK, html)
}

// HTMLStatus writes an HTML response with custom status.
func (c *Context) HTMLStatus(status int, html string) {
	c.written = true
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(status)
	c.Response.Write([]byte(html))
}

// JSON writes a JSON response with 200 status.
func (c *Context) JSON(data any) {
	c.JSONStatus(http.StatusOK, data)
}

// JSONStatus writes a JSON response with custom status.
func (c *Context) JSONStatus(status int, data any) {
	c.written = true
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(status)
	json.NewEncoder(c.Response).Encode(data)
}

// Error writes an error response.
func (c *Context) Error(err error) {
	c.ErrorStatus(http.StatusInternalServerError, err.Error())
}

// ErrorStatus writes an error response with custom status.
func (c *Context) ErrorStatus(status int, message string) {
	c.written = true
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(status)
	c.Response.Write([]byte(`<div class="error" role="alert">` + message + `</div>`))
}

// NotFound writes a 404 response.
func (c *Context) NotFound(message string) {
	if message == "" {
		message = "Not Found"
	}
	c.ErrorStatus(http.StatusNotFound, message)
}

// BadRequest writes a 400 response.
func (c *Context) BadRequest(message string) {
	if message == "" {
		message = "Bad Request"
	}
	c.ErrorStatus(http.StatusBadRequest, message)
}

// Trigger sends an HX-Trigger header for client-side events.
func (c *Context) Trigger(event string) {
	c.Response.Header().Set("HX-Trigger", event)
}

// TriggerAfterSettle sends an HX-Trigger-After-Settle header.
func (c *Context) TriggerAfterSettle(event string) {
	c.Response.Header().Set("HX-Trigger-After-Settle", event)
}

// TriggerAfterSwap sends an HX-Trigger-After-Swap header.
func (c *Context) TriggerAfterSwap(event string) {
	c.Response.Header().Set("HX-Trigger-After-Swap", event)
}

// Redirect sends a redirect response.
// For HTMX requests, uses HX-Redirect. Otherwise, standard redirect.
func (c *Context) Redirect(url string) {
	c.written = true
	if c.IsHTMX() {
		c.Response.Header().Set("HX-Redirect", url)
		c.Response.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(c.Response, c.Request, url, http.StatusSeeOther)
	}
}

// PushURL tells HTMX to push a new URL to the browser history.
func (c *Context) PushURL(url string) {
	c.Response.Header().Set("HX-Push-Url", url)
}

// ReplaceURL tells HTMX to replace the current URL in browser history.
func (c *Context) ReplaceURL(url string) {
	c.Response.Header().Set("HX-Replace-Url", url)
}

// Retarget changes the target element for the swap.
func (c *Context) Retarget(selector string) {
	c.Response.Header().Set("HX-Retarget", selector)
}

// Reswap changes the swap strategy.
func (c *Context) Reswap(strategy string) {
	c.Response.Header().Set("HX-Reswap", strategy)
}

// Reselect changes which part of the response to use.
func (c *Context) Reselect(selector string) {
	c.Response.Header().Set("HX-Reselect", selector)
}

// Refresh tells HTMX to do a full page refresh.
func (c *Context) Refresh() {
	c.written = true
	c.Response.Header().Set("HX-Refresh", "true")
	c.Response.WriteHeader(http.StatusOK)
}

// NoContent writes a 204 No Content response.
func (c *Context) NoContent() {
	c.written = true
	c.Response.WriteHeader(http.StatusNoContent)
}

// Written returns true if a response has been written.
func (c *Context) Written() bool {
	return c.written
}

// Bind decodes JSON body into the given struct.
func (c *Context) Bind(v any) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}
