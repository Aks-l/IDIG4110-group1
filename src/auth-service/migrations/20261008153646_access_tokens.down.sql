-- Migration: access_tokens (rollback)
-- Created: 2026-10-08T15:36:46Z

BEGIN;

DROP TABLE IF EXISTS auth.sessions;

DROP INDEX IF EXISTS idx_session_user_id;
DROP INDEX IF EXISTS idx_session_expires_at;

COMMIT;
