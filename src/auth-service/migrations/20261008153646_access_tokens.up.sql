-- Migration: access_tokens
-- Created: 2026-10-08T15:36:46Z
-- Description: Add description here

BEGIN;

CREATE TABLE auth.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL
        REFERENCES auth.users(id)
        ON DELETE CASCADE,

    refresh_token_hash TEXT NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;
