package websocket

import (
	"sync"
	"time"
)

// Session represents a virtual WebSocket connection.
// Each WebSocket connection from the WebView creates one session.
type Session struct {
	ID        string
	URL       string
	CreatedAt time.Time

	// SendChan receives envelopes to be sent to the WebView.
	// The mobile bridge reads from this channel.
	SendChan chan *Envelope

	// Handler processes incoming messages.
	Handler MessageHandler

	// Pending tracks requests awaiting responses.
	pending   map[string]*pendingRequest
	pendingMu sync.RWMutex

	// Metadata allows storing arbitrary data with the session.
	metadata   map[string]any
	metadataMu sync.RWMutex

	// closed tracks if the session has been closed.
	closed bool
	mu     sync.RWMutex
}

type pendingRequest struct {
	Request   *Request
	Timestamp time.Time
}

// MessageHandler processes WebSocket messages for a session.
type MessageHandler interface {
	// OnConnect is called when a WebSocket connection is established.
	OnConnect(session *Session) error

	// OnMessage is called when a message is received from the client.
	// Returns an envelope to send back, or nil for no response.
	OnMessage(session *Session, req *Request) (*Envelope, error)

	// OnClose is called when the connection is closed.
	OnClose(session *Session)
}

// MessageHandlerFunc is a function adapter for simple handlers.
type MessageHandlerFunc func(session *Session, req *Request) (*Envelope, error)

// OnConnect implements MessageHandler (no-op).
func (f MessageHandlerFunc) OnConnect(session *Session) error {
	return nil
}

// OnMessage implements MessageHandler.
func (f MessageHandlerFunc) OnMessage(session *Session, req *Request) (*Envelope, error) {
	return f(session, req)
}

// OnClose implements MessageHandler (no-op).
func (f MessageHandlerFunc) OnClose(session *Session) {
}

// NewSession creates a new WebSocket session.
func NewSession(id, url string, handler MessageHandler) *Session {
	return &Session{
		ID:        id,
		URL:       url,
		CreatedAt: time.Now(),
		SendChan:  make(chan *Envelope, 100), // Buffered to prevent blocking
		Handler:   handler,
		pending:   make(map[string]*pendingRequest),
		metadata:  make(map[string]any),
	}
}

// Send queues an envelope to be sent to the client.
func (s *Session) Send(envelope *Envelope) bool {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return false
	}
	s.mu.RUnlock()

	select {
	case s.SendChan <- envelope:
		return true
	default:
		// Channel full, drop the message
		return false
	}
}

// SendHTML sends an HTML fragment to a target element.
func (s *Session) SendHTML(target, html string) bool {
	return s.Send(HTMLEnvelope(target, html))
}

// Reply sends a response matching a specific request.
func (s *Session) Reply(requestID, html string) bool {
	s.clearPending(requestID)
	return s.Send(ReplyEnvelope(requestID, html))
}

// HandleMessage processes an incoming message from the client.
func (s *Session) HandleMessage(data []byte) (*Envelope, error) {
	req, err := ParseRequest(data)
	if err != nil {
		return nil, err
	}

	// Track pending request for response matching
	if req.RequestID != "" {
		s.trackPending(req)
	}

	if s.Handler != nil {
		return s.Handler.OnMessage(s, req)
	}
	return nil, nil
}

// Close marks the session as closed and cleans up.
func (s *Session) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	close(s.SendChan)

	if s.Handler != nil {
		s.Handler.OnClose(s)
	}
}

// IsClosed returns true if the session has been closed.
func (s *Session) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// Set stores metadata on the session.
func (s *Session) Set(key string, value any) {
	s.metadataMu.Lock()
	defer s.metadataMu.Unlock()
	s.metadata[key] = value
}

// Get retrieves metadata from the session.
func (s *Session) Get(key string) (any, bool) {
	s.metadataMu.RLock()
	defer s.metadataMu.RUnlock()
	v, ok := s.metadata[key]
	return v, ok
}

// GetString retrieves string metadata.
func (s *Session) GetString(key string) string {
	if v, ok := s.Get(key); ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt retrieves int metadata.
func (s *Session) GetInt(key string) int {
	if v, ok := s.Get(key); ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

// Delete removes metadata.
func (s *Session) Delete(key string) {
	s.metadataMu.Lock()
	defer s.metadataMu.Unlock()
	delete(s.metadata, key)
}

func (s *Session) trackPending(req *Request) {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()
	s.pending[req.RequestID] = &pendingRequest{
		Request:   req,
		Timestamp: time.Now(),
	}
}

func (s *Session) clearPending(requestID string) {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()
	delete(s.pending, requestID)
}

// GetPendingRequest retrieves a pending request by ID.
func (s *Session) GetPendingRequest(requestID string) *Request {
	s.pendingMu.RLock()
	defer s.pendingMu.RUnlock()
	if p, ok := s.pending[requestID]; ok {
		return p.Request
	}
	return nil
}

// CleanupExpiredPending removes pending requests older than ttl.
func (s *Session) CleanupExpiredPending(ttl time.Duration) {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()

	now := time.Now()
	for id, p := range s.pending {
		if now.Sub(p.Timestamp) > ttl {
			delete(s.pending, id)
		}
	}
}
