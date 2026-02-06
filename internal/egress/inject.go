package egress

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dop251/goja"
	"github.com/fazt-sh/fazt/internal/timeout"
)

// InjectNetNamespace adds the fazt.net namespace with fetch to a Goja VM.
// Must be called after fazt object already exists on the VM.
func InjectNetNamespace(vm *goja.Runtime, proxy *EgressProxy, appID string,
	ctx context.Context, budget *timeout.Budget) error {

	netObj := vm.NewObject()
	callCount := 0

	netObj.Set("fetch", func(call goja.FunctionCall) goja.Value {
		callCount++
		if callCount > proxy.callLimit {
			panic(vm.NewGoError(&EgressError{
				Code:    CodeLimit,
				Message: fmt.Sprintf("fetch limit exceeded (%d calls)", proxy.callLimit),
			}))
		}

		if len(call.Arguments) == 0 {
			panic(vm.NewGoError(errNet("fetch requires a URL argument")))
		}

		rawURL := call.Argument(0).String()
		opts := parseJSOptions(vm, call)

		// Get net context from budget
		netCtx, cancel, err := budget.NetContext(ctx)
		if err != nil {
			panic(vm.NewGoError(&EgressError{
				Code:      CodeBudget,
				Message:   err.Error(),
				Retryable: true,
			}))
		}
		defer cancel()

		resp, err := proxy.Fetch(netCtx, appID, rawURL, opts)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return responseToJS(vm, resp)
	})

	// Get existing fazt object and add net namespace
	faztVal := vm.Get("fazt")
	if faztVal == nil || goja.IsUndefined(faztVal) {
		return fmt.Errorf("fazt object not found on VM")
	}
	fazt := faztVal.ToObject(vm)
	fazt.Set("net", netObj)
	return nil
}

// parseJSOptions extracts FetchOptions from the second JS argument.
func parseJSOptions(vm *goja.Runtime, call goja.FunctionCall) FetchOptions {
	opts := FetchOptions{}
	if len(call.Arguments) < 2 {
		return opts
	}

	arg := call.Argument(1)
	if arg == nil || goja.IsUndefined(arg) || goja.IsNull(arg) {
		return opts
	}

	obj := arg.ToObject(vm)
	if obj == nil {
		return opts
	}

	if v := obj.Get("method"); v != nil && !goja.IsUndefined(v) {
		opts.Method = v.String()
	}
	if v := obj.Get("body"); v != nil && !goja.IsUndefined(v) {
		opts.Body = v.String()
	}
	if v := obj.Get("auth"); v != nil && !goja.IsUndefined(v) {
		opts.Auth = v.String()
	}
	if v := obj.Get("headers"); v != nil && !goja.IsUndefined(v) {
		headersObj := v.ToObject(vm)
		if headersObj != nil {
			opts.Headers = make(map[string]string)
			for _, key := range headersObj.Keys() {
				val := headersObj.Get(key)
				if val != nil && !goja.IsUndefined(val) {
					opts.Headers[key] = val.String()
				}
			}
		}
	}

	return opts
}

// responseToJS converts a FetchResponse to a Goja-compatible JS object.
func responseToJS(vm *goja.Runtime, resp *FetchResponse) goja.Value {
	obj := vm.NewObject()
	obj.Set("status", resp.Status)
	obj.Set("ok", resp.OK)
	obj.Set("headers", resp.Headers)
	obj.Set("text", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(resp.Text())
	})
	obj.Set("json", func(call goja.FunctionCall) goja.Value {
		var data interface{}
		if err := json.Unmarshal(resp.body, &data); err != nil {
			panic(vm.NewGoError(fmt.Errorf("invalid JSON: %w", err)))
		}
		return vm.ToValue(data)
	})
	return obj
}
