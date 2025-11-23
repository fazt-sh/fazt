package hosting

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

// SiteHub manages WebSocket connections for a single site
type SiteHub struct {
	siteID     string
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
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
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}

	hubManager.hubs[siteID] = hub
	go hub.run()

	return hub
}

// run handles the hub's event loop
func (h *SiteHub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
			log.Printf("[WS:%s] Client connected (%d total)", h.siteID, len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("[WS:%s] Client disconnected (%d remaining)", h.siteID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *SiteHub) Broadcast(message string) {
	select {
	case h.broadcast <- []byte(message):
	default:
		// Channel full, drop message
		log.Printf("[WS:%s] Broadcast channel full, dropping message", h.siteID)
	}
}

// ClientCount returns the number of connected clients
func (h *SiteHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HandleWebSocket upgrades HTTP connections to WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request, siteID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS:%s] Upgrade error: %v", siteID, err)
		return
	}

	hub := GetHub(siteID)
	hub.register <- conn

	// Read loop - handle client messages and disconnection
	go func() {
		defer func() {
			hub.unregister <- conn
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			// Echo messages to all clients (broadcast)
			hub.Broadcast(string(message))
		}
	}()
}
