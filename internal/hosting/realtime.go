package hosting

import (
	"fmt"

	"github.com/dop251/goja"
)

// InjectRealtimeNamespace adds fazt.realtime.* to a Goja VM.
// This allows serverless handlers to broadcast messages to WebSocket clients.
func InjectRealtimeNamespace(vm *goja.Runtime, siteID string) error {
	// Get or create fazt object
	faztVal := vm.Get("fazt")
	var fazt *goja.Object
	if faztVal == nil || goja.IsUndefined(faztVal) {
		fazt = vm.NewObject()
		vm.Set("fazt", fazt)
	} else {
		fazt = faztVal.ToObject(vm)
	}

	rt := vm.NewObject()
	rt.Set("broadcast", makeBroadcast(vm, siteID))
	rt.Set("broadcastAll", makeBroadcastAll(vm, siteID))
	rt.Set("subscribers", makeSubscribers(vm, siteID))
	rt.Set("count", makeCount(vm, siteID))
	rt.Set("kick", makeKick(vm, siteID))

	fazt.Set("realtime", rt)
	return nil
}

// makeBroadcast creates fazt.realtime.broadcast(channel, data)
func makeBroadcast(vm *goja.Runtime, siteID string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("realtime.broadcast requires channel and data")))
		}

		channel := call.Argument(0).String()
		data := call.Argument(1).Export()

		hub := GetHub(siteID)
		hub.BroadcastToChannel(channel, data)

		return goja.Undefined()
	}
}

// makeBroadcastAll creates fazt.realtime.broadcastAll(data)
func makeBroadcastAll(vm *goja.Runtime, siteID string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("realtime.broadcastAll requires data")))
		}

		data := call.Argument(0).Export()

		hub := GetHub(siteID)
		hub.BroadcastAll(data)

		return goja.Undefined()
	}
}

// makeSubscribers creates fazt.realtime.subscribers(channel)
func makeSubscribers(vm *goja.Runtime, siteID string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("realtime.subscribers requires channel")))
		}

		channel := call.Argument(0).String()

		hub := GetHub(siteID)
		subscribers := hub.GetSubscribers(channel)

		return vm.ToValue(subscribers)
	}
}

// makeCount creates fazt.realtime.count(channel)
func makeCount(vm *goja.Runtime, siteID string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		hub := GetHub(siteID)

		// If channel provided, count subscribers in that channel
		// Otherwise, count total connected clients
		if len(call.Arguments) >= 1 && !goja.IsUndefined(call.Argument(0)) {
			channel := call.Argument(0).String()
			return vm.ToValue(hub.ChannelCount(channel))
		}

		return vm.ToValue(hub.ClientCount())
	}
}

// makeKick creates fazt.realtime.kick(clientId, reason)
func makeKick(vm *goja.Runtime, siteID string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("realtime.kick requires clientId")))
		}

		clientID := call.Argument(0).String()

		reason := ""
		if len(call.Arguments) >= 2 && !goja.IsUndefined(call.Argument(1)) {
			reason = call.Argument(1).String()
		}

		hub := GetHub(siteID)
		kicked := hub.KickClient(clientID, reason)

		return vm.ToValue(kicked)
	}
}
