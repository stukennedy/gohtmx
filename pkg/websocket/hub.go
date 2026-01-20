package websocket

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrSessionNotFound is returned when a session doesn't exist.
	ErrSessionNotFound = errors.New("websocket session not found")

	// ErrSessionClosed is returned when trying to use a closed session.
	ErrSessionClosed = errors.New("websocket session closed")

	// ErrNoHandler is returned when no handler is registered for a URL.
	ErrNoHandler = errors.New("no handler registered for URL")
)

// Hub manages all WebSocket sessions and message routing.
type Hub struct {
	sessions    map[string]*Session
	handlers    map[string]MessageHandler // URL pattern → handler
	defaultHandler MessageHandler
	sessionsMu  sync.RWMutex
	handlersMu  sync.RWMutex
	counter     uint64

	// Callback for when sessions are created/destroyed
	onSessionCreated  func(session *Session)
	onSessionDestroyed func(session *Session)
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]*Session),
		handlers: make(map[string]MessageHandler),
	}
}

// Handle registers a handler for a URL pattern.
// Patterns can be exact ("/ws/chat") or prefix ("/ws/").
func (h *Hub) Handle(pattern string, handler MessageHandler) {
	h.handlersMu.Lock()
	defer h.handlersMu.Unlock()
	h.handlers[pattern] = handler
}

// HandleFunc registers a function handler for a URL pattern.
func (h *Hub) HandleFunc(pattern string, handler func(*Session, *Request) (*Envelope, error)) {
	h.Handle(pattern, MessageHandlerFunc(handler))
}

// SetDefaultHandler sets the handler for URLs that don't match any pattern.
func (h *Hub) SetDefaultHandler(handler MessageHandler) {
	h.handlersMu.Lock()
	defer h.handlersMu.Unlock()
	h.defaultHandler = handler
}

// OnSessionCreated sets a callback for when sessions are created.
func (h *Hub) OnSessionCreated(fn func(*Session)) {
	h.onSessionCreated = fn
}

// OnSessionDestroyed sets a callback for when sessions are destroyed.
func (h *Hub) OnSessionDestroyed(fn func(*Session)) {
	h.onSessionDestroyed = fn
}

// Connect creates a new session for the given URL.
// Returns the session ID and the session.
func (h *Hub) Connect(url string) (*Session, error) {
	handler := h.findHandler(url)
	if handler == nil && h.defaultHandler == nil {
		return nil, ErrNoHandler
	}
	if handler == nil {
		handler = h.defaultHandler
	}

	sessionID := h.generateSessionID()
	session := NewSession(sessionID, url, handler)

	h.sessionsMu.Lock()
	h.sessions[sessionID] = session
	h.sessionsMu.Unlock()

	// Call OnConnect
	if err := handler.OnConnect(session); err != nil {
		h.sessionsMu.Lock()
		delete(h.sessions, sessionID)
		h.sessionsMu.Unlock()
		return nil, err
	}

	if h.onSessionCreated != nil {
		h.onSessionCreated(session)
	}

	return session, nil
}

// ConnectWithID creates a session with a specific ID (for reconnection).
func (h *Hub) ConnectWithID(sessionID, url string) (*Session, error) {
	handler := h.findHandler(url)
	if handler == nil && h.defaultHandler == nil {
		return nil, ErrNoHandler
	}
	if handler == nil {
		handler = h.defaultHandler
	}

	session := NewSession(sessionID, url, handler)

	h.sessionsMu.Lock()
	// If session already exists, close the old one
	if old, exists := h.sessions[sessionID]; exists {
		old.Close()
	}
	h.sessions[sessionID] = session
	h.sessionsMu.Unlock()

	if err := handler.OnConnect(session); err != nil {
		h.sessionsMu.Lock()
		delete(h.sessions, sessionID)
		h.sessionsMu.Unlock()
		return nil, err
	}

	if h.onSessionCreated != nil {
		h.onSessionCreated(session)
	}

	return session, nil
}

// Disconnect closes and removes a session.
func (h *Hub) Disconnect(sessionID string) {
	h.sessionsMu.Lock()
	session, exists := h.sessions[sessionID]
	if exists {
		delete(h.sessions, sessionID)
	}
	h.sessionsMu.Unlock()

	if exists {
		session.Close()
		if h.onSessionDestroyed != nil {
			h.onSessionDestroyed(session)
		}
	}
}

// GetSession returns a session by ID.
func (h *Hub) GetSession(sessionID string) (*Session, bool) {
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()
	s, ok := h.sessions[sessionID]
	return s, ok
}

// HandleMessage processes an incoming message for a session.
func (h *Hub) HandleMessage(sessionID string, data []byte) (*Envelope, error) {
	session, ok := h.GetSession(sessionID)
	if !ok {
		return nil, ErrSessionNotFound
	}
	if session.IsClosed() {
		return nil, ErrSessionClosed
	}
	return session.HandleMessage(data)
}

// Send sends an envelope to a specific session.
func (h *Hub) Send(sessionID string, envelope *Envelope) error {
	session, ok := h.GetSession(sessionID)
	if !ok {
		return ErrSessionNotFound
	}
	if !session.Send(envelope) {
		return ErrSessionClosed
	}
	return nil
}

// SendHTML sends an HTML fragment to a session.
func (h *Hub) SendHTML(sessionID, target, html string) error {
	return h.Send(sessionID, HTMLEnvelope(target, html))
}

// Broadcast sends an envelope to all sessions.
func (h *Hub) Broadcast(envelope *Envelope) {
	h.sessionsMu.RLock()
	sessions := make([]*Session, 0, len(h.sessions))
	for _, s := range h.sessions {
		sessions = append(sessions, s)
	}
	h.sessionsMu.RUnlock()

	for _, s := range sessions {
		s.Send(envelope)
	}
}

// BroadcastHTML sends HTML to all sessions.
func (h *Hub) BroadcastHTML(target, html string) {
	h.Broadcast(HTMLEnvelope(target, html))
}

// BroadcastToURL sends to all sessions connected to URLs matching the pattern.
func (h *Hub) BroadcastToURL(urlPattern string, envelope *Envelope) {
	h.sessionsMu.RLock()
	sessions := make([]*Session, 0)
	for _, s := range h.sessions {
		if h.matchURL(s.URL, urlPattern) {
			sessions = append(sessions, s)
		}
	}
	h.sessionsMu.RUnlock()

	for _, s := range sessions {
		s.Send(envelope)
	}
}

// Sessions returns the number of active sessions.
func (h *Hub) SessionCount() int {
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()
	return len(h.sessions)
}

// SessionsForURL returns sessions connected to URLs matching the pattern.
func (h *Hub) SessionsForURL(urlPattern string) []*Session {
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()

	var result []*Session
	for _, s := range h.sessions {
		if h.matchURL(s.URL, urlPattern) {
			result = append(result, s)
		}
	}
	return result
}

// AllSessions returns all active sessions.
func (h *Hub) AllSessions() []*Session {
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()

	result := make([]*Session, 0, len(h.sessions))
	for _, s := range h.sessions {
		result = append(result, s)
	}
	return result
}

// CleanupExpired removes stale pending requests from all sessions.
func (h *Hub) CleanupExpired(ttl time.Duration) {
	h.sessionsMu.RLock()
	sessions := make([]*Session, 0, len(h.sessions))
	for _, s := range h.sessions {
		sessions = append(sessions, s)
	}
	h.sessionsMu.RUnlock()

	for _, s := range sessions {
		s.CleanupExpiredPending(ttl)
	}
}

// Close closes all sessions and cleans up the hub.
func (h *Hub) Close() {
	h.sessionsMu.Lock()
	sessions := make([]*Session, 0, len(h.sessions))
	for _, s := range h.sessions {
		sessions = append(sessions, s)
	}
	h.sessions = make(map[string]*Session)
	h.sessionsMu.Unlock()

	for _, s := range sessions {
		s.Close()
		if h.onSessionDestroyed != nil {
			h.onSessionDestroyed(s)
		}
	}
}

func (h *Hub) findHandler(url string) MessageHandler {
	h.handlersMu.RLock()
	defer h.handlersMu.RUnlock()

	// Exact match first
	if handler, ok := h.handlers[url]; ok {
		return handler
	}

	// Prefix match
	for pattern, handler := range h.handlers {
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(url, pattern) {
			return handler
		}
	}

	// Path extraction (remove protocol and host)
	path := extractPath(url)
	if handler, ok := h.handlers[path]; ok {
		return handler
	}

	for pattern, handler := range h.handlers {
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(path, pattern) {
			return handler
		}
	}

	return nil
}

func (h *Hub) matchURL(url, pattern string) bool {
	if url == pattern {
		return true
	}
	if strings.HasSuffix(pattern, "/") && strings.HasPrefix(url, pattern) {
		return true
	}
	path := extractPath(url)
	if path == pattern {
		return true
	}
	if strings.HasSuffix(pattern, "/") && strings.HasPrefix(path, pattern) {
		return true
	}
	return false
}

func (h *Hub) generateSessionID() string {
	id := atomic.AddUint64(&h.counter, 1)
	return "ws_" + time.Now().Format("20060102150405") + "_" + itoa(id)
}

func extractPath(url string) string {
	// Remove protocol
	if idx := strings.Index(url, "://"); idx != -1 {
		url = url[idx+3:]
	}
	// Remove host
	if idx := strings.Index(url, "/"); idx != -1 {
		return url[idx:]
	}
	return "/"
}

func itoa(n uint64) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for n > 0 {
		buf[i] = digits[n%10]
		n /= 10
		i--
	}
	return string(buf[i+1:])
}
