package hosting

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// StressTestConfig defines parameters for stress testing
type StressTestConfig struct {
	NumClients       int
	NumChannels      int
	MessagesPerTest  int
	SubscribersPerCh int
}

func TestStressConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	configs := []struct {
		name       string
		numClients int
	}{
		{"100 clients", 100},
		{"500 clients", 500},
		{"1000 clients", 1000},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			hub := GetHub(fmt.Sprintf("stress-conn-%d", cfg.numClients))
			defer func() {
				hub.mu.Lock()
				hub.clients = make(map[string]*Client)
				hub.channels = make(map[string]map[string]bool)
				hub.mu.Unlock()
			}()

			var wg sync.WaitGroup
			var connected int64
			start := time.Now()

			// Spawn clients concurrently
			for i := 0; i < cfg.numClients; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					client := createTestClient(hub, fmt.Sprintf("stress-client-%d", id))
					hub.mu.Lock()
					hub.clients[client.ID] = client
					hub.mu.Unlock()
					atomic.AddInt64(&connected, 1)
				}(i)
			}

			wg.Wait()
			elapsed := time.Since(start)

			if hub.ClientCount() != cfg.numClients {
				t.Errorf("Expected %d clients, got %d", cfg.numClients, hub.ClientCount())
			}

			t.Logf("Connected %d clients in %v (%.0f/sec)",
				cfg.numClients, elapsed, float64(cfg.numClients)/elapsed.Seconds())
		})
	}
}

func TestStressChannelSubscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	hub := GetHub("stress-subscribe")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	numClients := 100
	numChannels := 50

	// Create clients
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = createTestClient(hub, fmt.Sprintf("sub-client-%d", i))
		hub.mu.Lock()
		hub.clients[clients[i].ID] = clients[i]
		hub.mu.Unlock()
	}

	// Subscribe each client to multiple channels concurrently
	var wg sync.WaitGroup
	start := time.Now()

	for _, client := range clients {
		for j := 0; j < numChannels; j++ {
			wg.Add(1)
			go func(c *Client, ch string) {
				defer wg.Done()
				hub.subscribe(c, ch)
			}(client, fmt.Sprintf("channel-%d", j))
		}
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalSubs := numClients * numChannels
	t.Logf("Completed %d subscriptions in %v (%.0f/sec)",
		totalSubs, elapsed, float64(totalSubs)/elapsed.Seconds())

	// Verify counts
	for j := 0; j < numChannels; j++ {
		ch := fmt.Sprintf("channel-%d", j)
		if count := hub.ChannelCount(ch); count != numClients {
			t.Errorf("Channel %s has %d subscribers, expected %d", ch, count, numClients)
		}
	}
}

func TestStressMessageThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	hub := GetHub("stress-throughput")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	numClients := 100
	numMessages := 1000
	channel := "throughput-test"

	// Create and subscribe clients
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = createTestClient(hub, fmt.Sprintf("tp-client-%d", i))
		hub.mu.Lock()
		hub.clients[clients[i].ID] = clients[i]
		hub.mu.Unlock()
		hub.subscribe(clients[i], channel)
	}

	// Track received messages
	var received int64

	// Start receivers
	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			for {
				select {
				case <-c.Send:
					atomic.AddInt64(&received, 1)
				case <-time.After(500 * time.Millisecond):
					return
				}
			}
		}(client)
	}

	// Broadcast messages
	start := time.Now()
	for i := 0; i < numMessages; i++ {
		hub.BroadcastToChannel(channel, map[string]interface{}{
			"seq":  i,
			"data": "test payload",
		})
	}

	// Wait for receivers with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Log("Timeout waiting for receivers")
	}

	elapsed := time.Since(start)
	expectedTotal := int64(numMessages * numClients)

	t.Logf("Sent %d messages to %d clients in %v", numMessages, numClients, elapsed)
	t.Logf("Total messages delivered: %d/%d (%.1f%%)",
		received, expectedTotal, float64(received)/float64(expectedTotal)*100)
	t.Logf("Throughput: %.0f messages/sec", float64(received)/elapsed.Seconds())

	// Allow some message loss due to timing and buffer limits
	// In high-throughput scenarios, 85% delivery is acceptable
	minExpected := int64(float64(expectedTotal) * 0.85)
	if received < minExpected {
		t.Errorf("Too many messages lost: received %d, expected at least %d", received, minExpected)
	}
}

func TestStressChannelFanout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	configs := []struct {
		name        string
		subscribers int
	}{
		{"10 subscribers", 10},
		{"100 subscribers", 100},
		{"500 subscribers", 500},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			hub := GetHub(fmt.Sprintf("stress-fanout-%d", cfg.subscribers))
			defer func() {
				hub.mu.Lock()
				hub.clients = make(map[string]*Client)
				hub.channels = make(map[string]map[string]bool)
				hub.mu.Unlock()
			}()

			channel := "fanout"

			// Create and subscribe clients
			clients := make([]*Client, cfg.subscribers)
			for i := 0; i < cfg.subscribers; i++ {
				clients[i] = createTestClient(hub, fmt.Sprintf("fanout-client-%d", i))
				hub.mu.Lock()
				hub.clients[clients[i].ID] = clients[i]
				hub.mu.Unlock()
				hub.subscribe(clients[i], channel)
			}

			// Measure single broadcast latency
			numBroadcasts := 100
			start := time.Now()

			for i := 0; i < numBroadcasts; i++ {
				hub.BroadcastToChannel(channel, map[string]interface{}{
					"msg": i,
				})
			}

			elapsed := time.Since(start)
			avgLatency := elapsed / time.Duration(numBroadcasts)

			t.Logf("Fanout to %d subscribers: %d broadcasts in %v (avg %v/broadcast)",
				cfg.subscribers, numBroadcasts, elapsed, avgLatency)

			// Drain client channels
			for _, client := range clients {
			drain:
				for {
					select {
					case <-client.Send:
					default:
						break drain
					}
				}
			}
		})
	}
}

func TestStressConcurrentBroadcasts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	hub := GetHub("stress-concurrent-broadcast")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	numClients := 50
	numChannels := 10
	numBroadcasters := 10
	messagesPerBroadcaster := 100

	// Create clients and distribute across channels
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = createTestClient(hub, fmt.Sprintf("conc-client-%d", i))
		hub.mu.Lock()
		hub.clients[clients[i].ID] = clients[i]
		hub.mu.Unlock()
		// Subscribe to a subset of channels
		for j := 0; j < 3; j++ {
			ch := fmt.Sprintf("conc-channel-%d", (i+j)%numChannels)
			hub.subscribe(clients[i], ch)
		}
	}

	// Track total messages
	var sent int64
	var wg sync.WaitGroup

	start := time.Now()

	// Start concurrent broadcasters
	for b := 0; b < numBroadcasters; b++ {
		wg.Add(1)
		go func(broadcasterID int) {
			defer wg.Done()
			for m := 0; m < messagesPerBroadcaster; m++ {
				ch := fmt.Sprintf("conc-channel-%d", m%numChannels)
				hub.BroadcastToChannel(ch, map[string]interface{}{
					"from": broadcasterID,
					"seq":  m,
				})
				atomic.AddInt64(&sent, 1)
			}
		}(b)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Concurrent broadcasts: %d messages from %d broadcasters in %v (%.0f/sec)",
		sent, numBroadcasters, elapsed, float64(sent)/elapsed.Seconds())
}

func TestStressMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	hub := GetHub("stress-memory")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	numClients := 1000
	numChannels := 100

	// Create clients
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = createTestClient(hub, fmt.Sprintf("mem-client-%d", i))
		hub.mu.Lock()
		hub.clients[clients[i].ID] = clients[i]
		hub.mu.Unlock()
	}

	// Subscribe to channels
	for i, client := range clients {
		for j := 0; j < 10; j++ {
			ch := fmt.Sprintf("mem-channel-%d", (i+j)%numChannels)
			hub.subscribe(client, ch)
		}
	}

	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	allocatedMB := float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024
	perClientKB := float64(memAfter.Alloc-memBefore.Alloc) / float64(numClients) / 1024

	t.Logf("Memory for %d clients with %d subscriptions each:", numClients, 10)
	t.Logf("  Total allocated: %.2f MB", allocatedMB)
	t.Logf("  Per client: %.2f KB", perClientKB)

	// Sanity check - should be reasonable
	if perClientKB > 50 {
		t.Errorf("Memory per client too high: %.2f KB (expected < 50 KB)", perClientKB)
	}
}

func TestStressRapidSubscribeUnsubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	hub := GetHub("stress-sub-unsub")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	numClients := 50
	iterations := 100

	// Create clients
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = createTestClient(hub, fmt.Sprintf("rapid-client-%d", i))
		hub.mu.Lock()
		hub.clients[clients[i].ID] = clients[i]
		hub.mu.Unlock()
	}

	var wg sync.WaitGroup
	var ops int64
	start := time.Now()

	// Rapid subscribe/unsubscribe cycles
	for _, client := range clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				ch := fmt.Sprintf("rapid-channel-%d", i%10)
				hub.subscribe(c, ch)
				atomic.AddInt64(&ops, 1)
				hub.unsubscribe(c, ch)
				atomic.AddInt64(&ops, 1)
			}
		}(client)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Rapid sub/unsub: %d operations in %v (%.0f ops/sec)",
		ops, elapsed, float64(ops)/elapsed.Seconds())

	// Verify no stale subscriptions
	hub.mu.RLock()
	channelCount := len(hub.channels)
	hub.mu.RUnlock()

	if channelCount > 0 {
		t.Logf("Warning: %d channels still exist after unsubscribe cycles", channelCount)
	}
}

func TestStressBroadcastUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	hub := GetHub("stress-under-load")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	numClients := 100
	channel := "load-test"
	bufferSize := 50

	// Create clients with moderate buffers
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = &Client{
			ID:          fmt.Sprintf("load-client-%d", i),
			Hub:         hub,
			Channels:    make(map[string]bool),
			Send:        make(chan []byte, bufferSize),
			ConnectedAt: time.Now(),
		}
		hub.mu.Lock()
		hub.clients[clients[i].ID] = clients[i]
		hub.mu.Unlock()
		hub.subscribe(clients[i], channel)
	}

	// Track received messages
	var received int64
	var wg sync.WaitGroup

	// Start consumer goroutines that drain at a realistic pace
	for _, client := range clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			for {
				select {
				case <-c.Send:
					atomic.AddInt64(&received, 1)
				case <-time.After(200 * time.Millisecond):
					return
				}
			}
		}(client)
	}

	// Send messages at a controlled rate to simulate realistic load
	numMessages := 200
	var sent int64
	start := time.Now()

	for i := 0; i < numMessages; i++ {
		hub.BroadcastToChannel(channel, map[string]interface{}{
			"seq":     i,
			"payload": "test data under load",
		})
		atomic.AddInt64(&sent, 1)
		// Small delay to allow consumers to keep up
		if i%10 == 0 {
			time.Sleep(time.Microsecond * 100)
		}
	}

	// Wait for consumers to finish
	wg.Wait()
	elapsed := time.Since(start)

	totalExpected := int64(numMessages * numClients)
	deliveryRate := float64(received) / float64(totalExpected) * 100

	t.Logf("Broadcast under load: %d messages to %d clients in %v",
		numMessages, numClients, elapsed)
	t.Logf("Messages delivered: %d/%d (%.1f%%)",
		received, totalExpected, deliveryRate)
	t.Logf("Throughput: %.0f messages/sec", float64(received)/elapsed.Seconds())

	// With consumers draining, we should deliver most messages
	if deliveryRate < 80 {
		t.Errorf("Delivery rate too low: %.1f%% (expected >= 80%%)", deliveryRate)
	}
}

// BenchmarkBroadcastToChannel measures broadcast performance
func BenchmarkBroadcastToChannel(b *testing.B) {
	hub := GetHub("bench-broadcast")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	// Setup: 100 clients subscribed to channel
	channel := "bench"
	for i := 0; i < 100; i++ {
		client := createTestClient(hub, fmt.Sprintf("bench-client-%d", i))
		hub.mu.Lock()
		hub.clients[client.ID] = client
		hub.mu.Unlock()
		hub.subscribe(client, channel)
	}

	payload := map[string]interface{}{"test": "data", "seq": 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.BroadcastToChannel(channel, payload)
	}
}

// BenchmarkSubscribe measures subscription performance
func BenchmarkSubscribe(b *testing.B) {
	hub := GetHub("bench-subscribe")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	client := createTestClient(hub, "bench-client")
	hub.mu.Lock()
	hub.clients[client.ID] = client
	hub.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := fmt.Sprintf("bench-channel-%d", i%100)
		hub.subscribe(client, ch)
	}
}

// BenchmarkGetSubscribers measures subscriber lookup performance
func BenchmarkGetSubscribers(b *testing.B) {
	hub := GetHub("bench-getsubs")
	defer func() {
		hub.mu.Lock()
		hub.clients = make(map[string]*Client)
		hub.channels = make(map[string]map[string]bool)
		hub.mu.Unlock()
	}()

	// Setup: 100 clients subscribed to channel
	channel := "bench"
	for i := 0; i < 100; i++ {
		client := createTestClient(hub, fmt.Sprintf("getsub-client-%d", i))
		hub.mu.Lock()
		hub.clients[client.ID] = client
		hub.mu.Unlock()
		hub.subscribe(client, channel)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hub.GetSubscribers(channel)
	}
}
