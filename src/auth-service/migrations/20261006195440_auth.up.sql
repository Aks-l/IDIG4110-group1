-- Migration: auth
-- Created: 2026-10-06T19:54:40Z
-- Description: Add description here

BEGIN;

CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth.homes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth.home_members (
    user_id UUID NOT NULL REFERENCES auth.users(id),
    home_id UUID NOT NULL REFERENCES auth.homes(id),
    PRIMARY KEY (user_id, home_id)
);

COMMIT;
