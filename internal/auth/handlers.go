package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/api"
)

// Handler wraps the auth service for HTTP handlers
type Handler struct {
	service *Service
	mux     *http.ServeMux
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	h := &Handler{
		service: service,
		mux:     http.NewServeMux(),
	}
	// Register routes on internal mux for ServeHTTP
	h.registerInternalRoutes()
	return h
}

// ServeHTTP implements http.Handler for standalone auth routing
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// registerInternalRoutes registers auth routes on the internal mux
func (h *Handler) registerInternalRoutes() {
	// Public routes (OAuth flow)
	h.mux.HandleFunc("GET /auth/login", h.LoginPage)
	h.mux.HandleFunc("GET /auth/login/{provider}", h.StartLogin)
	h.mux.HandleFunc("GET /auth/callback/{provider}", h.Callback)
	h.mux.HandleFunc("GET /auth/session", h.Session)
	h.mux.HandleFunc("POST /auth/logout", h.Logout)

	// Dev login routes (local only)
	h.mux.HandleFunc("GET /auth/dev/login", h.DevLoginForm)
	h.mux.HandleFunc("POST /auth/dev/callback", h.DevLoginCallback)

	// Invite routes
	h.mux.HandleFunc("GET /auth/invite/{code}", h.InvitePage)
	h.mux.HandleFunc("POST /auth/invite/{code}", h.RedeemInvite)
}

// RegisterRoutes registers auth routes on a mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Public routes
	mux.HandleFunc("GET /auth/login", h.LoginPage)
	mux.HandleFunc("GET /auth/login/{provider}", h.StartLogin)
	mux.HandleFunc("GET /auth/callback/{provider}", h.Callback)
	mux.HandleFunc("GET /auth/session", h.Session)
	mux.HandleFunc("POST /auth/logout", h.Logout)

	// Dev login routes (local only)
	mux.HandleFunc("GET /auth/dev/login", h.DevLoginForm)
	mux.HandleFunc("POST /auth/dev/callback", h.DevLoginCallback)

	// Invite routes
	mux.HandleFunc("GET /auth/invite/{code}", h.InvitePage)
	mux.HandleFunc("POST /auth/invite/{code}", h.RedeemInvite)

	// Admin routes (require authentication)
	mux.HandleFunc("GET /auth/users", h.ListUsers)
	mux.HandleFunc("GET /auth/users/{id}", h.GetUser)
	mux.HandleFunc("PATCH /auth/users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /auth/users/{id}", h.DeleteUser)
	mux.HandleFunc("POST /auth/invite", h.CreateInvite)
	mux.HandleFunc("GET /auth/invites", h.ListInvites)
	mux.HandleFunc("GET /auth/providers", h.ListProvidersHandler)

	// API routes (require API token - used by CLI)
	mux.HandleFunc("GET /api/auth/providers", h.APIListProviders)
	mux.HandleFunc("PUT /api/auth/providers/{name}", h.APIConfigureProvider)
}

// LoginPage renders the login page with provider buttons
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	redirectTo := r.URL.Query().Get("redirect")
	if redirectTo == "" {
		redirectTo = "/"
	}

	// Get enabled providers
	providers, err := h.service.GetEnabledProviders()
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// If no providers configured and not in local mode, show setup message
	if len(providers) == 0 && !IsLocalMode(r) {
		h.renderLoginPageWithRequest(w, r, nil, redirectTo, "No login providers configured. Contact the administrator.")
		return
	}

	h.renderLoginPageWithRequest(w, r, providers, redirectTo, "")
}

// StartLogin initiates the OAuth flow for a provider
func (h *Handler) StartLogin(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	if providerName == "" {
		api.BadRequest(w, "provider required")
		return
	}

	redirectTo := r.URL.Query().Get("redirect")
	if redirectTo == "" {
		redirectTo = "/"
	}
	appID := r.URL.Query().Get("app")

	// Build callback URL - always use root domain for OAuth
	scheme := "https"
	if !h.service.IsSecure() {
		scheme = "http"
	}
	// Use configured domain (root), not request host (could be subdomain)
	callbackURL := fmt.Sprintf("%s://%s/auth/callback/%s", scheme, h.service.Domain(), providerName)

	// Start OAuth flow
	authURL, err := h.service.StartOAuthFlow(providerName, redirectTo, appID, callbackURL)
	if err != nil {
		if err == ErrProviderDisabled {
			api.BadRequest(w, "Provider not enabled")
			return
		}
		api.InternalError(w, err)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth callback
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")

	// Check for errors from provider
	if errCode := r.URL.Query().Get("error"); errCode != "" {
		errDesc := r.URL.Query().Get("error_description")
		h.renderErrorPage(w, fmt.Sprintf("Authentication failed: %s - %s", errCode, errDesc))
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		h.renderErrorPage(w, "Invalid callback parameters")
		return
	}

	// Build callback URL (must match the one used in StartLogin)
	scheme := "https"
	if !h.service.IsSecure() {
		scheme = "http"
	}
	// Use configured domain (root), not request host (could be subdomain)
	callbackURL := fmt.Sprintf("%s://%s/auth/callback/%s", scheme, h.service.Domain(), providerName)

	// Complete OAuth flow
	sessionToken, _, redirectTo, err := h.service.CompleteOAuthFlow(providerName, code, state, callbackURL)
	if err != nil {
		h.renderErrorPage(w, "Authentication failed: "+err.Error())
		return
	}

	// Set session cookie
	http.SetCookie(w, h.service.SessionCookie(sessionToken, int(DefaultSessionTTL.Seconds())))

	// Redirect to original destination
	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
}

// Session returns the current session info
func (h *Handler) Session(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.GetSessionFromRequest(r)
	if err != nil {
		api.Success(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"authenticated": true,
		"user":          user,
	})
}

// Logout clears the session
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("fazt_session")
	if err == nil && cookie.Value != "" {
		h.service.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, h.service.ClearSessionCookie())

	// Check if this is an API request or browser request
	if r.Header.Get("Accept") == "application/json" {
		api.Success(w, http.StatusOK, map[string]interface{}{
			"message": "Logged out successfully",
		})
		return
	}

	// Redirect to login page
	http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
}

// Admin handlers

// ListUsers returns all users (admin only)
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.GetSessionFromRequest(r)
	if err != nil || !user.IsAdmin() {
		api.Forbidden(w, "Admin access required")
		return
	}

	users, err := h.service.ListUsers()
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, users)
}

// GetUser returns a specific user (admin only)
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.service.GetSessionFromRequest(r)
	if err != nil || !currentUser.IsAdmin() {
		api.Forbidden(w, "Admin access required")
		return
	}

	userID := r.PathValue("id")
	user, err := h.service.GetUserByID(userID)
	if err == ErrUserNotFound {
		api.NotFound(w, "USER_NOT_FOUND", "User not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, user)
}

// UpdateUser updates a user (admin only)
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.service.GetSessionFromRequest(r)
	if err != nil || !currentUser.IsAdmin() {
		api.Forbidden(w, "Admin access required")
		return
	}

	userID := r.PathValue("id")

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.InvalidJSON(w, "Invalid request body")
		return
	}

	// Only owner can change roles
	if req.Role != "" {
		if !currentUser.IsOwner() {
			api.Forbidden(w, "Only owner can change roles")
			return
		}
		// Don't allow demoting the owner
		targetUser, err := h.service.GetUserByID(userID)
		if err != nil {
			api.NotFound(w, "USER_NOT_FOUND", "User not found")
			return
		}
		if targetUser.IsOwner() && req.Role != "owner" {
			api.BadRequest(w, "Cannot demote the owner")
			return
		}
		if err := h.service.UpdateUserRole(userID, req.Role); err != nil {
			api.BadRequest(w, err.Error())
			return
		}
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, user)
}

// DeleteUser removes a user (admin only)
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.service.GetSessionFromRequest(r)
	if err != nil || !currentUser.IsAdmin() {
		api.Forbidden(w, "Admin access required")
		return
	}

	userID := r.PathValue("id")

	// Don't allow deleting the owner
	targetUser, err := h.service.GetUserByID(userID)
	if err == ErrUserNotFound {
		api.NotFound(w, "USER_NOT_FOUND", "User not found")
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	if targetUser.IsOwner() {
		api.BadRequest(w, "Cannot delete the owner")
		return
	}

	// Only owner can delete users
	if !currentUser.IsOwner() {
		api.Forbidden(w, "Only owner can delete users")
		return
	}

	if err := h.service.DeleteUser(userID); err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"message": "User deleted",
	})
}

// ListProvidersHandler returns configured providers
func (h *Handler) ListProvidersHandler(w http.ResponseWriter, r *http.Request) {
	// For login page, only show enabled providers
	providers, err := h.service.GetEnabledProviders()
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Build response with display names
	var result []map[string]interface{}
	for _, cfg := range providers {
		if provider, ok := Providers[cfg.Name]; ok {
			result = append(result, map[string]interface{}{
				"name":         cfg.Name,
				"display_name": provider.DisplayName,
				"enabled":      cfg.Enabled,
			})
		}
	}

	api.Success(w, http.StatusOK, result)
}

// HTML rendering helpers

func (h *Handler) renderLoginPage(w http.ResponseWriter, providers []*ProviderConfig, redirectTo, errorMsg string) {
	h.renderLoginPageWithRequest(w, nil, providers, redirectTo, errorMsg)
}

func (h *Handler) renderLoginPageWithRequest(w http.ResponseWriter, r *http.Request, providers []*ProviderConfig, redirectTo, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var providerButtons strings.Builder
	for _, cfg := range providers {
		provider := Providers[cfg.Name]
		if provider == nil {
			continue
		}
		loginURL := fmt.Sprintf("/auth/login/%s?redirect=%s", cfg.Name, redirectTo)
		providerButtons.WriteString(fmt.Sprintf(`
      <a href="%s" class="provider-btn %s">
        Continue with %s
      </a>`, loginURL, cfg.Name, provider.DisplayName))
	}

	// Add dev login option if in local mode
	devLoginHTML := ""
	if r != nil && IsLocalMode(r) {
		devLoginHTML = fmt.Sprintf(`
    <div class="divider"><span>Development</span></div>
    <a href="/auth/dev/login?redirect=%s" class="provider-btn dev">
      Dev Login
    </a>
    <p class="hint">Simulates OAuth for local testing</p>`, redirectTo)
	}

	errorHTML := ""
	if errorMsg != "" {
		errorHTML = fmt.Sprintf(`<div class="error">%s</div>`, errorMsg)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Sign In - Fazt</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #0a0a0a;
      color: #fff;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 20px;
    }
    .container {
      width: 100%%;
      max-width: 400px;
      background: #141414;
      border: 1px solid #333;
      border-radius: 12px;
      padding: 40px;
    }
    h1 {
      font-size: 24px;
      font-weight: 600;
      margin-bottom: 8px;
      text-align: center;
    }
    .subtitle {
      color: #888;
      text-align: center;
      margin-bottom: 32px;
    }
    .error {
      background: #3b1c1c;
      border: 1px solid #5c2626;
      color: #f87171;
      padding: 12px;
      border-radius: 8px;
      margin-bottom: 24px;
      font-size: 14px;
    }
    .providers {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }
    .provider-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 14px 20px;
      border-radius: 8px;
      text-decoration: none;
      font-weight: 500;
      transition: opacity 0.2s;
    }
    .provider-btn:hover { opacity: 0.9; }
    .provider-btn.google {
      background: #fff;
      color: #333;
    }
    .provider-btn.github {
      background: #333;
      color: #fff;
    }
    .provider-btn.discord {
      background: #5865F2;
      color: #fff;
    }
    .provider-btn.microsoft {
      background: #0078D4;
      color: #fff;
    }
    .provider-btn.dev {
      background: #444;
      color: #fff;
    }
    .divider {
      display: flex;
      align-items: center;
      margin: 24px 0 16px;
      color: #666;
      font-size: 12px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    .divider::before,
    .divider::after {
      content: "";
      flex: 1;
      height: 1px;
      background: #333;
    }
    .divider span {
      padding: 0 12px;
    }
    .hint {
      color: #666;
      font-size: 12px;
      text-align: center;
      margin-top: 8px;
    }
    .footer {
      margin-top: 32px;
      text-align: center;
      color: #666;
      font-size: 12px;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Sign In</h1>
    <p class="subtitle">to %s</p>
    %s
    <div class="providers">
      %s
    </div>
    %s
    <p class="footer">Powered by Fazt</p>
  </div>
</body>
</html>`, h.service.Domain(), errorHTML, providerButtons.String(), devLoginHTML)

	w.Write([]byte(html))
}

func (h *Handler) renderErrorPage(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Error - Fazt</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #0a0a0a;
      color: #fff;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 20px;
    }
    .container {
      width: 100%%;
      max-width: 400px;
      background: #141414;
      border: 1px solid #333;
      border-radius: 12px;
      padding: 40px;
      text-align: center;
    }
    h1 {
      font-size: 24px;
      font-weight: 600;
      margin-bottom: 16px;
      color: #f87171;
    }
    .message {
      color: #ccc;
      margin-bottom: 24px;
    }
    a {
      display: inline-block;
      padding: 12px 24px;
      background: #333;
      color: #fff;
      text-decoration: none;
      border-radius: 8px;
    }
    a:hover { background: #444; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Authentication Error</h1>
    <p class="message">%s</p>
    <a href="/auth/login">Try Again</a>
  </div>
</body>
</html>`, errorMsg)

	w.Write([]byte(html))
}

// API handlers (require Bearer token authentication)

// APIListProviders returns all configured providers (for CLI)
func (h *Handler) APIListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.service.ListProviders()
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Include all provider info for admin view
	var result []map[string]interface{}
	for _, cfg := range providers {
		displayName := cfg.Name
		if provider, ok := Providers[cfg.Name]; ok {
			displayName = provider.DisplayName
		}
		clientIDDisplay := cfg.ClientID
		if len(clientIDDisplay) > 20 {
			clientIDDisplay = clientIDDisplay[:17] + "..."
		}
		result = append(result, map[string]interface{}{
			"name":         cfg.Name,
			"display_name": displayName,
			"enabled":      cfg.Enabled,
			"client_id":    clientIDDisplay,
		})
	}

	api.Success(w, http.StatusOK, result)
}

// APIConfigureProvider configures an OAuth provider (for CLI)
func (h *Handler) APIConfigureProvider(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("name")
	if providerName == "" {
		api.BadRequest(w, "provider name required")
		return
	}

	// Validate provider name
	if _, ok := Providers[providerName]; !ok {
		api.BadRequest(w, "unknown provider: "+providerName)
		return
	}

	var req struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Enable       *bool  `json:"enable,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.InvalidJSON(w, "Invalid request body")
		return
	}

	// If credentials provided, set them
	if req.ClientID != "" && req.ClientSecret != "" {
		if err := h.service.SetProviderConfig(providerName, req.ClientID, req.ClientSecret); err != nil {
			api.InternalError(w, err)
			return
		}
	}

	// Handle enable/disable
	if req.Enable != nil {
		if *req.Enable {
			if err := h.service.EnableProvider(providerName); err != nil {
				api.BadRequest(w, err.Error())
				return
			}
		} else {
			if err := h.service.DisableProvider(providerName); err != nil {
				api.InternalError(w, err)
				return
			}
		}
	}

	// Return current status
	cfg, err := h.service.GetProviderConfig(providerName)
	if err == ErrProviderDisabled {
		api.Success(w, http.StatusOK, map[string]interface{}{
			"name":       providerName,
			"configured": false,
		})
		return
	}
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"name":       cfg.Name,
		"enabled":    cfg.Enabled,
		"configured": true,
	})
}
