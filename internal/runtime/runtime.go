// Package runtime provides a JavaScript execution environment for serverless functions.
// It uses Goja (a pure Go JavaScript engine) to execute api/main.js files.
package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

const (
	DefaultTimeout = 1 * time.Second // Increased for KV operations with larger data
	MaxPoolSize    = 10
)

// Runtime manages JavaScript execution.
type Runtime struct {
	pool     chan *goja.Runtime
	poolSize int
	timeout  time.Duration
	mu       sync.RWMutex
}

// Request represents an HTTP request passed to JavaScript.
type Request struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Query   map[string]string      `json:"query"`
	Headers map[string]string      `json:"headers"`
	Body    interface{}            `json:"body"`
}

// Response represents the response from JavaScript execution.
type Response struct {
	Status  int                    `json:"status"`
	Headers map[string]string      `json:"headers"`
	Body    interface{}            `json:"body"`
}

// ExecuteResult contains the result of JavaScript execution.
type ExecuteResult struct {
	Response *Response
	Logs     []LogEntry
	Error    error
	Duration time.Duration
}

// LogEntry represents a console log entry.
type LogEntry struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// NewRuntime creates a new JavaScript runtime manager.
func NewRuntime(poolSize int, timeout time.Duration) *Runtime {
	if poolSize <= 0 {
		poolSize = MaxPoolSize
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	r := &Runtime{
		pool:     make(chan *goja.Runtime, poolSize),
		poolSize: poolSize,
		timeout:  timeout,
	}

	// Pre-warm the pool
	for i := 0; i < poolSize; i++ {
		vm := goja.New()
		r.pool <- vm
	}

	return r
}

// getVM gets a VM from the pool or creates a new one.
func (r *Runtime) getVM() *goja.Runtime {
	select {
	case vm := <-r.pool:
		return vm
	default:
		return goja.New()
	}
}

// returnVM returns a VM to the pool.
func (r *Runtime) returnVM(vm *goja.Runtime) {
	select {
	case r.pool <- vm:
	default:
		// Pool is full, discard VM
	}
}

// Execute runs JavaScript code with the given request context.
func (r *Runtime) Execute(ctx context.Context, code string, req *Request) *ExecuteResult {
	start := time.Now()
	result := &ExecuteResult{
		Logs: make([]LogEntry, 0),
	}

	vm := r.getVM()
	defer r.returnVM(vm)

	// Set up timeout
	done := make(chan struct{})
	var timeoutCtx context.Context
	var cancel context.CancelFunc

	if r.timeout > 0 {
		timeoutCtx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	} else {
		timeoutCtx = ctx
	}

	// Interrupt handler for timeout
	go func() {
		select {
		case <-timeoutCtx.Done():
			vm.Interrupt("execution timeout")
		case <-done:
		}
	}()

	defer func() {
		close(done)
		vm.ClearInterrupt()
	}()

	// Inject globals
	if err := r.injectGlobals(vm, req, result); err != nil {
		result.Error = fmt.Errorf("failed to inject globals: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Execute the code
	value, err := vm.RunString(code)
	result.Duration = time.Since(start)

	if err != nil {
		if jserr, ok := err.(*goja.InterruptedError); ok {
			result.Error = fmt.Errorf("execution timeout: %v", jserr.Value())
		} else {
			result.Error = fmt.Errorf("javascript error: %w", err)
		}
		return result
	}

	// Process the return value
	result.Response = r.extractResponse(vm, value)
	return result
}

// injectGlobals sets up the JavaScript execution environment.
func (r *Runtime) injectGlobals(vm *goja.Runtime, req *Request, result *ExecuteResult) error {
	// Inject request object
	reqObj := vm.NewObject()
	reqObj.Set("method", req.Method)
	reqObj.Set("path", req.Path)
	reqObj.Set("query", req.Query)
	reqObj.Set("headers", req.Headers)
	reqObj.Set("body", req.Body)
	vm.Set("request", reqObj)

	// Inject respond helper
	vm.Set("respond", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(map[string]interface{}{
				"status": 200,
				"body":   nil,
			})
		}

		first := call.Argument(0)

		// respond(body) - just a body
		if len(call.Arguments) == 1 {
			// Check if it's a number (status code with no body)
			exported := first.Export()
			switch num := exported.(type) {
			case int64:
				return vm.ToValue(map[string]interface{}{
					"status": int(num),
					"body":   nil,
				})
			case float64:
				return vm.ToValue(map[string]interface{}{
					"status": int(num),
					"body":   nil,
				})
			case int:
				return vm.ToValue(map[string]interface{}{
					"status": num,
					"body":   nil,
				})
			}
			// Otherwise it's a body with 200 status
			return vm.ToValue(map[string]interface{}{
				"status": 200,
				"body":   first.Export(),
			})
		}

		// respond(status, body)
		status := 200
		if s, ok := first.Export().(int64); ok {
			status = int(s)
		}

		return vm.ToValue(map[string]interface{}{
			"status": status,
			"body":   call.Argument(1).Export(),
		})
	})

	// Inject console
	console := vm.NewObject()

	makeLogger := func(level string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			var parts []string
			for _, arg := range call.Arguments {
				parts = append(parts, fmt.Sprintf("%v", arg.Export()))
			}
			msg := ""
			if len(parts) > 0 {
				msg = parts[0]
				if len(parts) > 1 {
					msg = fmt.Sprintf(msg, toInterfaceSlice(parts[1:])...)
				}
			}
			result.Logs = append(result.Logs, LogEntry{
				Level:   level,
				Message: msg,
				Time:    time.Now(),
			})
			return goja.Undefined()
		}
	}

	console.Set("log", makeLogger("info"))
	console.Set("info", makeLogger("info"))
	console.Set("warn", makeLogger("warn"))
	console.Set("error", makeLogger("error"))
	console.Set("debug", makeLogger("debug"))
	vm.Set("console", console)

	return nil
}

// extractResponse converts the JavaScript return value to a Response.
func (r *Runtime) extractResponse(vm *goja.Runtime, value goja.Value) *Response {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return &Response{Status: 200}
	}

	exported := value.Export()

	// Handle the response object format
	if m, ok := exported.(map[string]interface{}); ok {
		resp := &Response{
			Status:  200,
			Headers: make(map[string]string),
		}

		if status, ok := m["status"].(int64); ok {
			resp.Status = int(status)
		} else if status, ok := m["status"].(int); ok {
			resp.Status = status
		} else if status, ok := m["status"].(float64); ok {
			resp.Status = int(status)
		}

		if body, ok := m["body"]; ok {
			resp.Body = body
		}

		if headers, ok := m["headers"].(map[string]interface{}); ok {
			for k, v := range headers {
				resp.Headers[k] = fmt.Sprintf("%v", v)
			}
		}

		return resp
	}

	// Raw return value becomes the body
	return &Response{
		Status: 200,
		Body:   exported,
	}
}

// ExecuteWithFiles runs JavaScript with file loading support.
func (r *Runtime) ExecuteWithFiles(ctx context.Context, mainCode string, req *Request, fileLoader FileLoader) *ExecuteResult {
	start := time.Now()
	result := &ExecuteResult{
		Logs: make([]LogEntry, 0),
	}

	vm := r.getVM()
	defer r.returnVM(vm)

	// Set up timeout
	done := make(chan struct{})
	var timeoutCtx context.Context
	var cancel context.CancelFunc

	if r.timeout > 0 {
		timeoutCtx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	} else {
		timeoutCtx = ctx
	}

	// Interrupt handler for timeout
	go func() {
		select {
		case <-timeoutCtx.Done():
			vm.Interrupt("execution timeout")
		case <-done:
		}
	}()

	defer func() {
		close(done)
		vm.ClearInterrupt()
	}()

	// Inject globals
	if err := r.injectGlobals(vm, req, result); err != nil {
		result.Error = fmt.Errorf("failed to inject globals: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Inject require with file loading
	moduleCache := make(map[string]goja.Value)
	r.injectRequire(vm, fileLoader, "api", moduleCache)

	// Execute the code
	value, err := vm.RunString(mainCode)
	result.Duration = time.Since(start)

	if err != nil {
		if jserr, ok := err.(*goja.InterruptedError); ok {
			result.Error = fmt.Errorf("execution timeout: %v", jserr.Value())
		} else {
			result.Error = fmt.Errorf("javascript error: %w", err)
		}
		return result
	}

	// Process the return value
	result.Response = r.extractResponse(vm, value)
	return result
}

// FileLoader is a function that loads file content by path.
type FileLoader func(path string) (string, error)

// injectRequire adds the require() function for module loading.
func (r *Runtime) injectRequire(vm *goja.Runtime, loader FileLoader, basePath string, cache map[string]goja.Value) {
	vm.Set("require", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewGoError(fmt.Errorf("require() needs a path argument")))
		}

		path := call.Argument(0).String()

		// 1. Check stdlib first for bare specifiers (not relative paths)
		if !isRelativePath(path) {
			if source, ok := GetStdlibModule(path); ok {
				cacheKey := "stdlib:" + path
				if cached, ok := cache[cacheKey]; ok {
					return cached
				}

				// Wrap in module pattern
				moduleCode := fmt.Sprintf(`
(function() {
    var module = { exports: {} };
    var exports = module.exports;
    %s
    return module.exports;
})()
`, source)

				result, err := vm.RunString(moduleCode)
				if err != nil {
					panic(vm.NewGoError(fmt.Errorf("stdlib %s error: %w", path, err)))
				}

				cache[cacheKey] = result
				return result
			}
		}

		// 2. Fall back to local file resolution
		resolved := resolvePath(basePath, path)

		// Check cache
		if cached, ok := cache[resolved]; ok {
			return cached
		}

		// Load the file
		content, err := loader(resolved)
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("cannot require '%s': %w", path, err)))
		}

		// Wrap in module pattern
		moduleCode := fmt.Sprintf(`
(function() {
    var module = { exports: {} };
    var exports = module.exports;
    %s
    return module.exports;
})()
`, content)

		// Execute
		result, err := vm.RunString(moduleCode)
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("error in '%s': %w", path, err)))
		}

		// Cache the result
		cache[resolved] = result
		return result
	})
}

// isRelativePath checks if a path starts with ./ or ../
func isRelativePath(path string) bool {
	return len(path) >= 2 && (path[:2] == "./" || (len(path) >= 3 && path[:3] == "../"))
}

// resolvePath resolves a relative require path.
func resolvePath(basePath, requirePath string) string {
	// Handle relative paths
	if len(requirePath) >= 2 && requirePath[:2] == "./" {
		return basePath + "/" + requirePath[2:]
	}
	if len(requirePath) >= 3 && requirePath[:3] == "../" {
		// For now, don't allow escaping the base path
		return basePath + "/" + requirePath
	}
	// Bare specifier - treat as relative to base
	return basePath + "/" + requirePath
}

// toInterfaceSlice converts a string slice to interface slice.
func toInterfaceSlice(ss []string) []interface{} {
	result := make([]interface{}, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
}

// ResponseToJSON converts a Response to JSON bytes.
func ResponseToJSON(resp *Response) ([]byte, error) {
	return json.Marshal(resp.Body)
}

// VMInjector is a function that injects additional globals into a VM.
type VMInjector func(vm *goja.Runtime) error

// ExecuteWithInjectors runs JavaScript with file loading and custom injectors.
func (r *Runtime) ExecuteWithInjectors(ctx context.Context, mainCode string, req *Request, fileLoader FileLoader, injectors ...VMInjector) *ExecuteResult {
	start := time.Now()
	result := &ExecuteResult{
		Logs: make([]LogEntry, 0),
	}

	vm := r.getVM()
	defer r.returnVM(vm)

	// Set up timeout
	done := make(chan struct{})
	var timeoutCtx context.Context
	var cancel context.CancelFunc

	if r.timeout > 0 {
		timeoutCtx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	} else {
		timeoutCtx = ctx
	}

	// Interrupt handler for timeout
	go func() {
		select {
		case <-timeoutCtx.Done():
			vm.Interrupt("execution timeout")
		case <-done:
		}
	}()

	defer func() {
		close(done)
		vm.ClearInterrupt()
	}()

	// Inject globals
	if err := r.injectGlobals(vm, req, result); err != nil {
		result.Error = fmt.Errorf("failed to inject globals: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Inject require with file loading
	moduleCache := make(map[string]goja.Value)
	r.injectRequire(vm, fileLoader, "api", moduleCache)

	// Run custom injectors
	for _, injector := range injectors {
		if err := injector(vm); err != nil {
			result.Error = fmt.Errorf("failed to inject: %w", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Execute the code
	value, err := vm.RunString(mainCode)
	result.Duration = time.Since(start)

	if err != nil {
		if jserr, ok := err.(*goja.InterruptedError); ok {
			result.Error = fmt.Errorf("execution timeout: %v", jserr.Value())
		} else {
			result.Error = fmt.Errorf("javascript error: %w", err)
		}
		return result
	}

	// Process the return value
	result.Response = r.extractResponse(vm, value)
	return result
}
