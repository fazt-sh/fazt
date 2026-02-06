package egress

import (
	"database/sql"
	"strings"
	"sync"
	"time"
)

// AllowlistEntry represents a single domain allowlist entry.
type AllowlistEntry struct {
	ID          int64
	Domain      string
	AppID       string // empty = global
	HTTPSOnly   bool
	RateLimit   int   // req/min, 0 = use system default
	RateBurst   int   // 0 = use system default
	MaxResponse int64 // 0 = use system default
	TimeoutMs   int   // 0 = use system default
	CacheTTL    int   // seconds, 0 = no cache
	CreatedAt   int64
}

// Allowlist provides domain-level access control for egress requests.
type Allowlist struct {
	db       *sql.DB
	cache    map[string][]AllowlistEntry // keyed by appID ("" = global)
	mu       sync.RWMutex
	loadedAt time.Time
	ttl      time.Duration
}

// NewAllowlist creates an Allowlist backed by the given database.
func NewAllowlist(db *sql.DB) *Allowlist {
	return &Allowlist{
		db:  db,
		ttl: 30 * time.Second,
	}
}

// IsAllowed returns true if the domain is allowed for the given app.
// Checks app-scoped entries first, then global entries.
func (a *Allowlist) IsAllowed(domain string, appID string) bool {
	domain = canonicalizeHost(domain)
	entries := a.getEntries(appID)

	// Check app-scoped entries
	for _, e := range entries {
		if matchDomain(e.Domain, domain) {
			return true
		}
	}

	// Check global entries if appID is not empty
	if appID != "" {
		global := a.getEntries("")
		for _, e := range global {
			if matchDomain(e.Domain, domain) {
				return true
			}
		}
	}

	return false
}

// entryFor returns the matching AllowlistEntry for a domain, or nil.
func (a *Allowlist) entryFor(domain string, appID string) *AllowlistEntry {
	domain = canonicalizeHost(domain)

	// Check app-scoped first
	entries := a.getEntries(appID)
	for _, e := range entries {
		if matchDomain(e.Domain, domain) {
			return &e
		}
	}

	// Check global
	if appID != "" {
		global := a.getEntries("")
		for _, e := range global {
			if matchDomain(e.Domain, domain) {
				return &e
			}
		}
	}

	return nil
}

// Add adds a domain to the allowlist.
func (a *Allowlist) Add(domain string, appID string, httpsOnly bool) error {
	domain = canonicalizeHost(domain)

	// Reject bare wildcards
	if domain == "*" {
		return errBlocked("bare wildcard (*) not allowed")
	}

	appIDVal := sql.NullString{}
	if appID != "" {
		appIDVal = sql.NullString{String: appID, Valid: true}
	}

	_, err := a.db.Exec(`
		INSERT INTO net_allowlist (domain, app_id, https_only)
		VALUES (?, ?, ?)
		ON CONFLICT(domain, app_id) DO UPDATE SET https_only = excluded.https_only
	`, domain, appIDVal, boolToInt(httpsOnly))
	if err != nil {
		return err
	}

	a.invalidateCache()
	return nil
}

// Remove removes a domain from the allowlist.
func (a *Allowlist) Remove(domain string, appID string) error {
	domain = canonicalizeHost(domain)

	appIDVal := sql.NullString{}
	if appID != "" {
		appIDVal = sql.NullString{String: appID, Valid: true}
	}

	result, err := a.db.Exec(`
		DELETE FROM net_allowlist WHERE domain = ? AND app_id IS ?
	`, domain, appIDVal)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errBlocked("domain not found in allowlist")
	}

	a.invalidateCache()
	return nil
}

// List returns all allowlist entries for an app (or global if appID is "").
func (a *Allowlist) List(appID string) ([]AllowlistEntry, error) {
	var rows *sql.Rows
	var err error

	if appID == "" {
		rows, err = a.db.Query(`
			SELECT id, domain, COALESCE(app_id, ''), https_only,
			       COALESCE(rate_limit, 0), COALESCE(rate_burst, 0),
			       COALESCE(max_response, 0), COALESCE(timeout_ms, 0),
			       COALESCE(cache_ttl, 0), created_at
			FROM net_allowlist ORDER BY domain
		`)
	} else {
		rows, err = a.db.Query(`
			SELECT id, domain, COALESCE(app_id, ''), https_only,
			       COALESCE(rate_limit, 0), COALESCE(rate_burst, 0),
			       COALESCE(max_response, 0), COALESCE(timeout_ms, 0),
			       COALESCE(cache_ttl, 0), created_at
			FROM net_allowlist WHERE app_id = ? OR app_id IS NULL
			ORDER BY domain
		`, appID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AllowlistEntry
	for rows.Next() {
		var e AllowlistEntry
		var httpsOnly int
		if err := rows.Scan(&e.ID, &e.Domain, &e.AppID, &httpsOnly,
			&e.RateLimit, &e.RateBurst, &e.MaxResponse, &e.TimeoutMs,
			&e.CacheTTL, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.HTTPSOnly = httpsOnly != 0
		entries = append(entries, e)
	}
	return entries, nil
}

// getEntries returns cached entries for the given appID.
func (a *Allowlist) getEntries(appID string) []AllowlistEntry {
	a.mu.RLock()
	if a.cache != nil && time.Since(a.loadedAt) < a.ttl {
		entries := a.cache[appID]
		a.mu.RUnlock()
		return entries
	}
	a.mu.RUnlock()

	a.reload()

	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cache[appID]
}

// reload loads all entries from the database into the cache.
func (a *Allowlist) reload() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring write lock
	if a.cache != nil && time.Since(a.loadedAt) < a.ttl {
		return
	}

	cache := make(map[string][]AllowlistEntry)

	rows, err := a.db.Query(`
		SELECT id, domain, COALESCE(app_id, ''), https_only,
		       COALESCE(rate_limit, 0), COALESCE(rate_burst, 0),
		       COALESCE(max_response, 0), COALESCE(timeout_ms, 0),
		       COALESCE(cache_ttl, 0), created_at
		FROM net_allowlist
	`)
	if err != nil {
		a.cache = cache
		a.loadedAt = time.Now()
		return
	}
	defer rows.Close()

	for rows.Next() {
		var e AllowlistEntry
		var httpsOnly int
		if err := rows.Scan(&e.ID, &e.Domain, &e.AppID, &httpsOnly,
			&e.RateLimit, &e.RateBurst, &e.MaxResponse, &e.TimeoutMs,
			&e.CacheTTL, &e.CreatedAt); err != nil {
			continue
		}
		e.HTTPSOnly = httpsOnly != 0
		cache[e.AppID] = append(cache[e.AppID], e)
	}

	a.cache = cache
	a.loadedAt = time.Now()
}

// invalidateCache forces a reload on next access.
func (a *Allowlist) invalidateCache() {
	a.mu.Lock()
	a.cache = nil
	a.mu.Unlock()
}

// matchDomain checks if a pattern matches a domain.
// Exact: "api.stripe.com" matches "api.stripe.com"
// Wildcard: "*.googleapis.com" matches "maps.googleapis.com" but not "googleapis.com"
func matchDomain(pattern, domain string) bool {
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // ".googleapis.com"
		return strings.HasSuffix(domain, suffix) && domain != pattern[2:]
	}
	return pattern == domain
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
