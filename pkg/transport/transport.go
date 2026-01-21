// Package transport provides a unified abstraction for communication between
// the WebView frontend and Go backend handlers. It supports both network-based
// (LoopbackTransport) and in-memory (InProcessTransport) implementations.
package transport

import (
	"context"
	"errors"

	"github.com/stukennedy/irgo/pkg/core"
)

var (
	// ErrTransportClosed is returned when operations are attempted on a closed transport.
	ErrTransportClosed = errors.New("transport closed")

	// ErrChannelClosed is returned when operations are attempted on a closed channel.
	ErrChannelClosed = errors.New("channel closed")

	// ErrChannelFull is returned when the channel buffer is full and would block.
	ErrChannelFull = errors.New("channel buffer full")

	// ErrNoHandler is returned when no handler is registered for a URL pattern.
	ErrNoHandler = errors.New("no handler registered for URL")
)

// Transport abstracts the communication layer between WebView and Go handlers.
// Implementations include LoopbackTransport (real HTTP on localhost) and
// InProcessTransport (in-memory, no network I/O).
type Transport interface {
	// HandleRequest processes an HTTP-like request and returns a response.
	// This is the primary method for request/response communication.
	HandleRequest(ctx context.Context, req *core.Request) (*core.Response, error)

	// OpenChannel creates a bidirectional channel for WebSocket-like communication.
	// The url parameter matches against registered channel handlers.
	OpenChannel(ctx context.Context, url string) (Channel, error)

	// RegisterChannelHandler sets the handler for channels matching a URL pattern.
	// Patterns can be exact ("/ws/chat") or prefix ("/ws/").
	RegisterChannelHandler(pattern string, handler ChannelHandler)

	// SetDefaultChannelHandler sets the fallback handler for unmatched patterns.
	SetDefaultChannelHandler(handler ChannelHandler)

	// Start initializes the transport. For LoopbackTransport, this starts
	// the HTTP server. For InProcessTransport, this is a no-op.
	Start() error

	// Stop gracefully shuts down the transport.
	Stop(ctx context.Context) error

	// Config returns the transport configuration.
	// Returns nil for transports that don't require configuration.
	Config() *Config
}

// Config holds transport configuration.
type Config struct {
	// Security settings (LoopbackTransport only)
	Secret         string   // Per-launch authentication secret
	AllowedOrigins []string // Origins allowed for CORS/security

	// Server settings (LoopbackTransport only)
	Port    int    // Port number (0 for auto-select)
	Address string // Bind address (always "127.0.0.1" for security)

	// Channel settings
	ChannelBufferSize int // Buffer size for channel messages (default: 100)
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Address:           "127.0.0.1",
		ChannelBufferSize: 100,
	}
}

// Option configures a Transport.
type Option func(*Config)

// WithPort sets the server port (LoopbackTransport only).
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithSecret sets the authentication secret.
func WithSecret(secret string) Option {
	return func(c *Config) {
		c.Secret = secret
	}
}

// WithAllowedOrigins sets the allowed origins for CORS/security.
func WithAllowedOrigins(origins ...string) Option {
	return func(c *Config) {
		c.AllowedOrigins = origins
	}
}

// WithChannelBufferSize sets the channel message buffer size.
func WithChannelBufferSize(size int) Option {
	return func(c *Config) {
		c.ChannelBufferSize = size
	}
}
