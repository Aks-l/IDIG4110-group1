-- Migration: auth
-- Created: 2026-10-06T15:08:55Z
-- Description: Add description here

BEGIN;

-- Add your migration SQL here
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth.homes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth.home_members (
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    home_id UUID NOT NULL REFERENCES auth.homes(id) ON DELETE CASCADE,

    PRIMARY KEY (user_id, home_id)
);

COMMIT;
