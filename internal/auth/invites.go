package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/api"
)

const (
	// DefaultInviteExpiry is 7 days
	DefaultInviteExpiry = 7 * 24 * time.Hour
	// InviteCodeLength is the length of invite codes
	InviteCodeLength = 8
)

// Invite represents an invite code
type Invite struct {
	Code      string  `json:"code"`
	Role      string  `json:"role"`
	CreatedBy string  `json:"created_by"`
	CreatedAt int64   `json:"created_at"`
	ExpiresAt *int64  `json:"expires_at,omitempty"`
	MaxUses   int     `json:"max_uses"`
	UseCount  int     `json:"use_count"`
	UsedBy    *string `json:"used_by,omitempty"`
	UsedAt    *int64  `json:"used_at,omitempty"`
}

// IsValid checks if the invite is still valid
func (i *Invite) IsValid() bool {
	if i.MaxUses > 0 && i.UseCount >= i.MaxUses {
		return false
	}
	if i.ExpiresAt != nil && time.Now().Unix() > *i.ExpiresAt {
		return false
	}
	return true
}

// CreateInvite creates a new invite code
func (s *Service) CreateInvite(role, createdBy string, maxUses int, expiry *time.Duration) (*Invite, error) {
	// Generate a short, readable code
	code, err := generateInviteCode(InviteCodeLength)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	var expiresAt *int64
	if expiry != nil {
		exp := now + int64(expiry.Seconds())
		expiresAt = &exp
	} else {
		// Default expiry
		exp := now + int64(DefaultInviteExpiry.Seconds())
		expiresAt = &exp
	}

	if maxUses == 0 {
		maxUses = 1 // Default to single use
	}

	_, err = s.db.Exec(`
		INSERT INTO auth_invites (code, role, created_by, created_at, expires_at, max_uses)
		VALUES (?, ?, ?, ?, ?, ?)
	`, code, role, createdBy, now, expiresAt, maxUses)

	if err != nil {
		return nil, err
	}

	return &Invite{
		Code:      code,
		Role:      role,
		CreatedBy: createdBy,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		MaxUses:   maxUses,
		UseCount:  0,
	}, nil
}

// GetInvite retrieves an invite by code
func (s *Service) GetInvite(code string) (*Invite, error) {
	var invite Invite
	var expiresAt, usedAt sql.NullInt64
	var usedBy sql.NullString

	err := s.db.QueryRow(`
		SELECT code, role, created_by, created_at, expires_at, max_uses, use_count, used_by, used_at
		FROM auth_invites WHERE code = ?
	`, code).Scan(
		&invite.Code, &invite.Role, &invite.CreatedBy, &invite.CreatedAt,
		&expiresAt, &invite.MaxUses, &invite.UseCount, &usedBy, &usedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidInvite
	}
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Int64
	}
	if usedBy.Valid {
		invite.UsedBy = &usedBy.String
	}
	if usedAt.Valid {
		invite.UsedAt = &usedAt.Int64
	}

	return &invite, nil
}

// RedeemInvite marks an invite as used and creates a user
func (s *Service) RedeemInvite(code, email, name, password string) (*User, error) {
	invite, err := s.GetInvite(code)
	if err != nil {
		return nil, err
	}

	if !invite.IsValid() {
		return nil, ErrInviteUsed
	}

	// Create the user
	user, err := s.CreatePasswordUser(email, name, password, invite.CreatedBy)
	if err != nil {
		return nil, err
	}

	// Update the user's role if invite specifies non-default role
	if invite.Role != "user" && invite.Role != "" {
		s.UpdateUserRole(user.ID, invite.Role)
		user.Role = invite.Role
	}

	// Mark invite as used
	now := time.Now().Unix()
	s.db.Exec(`
		UPDATE auth_invites
		SET use_count = use_count + 1, used_by = ?, used_at = ?
		WHERE code = ?
	`, user.ID, now, code)

	return user, nil
}

// ListInvites returns all invites
func (s *Service) ListInvites() ([]*Invite, error) {
	rows, err := s.db.Query(`
		SELECT code, role, created_by, created_at, expires_at, max_uses, use_count, used_by, used_at
		FROM auth_invites ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*Invite
	for rows.Next() {
		var invite Invite
		var expiresAt, usedAt sql.NullInt64
		var usedBy sql.NullString

		err := rows.Scan(
			&invite.Code, &invite.Role, &invite.CreatedBy, &invite.CreatedAt,
			&expiresAt, &invite.MaxUses, &invite.UseCount, &usedBy, &usedAt,
		)
		if err != nil {
			continue
		}

		if expiresAt.Valid {
			invite.ExpiresAt = &expiresAt.Int64
		}
		if usedBy.Valid {
			invite.UsedBy = &usedBy.String
		}
		if usedAt.Valid {
			invite.UsedAt = &usedAt.Int64
		}

		invites = append(invites, &invite)
	}

	return invites, nil
}

// DeleteInvite removes an invite
func (s *Service) DeleteInvite(code string) error {
	result, err := s.db.Exec(`DELETE FROM auth_invites WHERE code = ?`, code)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrInvalidInvite
	}
	return nil
}

// generateInviteCode generates a short, readable invite code
func generateInviteCode(length int) (string, error) {
	// Use only uppercase letters and numbers, excluding confusing chars (0, O, I, 1, L)
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

	token, err := generateToken(length)
	if err != nil {
		return "", err
	}

	// Convert token bytes to charset
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = charset[int(token[i])%len(charset)]
	}

	return string(code), nil
}

// HTTP handlers for invites

// InvitePage renders the invite signup form
func (h *Handler) InvitePage(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		h.renderErrorPage(w, "Invalid invite code")
		return
	}

	invite, err := h.service.GetInvite(code)
	if err != nil {
		h.renderErrorPage(w, "Invalid or expired invite code")
		return
	}

	if !invite.IsValid() {
		h.renderErrorPage(w, "This invite code has expired or been used")
		return
	}

	h.renderInvitePage(w, code, "")
}

// RedeemInvite processes the invite signup form
func (h *Handler) RedeemInvite(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	// Check content type for JSON vs form submission
	contentType := r.Header.Get("Content-Type")
	isJSON := strings.Contains(contentType, "application/json")

	var email, name, password string

	if isJSON {
		var req struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			api.InvalidJSON(w, "Invalid request body")
			return
		}
		email = req.Email
		name = req.Name
		password = req.Password
	} else {
		r.ParseForm()
		email = r.FormValue("email")
		name = r.FormValue("name")
		password = r.FormValue("password")
	}

	// Validate
	if email == "" {
		if isJSON {
			api.BadRequest(w, "Email is required")
		} else {
			h.renderInvitePage(w, code, "Email is required")
		}
		return
	}
	if password == "" {
		if isJSON {
			api.BadRequest(w, "Password is required")
		} else {
			h.renderInvitePage(w, code, "Password is required")
		}
		return
	}
	if len(password) < 8 {
		if isJSON {
			api.BadRequest(w, "Password must be at least 8 characters")
		} else {
			h.renderInvitePage(w, code, "Password must be at least 8 characters")
		}
		return
	}

	// Redeem invite
	user, err := h.service.RedeemInvite(code, email, name, password)
	if err != nil {
		errMsg := "Failed to create account"
		if err == ErrUserExists {
			errMsg = "An account with this email already exists"
		} else if err == ErrInvalidInvite || err == ErrInviteUsed {
			errMsg = "Invalid or expired invite code"
		}
		if isJSON {
			api.BadRequest(w, errMsg)
		} else {
			h.renderInvitePage(w, code, errMsg)
		}
		return
	}

	// Create session
	sessionToken, err := h.service.CreateSession(user.ID)
	if err != nil {
		if isJSON {
			api.InternalError(w, err)
		} else {
			h.renderErrorPage(w, "Failed to create session")
		}
		return
	}

	// Set session cookie
	http.SetCookie(w, h.service.SessionCookie(sessionToken, int(DefaultSessionTTL.Seconds())))

	if isJSON {
		api.Success(w, http.StatusCreated, map[string]interface{}{
			"user":    user,
			"message": "Account created successfully",
		})
	} else {
		// Redirect to home
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// CreateInvite creates a new invite (admin only)
func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.GetSessionFromRequest(r)
	if err != nil || !user.IsAdmin() {
		api.Forbidden(w, "Admin access required")
		return
	}

	var req struct {
		Role      string `json:"role"`
		MaxUses   int    `json:"max_uses"`
		ExpiryDays int    `json:"expiry_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use defaults
		req.Role = "user"
		req.MaxUses = 1
	}

	if req.Role == "" {
		req.Role = "user"
	}

	// Only owner can create admin invites
	if req.Role == "admin" || req.Role == "owner" {
		if !user.IsOwner() {
			api.Forbidden(w, "Only owner can create admin invites")
			return
		}
	}

	var expiry *time.Duration
	if req.ExpiryDays > 0 {
		d := time.Duration(req.ExpiryDays) * 24 * time.Hour
		expiry = &d
	}

	invite, err := h.service.CreateInvite(req.Role, user.ID, req.MaxUses, expiry)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Build full URL
	scheme := "https"
	if !h.service.IsSecure() {
		scheme = "http"
	}
	inviteURL := fmt.Sprintf("%s://%s/auth/invite/%s", scheme, h.service.Domain(), invite.Code)

	api.Success(w, http.StatusCreated, map[string]interface{}{
		"code": invite.Code,
		"url":  inviteURL,
		"role": invite.Role,
	})
}

// ListInvites returns all invites (admin only)
func (h *Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.GetSessionFromRequest(r)
	if err != nil || !user.IsAdmin() {
		api.Forbidden(w, "Admin access required")
		return
	}

	invites, err := h.service.ListInvites()
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, invites)
}

// renderInvitePage renders the invite signup form
func (h *Handler) renderInvitePage(w http.ResponseWriter, code, errorMsg string) {
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
  <title>Create Account - Fazt</title>
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
    form {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    label {
      display: block;
      font-size: 14px;
      color: #888;
      margin-bottom: 6px;
    }
    input {
      width: 100%%;
      padding: 12px 16px;
      background: #1a1a1a;
      border: 1px solid #333;
      border-radius: 8px;
      color: #fff;
      font-size: 16px;
    }
    input:focus {
      outline: none;
      border-color: #666;
    }
    button {
      padding: 14px 20px;
      background: #fff;
      color: #000;
      border: none;
      border-radius: 8px;
      font-weight: 600;
      font-size: 16px;
      cursor: pointer;
      margin-top: 8px;
    }
    button:hover {
      background: #e0e0e0;
    }
    .footer {
      margin-top: 24px;
      text-align: center;
      color: #666;
      font-size: 12px;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Create Account</h1>
    <p class="subtitle">on %s</p>
    %s
    <form method="POST" action="/auth/invite/%s">
      <div>
        <label for="email">Email</label>
        <input type="email" id="email" name="email" required placeholder="you@example.com">
      </div>
      <div>
        <label for="name">Name</label>
        <input type="text" id="name" name="name" placeholder="Your name (optional)">
      </div>
      <div>
        <label for="password">Password</label>
        <input type="password" id="password" name="password" required minlength="8" placeholder="At least 8 characters">
      </div>
      <button type="submit">Create Account</button>
    </form>
    <p class="footer">Powered by Fazt</p>
  </div>
</body>
</html>`, h.service.Domain(), errorHTML, code)

	w.Write([]byte(html))
}
