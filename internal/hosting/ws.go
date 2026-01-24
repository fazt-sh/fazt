package hosting

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 10 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = 30 * time.Second

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// checkOrigin validates WebSocket origin against allowed patterns
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // Allow connections without Origin header (non-browser clients)
	}

	// Allow localhost for development
	if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
		return true
	}

	// Allow same-host connections
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	if strings.Contains(origin, host) {
		return true
	}

	log.Printf("[WS] Rejected origin: %s (host: %s)", origin, r.Host)
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

// InboundMessage represents messages from the client
type InboundMessage struct {
	Type    string `json:"type"`              // subscribe, unsubscribe, pong
	Channel string `json:"channel,omitempty"` // channel name for subscribe/unsubscribe
}

// OutboundMessage represents messages to the client
type OutboundMessage struct {
	Type      string      `json:"type"`                // subscribed, unsubscribed, message, ping, error
	Channel   string      `json:"channel,omitempty"`   // channel for subscribed/unsubscribed/message
	Data      interface{} `json:"data,omitempty"`      // payload for message type
	Timestamp int64       `json:"timestamp,omitempty"` // unix millis for message type
	Error     string      `json:"error,omitempty"`     // error message
}

// Client represents a WebSocket connection
type Client struct {
	ID          string
	Conn        *websocket.Conn
	Hub         *SiteHub
	Channels    map[string]bool
	Send        chan []byte
	ConnectedAt time.Time
	mu          sync.RWMutex
}

// SiteHub manages WebSocket connections for a single site
type SiteHub struct {
	siteID     string
	clients    map[string]*Client          // clientID -> Client
	channels   map[string]map[string]bool  // channel -> clientIDs
	broadcast  chan []byte                 // broadcast to all clients (legacy)
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
	mu         sync.RWMutex
}

// HubManager manages hubs for all sites
type HubManager struct {
	hubs map[string]*SiteHub
	mu   sync.RWMutex
}

var hubManager = &HubManager{
	hubs: make(map[string]*SiteHub),
}

// GetHub returns the hub for a site, creating one if needed
func GetHub(siteID string) *SiteHub {
	hubManager.mu.Lock()
	defer hubManager.mu.Unlock()

	if hub, exists := hubManager.hubs[siteID]; exists {
		return hub
	}

	hub := &SiteHub{
		siteID:     siteID,
		clients:    make(map[string]*Client),
		channels:   make(map[string]map[string]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}

	hubManager.hubs[siteID] = hub
	go hub.run()

	return hub
}

// RemoveHub stops and removes a hub for a site (call when site is deleted)
func RemoveHub(siteID string) {
	hubManager.mu.Lock()
	defer hubManager.mu.Unlock()

	if hub, exists := hubManager.hubs[siteID]; exists {
		hub.Stop()
		delete(hubManager.hubs, siteID)
		log.Printf("[WS:%s] Hub removed", siteID)
	}
}

// run handles the hub's event loop
func (h *SiteHub) run() {
	for {
		select {
		case <-h.done:
			// Shutdown: close all client connections
			h.mu.Lock()
			for _, client := range h.clients {
				close(client.Send)
				client.Conn.Close()
			}
			h.clients = make(map[string]*Client)
			h.channels = make(map[string]map[string]bool)
			h.mu.Unlock()
			log.Printf("[WS:%s] Hub shutdown complete", h.siteID)
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("[WS:%s] Client %s connected (%d total)", h.siteID, client.ID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				// Remove from all channels
				for channel := range client.Channels {
					if subs, exists := h.channels[channel]; exists {
						delete(subs, client.ID)
						if len(subs) == 0 {
							delete(h.channels, channel)
						}
					}
				}
				close(client.Send)
				delete(h.clients, client.ID)
			}
			h.mu.Unlock()
			log.Printf("[WS:%s] Client %s disconnected (%d remaining)", h.siteID, client.ID, len(h.clients))

		case message := <-h.broadcast:
			// Legacy: broadcast to all clients
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Channel full, skip this client
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Stop signals the hub to shutdown
func (h *SiteHub) Stop() {
	close(h.done)
}

// Broadcast sends a message to all connected clients (legacy)
func (h *SiteHub) Broadcast(message string) {
	select {
	case h.broadcast <- []byte(message):
	default:
		// Channel full, drop message
		log.Printf("[WS:%s] Broadcast channel full, dropping message", h.siteID)
	}
}

// BroadcastAll sends data to all connected clients with proper message format
func (h *SiteHub) BroadcastAll(data interface{}) {
	msg := OutboundMessage{
		Type:      "message",
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WS:%s] Failed to marshal broadcast message: %v", h.siteID, err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- payload:
		default:
			// Channel full, skip
		}
	}
}

// BroadcastToChannel sends data to all clients subscribed to a channel
func (h *SiteHub) BroadcastToChannel(channel string, data interface{}) {
	msg := OutboundMessage{
		Type:      "message",
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WS:%s] Failed to marshal channel message: %v", h.siteID, err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	subscribers, exists := h.channels[channel]
	if !exists {
		return
	}

	for clientID := range subscribers {
		if client, ok := h.clients[clientID]; ok {
			select {
			case client.Send <- payload:
			default:
				// Channel full, skip
			}
		}
	}
}

// GetSubscribers returns all client IDs subscribed to a channel
func (h *SiteHub) GetSubscribers(channel string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subscribers, exists := h.channels[channel]
	if !exists {
		return []string{}
	}

	result := make([]string, 0, len(subscribers))
	for clientID := range subscribers {
		result = append(result, clientID)
	}
	return result
}

// ChannelCount returns the number of subscribers in a channel
func (h *SiteHub) ChannelCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if subscribers, exists := h.channels[channel]; exists {
		return len(subscribers)
	}
	return 0
}

// ClientCount returns the number of connected clients
func (h *SiteHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// KickClient disconnects a client with an optional reason
func (h *SiteHub) KickClient(clientID string, reason string) bool {
	h.mu.RLock()
	client, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	// Send error message before closing
	if reason != "" {
		msg := OutboundMessage{
			Type:  "error",
			Error: reason,
		}
		if payload, err := json.Marshal(msg); err == nil {
			select {
			case client.Send <- payload:
			default:
			}
		}
	}

	// Close connection - this will trigger unregister via readPump
	client.Conn.Close()
	return true
}

// subscribe adds a client to a channel
func (h *SiteHub) subscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to hub's channel map
	if _, exists := h.channels[channel]; !exists {
		h.channels[channel] = make(map[string]bool)
	}
	h.channels[channel][client.ID] = true

	// Add to client's channel set
	client.mu.Lock()
	client.Channels[channel] = true
	client.mu.Unlock()

	log.Printf("[WS:%s] Client %s subscribed to %s", h.siteID, client.ID, channel)
}

// unsubscribe removes a client from a channel
func (h *SiteHub) unsubscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, exists := h.channels[channel]; exists {
		delete(subs, client.ID)
		if len(subs) == 0 {
			delete(h.channels, channel)
		}
	}

	client.mu.Lock()
	delete(client.Channels, channel)
	client.mu.Unlock()

	log.Printf("[WS:%s] Client %s unsubscribed from %s", h.siteID, client.ID, channel)
}

// generateClientID creates a random client ID
func generateClientID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// HandleWebSocket upgrades HTTP connections to WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request, siteID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS:%s] Upgrade error: %v", siteID, err)
		return
	}

	hub := GetHub(siteID)

	client := &Client{
		ID:          generateClientID(),
		Conn:        conn,
		Hub:         hub,
		Channels:    make(map[string]bool),
		Send:        make(chan []byte, 256),
		ConnectedAt: time.Now(),
	}

	hub.register <- client

	// Start goroutines for read/write pumps
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait + pingPeriod))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait + pingPeriod))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS:%s] Client %s read error: %v", c.Hub.siteID, c.ID, err)
			}
			break
		}

		// Parse inbound message
		var msg InboundMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.sendError("Invalid JSON message")
			continue
		}

		// Handle message types
		switch msg.Type {
		case "subscribe":
			if msg.Channel == "" {
				c.sendError("Channel required for subscribe")
				continue
			}
			c.Hub.subscribe(c, msg.Channel)
			c.sendJSON(OutboundMessage{Type: "subscribed", Channel: msg.Channel})

		case "unsubscribe":
			if msg.Channel == "" {
				c.sendError("Channel required for unsubscribe")
				continue
			}
			c.Hub.unsubscribe(c, msg.Channel)
			c.sendJSON(OutboundMessage{Type: "unsubscribed", Channel: msg.Channel})

		case "pong":
			// Response to our ping - already handled by SetPongHandler for websocket pings
			// This handles application-level pong for clients that can't send WS pongs
			c.Conn.SetReadDeadline(time.Now().Add(pongWait + pingPeriod))

		default:
			c.sendError("Unknown message type: " + msg.Type)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Write any queued messages
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// Send application-level ping
			c.sendJSON(OutboundMessage{Type: "ping"})

			// Also send WebSocket ping for clients that support it
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendJSON sends a JSON message to the client
func (c *Client) sendJSON(msg OutboundMessage) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.Send <- payload:
	default:
		// Channel full
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	c.sendJSON(OutboundMessage{Type: "error", Error: errMsg})
}
