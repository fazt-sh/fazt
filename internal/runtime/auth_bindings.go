package runtime

import (
	"fmt"

	"github.com/dop251/goja"
)

// UserInfo represents user information for JS bindings (avoids import cycle with auth package)
type UserInfo struct {
	ID       string
	Email    string
	Name     string
	Picture  string
	Role     string
	Provider string
}

// InjectAuthNamespace adds fazt.auth.* functions to a Goja VM
func InjectAuthNamespace(vm *goja.Runtime, authCtx *AuthContext, app *AppContext) error {
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

	// Extract user info if available
	var userInfo *UserInfo
	if authCtx != nil && authCtx.User != nil {
		// The User is passed as interface{}, we need to extract fields
		// This is a workaround to avoid import cycles
		if u, ok := authCtx.User.(*UserInfo); ok {
			userInfo = u
		} else if umap, ok := authCtx.User.(map[string]interface{}); ok {
			userInfo = &UserInfo{}
			if id, ok := umap["id"].(string); ok {
				userInfo.ID = id
			}
			if email, ok := umap["email"].(string); ok {
				userInfo.Email = email
			}
			if name, ok := umap["name"].(string); ok {
				userInfo.Name = name
			}
			if picture, ok := umap["picture"].(string); ok {
				userInfo.Picture = picture
			}
			if role, ok := umap["role"].(string); ok {
				userInfo.Role = role
			}
			if provider, ok := umap["provider"].(string); ok {
				userInfo.Provider = provider
			}
		}
	}

	// fazt.auth.getUser() - returns the current user or null
	authObj.Set("getUser", func(call goja.FunctionCall) goja.Value {
		if userInfo == nil {
			return goja.Null()
		}
		return vm.ToValue(map[string]interface{}{
			"id":       userInfo.ID,
			"email":    userInfo.Email,
			"name":     userInfo.Name,
			"picture":  userInfo.Picture,
			"role":     userInfo.Role,
			"provider": userInfo.Provider,
		})
	})

	// fazt.auth.isLoggedIn() - returns true if user is authenticated
	authObj.Set("isLoggedIn", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(userInfo != nil)
	})

	// fazt.auth.isOwner() - returns true if user is the owner
	authObj.Set("isOwner", func(call goja.FunctionCall) goja.Value {
		if userInfo == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(userInfo.Role == "owner")
	})

	// fazt.auth.isAdmin() - returns true if user is admin or owner
	authObj.Set("isAdmin", func(call goja.FunctionCall) goja.Value {
		if userInfo == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(userInfo.Role == "owner" || userInfo.Role == "admin")
	})

	// fazt.auth.hasRole(role) - checks if user has the specified role
	authObj.Set("hasRole", func(call goja.FunctionCall) goja.Value {
		if userInfo == nil {
			return vm.ToValue(false)
		}
		if len(call.Arguments) < 1 {
			return vm.ToValue(false)
		}
		role := call.Argument(0).String()
		switch role {
		case "owner":
			return vm.ToValue(userInfo.Role == "owner")
		case "admin":
			return vm.ToValue(userInfo.Role == "owner" || userInfo.Role == "admin")
		case "user":
			return vm.ToValue(true) // All authenticated users have "user" role
		default:
			return vm.ToValue(userInfo.Role == role)
		}
	})

	// fazt.auth.requireLogin() - throws redirect if not logged in
	authObj.Set("requireLogin", func(call goja.FunctionCall) goja.Value {
		if userInfo != nil {
			return goja.Undefined()
		}

		// Build login URL with redirect back to current app
		appID := ""
		if app != nil {
			appID = app.ID
		}
		loginURL := fmt.Sprintf("/auth/login?redirect=/&app=%s", appID)
		panic(vm.NewGoError(&AuthRedirectError{URL: loginURL}))
	})

	// fazt.auth.requireRole(role) - throws 403 if user doesn't have role
	authObj.Set("requireRole", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("requireRole requires a role argument")))
		}
		role := call.Argument(0).String()

		appID := ""
		if app != nil {
			appID = app.ID
		}

		if userInfo == nil {
			loginURL := fmt.Sprintf("/auth/login?redirect=/&app=%s", appID)
			panic(vm.NewGoError(&AuthRedirectError{URL: loginURL}))
		}

		hasRole := false
		switch role {
		case "owner":
			hasRole = userInfo.Role == "owner"
		case "admin":
			hasRole = userInfo.Role == "owner" || userInfo.Role == "admin"
		case "user":
			hasRole = true
		default:
			hasRole = userInfo.Role == role
		}

		if !hasRole {
			panic(vm.NewGoError(&AuthForbiddenError{Role: role}))
		}

		return goja.Undefined()
	})

	// fazt.auth.requireOwner() - throws 403 if user is not the owner
	authObj.Set("requireOwner", func(call goja.FunctionCall) goja.Value {
		appID := ""
		if app != nil {
			appID = app.ID
		}

		if userInfo == nil {
			loginURL := fmt.Sprintf("/auth/login?redirect=/&app=%s", appID)
			panic(vm.NewGoError(&AuthRedirectError{URL: loginURL}))
		}
		if userInfo.Role != "owner" {
			panic(vm.NewGoError(&AuthForbiddenError{Role: "owner"}))
		}
		return goja.Undefined()
	})

	// fazt.auth.requireAdmin() - throws 403 if user is not admin
	authObj.Set("requireAdmin", func(call goja.FunctionCall) goja.Value {
		appID := ""
		if app != nil {
			appID = app.ID
		}

		if userInfo == nil {
			loginURL := fmt.Sprintf("/auth/login?redirect=/&app=%s", appID)
			panic(vm.NewGoError(&AuthRedirectError{URL: loginURL}))
		}
		if userInfo.Role != "owner" && userInfo.Role != "admin" {
			panic(vm.NewGoError(&AuthForbiddenError{Role: "admin"}))
		}
		return goja.Undefined()
	})

	// fazt.auth.getLoginURL(redirect) - returns the login URL
	authObj.Set("getLoginURL", func(call goja.FunctionCall) goja.Value {
		redirect := "/"
		if len(call.Arguments) > 0 {
			redirect = call.Argument(0).String()
		}
		appID := ""
		if app != nil {
			appID = app.ID
		}
		loginURL := fmt.Sprintf("/auth/login?redirect=%s&app=%s", redirect, appID)
		return vm.ToValue(loginURL)
	})

	// fazt.auth.getLogoutURL() - returns the logout URL
	authObj.Set("getLogoutURL", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue("/auth/logout")
	})

	fazt.Set("auth", authObj)
	return nil
}

// AuthRedirectError is thrown when auth requires redirect
type AuthRedirectError struct {
	URL string
}

func (e *AuthRedirectError) Error() string {
	return fmt.Sprintf("auth redirect to %s", e.URL)
}

// AuthForbiddenError is thrown when user lacks required role
type AuthForbiddenError struct {
	Role string
}

func (e *AuthForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: requires role '%s'", e.Role)
}

// IsAuthRedirectError checks if error is an auth redirect
func IsAuthRedirectError(err error) (*AuthRedirectError, bool) {
	if re, ok := err.(*AuthRedirectError); ok {
		return re, true
	}
	return nil, false
}

// IsAuthForbiddenError checks if error is an auth forbidden
func IsAuthForbiddenError(err error) (*AuthForbiddenError, bool) {
	if fe, ok := err.(*AuthForbiddenError); ok {
		return fe, true
	}
	return nil, false
}
