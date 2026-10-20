-- Migration: access_tokens (rollback)
-- Created: 2026-10-08T15:36:46Z

BEGIN;

DROP TABLE IF EXISTS auth.sessions;

COMMIT;
