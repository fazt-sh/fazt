package auth

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/fazt-sh/fazt/internal/api"
)

// IsLocalMode detects if we're running in local development mode (no HTTPS)
func IsLocalMode(r *http.Request) bool {
	// Primary check: no TLS
	if r.TLS == nil {
		return true
	}

	// Secondary: known local domains
	host := strings.ToLower(r.Host)
	localPatterns := []string{
		"localhost",
		"127.0.0.1",
		".nip.io",
		".local",
		".internal",
	}

	for _, pattern := range localPatterns {
		if strings.Contains(host, pattern) {
			return true
		}
	}

	return false
}

// hashEmail creates a deterministic ID from email for dev users
func hashEmail(email string) string {
	h := sha256.Sum256([]byte(email))
	return hex.EncodeToString(h[:8]) // 16 char hex string
}

// gravatarURL generates a Gravatar URL for an email (with fallback to initials)
func gravatarURL(email string) string {
	// MD5 hash of lowercase email for Gravatar
	email = strings.ToLower(strings.TrimSpace(email))
	h := md5.Sum([]byte(email))
	hash := hex.EncodeToString(h[:])
	// Use "retro" style as fallback (generates unique geometric patterns)
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?d=retro&s=200", hash)
}

// DevLoginForm shows the dev login form (local only)
func (h *Handler) DevLoginForm(w http.ResponseWriter, r *http.Request) {
	if !IsLocalMode(r) {
		api.Forbidden(w, "Dev login only available locally")
		return
	}

	redirectTo := r.URL.Query().Get("redirect")
	if redirectTo == "" {
		redirectTo = "/"
	}

	h.renderDevLoginPage(w, redirectTo, "")
}

// DevLoginCallback processes the dev login form (local only)
func (h *Handler) DevLoginCallback(w http.ResponseWriter, r *http.Request) {
	if !IsLocalMode(r) {
		api.Forbidden(w, "Dev login only available locally")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderDevLoginPage(w, "/", "Invalid form data")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	name := strings.TrimSpace(r.FormValue("name"))
	role := r.FormValue("role")
	redirectTo := r.FormValue("redirect")

	// Defaults
	if email == "" {
		email = "dev@example.com"
	}
	if name == "" {
		name = "Dev User"
	}
	if role == "" {
		role = "user"
	}
	if redirectTo == "" {
		redirectTo = "/"
	}

	// Validate role
	if role != "user" && role != "admin" && role != "owner" {
		role = "user"
	}

	// Create provider ID from email hash
	providerID := "dev_" + hashEmail(email)

	// Dev users don't get profile pictures by default (show initials in UI)
	picture := ""

	// Try to find existing user by provider
	user, err := h.service.GetUserByProvider("dev", providerID)
	if err == ErrUserNotFound {
		// Try by email (for account linking)
		user, err = h.service.GetUserByEmail(email)
		if err == ErrUserNotFound {
			// Create new user without picture (UI will show initials)
			user, err = h.service.CreateUser(email, name, picture, "dev", &providerID)
			if err != nil {
				h.renderDevLoginPage(w, redirectTo, "Failed to create user: "+err.Error())
				return
			}
		} else if err != nil {
			h.renderDevLoginPage(w, redirectTo, "Failed to lookup user: "+err.Error())
			return
		}
	} else if err != nil {
		h.renderDevLoginPage(w, redirectTo, "Failed to lookup user: "+err.Error())
		return
	}

	// Update role if different from current
	if user.Role != role {
		if err := h.service.UpdateUserRole(user.ID, role); err != nil {
			h.renderDevLoginPage(w, redirectTo, "Failed to update role: "+err.Error())
			return
		}
	}

	// Update profile if name changed
	if user.Name != name {
		h.service.UpdateUserProfile(user.ID, name, "")
	}

	// Create session
	sessionToken, err := h.service.CreateSession(user.ID)
	if err != nil {
		h.renderDevLoginPage(w, redirectTo, "Failed to create session: "+err.Error())
		return
	}

	// Set session cookie
	http.SetCookie(w, h.service.SessionCookie(sessionToken, int(DefaultSessionTTL.Seconds())))

	// Redirect to original destination
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// renderDevLoginPage renders the dev login form
func (h *Handler) renderDevLoginPage(w http.ResponseWriter, redirectTo, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	errorHTML := ""
	if errorMsg != "" {
		errorHTML = fmt.Sprintf(`<div class="error">%s</div>`, errorMsg)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Dev Login - Fazt</title>
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
      margin-bottom: 24px;
    }
    .warning {
      background: #3b2f1c;
      border: 1px solid #5c4d26;
      color: #fbbf24;
      padding: 12px;
      border-radius: 8px;
      margin-bottom: 24px;
      font-size: 14px;
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
    .form-group {
      margin-bottom: 16px;
    }
    label {
      display: block;
      margin-bottom: 6px;
      font-weight: 500;
      color: #ccc;
    }
    input, select {
      width: 100%%;
      padding: 12px;
      background: #1a1a1a;
      border: 1px solid #333;
      border-radius: 8px;
      color: #fff;
      font-size: 16px;
    }
    input:focus, select:focus {
      outline: none;
      border-color: #666;
    }
    button {
      width: 100%%;
      padding: 14px;
      background: #666;
      color: #fff;
      border: none;
      border-radius: 8px;
      font-size: 16px;
      font-weight: 500;
      cursor: pointer;
      margin-top: 8px;
    }
    button:hover {
      background: #777;
    }
    .back-link {
      display: block;
      text-align: center;
      margin-top: 16px;
      color: #888;
      text-decoration: none;
    }
    .back-link:hover {
      color: #fff;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Dev Login</h1>
    <p class="subtitle">Local development only</p>

    <div class="warning">
      This creates a real session for testing. Same as production OAuth.
    </div>

    %s

    <form action="/auth/dev/callback" method="POST">
      <input type="hidden" name="redirect" value="%s">

      <div class="form-group">
        <label>Email</label>
        <input type="email" name="email" value="dev@example.com" required>
      </div>

      <div class="form-group">
        <label>Name</label>
        <input type="text" name="name" value="Dev User" required>
      </div>

      <div class="form-group">
        <label>Role</label>
        <select name="role">
          <option value="user">User</option>
          <option value="admin">Admin</option>
          <option value="owner">Owner</option>
        </select>
      </div>

      <button type="submit">Sign In</button>
    </form>

    <a href="/auth/login?redirect=%s" class="back-link">Back to login</a>
  </div>
</body>
</html>`, errorHTML, redirectTo, redirectTo)

	w.Write([]byte(html))
}
