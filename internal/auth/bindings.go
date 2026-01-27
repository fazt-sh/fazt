package auth

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
)

// RequestContext holds authentication context for a request
type RequestContext struct {
	User      *User
	SessionID string
	AppID     string
	Domain    string
}

// RedirectError is thrown when requireLogin needs to redirect
type RedirectError struct {
	URL string
}

func (e *RedirectError) Error() string {
	return fmt.Sprintf("redirect to %s", e.URL)
}

// InjectAuthNamespace adds fazt.auth.* functions to a Goja VM
func InjectAuthNamespace(vm *goja.Runtime, service *Service, reqCtx *RequestContext, ctx context.Context) error {
	// Get or create fazt object
	faztVal := vm.Get("fazt")
	var fazt *goja.Object
	if faztVal == nil || goja.IsUndefined(faztVal) {
		fazt = vm.NewObject()
		vm.Set("fazt", fazt)
	} else {
		fazt = faztVal.ToObject(vm)
	}

	authObj := vm.NewObject()

	// fazt.auth.getUser() - returns the current user or null
	authObj.Set("getUser", func(call goja.FunctionCall) goja.Value {
		if reqCtx.User == nil {
			return goja.Null()
		}
		return vm.ToValue(map[string]interface{}{
			"id":       reqCtx.User.ID,
			"email":    reqCtx.User.Email,
			"name":     reqCtx.User.Name,
			"picture":  reqCtx.User.Picture,
			"role":     reqCtx.User.Role,
			"provider": reqCtx.User.Provider,
		})
	})

	// fazt.auth.isLoggedIn() - returns true if user is authenticated
	authObj.Set("isLoggedIn", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(reqCtx.User != nil)
	})

	// fazt.auth.isOwner() - returns true if user is the owner
	authObj.Set("isOwner", func(call goja.FunctionCall) goja.Value {
		if reqCtx.User == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(reqCtx.User.IsOwner())
	})

	// fazt.auth.isAdmin() - returns true if user is admin or owner
	authObj.Set("isAdmin", func(call goja.FunctionCall) goja.Value {
		if reqCtx.User == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(reqCtx.User.IsAdmin())
	})

	// fazt.auth.hasRole(role) - checks if user has the specified role
	authObj.Set("hasRole", func(call goja.FunctionCall) goja.Value {
		if reqCtx.User == nil {
			return vm.ToValue(false)
		}
		if len(call.Arguments) < 1 {
			return vm.ToValue(false)
		}
		role := call.Argument(0).String()
		switch role {
		case "owner":
			return vm.ToValue(reqCtx.User.IsOwner())
		case "admin":
			return vm.ToValue(reqCtx.User.IsAdmin())
		case "user":
			return vm.ToValue(true) // All authenticated users have "user" role
		default:
			return vm.ToValue(reqCtx.User.Role == role)
		}
	})

	// fazt.auth.requireLogin() - throws redirect if not logged in
	authObj.Set("requireLogin", func(call goja.FunctionCall) goja.Value {
		if reqCtx.User != nil {
			return goja.Undefined()
		}

		// Build login URL with redirect back to current app
		loginURL := fmt.Sprintf("/auth/login?redirect=/&app=%s", reqCtx.AppID)
		panic(vm.NewGoError(&RedirectError{URL: loginURL}))
	})

	// fazt.auth.requireRole(role) - throws 403 if user doesn't have role
	authObj.Set("requireRole", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("requireRole requires a role argument")))
		}
		role := call.Argument(0).String()

		if reqCtx.User == nil {
			loginURL := fmt.Sprintf("/auth/login?redirect=/&app=%s", reqCtx.AppID)
			panic(vm.NewGoError(&RedirectError{URL: loginURL}))
		}

		hasRole := false
		switch role {
		case "owner":
			hasRole = reqCtx.User.IsOwner()
		case "admin":
			hasRole = reqCtx.User.IsAdmin()
		case "user":
			hasRole = true
		default:
			hasRole = reqCtx.User.Role == role
		}

		if !hasRole {
			panic(vm.NewGoError(fmt.Errorf("forbidden: requires role '%s'", role)))
		}

		return goja.Undefined()
	})

	// fazt.auth.requireOwner() - throws 403 if user is not the owner
	authObj.Set("requireOwner", func(call goja.FunctionCall) goja.Value {
		if reqCtx.User == nil {
			loginURL := fmt.Sprintf("/auth/login?redirect=/&app=%s", reqCtx.AppID)
			panic(vm.NewGoError(&RedirectError{URL: loginURL}))
		}
		if !reqCtx.User.IsOwner() {
			panic(vm.NewGoError(fmt.Errorf("forbidden: owner access required")))
		}
		return goja.Undefined()
	})

	// fazt.auth.requireAdmin() - throws 403 if user is not admin
	authObj.Set("requireAdmin", func(call goja.FunctionCall) goja.Value {
		if reqCtx.User == nil {
			loginURL := fmt.Sprintf("/auth/login?redirect=/&app=%s", reqCtx.AppID)
			panic(vm.NewGoError(&RedirectError{URL: loginURL}))
		}
		if !reqCtx.User.IsAdmin() {
			panic(vm.NewGoError(fmt.Errorf("forbidden: admin access required")))
		}
		return goja.Undefined()
	})

	// fazt.auth.getLoginURL(redirect) - returns the login URL
	authObj.Set("getLoginURL", func(call goja.FunctionCall) goja.Value {
		redirect := "/"
		if len(call.Arguments) > 0 {
			redirect = call.Argument(0).String()
		}
		loginURL := fmt.Sprintf("/auth/login?redirect=%s&app=%s", redirect, reqCtx.AppID)
		return vm.ToValue(loginURL)
	})

	// fazt.auth.getLogoutURL(redirect) - returns the logout URL
	authObj.Set("getLogoutURL", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue("/auth/logout")
	})

	fazt.Set("auth", authObj)
	return nil
}

// AuthInjector creates an injector function for the runtime
func AuthInjector(service *Service, user *User, sessionID, appID, domain string) func(vm *goja.Runtime) error {
	return func(vm *goja.Runtime) error {
		reqCtx := &RequestContext{
			User:      user,
			SessionID: sessionID,
			AppID:     appID,
			Domain:    domain,
		}
		return InjectAuthNamespace(vm, service, reqCtx, context.Background())
	}
}

// IsRedirectError checks if an error is a redirect error
func IsRedirectError(err error) (*RedirectError, bool) {
	if re, ok := err.(*RedirectError); ok {
		return re, true
	}
	return nil, false
}
