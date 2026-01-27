-- Migration 015: Multi-user authentication
-- OAuth providers, users, sessions, states, and invites

-- OAuth providers configuration
CREATE TABLE IF NOT EXISTS auth_providers (
    name TEXT PRIMARY KEY,           -- 'google', 'github', 'discord', 'microsoft'
    enabled INTEGER DEFAULT 0,
    client_id TEXT,
    client_secret TEXT,              -- Encrypted
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- Users
CREATE TABLE IF NOT EXISTS auth_users (
    id TEXT PRIMARY KEY,             -- UUID
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    picture TEXT,
    provider TEXT NOT NULL,          -- 'google', 'github', 'discord', 'microsoft', 'password'
    provider_id TEXT,                -- External ID (null for password)
    password_hash TEXT,              -- Only if provider='password'
    role TEXT DEFAULT 'user',        -- 'owner', 'admin', 'user'
    invited_by TEXT,                 -- User ID or 'owner'
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    last_login INTEGER
);

CREATE INDEX IF NOT EXISTS idx_auth_users_email ON auth_users(email);
CREATE INDEX IF NOT EXISTS idx_auth_users_provider ON auth_users(provider, provider_id);

-- Sessions (SQLite-backed for persistence across restarts)
CREATE TABLE IF NOT EXISTS auth_sessions (
    token_hash TEXT PRIMARY KEY,     -- SHA-256 of session token
    user_id TEXT NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    expires_at INTEGER NOT NULL,
    last_seen INTEGER
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user ON auth_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires ON auth_sessions(expires_at);

-- OAuth state tokens (temporary, CSRF protection)
CREATE TABLE IF NOT EXISTS auth_states (
    state TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    redirect_to TEXT,                -- Where to redirect after auth
    app_id TEXT,                     -- Which app initiated the auth
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    expires_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_states_expires ON auth_states(expires_at);

-- Invite codes
CREATE TABLE IF NOT EXISTS auth_invites (
    code TEXT PRIMARY KEY,
    role TEXT DEFAULT 'user',
    created_by TEXT NOT NULL,        -- User ID or 'owner'
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    expires_at INTEGER,
    max_uses INTEGER DEFAULT 1,
    use_count INTEGER DEFAULT 0,
    used_by TEXT,                    -- Last user ID that used it
    used_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_auth_invites_expires ON auth_invites(expires_at);
