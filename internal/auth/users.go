package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/fazt-sh/fazt/internal/appid"
	"golang.org/x/crypto/bcrypt"
)

// User represents an authenticated user
type User struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	Name       string  `json:"name,omitempty"`
	Picture    string  `json:"picture,omitempty"`
	Provider   string  `json:"provider"`
	ProviderID *string `json:"-"` // Not exposed in JSON
	Role       string  `json:"role"`
	InvitedBy  *string `json:"invited_by,omitempty"`
	CreatedAt  int64   `json:"created_at"`
	LastLogin  *int64  `json:"last_login,omitempty"`
}

// IsOwner returns true if the user has owner role
func (u *User) IsOwner() bool {
	return u.Role == "owner"
}

// IsAdmin returns true if the user has admin or owner role
func (u *User) IsAdmin() bool {
	return u.Role == "owner" || u.Role == "admin"
}

// CreateUser creates a new user in the database
func (s *Service) CreateUser(email, name, picture, provider string, providerID *string) (*User, error) {
	// Check if user already exists
	existing, _ := s.GetUserByEmail(email)
	if existing != nil {
		return nil, ErrUserExists
	}

	id := appid.GenerateUser()
	now := time.Now().Unix()

	// OAuth users are always regular users
	// Owner is the server admin (separate auth flow)
	role := "user"

	_, err := s.db.Exec(`
		INSERT INTO auth_users (id, email, name, picture, provider, provider_id, role, created_at, last_login)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, email, name, picture, provider, providerID, role, now, now)

	if err != nil {
		return nil, err
	}

	return &User{
		ID:         id,
		Email:      email,
		Name:       name,
		Picture:    picture,
		Provider:   provider,
		ProviderID: providerID,
		Role:       role,
		CreatedAt:  now,
		LastLogin:  &now,
	}, nil
}

// CreatePasswordUser creates a user with password authentication
func (s *Service) CreatePasswordUser(email, name, password, invitedBy string) (*User, error) {
	// Check if user already exists
	existing, _ := s.GetUserByEmail(email)
	if existing != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	id := appid.GenerateUser()
	now := time.Now().Unix()

	// Password users via invite are always regular users
	// Owner is the server admin (separate auth flow)
	role := "user"

	var invitedByPtr *string
	if invitedBy != "" {
		invitedByPtr = &invitedBy
	}

	_, err = s.db.Exec(`
		INSERT INTO auth_users (id, email, name, provider, password_hash, role, invited_by, created_at, last_login)
		VALUES (?, ?, ?, 'password', ?, ?, ?, ?, ?)
	`, id, email, name, string(hash), role, invitedByPtr, now, now)

	if err != nil {
		return nil, err
	}

	return &User{
		ID:        id,
		Email:     email,
		Name:      name,
		Provider:  "password",
		Role:      role,
		InvitedBy: invitedByPtr,
		CreatedAt: now,
		LastLogin: &now,
	}, nil
}

// GetUserByID retrieves a user by their ID
func (s *Service) GetUserByID(id string) (*User, error) {
	var user User
	var providerID sql.NullString
	var invitedBy sql.NullString
	var lastLogin sql.NullInt64

	err := s.db.QueryRow(`
		SELECT id, email, name, picture, provider, provider_id, role, invited_by, created_at, last_login
		FROM auth_users WHERE id = ?
	`, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Picture,
		&user.Provider, &providerID, &user.Role, &invitedBy,
		&user.CreatedAt, &lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if providerID.Valid {
		user.ProviderID = &providerID.String
	}
	if invitedBy.Valid {
		user.InvitedBy = &invitedBy.String
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Int64
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (s *Service) GetUserByEmail(email string) (*User, error) {
	var user User
	var providerID sql.NullString
	var invitedBy sql.NullString
	var lastLogin sql.NullInt64

	err := s.db.QueryRow(`
		SELECT id, email, name, picture, provider, provider_id, role, invited_by, created_at, last_login
		FROM auth_users WHERE email = ?
	`, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Picture,
		&user.Provider, &providerID, &user.Role, &invitedBy,
		&user.CreatedAt, &lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if providerID.Valid {
		user.ProviderID = &providerID.String
	}
	if invitedBy.Valid {
		user.InvitedBy = &invitedBy.String
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Int64
	}

	return &user, nil
}

// GetUserByProvider retrieves a user by their OAuth provider and provider ID
func (s *Service) GetUserByProvider(provider, providerID string) (*User, error) {
	var user User
	var pid sql.NullString
	var invitedBy sql.NullString
	var lastLogin sql.NullInt64

	err := s.db.QueryRow(`
		SELECT id, email, name, picture, provider, provider_id, role, invited_by, created_at, last_login
		FROM auth_users WHERE provider = ? AND provider_id = ?
	`, provider, providerID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Picture,
		&user.Provider, &pid, &user.Role, &invitedBy,
		&user.CreatedAt, &lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if pid.Valid {
		user.ProviderID = &pid.String
	}
	if invitedBy.Valid {
		user.InvitedBy = &invitedBy.String
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Int64
	}

	return &user, nil
}

// UpdateLastLogin updates the user's last login timestamp
func (s *Service) UpdateLastLogin(userID string) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(`UPDATE auth_users SET last_login = ? WHERE id = ?`, now, userID)
	return err
}

// UpdateUserRole updates a user's role
func (s *Service) UpdateUserRole(userID, role string) error {
	// Validate role
	if role != "owner" && role != "admin" && role != "user" {
		return errors.New("invalid role: must be owner, admin, or user")
	}
	_, err := s.db.Exec(`UPDATE auth_users SET role = ? WHERE id = ?`, role, userID)
	return err
}

// UpdateUserProfile updates a user's name and picture
func (s *Service) UpdateUserProfile(userID, name, picture string) error {
	_, err := s.db.Exec(`UPDATE auth_users SET name = ?, picture = ? WHERE id = ?`, name, picture, userID)
	return err
}

// DeleteUser removes a user from the database
func (s *Service) DeleteUser(userID string) error {
	// First delete all sessions for this user
	s.db.Exec(`DELETE FROM auth_sessions WHERE user_id = ?`, userID)

	result, err := s.db.Exec(`DELETE FROM auth_users WHERE id = ?`, userID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsers returns all users
func (s *Service) ListUsers() ([]*User, error) {
	rows, err := s.db.Query(`
		SELECT id, email, name, picture, provider, provider_id, role, invited_by, created_at, last_login
		FROM auth_users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		var providerID sql.NullString
		var invitedBy sql.NullString
		var lastLogin sql.NullInt64

		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.Picture,
			&user.Provider, &providerID, &user.Role, &invitedBy,
			&user.CreatedAt, &lastLogin,
		)
		if err != nil {
			continue
		}

		if providerID.Valid {
			user.ProviderID = &providerID.String
		}
		if invitedBy.Valid {
			user.InvitedBy = &invitedBy.String
		}
		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Int64
		}

		users = append(users, &user)
	}

	return users, nil
}

// VerifyPassword checks if the provided password matches the user's stored hash
func (s *Service) VerifyPassword(userID, password string) (bool, error) {
	var hash sql.NullString
	err := s.db.QueryRow(`SELECT password_hash FROM auth_users WHERE id = ?`, userID).Scan(&hash)
	if err == sql.ErrNoRows {
		return false, ErrUserNotFound
	}
	if err != nil {
		return false, err
	}
	if !hash.Valid || hash.String == "" {
		return false, errors.New("user does not have password authentication")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(password))
	return err == nil, nil
}

// GetOwner returns the owner user (first user)
func (s *Service) GetOwner() (*User, error) {
	var user User
	var providerID sql.NullString
	var invitedBy sql.NullString
	var lastLogin sql.NullInt64

	err := s.db.QueryRow(`
		SELECT id, email, name, picture, provider, provider_id, role, invited_by, created_at, last_login
		FROM auth_users WHERE role = 'owner' LIMIT 1
	`).Scan(
		&user.ID, &user.Email, &user.Name, &user.Picture,
		&user.Provider, &providerID, &user.Role, &invitedBy,
		&user.CreatedAt, &lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if providerID.Valid {
		user.ProviderID = &providerID.String
	}
	if invitedBy.Valid {
		user.InvitedBy = &invitedBy.String
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Int64
	}

	return &user, nil
}

// CountUsers returns the total number of users
func (s *Service) CountUsers() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM auth_users`).Scan(&count)
	return count, err
}

