package hosting

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHubManager(t *testing.T) {
	// Get hub for a site
	hub1 := GetHub("site1")
	if hub1 == nil {
		t.Fatal("GetHub() returned nil")
	}

	// Get same hub again
	hub1Again := GetHub("site1")
	if hub1 != hub1Again {
		t.Error("GetHub() should return same hub for same site")
	}

	// Get different hub for different site
	hub2 := GetHub("site2")
	if hub1 == hub2 {
		t.Error("GetHub() should return different hubs for different sites")
	}
}

func TestHubClientCount(t *testing.T) {
	hub := GetHub("test-count")

	// Initially should be 0
	if count := hub.ClientCount(); count != 0 {
		t.Errorf("ClientCount() = %d, want 0", count)
	}
}

func TestHubBroadcast(t *testing.T) {
	hub := GetHub("test-broadcast")

	// Broadcast should not block even with no clients
	done := make(chan bool, 1)
	go func() {
		hub.Broadcast("test message")
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Broadcast() blocked with no clients")
	}
}

func TestSiteIsolation(t *testing.T) {
	// Each site should have its own hub
	sites := []string{"isolated1", "isolated2", "isolated3"}
	hubs := make([]*SiteHub, len(sites))

	for i, site := range sites {
		hubs[i] = GetHub(site)
		if hubs[i].siteID != site {
			t.Errorf("hub.siteID = %q, want %q", hubs[i].siteID, site)
		}
	}

	// Verify they're all different
	for i := 0; i < len(hubs); i++ {
		for j := i + 1; j < len(hubs); j++ {
			if hubs[i] == hubs[j] {
				t.Errorf("hubs[%d] == hubs[%d], should be different", i, j)
			}
		}
	}
}

// createTestClient creates a mock client for testing
func createTestClient(hub *SiteHub, id string) *Client {
	return &Client{
		ID:          id,
		Hub:         hub,
		Channels:    make(map[string]bool),
		Send:        make(chan []byte, 256),
		ConnectedAt: time.Now(),
	}
}

func TestChannelSubscription(t *testing.T) {
	hub := GetHub("test-channels")

	// Create test client
	client := createTestClient(hub, "client1")

	// Register client
	hub.mu.Lock()
	hub.clients[client.ID] = client
	hub.mu.Unlock()

	// Subscribe to channel
	hub.subscribe(client, "chat")

	// Verify client is subscribed
	client.mu.RLock()
	if !client.Channels["chat"] {
		t.Error("Client should be subscribed to 'chat' channel")
	}
	client.mu.RUnlock()

	// Verify hub has channel
	if count := hub.ChannelCount("chat"); count != 1 {
		t.Errorf("ChannelCount('chat') = %d, want 1", count)
	}

	// Verify GetSubscribers
	subs := hub.GetSubscribers("chat")
	if len(subs) != 1 || subs[0] != "client1" {
		t.Errorf("GetSubscribers('chat') = %v, want ['client1']", subs)
	}

	// Unsubscribe
	hub.unsubscribe(client, "chat")

	// Verify unsubscribed
	client.mu.RLock()
	if client.Channels["chat"] {
		t.Error("Client should not be subscribed to 'chat' after unsubscribe")
	}
	client.mu.RUnlock()

	if count := hub.ChannelCount("chat"); count != 0 {
		t.Errorf("ChannelCount('chat') after unsubscribe = %d, want 0", count)
	}

	// Cleanup
	hub.mu.Lock()
	delete(hub.clients, client.ID)
	hub.mu.Unlock()
}

func TestChannelIsolation(t *testing.T) {
	hub := GetHub("test-channel-isolation")

	// Create two clients
	client1 := createTestClient(hub, "iso-client1")
	client2 := createTestClient(hub, "iso-client2")

	// Register clients
	hub.mu.Lock()
	hub.clients[client1.ID] = client1
	hub.clients[client2.ID] = client2
	hub.mu.Unlock()

	// Subscribe client1 to channelA, client2 to channelB
	hub.subscribe(client1, "channelA")
	hub.subscribe(client2, "channelB")

	// Broadcast to channelA
	hub.BroadcastToChannel("channelA", map[string]string{"msg": "hello A"})

	// Give time for messages to be sent
	time.Sleep(10 * time.Millisecond)

	// Client1 should have received the message
	select {
	case msg := <-client1.Send:
		var parsed OutboundMessage
		if err := json.Unmarshal(msg, &parsed); err != nil {
			t.Errorf("Failed to parse message: %v", err)
		}
		if parsed.Channel != "channelA" {
			t.Errorf("Expected channel 'channelA', got %q", parsed.Channel)
		}
	default:
		t.Error("Client1 should have received message for channelA")
	}

	// Client2 should NOT have received any message
	select {
	case msg := <-client2.Send:
		t.Errorf("Client2 should not receive messages for channelA, got: %s", msg)
	default:
		// Expected - no message
	}

	// Cleanup
	hub.mu.Lock()
	delete(hub.clients, client1.ID)
	delete(hub.clients, client2.ID)
	hub.channels = make(map[string]map[string]bool)
	hub.mu.Unlock()
}

func TestBroadcastToChannel(t *testing.T) {
	hub := GetHub("test-broadcast-channel")

	// Create clients
	client1 := createTestClient(hub, "bc-client1")
	client2 := createTestClient(hub, "bc-client2")

	// Register and subscribe both to same channel
	hub.mu.Lock()
	hub.clients[client1.ID] = client1
	hub.clients[client2.ID] = client2
	hub.mu.Unlock()

	hub.subscribe(client1, "room1")
	hub.subscribe(client2, "room1")

	// Broadcast
	testData := map[string]interface{}{"action": "test", "value": 42}
	hub.BroadcastToChannel("room1", testData)

	// Both clients should receive
	for _, client := range []*Client{client1, client2} {
		select {
		case msg := <-client.Send:
			var parsed OutboundMessage
			if err := json.Unmarshal(msg, &parsed); err != nil {
				t.Errorf("Failed to parse message for %s: %v", client.ID, err)
			}
			if parsed.Type != "message" {
				t.Errorf("Expected type 'message', got %q", parsed.Type)
			}
			if parsed.Channel != "room1" {
				t.Errorf("Expected channel 'room1', got %q", parsed.Channel)
			}
			if parsed.Timestamp == 0 {
				t.Error("Expected timestamp to be set")
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Client %s did not receive message", client.ID)
		}
	}

	// Cleanup
	hub.mu.Lock()
	delete(hub.clients, client1.ID)
	delete(hub.clients, client2.ID)
	hub.channels = make(map[string]map[string]bool)
	hub.mu.Unlock()
}

func TestBroadcastAll(t *testing.T) {
	hub := GetHub("test-broadcast-all")

	// Create clients (not subscribed to any channel)
	client1 := createTestClient(hub, "ba-client1")
	client2 := createTestClient(hub, "ba-client2")

	// Register clients
	hub.mu.Lock()
	hub.clients[client1.ID] = client1
	hub.clients[client2.ID] = client2
	hub.mu.Unlock()

	// BroadcastAll
	testData := map[string]string{"announcement": "server restart"}
	hub.BroadcastAll(testData)

	// Both clients should receive regardless of channel subscription
	for _, client := range []*Client{client1, client2} {
		select {
		case msg := <-client.Send:
			var parsed OutboundMessage
			if err := json.Unmarshal(msg, &parsed); err != nil {
				t.Errorf("Failed to parse message for %s: %v", client.ID, err)
			}
			if parsed.Type != "message" {
				t.Errorf("Expected type 'message', got %q", parsed.Type)
			}
			// BroadcastAll should not have a channel
			if parsed.Channel != "" {
				t.Errorf("BroadcastAll message should not have channel, got %q", parsed.Channel)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Client %s did not receive message", client.ID)
		}
	}

	// Cleanup
	hub.mu.Lock()
	delete(hub.clients, client1.ID)
	delete(hub.clients, client2.ID)
	hub.mu.Unlock()
}

func TestGetSubscribersEmpty(t *testing.T) {
	hub := GetHub("test-subscribers-empty")

	// Non-existent channel should return empty slice
	subs := hub.GetSubscribers("nonexistent")
	if subs == nil {
		t.Error("GetSubscribers should return empty slice, not nil")
	}
	if len(subs) != 0 {
		t.Errorf("GetSubscribers('nonexistent') = %v, want empty slice", subs)
	}
}

func TestChannelCountEmpty(t *testing.T) {
	hub := GetHub("test-channel-count-empty")

	// Non-existent channel should return 0
	count := hub.ChannelCount("nonexistent")
	if count != 0 {
		t.Errorf("ChannelCount('nonexistent') = %d, want 0", count)
	}
}

func TestClientUnregisterCleansUpChannels(t *testing.T) {
	hub := GetHub("test-unregister-cleanup")

	// Create and register client
	client := createTestClient(hub, "cleanup-client")
	hub.mu.Lock()
	hub.clients[client.ID] = client
	hub.mu.Unlock()

	// Subscribe to multiple channels
	hub.subscribe(client, "channel1")
	hub.subscribe(client, "channel2")
	hub.subscribe(client, "channel3")

	// Verify subscribed
	if count := len(hub.channels); count != 3 {
		t.Errorf("Expected 3 channels, got %d", count)
	}

	// Simulate unregister by running the cleanup logic
	hub.mu.Lock()
	for channel := range client.Channels {
		if subs, exists := hub.channels[channel]; exists {
			delete(subs, client.ID)
			if len(subs) == 0 {
				delete(hub.channels, channel)
			}
		}
	}
	delete(hub.clients, client.ID)
	hub.mu.Unlock()

	// All channels should be cleaned up
	if count := len(hub.channels); count != 0 {
		t.Errorf("Expected 0 channels after unregister, got %d", count)
	}
}

func TestMessageProtocol(t *testing.T) {
	// Test OutboundMessage JSON marshaling
	msg := OutboundMessage{
		Type:      "message",
		Channel:   "chat",
		Data:      map[string]interface{}{"text": "hello"},
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal OutboundMessage: %v", err)
	}

	var parsed OutboundMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal OutboundMessage: %v", err)
	}

	if parsed.Type != "message" {
		t.Errorf("Type = %q, want 'message'", parsed.Type)
	}
	if parsed.Channel != "chat" {
		t.Errorf("Channel = %q, want 'chat'", parsed.Channel)
	}
}

func TestInboundMessageParsing(t *testing.T) {
	tests := []struct {
		input   string
		want    InboundMessage
		wantErr bool
	}{
		{
			input:   `{"type":"subscribe","channel":"chat"}`,
			want:    InboundMessage{Type: "subscribe", Channel: "chat"},
			wantErr: false,
		},
		{
			input:   `{"type":"unsubscribe","channel":"chat"}`,
			want:    InboundMessage{Type: "unsubscribe", Channel: "chat"},
			wantErr: false,
		},
		{
			input:   `{"type":"pong"}`,
			want:    InboundMessage{Type: "pong"},
			wantErr: false,
		},
		{
			input:   `invalid json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var msg InboundMessage
		err := json.Unmarshal([]byte(tt.input), &msg)

		if tt.wantErr {
			if err == nil {
				t.Errorf("Expected error for input %q", tt.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for input %q: %v", tt.input, err)
			continue
		}

		if msg.Type != tt.want.Type {
			t.Errorf("Type = %q, want %q", msg.Type, tt.want.Type)
		}
		if msg.Channel != tt.want.Channel {
			t.Errorf("Channel = %q, want %q", msg.Channel, tt.want.Channel)
		}
	}
}

func TestKickClientNonExistent(t *testing.T) {
	hub := GetHub("test-kick-nonexistent")

	// Kicking non-existent client should return false
	result := hub.KickClient("nonexistent-client", "test reason")
	if result {
		t.Error("KickClient should return false for non-existent client")
	}
}

func TestMultipleChannelSubscriptions(t *testing.T) {
	hub := GetHub("test-multi-channel")

	client := createTestClient(hub, "multi-client")
	hub.mu.Lock()
	hub.clients[client.ID] = client
	hub.mu.Unlock()

	// Subscribe to multiple channels
	channels := []string{"general", "random", "dev", "support"}
	for _, ch := range channels {
		hub.subscribe(client, ch)
	}

	// Verify all subscriptions
	client.mu.RLock()
	for _, ch := range channels {
		if !client.Channels[ch] {
			t.Errorf("Client should be subscribed to %q", ch)
		}
	}
	client.mu.RUnlock()

	// Verify each channel has the client
	for _, ch := range channels {
		if count := hub.ChannelCount(ch); count != 1 {
			t.Errorf("ChannelCount(%q) = %d, want 1", ch, count)
		}
	}

	// Cleanup
	hub.mu.Lock()
	for _, ch := range channels {
		delete(hub.channels, ch)
	}
	delete(hub.clients, client.ID)
	hub.mu.Unlock()
}

func TestGenerateClientID(t *testing.T) {
	// Generate multiple IDs and verify uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateClientID()
		if len(id) != 16 { // 8 bytes = 16 hex chars
			t.Errorf("generateClientID() = %q, want 16 chars", id)
		}
		if ids[id] {
			t.Errorf("generateClientID() produced duplicate: %s", id)
		}
		ids[id] = true
	}
}
