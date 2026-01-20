package core

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response represents the hypermedia response back to the WebView.
// Body contains HTML fragments for HTMX to swap.
type Response struct {
	Status  int    // HTTP status code (200, 404, 500, etc.)
	Headers string // JSON-encoded response headers
	Body    []byte // HTML fragment (or JSON for capability responses)
}

// NewResponse creates a new Response with the given status.
func NewResponse(status int) *Response {
	return &Response{
		Status:  status,
		Headers: "{}",
	}
}

// GetHeader returns a response header value.
// Header lookup is case-insensitive per HTTP spec.
func (r *Response) GetHeader(key string) string {
	if r.Headers == "" || r.Headers == "{}" {
		return ""
	}
	var headers map[string]string
	if err := json.Unmarshal([]byte(r.Headers), &headers); err != nil {
		return ""
	}
	// Try exact match first
	if v, ok := headers[key]; ok {
		return v
	}
	// Try canonical header key (HTTP headers are case-insensitive)
	canonical := http.CanonicalHeaderKey(key)
	return headers[canonical]
}

// SetHeader sets a response header.
func (r *Response) SetHeader(key, value string) {
	var headers map[string]string
	if r.Headers == "" || r.Headers == "{}" {
		headers = make(map[string]string)
	} else {
		if err := json.Unmarshal([]byte(r.Headers), &headers); err != nil {
			headers = make(map[string]string)
		}
	}
	headers[key] = value
	data, _ := json.Marshal(headers)
	r.Headers = string(data)
}

// GetHeaders returns all headers as a map.
func (r *Response) GetHeaders() map[string]string {
	if r.Headers == "" || r.Headers == "{}" {
		return make(map[string]string)
	}
	var headers map[string]string
	if err := json.Unmarshal([]byte(r.Headers), &headers); err != nil {
		return make(map[string]string)
	}
	return headers
}

// SetHeaders sets all headers from a map.
func (r *Response) SetHeaders(headers map[string]string) {
	data, _ := json.Marshal(headers)
	r.Headers = string(data)
}

// BodyString returns the body as a string.
func (r *Response) BodyString() string {
	return string(r.Body)
}

// SetBody sets the body from a string.
func (r *Response) SetBody(body string) {
	r.Body = []byte(body)
}

// HTMLResponse creates a response with HTML content.
func HTMLResponse(status int, html string) *Response {
	r := &Response{
		Status: status,
		Body:   []byte(html),
	}
	r.SetHeader("Content-Type", "text/html; charset=utf-8")
	return r
}

// JSONResponse creates a response with JSON content.
func JSONResponse(status int, data any) *Response {
	body, err := json.Marshal(data)
	if err != nil {
		return ErrorResponse(500, "JSON encoding error: "+err.Error())
	}
	r := &Response{
		Status: status,
		Body:   body,
	}
	r.SetHeader("Content-Type", "application/json")
	return r
}

// ErrorResponse creates an error response with the given status and message.
func ErrorResponse(status int, message string) *Response {
	html := fmt.Sprintf(`<div class="error" role="alert">%s</div>`, message)
	return HTMLResponse(status, html)
}

// RedirectResponse creates a redirect response.
// For HTMX requests, sets HX-Redirect header. Otherwise, sets Location header.
func RedirectResponse(url string, isHTMX bool) *Response {
	r := NewResponse(200)
	if isHTMX {
		r.SetHeader("HX-Redirect", url)
	} else {
		r.Status = 302
		r.SetHeader("Location", url)
	}
	return r
}

// TriggerResponse creates a response that triggers an HTMX event.
func TriggerResponse(status int, html string, event string) *Response {
	r := HTMLResponse(status, html)
	r.SetHeader("HX-Trigger", event)
	return r
}

// RetargetResponse creates a response that changes the target element.
func RetargetResponse(status int, html string, target string) *Response {
	r := HTMLResponse(status, html)
	r.SetHeader("HX-Retarget", target)
	return r
}

// ReswapResponse creates a response that changes the swap strategy.
func ReswapResponse(status int, html string, swap string) *Response {
	r := HTMLResponse(status, html)
	r.SetHeader("HX-Reswap", swap)
	return r
}

// NoContentResponse creates a 204 No Content response.
func NoContentResponse() *Response {
	return NewResponse(204)
}

// NotFoundResponse creates a 404 Not Found response.
func NotFoundResponse(message string) *Response {
	if message == "" {
		message = "Not Found"
	}
	return ErrorResponse(404, message)
}

// BadRequestResponse creates a 400 Bad Request response.
func BadRequestResponse(message string) *Response {
	if message == "" {
		message = "Bad Request"
	}
	return ErrorResponse(400, message)
}

// InternalErrorResponse creates a 500 Internal Server Error response.
func InternalErrorResponse(message string) *Response {
	if message == "" {
		message = "Internal Server Error"
	}
	return ErrorResponse(500, message)
}
