package auth

import (
	"database/sql"
	"time"
)

const (
	// StateExpiry is how long OAuth states are valid (10 minutes)
	StateExpiry = 10 * time.Minute
)

// OAuthState represents a pending OAuth flow
type OAuthState struct {
	State      string
	Provider   string
	RedirectTo string
	AppID      string
	CreatedAt  int64
	ExpiresAt  int64
}

// CreateState creates a new OAuth state token for CSRF protection
func (s *Service) CreateState(provider, redirectTo, appID string) (string, error) {
	state, err := generateToken(16)
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	expiresAt := now + int64(StateExpiry.Seconds())

	_, err = s.db.Exec(`
		INSERT INTO auth_states (state, provider, redirect_to, app_id, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, state, provider, redirectTo, appID, now, expiresAt)

	if err != nil {
		return "", err
	}

	return state, nil
}

// ValidateState validates and consumes an OAuth state token
func (s *Service) ValidateState(state string) (*OAuthState, error) {
	if state == "" {
		return nil, ErrInvalidState
	}

	var oauthState OAuthState
	var redirectTo, appID sql.NullString

	err := s.db.QueryRow(`
		SELECT state, provider, redirect_to, app_id, created_at, expires_at
		FROM auth_states WHERE state = ?
	`, state).Scan(
		&oauthState.State, &oauthState.Provider,
		&redirectTo, &appID,
		&oauthState.CreatedAt, &oauthState.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidState
	}
	if err != nil {
		return nil, err
	}

	// Always delete the state (one-time use)
	s.db.Exec(`DELETE FROM auth_states WHERE state = ?`, state)

	// Check if expired
	if time.Now().Unix() > oauthState.ExpiresAt {
		return nil, ErrInvalidState
	}

	if redirectTo.Valid {
		oauthState.RedirectTo = redirectTo.String
	}
	if appID.Valid {
		oauthState.AppID = appID.String
	}

	return &oauthState, nil
}

// HandleOAuthCallback processes an OAuth callback
// Returns the user (existing or newly created) and an error
func (s *Service) HandleOAuthCallback(providerName, code, redirectURI string) (*User, error) {
	// Exchange code for access token
	accessToken, err := s.ExchangeCode(providerName, code, redirectURI)
	if err != nil {
		return nil, err
	}

	// Fetch user info from provider
	userInfo, err := s.FetchUserInfo(providerName, accessToken)
	if err != nil {
		return nil, err
	}

	// Check if user already exists by provider ID
	user, err := s.GetUserByProvider(providerName, userInfo.ID)
	if err == nil {
		// User exists, update profile and return
		s.UpdateUserProfile(user.ID, userInfo.Name, userInfo.Picture)
		s.UpdateLastLogin(user.ID)
		return user, nil
	}

	// Check if user exists by email (linking accounts)
	user, err = s.GetUserByEmail(userInfo.Email)
	if err == nil {
		// User exists with same email but different provider
		// Update their provider info (link account)
		s.db.Exec(`
			UPDATE auth_users
			SET provider = ?, provider_id = ?, name = ?, picture = ?, last_login = ?
			WHERE id = ?
		`, providerName, userInfo.ID, userInfo.Name, userInfo.Picture, time.Now().Unix(), user.ID)
		return s.GetUserByID(user.ID)
	}

	// Create new user
	providerID := userInfo.ID
	return s.CreateUser(userInfo.Email, userInfo.Name, userInfo.Picture, providerName, &providerID)
}

// StartOAuthFlow initiates an OAuth flow for a provider
// Returns the authorization URL to redirect the user to
func (s *Service) StartOAuthFlow(providerName, redirectTo, appID, callbackURL string) (string, error) {
	// Create state token
	state, err := s.CreateState(providerName, redirectTo, appID)
	if err != nil {
		return "", err
	}

	// Build authorization URL
	authURL, err := s.BuildAuthURL(providerName, state, callbackURL)
	if err != nil {
		return "", err
	}

	return authURL, nil
}

// CompleteOAuthFlow processes the OAuth callback and creates a session
// Returns the session token
func (s *Service) CompleteOAuthFlow(providerName, code, state, callbackURL string) (string, *User, string, error) {
	// Validate state
	oauthState, err := s.ValidateState(state)
	if err != nil {
		return "", nil, "", err
	}

	// Verify provider matches
	if oauthState.Provider != providerName {
		return "", nil, "", ErrInvalidState
	}

	// Handle the OAuth callback
	user, err := s.HandleOAuthCallback(providerName, code, callbackURL)
	if err != nil {
		return "", nil, "", err
	}

	// Create session
	sessionToken, err := s.CreateSession(user.ID)
	if err != nil {
		return "", nil, "", err
	}

	return sessionToken, user, oauthState.RedirectTo, nil
}
