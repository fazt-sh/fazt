package runtime

import (
	"github.com/dop251/goja"
)

// AppContext contains information about the current app.
type AppContext struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EnvVars provides environment variable access.
type EnvVars map[string]string

// InjectFaztNamespace adds the fazt.* namespace to a VM.
func InjectFaztNamespace(vm *goja.Runtime, app *AppContext, env EnvVars, result *ExecuteResult) error {
	fazt := vm.NewObject()

	// fazt.app
	appObj := vm.NewObject()
	if app != nil {
		appObj.Set("id", app.ID)
		appObj.Set("name", app.Name)
	} else {
		appObj.Set("id", "")
		appObj.Set("name", "")
	}
	fazt.Set("app", appObj)

	// fazt.env
	envObj := vm.NewObject()
	envObj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		key := call.Argument(0).String()
		if val, ok := env[key]; ok {
			return vm.ToValue(val)
		}
		// Check for default value
		if len(call.Arguments) > 1 {
			return call.Argument(1)
		}
		return goja.Undefined()
	})
	envObj.Set("has", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(false)
		}
		key := call.Argument(0).String()
		_, ok := env[key]
		return vm.ToValue(ok)
	})
	fazt.Set("env", envObj)

	// fazt.log
	logObj := vm.NewObject()
	makeLogger := func(level string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				return goja.Undefined()
			}
			msg := call.Argument(0).String()
			result.Logs = append(result.Logs, LogEntry{
				Level:   level,
				Message: msg,
			})
			return goja.Undefined()
		}
	}
	logObj.Set("info", makeLogger("info"))
	logObj.Set("warn", makeLogger("warn"))
	logObj.Set("error", makeLogger("error"))
	logObj.Set("debug", makeLogger("debug"))
	fazt.Set("log", logObj)

	// fazt.version
	fazt.Set("version", "0.8.0")

	vm.Set("fazt", fazt)
	return nil
}
