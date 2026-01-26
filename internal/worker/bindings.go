package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// InjectWorkerNamespace adds fazt.worker.* to a Goja VM.
// This is called from serverless handlers to allow spawning background jobs.
func InjectWorkerNamespace(vm *goja.Runtime, appID string, ctx context.Context) error {
	// Get or create fazt object
	faztVal := vm.Get("fazt")
	var fazt *goja.Object
	if faztVal == nil || goja.IsUndefined(faztVal) {
		fazt = vm.NewObject()
		vm.Set("fazt", fazt)
	} else {
		fazt = faztVal.ToObject(vm)
	}

	workerObj := vm.NewObject()

	// fazt.worker.spawn(handler, options)
	workerObj.Set("spawn", makeWorkerSpawn(vm, appID))

	// fazt.worker.get(jobId)
	workerObj.Set("get", makeWorkerGet(vm, appID))

	// fazt.worker.list(options)
	workerObj.Set("list", makeWorkerList(vm, appID))

	// fazt.worker.cancel(jobId)
	workerObj.Set("cancel", makeWorkerCancel(vm))

	// fazt.worker.wait(jobId, options) - poll until done
	workerObj.Set("wait", makeWorkerWait(vm, ctx))

	fazt.Set("worker", workerObj)
	return nil
}

// makeWorkerSpawn creates the fazt.worker.spawn() function.
func makeWorkerSpawn(vm *goja.Runtime, appID string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("worker.spawn requires handler path")))
		}

		handler := call.Argument(0).String()

		// Parse options
		cfg := DefaultJobConfig()
		if len(call.Arguments) >= 2 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			opts := call.Argument(1).Export()
			if optsMap, ok := opts.(map[string]interface{}); ok {
				parseSpawnOptions(&cfg, optsMap)
			}
		}

		// Spawn the job
		job, err := Spawn(appID, handler, cfg)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(jobToJS(job))
	}
}

// parseSpawnOptions parses JS options into JobConfig.
func parseSpawnOptions(cfg *JobConfig, opts map[string]interface{}) {
	// data: { ... }
	if data, ok := opts["data"].(map[string]interface{}); ok {
		cfg.Data = data
	}

	// memory: '32MB' or '64MB'
	if mem, ok := opts["memory"].(string); ok {
		if bytes, err := ParseMemory(mem); err == nil {
			cfg.MemoryBytes = bytes
		}
	}

	// timeout: '5m' or '30s' or null
	if timeout, ok := opts["timeout"]; ok {
		if timeout == nil {
			cfg.Timeout = nil // Indefinite
		} else if timeoutStr, ok := timeout.(string); ok {
			if dur, err := ParseDuration(timeoutStr); err == nil {
				cfg.Timeout = dur
			}
		}
	}

	// daemon: true/false
	if daemon, ok := opts["daemon"].(bool); ok {
		cfg.Daemon = daemon
	}

	// retry: 3
	if retry, ok := opts["retry"]; ok {
		switch v := retry.(type) {
		case int64:
			cfg.MaxAttempts = int(v) + 1 // +1 because attempts includes first try
		case float64:
			cfg.MaxAttempts = int(v) + 1
		}
	}

	// retryDelay: '1m'
	if delay, ok := opts["retryDelay"].(string); ok {
		if dur, err := ParseDuration(delay); err == nil && dur != nil {
			cfg.RetryDelay = *dur
		}
	}

	// priority: 'low' | 'normal' | 'high'
	if priority, ok := opts["priority"].(string); ok {
		switch priority {
		case "low":
			cfg.Priority = -1
		case "normal":
			cfg.Priority = 0
		case "high":
			cfg.Priority = 1
		}
	}

	// uniqueKey: 'sync-user-123'
	if key, ok := opts["uniqueKey"].(string); ok {
		cfg.UniqueKey = key
	}
}

// makeWorkerGet creates the fazt.worker.get() function.
func makeWorkerGet(vm *goja.Runtime, appID string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("worker.get requires jobId")))
		}

		jobID := call.Argument(0).String()

		job, err := Get(jobID)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		// Verify app ownership
		if job.AppID != appID {
			panic(vm.NewGoError(fmt.Errorf("job not found")))
		}

		return vm.ToValue(jobToJS(job))
	}
}

// makeWorkerList creates the fazt.worker.list() function.
func makeWorkerList(vm *goja.Runtime, appID string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		var status *JobStatus
		limit := 50

		if len(call.Arguments) >= 1 && !goja.IsUndefined(call.Argument(0)) && !goja.IsNull(call.Argument(0)) {
			opts := call.Argument(0).Export()
			if optsMap, ok := opts.(map[string]interface{}); ok {
				if s, ok := optsMap["status"].(string); ok {
					st := JobStatus(s)
					status = &st
				}
				if l, ok := optsMap["limit"].(int64); ok {
					limit = int(l)
				} else if l, ok := optsMap["limit"].(float64); ok {
					limit = int(l)
				}
			}
		}

		jobs, err := List(appID, status, limit)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		result := make([]interface{}, len(jobs))
		for i, job := range jobs {
			result[i] = jobToJS(job)
		}

		return vm.ToValue(result)
	}
}

// makeWorkerCancel creates the fazt.worker.cancel() function.
func makeWorkerCancel(vm *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("worker.cancel requires jobId")))
		}

		jobID := call.Argument(0).String()

		if err := Cancel(jobID); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

// makeWorkerWait creates the fazt.worker.wait() function.
func makeWorkerWait(vm *goja.Runtime, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("worker.wait requires jobId")))
		}

		jobID := call.Argument(0).String()

		// Parse timeout option
		timeout := 5 * time.Minute
		if len(call.Arguments) >= 2 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			opts := call.Argument(1).Export()
			if optsMap, ok := opts.(map[string]interface{}); ok {
				if t, ok := optsMap["timeout"].(string); ok {
					if dur, err := ParseDuration(t); err == nil && dur != nil {
						timeout = *dur
					}
				}
			}
		}

		deadline := time.Now().Add(timeout)
		pollInterval := 100 * time.Millisecond

		for {
			select {
			case <-ctx.Done():
				panic(vm.NewGoError(fmt.Errorf("context cancelled")))
			default:
			}

			if time.Now().After(deadline) {
				panic(vm.NewGoError(fmt.Errorf("wait timeout exceeded")))
			}

			job, err := Get(jobID)
			if err != nil {
				panic(vm.NewGoError(err))
			}

			if job.Status == StatusDone || job.Status == StatusFailed || job.Status == StatusCancelled {
				return vm.ToValue(jobToJS(job))
			}

			time.Sleep(pollInterval)
			// Increase poll interval over time
			if pollInterval < time.Second {
				pollInterval = pollInterval * 2
			}
		}
	}
}

// jobToJS converts a Job to a JS-friendly map.
func jobToJS(job *Job) map[string]interface{} {
	result := map[string]interface{}{
		"id":       job.ID,
		"handler":  job.Handler,
		"status":   string(job.Status),
		"progress": int(job.Progress * 100), // Convert 0-1 to 0-100
		"attempt":  job.Attempt,
	}

	if job.Config.Data != nil {
		result["data"] = job.Config.Data
	}

	if job.Result != "" {
		result["result"] = job.Result
	}

	if job.Error != "" {
		result["error"] = job.Error
	}

	if len(job.Logs) > 0 {
		result["logs"] = job.Logs
	}

	if !job.CreatedAt.IsZero() {
		result["createdAt"] = job.CreatedAt.UnixMilli()
	}

	if !job.StartedAt.IsZero() {
		result["startedAt"] = job.StartedAt.UnixMilli()
	}

	if !job.DoneAt.IsZero() {
		result["completedAt"] = job.DoneAt.UnixMilli()
	}

	return result
}

// InjectJobContext adds job.* to a Goja VM for use inside worker handlers.
// This provides the job object with progress(), log(), checkpoint(), etc.
func InjectJobContext(vm *goja.Runtime, job *Job) error {
	jobObj := vm.NewObject()

	// Read-only properties
	jobObj.Set("id", job.ID)
	jobObj.Set("data", job.Config.Data)
	jobObj.Set("attempt", job.Attempt)
	jobObj.Set("memory", job.Config.MemoryBytes)
	jobObj.Set("daemon", job.Config.Daemon)

	// Dynamic cancelled property (getter)
	jobObj.DefineAccessorProperty("cancelled", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(job.IsCancelled())
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// job.progress(n) - report progress 0-100
	jobObj.Set("progress", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		var progress float64
		switch v := call.Argument(0).Export().(type) {
		case int64:
			progress = float64(v) / 100.0
		case float64:
			progress = v / 100.0
		}

		job.SetProgress(progress)
		return goja.Undefined()
	})

	// job.log(msg) - add log entry
	jobObj.Set("log", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		msg := fmt.Sprintf("%v", call.Argument(0).Export())
		job.AddLog(msg)
		return goja.Undefined()
	})

	// job.checkpoint(data) - save checkpoint for recovery
	jobObj.Set("checkpoint", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			// Get checkpoint (no args)
			checkpoint, err := job.GetCheckpoint()
			if err != nil {
				panic(vm.NewGoError(err))
			}
			if checkpoint == nil {
				return goja.Null()
			}
			return vm.ToValue(checkpoint)
		}

		// Set checkpoint (with args)
		data := call.Argument(0).Export()
		if err := job.SetCheckpoint(data); err != nil {
			panic(vm.NewGoError(err))
		}
		return goja.Undefined()
	})

	// job.getCheckpoint() - explicit getter for checkpoint
	jobObj.Set("getCheckpoint", func(call goja.FunctionCall) goja.Value {
		checkpoint, err := job.GetCheckpoint()
		if err != nil {
			panic(vm.NewGoError(err))
		}
		if checkpoint == nil {
			return goja.Null()
		}
		return vm.ToValue(checkpoint)
	})

	vm.Set("job", jobObj)
	return nil
}

// sleep helper for workers (synchronous sleep)
func InjectSleepHelper(vm *goja.Runtime) {
	vm.Set("sleep", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		var ms int64
		switch v := call.Argument(0).Export().(type) {
		case int64:
			ms = v
		case float64:
			ms = int64(v)
		}

		time.Sleep(time.Duration(ms) * time.Millisecond)
		return goja.Undefined()
	})
}
