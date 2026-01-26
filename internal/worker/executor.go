package worker

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dop251/goja"
	"github.com/fazt-sh/fazt/internal/debug"
	"github.com/fazt-sh/fazt/internal/storage"
)

// Executor executes worker JavaScript code with job context.
type Executor struct {
	db      *sql.DB
	storage *storage.Storage
}

// NewExecutor creates a new worker executor.
func NewExecutor(db *sql.DB) *Executor {
	return &Executor{
		db:      db,
		storage: storage.New(db),
	}
}

// Execute runs the worker code with job context injected.
func (e *Executor) Execute(ctx context.Context, job *Job, code string) (interface{}, error) {
	vm := goja.New()

	// Inject console
	injectConsole(vm, job)

	// Inject sleep helper
	InjectSleepHelper(vm)

	// Inject job context (job.id, job.data, job.progress(), etc.)
	if err := InjectJobContext(vm, job); err != nil {
		return nil, fmt.Errorf("failed to inject job context: %w", err)
	}

	// Inject storage namespace (fazt.storage.*)
	if err := storage.InjectStorageNamespace(vm, e.storage, job.AppID, ctx); err != nil {
		return nil, fmt.Errorf("failed to inject storage: %w", err)
	}

	// Inject worker namespace for spawning child jobs (fazt.worker.*)
	if err := InjectWorkerNamespace(vm, job.AppID, ctx); err != nil {
		return nil, fmt.Errorf("failed to inject worker namespace: %w", err)
	}

	// Set up interrupt on context cancellation
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt("job cancelled or timed out")
		case <-done:
		}
	}()

	defer func() {
		close(done)
		vm.ClearInterrupt()
	}()

	// Wrap code in module pattern and execute
	wrappedCode := fmt.Sprintf(`
(function() {
    var module = { exports: {} };
    var exports = module.exports;
    %s
    // If module.exports is a function, call it with job
    if (typeof module.exports === 'function') {
        return module.exports(job);
    }
    // If there's a handler function, call it
    if (typeof handler === 'function') {
        return handler(job);
    }
    return module.exports;
})()
`, code)

	value, err := vm.RunString(wrappedCode)
	if err != nil {
		// Check for interrupt
		if interruptErr, ok := err.(*goja.InterruptedError); ok {
			if job.IsCancelled() {
				return nil, fmt.Errorf("job cancelled")
			}
			return nil, fmt.Errorf("job interrupted: %v", interruptErr.Value())
		}
		return nil, err
	}

	// Mark as healthy if daemon and completed successfully
	if job.Config.Daemon {
		job.LastHealthyAt = job.StartedAt
	}

	// Extract return value
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, nil
	}

	return value.Export(), nil
}

// injectConsole adds console.log/warn/error that route to job logs.
func injectConsole(vm *goja.Runtime, job *Job) {
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
					msg = fmt.Sprintf("%s %v", msg, parts[1:])
				}
			}

			// Add to job logs
			job.AddLog(fmt.Sprintf("[%s] %s", level, msg))

			// Also log to debug
			debug.Log("worker", "job=%s [%s] %s", job.ID, level, msg)

			return goja.Undefined()
		}
	}

	console.Set("log", makeLogger("info"))
	console.Set("info", makeLogger("info"))
	console.Set("warn", makeLogger("warn"))
	console.Set("error", makeLogger("error"))
	console.Set("debug", makeLogger("debug"))
	vm.Set("console", console)
}

// SetupGlobalExecutor configures the global pool to use the executor.
func SetupGlobalExecutor(db *sql.DB) {
	executor := NewExecutor(db)
	SetExecutor(func(ctx context.Context, job *Job, code string) (interface{}, error) {
		return executor.Execute(ctx, job, code)
	})
}
