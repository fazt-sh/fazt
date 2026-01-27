package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// UserInfo represents user information from an OAuth provider
type UserInfo struct {
	ID      string
	Email   string
	Name    string
	Picture string
}

// OAuthProvider defines an OAuth 2.0 provider configuration
type OAuthProvider struct {
	Name         string
	DisplayName  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	ParseUser    func(data []byte) (*UserInfo, error)
	ExtraHeaders map[string]string // Extra headers for API calls
}

// ProviderConfig represents stored provider configuration
type ProviderConfig struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"-"` // Never serialize the secret
	CreatedAt    int64  `json:"created_at"`
}

// Providers contains all supported OAuth providers
var Providers = map[string]*OAuthProvider{
	"google": {
		Name:        "google",
		DisplayName: "Google",
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:    "https://oauth2.googleapis.com/token",
		UserInfoURL: "https://openidconnect.googleapis.com/v1/userinfo",
		Scopes:      []string{"openid", "email", "profile"},
		ParseUser:   parseGoogleUser,
	},
	"github": {
		Name:         "github",
		DisplayName:  "GitHub",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Scopes:       []string{"user:email"},
		ParseUser:    parseGitHubUser,
		ExtraHeaders: map[string]string{"Accept": "application/json"},
	},
	"discord": {
		Name:        "discord",
		DisplayName: "Discord",
		AuthURL:     "https://discord.com/api/oauth2/authorize",
		TokenURL:    "https://discord.com/api/oauth2/token",
		UserInfoURL: "https://discord.com/api/users/@me",
		Scopes:      []string{"identify", "email"},
		ParseUser:   parseDiscordUser,
	},
	"microsoft": {
		Name:        "microsoft",
		DisplayName: "Microsoft",
		AuthURL:     "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize",
		TokenURL:    "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
		UserInfoURL: "https://graph.microsoft.com/oidc/userinfo",
		Scopes:      []string{"openid", "email", "profile"},
		ParseUser:   parseMicrosoftUser,
	},
}

// User info parsers for each provider

func parseGoogleUser(data []byte) (*UserInfo, error) {
	var resp struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Email == "" {
		return nil, errors.New("email not provided by Google")
	}
	return &UserInfo{
		ID:      resp.Sub,
		Email:   resp.Email,
		Name:    resp.Name,
		Picture: resp.Picture,
	}, nil
}

func parseGitHubUser(data []byte) (*UserInfo, error) {
	var resp struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	name := resp.Name
	if name == "" {
		name = resp.Login
	}
	// Note: GitHub may not return email if user has it private
	// We'll need to make a separate call to /user/emails in that case
	return &UserInfo{
		ID:      fmt.Sprintf("%d", resp.ID),
		Email:   resp.Email,
		Name:    name,
		Picture: resp.AvatarURL,
	}, nil
}

func parseDiscordUser(data []byte) (*UserInfo, error) {
	var resp struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Discriminator string `json:"discriminator"`
		Email         string `json:"email"`
		Avatar        string `json:"avatar"`
		GlobalName    string `json:"global_name"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Email == "" {
		return nil, errors.New("email not provided by Discord")
	}
	name := resp.GlobalName
	if name == "" {
		name = resp.Username
	}
	var picture string
	if resp.Avatar != "" {
		picture = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", resp.ID, resp.Avatar)
	}
	return &UserInfo{
		ID:      resp.ID,
		Email:   resp.Email,
		Name:    name,
		Picture: picture,
	}, nil
}

func parseMicrosoftUser(data []byte) (*UserInfo, error) {
	var resp struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Email == "" {
		return nil, errors.New("email not provided by Microsoft")
	}
	return &UserInfo{
		ID:      resp.Sub,
		Email:   resp.Email,
		Name:    resp.Name,
		Picture: resp.Picture,
	}, nil
}

// Provider database operations

// GetProviderConfig retrieves provider configuration from the database
func (s *Service) GetProviderConfig(name string) (*ProviderConfig, error) {
	var cfg ProviderConfig
	var enabled int
	var clientSecret sql.NullString

	err := s.db.QueryRow(`
		SELECT name, enabled, client_id, client_secret, created_at
		FROM auth_providers WHERE name = ?
	`, name).Scan(&cfg.Name, &enabled, &cfg.ClientID, &clientSecret, &cfg.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrProviderDisabled
	}
	if err != nil {
		return nil, err
	}

	cfg.Enabled = enabled == 1
	if clientSecret.Valid {
		cfg.ClientSecret = clientSecret.String
	}

	return &cfg, nil
}

// SetProviderConfig creates or updates a provider configuration
func (s *Service) SetProviderConfig(name, clientID, clientSecret string) error {
	if _, ok := Providers[name]; !ok {
		return fmt.Errorf("unknown provider: %s", name)
	}

	now := time.Now().Unix()
	_, err := s.db.Exec(`
		INSERT INTO auth_providers (name, client_id, client_secret, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			client_id = excluded.client_id,
			client_secret = excluded.client_secret
	`, name, clientID, clientSecret, now)

	return err
}

// EnableProvider enables a provider
func (s *Service) EnableProvider(name string) error {
	result, err := s.db.Exec(`UPDATE auth_providers SET enabled = 1 WHERE name = ?`, name)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("provider %s not configured", name)
	}
	return nil
}

// DisableProvider disables a provider
func (s *Service) DisableProvider(name string) error {
	_, err := s.db.Exec(`UPDATE auth_providers SET enabled = 0 WHERE name = ?`, name)
	return err
}

// ListProviders returns all configured providers
func (s *Service) ListProviders() ([]*ProviderConfig, error) {
	rows, err := s.db.Query(`
		SELECT name, enabled, client_id, created_at
		FROM auth_providers ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*ProviderConfig
	for rows.Next() {
		var cfg ProviderConfig
		var enabled int
		err := rows.Scan(&cfg.Name, &enabled, &cfg.ClientID, &cfg.CreatedAt)
		if err != nil {
			continue
		}
		cfg.Enabled = enabled == 1
		providers = append(providers, &cfg)
	}

	return providers, nil
}

// GetEnabledProviders returns only enabled providers
func (s *Service) GetEnabledProviders() ([]*ProviderConfig, error) {
	rows, err := s.db.Query(`
		SELECT name, enabled, client_id, created_at
		FROM auth_providers WHERE enabled = 1 ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*ProviderConfig
	for rows.Next() {
		var cfg ProviderConfig
		var enabled int
		err := rows.Scan(&cfg.Name, &enabled, &cfg.ClientID, &cfg.CreatedAt)
		if err != nil {
			continue
		}
		cfg.Enabled = enabled == 1
		providers = append(providers, &cfg)
	}

	return providers, nil
}

// DeleteProvider removes a provider configuration
func (s *Service) DeleteProvider(name string) error {
	_, err := s.db.Exec(`DELETE FROM auth_providers WHERE name = ?`, name)
	return err
}

// OAuth flow helpers

// BuildAuthURL builds the OAuth authorization URL for a provider
func (s *Service) BuildAuthURL(providerName, state, redirectURI string) (string, error) {
	provider, ok := Providers[providerName]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", providerName)
	}

	cfg, err := s.GetProviderConfig(providerName)
	if err != nil {
		return "", err
	}
	if !cfg.Enabled {
		return "", ErrProviderDisabled
	}

	params := url.Values{}
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(provider.Scopes, " "))
	params.Set("state", state)

	// Provider-specific params
	if providerName == "google" {
		params.Set("access_type", "offline")
		params.Set("prompt", "select_account")
	}

	return provider.AuthURL + "?" + params.Encode(), nil
}

// ExchangeCode exchanges an authorization code for tokens
func (s *Service) ExchangeCode(providerName, code, redirectURI string) (string, error) {
	provider, ok := Providers[providerName]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", providerName)
	}

	cfg, err := s.GetProviderConfig(providerName)
	if err != nil {
		return "", err
	}

	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", provider.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("token error: %s", tokenResp.Error)
	}

	return tokenResp.AccessToken, nil
}

// FetchUserInfo fetches user information using an access token
func (s *Service) FetchUserInfo(providerName, accessToken string) (*UserInfo, error) {
	provider, ok := Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	req, err := http.NewRequest("GET", provider.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	for k, v := range provider.ExtraHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed: %s", string(body))
	}

	userInfo, err := provider.ParseUser(body)
	if err != nil {
		return nil, err
	}

	// For GitHub, if email is empty, fetch from /user/emails
	if providerName == "github" && userInfo.Email == "" {
		email, err := s.fetchGitHubEmail(accessToken)
		if err == nil && email != "" {
			userInfo.Email = email
		}
	}

	return userInfo, nil
}

// fetchGitHubEmail fetches the primary email from GitHub
func (s *Service) fetchGitHubEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Find primary verified email
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	// Fallback to first verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", errors.New("no verified email found")
}
