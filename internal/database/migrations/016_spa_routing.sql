-- Migration 016: SPA Routing Support
-- Add spa column to apps table for clean URL routing

ALTER TABLE apps ADD COLUMN spa INTEGER DEFAULT 0;
