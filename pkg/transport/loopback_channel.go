package transport

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

// LoopbackChannel wraps a real WebSocket connection to implement the Channel interface.
type LoopbackChannel struct {
	conn     *websocket.Conn
	url      string
	id       string
	incoming chan *Message
	done     chan struct{}

	metadata   map[string]any
	metadataMu sync.RWMutex

	closed    bool
	closeMu   sync.RWMutex
	closeOnce sync.Once
}

// newLoopbackChannel creates a new channel wrapping a WebSocket connection.
func newLoopbackChannel(conn *websocket.Conn, url string) *LoopbackChannel {
	ch := &LoopbackChannel{
		conn:     conn,
		url:      url,
		id:       generateChannelID(),
		incoming: make(chan *Message, 100),
		done:     make(chan struct{}),
		metadata: make(map[string]any),
	}

	// Start reader goroutine
	go ch.readLoop()

	return ch
}

// ID returns the unique channel identifier.
func (c *LoopbackChannel) ID() string {
	return c.id
}

// URL returns the connection URL.
func (c *LoopbackChannel) URL() string {
	return c.url
}

// Send sends a message through the WebSocket connection.
func (c *LoopbackChannel) Send(msg *Message) error {
	c.closeMu.RLock()
	if c.closed {
		c.closeMu.RUnlock()
		return ErrChannelClosed
	}
	c.closeMu.RUnlock()

	// Convert message to JSON and send
	data := map[string]any{
		"channel":    msg.Channel,
		"format":     msg.Format,
		"target":     msg.Target,
		"swap":       msg.Swap,
		"payload":    string(msg.Payload),
		"request_id": msg.ID,
	}

	if err := c.conn.WriteJSON(data); err != nil {
		return err
	}
	return nil
}

// Receive returns a channel for incoming messages.
func (c *LoopbackChannel) Receive() <-chan *Message {
	return c.incoming
}

// Close gracefully closes the channel.
func (c *LoopbackChannel) Close() error {
	c.closeOnce.Do(func() {
		c.closeMu.Lock()
		c.closed = true
		c.closeMu.Unlock()

		c.conn.Close()
		close(c.done)
		close(c.incoming)
	})
	return nil
}

// Done returns a channel that's closed when the channel terminates.
func (c *LoopbackChannel) Done() <-chan struct{} {
	return c.done
}

// Set stores metadata on the channel.
func (c *LoopbackChannel) Set(key string, value any) {
	c.metadataMu.Lock()
	defer c.metadataMu.Unlock()
	c.metadata[key] = value
}

// Get retrieves metadata from the channel.
func (c *LoopbackChannel) Get(key string) (any, bool) {
	c.metadataMu.RLock()
	defer c.metadataMu.RUnlock()
	v, ok := c.metadata[key]
	return v, ok
}

// SendStream sends messages from a stream with backpressure handling.
func (c *LoopbackChannel) SendStream(ctx context.Context, stream <-chan *Message) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return ErrChannelClosed
		case msg, ok := <-stream:
			if !ok {
				return nil // Stream completed
			}
			if err := c.Send(msg); err != nil {
				return err
			}
		}
	}
}

// readLoop reads messages from the WebSocket and delivers them to the incoming channel.
func (c *LoopbackChannel) readLoop() {
	defer c.Close()

	for {
		var data map[string]any
		if err := c.conn.ReadJSON(&data); err != nil {
			return
		}

		msg := &Message{
			Type:    getString(data, "type"),
			ID:      getString(data, "request_id"),
			Channel: getString(data, "channel"),
			Format:  getString(data, "format"),
			Target:  getString(data, "target"),
			Swap:    getString(data, "swap"),
		}

		if payload, ok := data["payload"].(string); ok {
			msg.Payload = []byte(payload)
		}

		if headers, ok := data["headers"].(map[string]any); ok {
			msg.Headers = make(map[string]string)
			for k, v := range headers {
				if s, ok := v.(string); ok {
					msg.Headers[k] = s
				}
			}
		}

		if values, ok := data["values"].(map[string]any); ok {
			msg.Values = values
		}

		select {
		case c.incoming <- msg:
		case <-c.done:
			return
		default:
			// Buffer full, drop message
		}
	}
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

var channelCounter uint64
var channelCounterMu sync.Mutex

func generateChannelID() string {
	channelCounterMu.Lock()
	channelCounter++
	id := channelCounter
	channelCounterMu.Unlock()
	return "ch_" + itoa64(id)
}

func itoa64(n uint64) string {
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

// Verify LoopbackChannel implements StreamingChannel
var _ StreamingChannel = (*LoopbackChannel)(nil)
