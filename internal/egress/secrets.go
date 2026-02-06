package egress

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Secret represents a stored credential for outbound HTTP injection.
type Secret struct {
	ID        int64
	AppID     string
	Name      string
	Value     string
	InjectAs  string // "bearer", "header", "query"
	InjectKey string // Header name or query param name
	Domain    string // Only inject for this domain (empty = any)
	CreatedAt int64
	UpdatedAt int64
}

// SecretsStore manages stored credentials for outbound HTTP requests.
type SecretsStore struct {
	db       *sql.DB
	cache    map[string][]Secret // keyed by appID ("" = global)
	mu       sync.RWMutex
	loadedAt time.Time
	ttl      time.Duration
}

// NewSecretsStore creates a SecretsStore backed by the given database.
func NewSecretsStore(db *sql.DB) *SecretsStore {
	return &SecretsStore{
		db:  db,
		ttl: 30 * time.Second,
	}
}

// Set creates or updates a secret.
func (s *SecretsStore) Set(name, value, injectAs, injectKey, domain, appID string) error {
	// Validate inject_as
	switch injectAs {
	case "bearer", "header", "query":
	default:
		return fmt.Errorf("invalid inject_as: %q (must be bearer, header, or query)", injectAs)
	}

	// Require inject_key for header and query
	if (injectAs == "header" || injectAs == "query") && injectKey == "" {
		return fmt.Errorf("inject_key required for inject_as=%q", injectAs)
	}

	appIDVal := sql.NullString{}
	if appID != "" {
		appIDVal = sql.NullString{String: appID, Valid: true}
	}

	domainVal := sql.NullString{}
	if domain != "" {
		domainVal = sql.NullString{String: canonicalizeHost(domain), Valid: true}
	}

	_, err := s.db.Exec(`
		INSERT INTO net_secrets (app_id, name, value, inject_as, inject_key, domain)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(app_id, name) DO UPDATE SET
			value = excluded.value,
			inject_as = excluded.inject_as,
			inject_key = excluded.inject_key,
			domain = excluded.domain,
			updated_at = unixepoch()
	`, appIDVal, name, value, injectAs, injectKey, domainVal)
	if err != nil {
		return err
	}

	s.invalidateCache()
	return nil
}

// Remove deletes a secret.
func (s *SecretsStore) Remove(name, appID string) error {
	appIDVal := sql.NullString{}
	if appID != "" {
		appIDVal = sql.NullString{String: appID, Valid: true}
	}

	result, err := s.db.Exec(`DELETE FROM net_secrets WHERE name = ? AND app_id IS ?`, name, appIDVal)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("secret %q not found", name)
	}

	s.invalidateCache()
	return nil
}

// List returns all secrets (with values masked).
func (s *SecretsStore) List(appID string) ([]Secret, error) {
	var rows *sql.Rows
	var err error

	if appID == "" {
		rows, err = s.db.Query(`
			SELECT id, COALESCE(app_id, ''), name, value, inject_as,
			       COALESCE(inject_key, ''), COALESCE(domain, ''),
			       created_at, updated_at
			FROM net_secrets ORDER BY name
		`)
	} else {
		rows, err = s.db.Query(`
			SELECT id, COALESCE(app_id, ''), name, value, inject_as,
			       COALESCE(inject_key, ''), COALESCE(domain, ''),
			       created_at, updated_at
			FROM net_secrets WHERE app_id = ? OR app_id IS NULL
			ORDER BY name
		`, appID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []Secret
	for rows.Next() {
		var sec Secret
		if err := rows.Scan(&sec.ID, &sec.AppID, &sec.Name, &sec.Value,
			&sec.InjectAs, &sec.InjectKey, &sec.Domain,
			&sec.CreatedAt, &sec.UpdatedAt); err != nil {
			return nil, err
		}
		secrets = append(secrets, sec)
	}
	return secrets, nil
}

// Lookup finds a secret by name, checking app-scoped first then global.
func (s *SecretsStore) Lookup(name, appID string) (*Secret, error) {
	secrets := s.getSecrets(appID)
	for _, sec := range secrets {
		if sec.Name == name {
			return &sec, nil
		}
	}
	// Check global if app-scoped miss
	if appID != "" {
		globals := s.getSecrets("")
		for _, sec := range globals {
			if sec.Name == name {
				return &sec, nil
			}
		}
	}
	return nil, errAuth(fmt.Sprintf("secret %q not found", name))
}

// InjectAuth applies a secret to an HTTP request based on its inject_as type.
func (s *SecretsStore) InjectAuth(req *http.Request, secret *Secret, targetDomain string) error {
	// Check domain restriction
	if secret.Domain != "" && secret.Domain != targetDomain {
		return errAuth(fmt.Sprintf("secret %q restricted to domain %s, not %s",
			secret.Name, secret.Domain, targetDomain))
	}

	switch secret.InjectAs {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+secret.Value)
	case "header":
		req.Header.Set(secret.InjectKey, secret.Value)
	case "query":
		q := req.URL.Query()
		q.Set(secret.InjectKey, secret.Value)
		req.URL.RawQuery = q.Encode()
	default:
		return errAuth(fmt.Sprintf("unknown inject_as: %q", secret.InjectAs))
	}

	return nil
}

func (s *SecretsStore) getSecrets(appID string) []Secret {
	s.mu.RLock()
	if s.cache != nil && time.Since(s.loadedAt) < s.ttl {
		secrets := s.cache[appID]
		s.mu.RUnlock()
		return secrets
	}
	s.mu.RUnlock()

	s.reload()

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[appID]
}

func (s *SecretsStore) reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache != nil && time.Since(s.loadedAt) < s.ttl {
		return
	}

	cache := make(map[string][]Secret)

	rows, err := s.db.Query(`
		SELECT id, COALESCE(app_id, ''), name, value, inject_as,
		       COALESCE(inject_key, ''), COALESCE(domain, ''),
		       created_at, updated_at
		FROM net_secrets
	`)
	if err != nil {
		s.cache = cache
		s.loadedAt = time.Now()
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sec Secret
		if err := rows.Scan(&sec.ID, &sec.AppID, &sec.Name, &sec.Value,
			&sec.InjectAs, &sec.InjectKey, &sec.Domain,
			&sec.CreatedAt, &sec.UpdatedAt); err != nil {
			continue
		}
		cache[sec.AppID] = append(cache[sec.AppID], sec)
	}

	s.cache = cache
	s.loadedAt = time.Now()
}

func (s *SecretsStore) invalidateCache() {
	s.mu.Lock()
	s.cache = nil
	s.mu.Unlock()
}

// MaskValue returns a masked version of a secret value for display.
func MaskValue(value string) string {
	if len(value) <= 6 {
		return strings.Repeat("*", len(value))
	}
	return value[:3] + strings.Repeat("*", len(value)-6) + value[len(value)-3:]
}

// InjectSecretIntoRequest looks up a secret and injects it into a URL/request.
// Used by the proxy when opts.Auth is set.
func InjectSecretIntoRequest(secrets *SecretsStore, req *http.Request,
	authName, appID, targetDomain string) error {

	if secrets == nil || authName == "" {
		return nil
	}

	secret, err := secrets.Lookup(authName, appID)
	if err != nil {
		return err
	}

	return secrets.InjectAuth(req, secret, targetDomain)
}

