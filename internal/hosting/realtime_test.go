package hosting

import (
	"testing"

	"github.com/dop251/goja"
)

func TestInjectRealtimeNamespace(t *testing.T) {
	vm := goja.New()
	siteID := "test-realtime-site"

	err := InjectRealtimeNamespace(vm, siteID)
	if err != nil {
		t.Fatalf("InjectRealtimeNamespace() error = %v", err)
	}

	// Check that fazt.realtime exists
	val, err := vm.RunString("typeof fazt.realtime")
	if err != nil {
		t.Fatalf("Failed to access fazt.realtime: %v", err)
	}
	if val.String() != "object" {
		t.Errorf("fazt.realtime should be object, got %s", val.String())
	}

	// Check that methods exist
	methods := []string{"broadcast", "broadcastAll", "subscribers", "count", "kick"}
	for _, method := range methods {
		val, err := vm.RunString("typeof fazt.realtime." + method)
		if err != nil {
			t.Errorf("Failed to access fazt.realtime.%s: %v", method, err)
			continue
		}
		if val.String() != "function" {
			t.Errorf("fazt.realtime.%s should be function, got %s", method, val.String())
		}
	}
}

func TestRealtimeBroadcast(t *testing.T) {
	vm := goja.New()
	siteID := "test-rt-broadcast"

	// Create a client to receive messages
	hub := GetHub(siteID)
	client := createTestClient(hub, "rt-test-client")
	hub.mu.Lock()
	hub.clients[client.ID] = client
	hub.mu.Unlock()
	hub.subscribe(client, "test-channel")

	// Inject realtime namespace
	err := InjectRealtimeNamespace(vm, siteID)
	if err != nil {
		t.Fatalf("InjectRealtimeNamespace() error = %v", err)
	}

	// Call broadcast
	_, err = vm.RunString(`fazt.realtime.broadcast("test-channel", { msg: "hello from JS" })`)
	if err != nil {
		t.Fatalf("fazt.realtime.broadcast() error = %v", err)
	}

	// Verify client received message
	select {
	case msg := <-client.Send:
		if len(msg) == 0 {
			t.Error("Received empty message")
		}
	default:
		t.Error("Client did not receive message")
	}

	// Cleanup
	hub.mu.Lock()
	delete(hub.clients, client.ID)
	delete(hub.channels, "test-channel")
	hub.mu.Unlock()
}

func TestRealtimeBroadcastAll(t *testing.T) {
	vm := goja.New()
	siteID := "test-rt-broadcast-all"

	// Create clients (no channel subscription needed)
	hub := GetHub(siteID)
	client1 := createTestClient(hub, "rt-all-client1")
	client2 := createTestClient(hub, "rt-all-client2")
	hub.mu.Lock()
	hub.clients[client1.ID] = client1
	hub.clients[client2.ID] = client2
	hub.mu.Unlock()

	// Inject realtime namespace
	err := InjectRealtimeNamespace(vm, siteID)
	if err != nil {
		t.Fatalf("InjectRealtimeNamespace() error = %v", err)
	}

	// Call broadcastAll
	_, err = vm.RunString(`fazt.realtime.broadcastAll({ announcement: "shutdown" })`)
	if err != nil {
		t.Fatalf("fazt.realtime.broadcastAll() error = %v", err)
	}

	// Verify both clients received message
	for _, client := range []*Client{client1, client2} {
		select {
		case msg := <-client.Send:
			if len(msg) == 0 {
				t.Errorf("Client %s received empty message", client.ID)
			}
		default:
			t.Errorf("Client %s did not receive message", client.ID)
		}
	}

	// Cleanup
	hub.mu.Lock()
	delete(hub.clients, client1.ID)
	delete(hub.clients, client2.ID)
	hub.mu.Unlock()
}

func TestRealtimeSubscribers(t *testing.T) {
	vm := goja.New()
	siteID := "test-rt-subscribers"

	// Create and subscribe clients
	hub := GetHub(siteID)
	client1 := createTestClient(hub, "rt-sub-client1")
	client2 := createTestClient(hub, "rt-sub-client2")
	hub.mu.Lock()
	hub.clients[client1.ID] = client1
	hub.clients[client2.ID] = client2
	hub.mu.Unlock()
	hub.subscribe(client1, "sub-channel")
	hub.subscribe(client2, "sub-channel")

	// Inject realtime namespace
	err := InjectRealtimeNamespace(vm, siteID)
	if err != nil {
		t.Fatalf("InjectRealtimeNamespace() error = %v", err)
	}

	// Call subscribers
	val, err := vm.RunString(`fazt.realtime.subscribers("sub-channel")`)
	if err != nil {
		t.Fatalf("fazt.realtime.subscribers() error = %v", err)
	}

	// Should return array of client IDs
	exported := val.Export()
	switch subs := exported.(type) {
	case []interface{}:
		if len(subs) != 2 {
			t.Errorf("Expected 2 subscribers, got %d", len(subs))
		}
	case []string:
		if len(subs) != 2 {
			t.Errorf("Expected 2 subscribers, got %d", len(subs))
		}
	default:
		t.Fatalf("Expected array, got %T", exported)
	}

	// Cleanup
	hub.mu.Lock()
	delete(hub.clients, client1.ID)
	delete(hub.clients, client2.ID)
	delete(hub.channels, "sub-channel")
	hub.mu.Unlock()
}

func TestRealtimeCount(t *testing.T) {
	vm := goja.New()
	siteID := "test-rt-count"

	// Create hub with clients
	hub := GetHub(siteID)
	client1 := createTestClient(hub, "rt-count-client1")
	client2 := createTestClient(hub, "rt-count-client2")
	hub.mu.Lock()
	hub.clients[client1.ID] = client1
	hub.clients[client2.ID] = client2
	hub.mu.Unlock()
	hub.subscribe(client1, "count-channel")

	// Inject realtime namespace
	err := InjectRealtimeNamespace(vm, siteID)
	if err != nil {
		t.Fatalf("InjectRealtimeNamespace() error = %v", err)
	}

	// Count without channel (total clients)
	val, err := vm.RunString(`fazt.realtime.count()`)
	if err != nil {
		t.Fatalf("fazt.realtime.count() error = %v", err)
	}
	if val.ToInteger() != 2 {
		t.Errorf("Expected total count 2, got %d", val.ToInteger())
	}

	// Count with channel
	val, err = vm.RunString(`fazt.realtime.count("count-channel")`)
	if err != nil {
		t.Fatalf("fazt.realtime.count('count-channel') error = %v", err)
	}
	if val.ToInteger() != 1 {
		t.Errorf("Expected channel count 1, got %d", val.ToInteger())
	}

	// Cleanup
	hub.mu.Lock()
	delete(hub.clients, client1.ID)
	delete(hub.clients, client2.ID)
	delete(hub.channels, "count-channel")
	hub.mu.Unlock()
}

func TestRealtimeBroadcastMissingArgs(t *testing.T) {
	vm := goja.New()
	siteID := "test-rt-missing-args"

	err := InjectRealtimeNamespace(vm, siteID)
	if err != nil {
		t.Fatalf("InjectRealtimeNamespace() error = %v", err)
	}

	// broadcast without args should panic/error
	_, err = vm.RunString(`
		try {
			fazt.realtime.broadcast();
			"no-error";
		} catch (e) {
			"error";
		}
	`)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
