-- Migration 011: Source Tracking for Git Integration
-- Adds columns to track app installation source for git-based apps

-- Add source tracking columns to apps table
-- source column already exists (deploy, git, template, system)
-- Adding detailed tracking for git-sourced apps

ALTER TABLE apps ADD COLUMN source_url TEXT;
ALTER TABLE apps ADD COLUMN source_ref TEXT;
ALTER TABLE apps ADD COLUMN source_commit TEXT;
ALTER TABLE apps ADD COLUMN installed_at DATETIME;

-- Index for finding git-sourced apps (for upgrade checking)
CREATE INDEX IF NOT EXISTS idx_apps_source_git
    ON apps(source) WHERE source = 'git';
