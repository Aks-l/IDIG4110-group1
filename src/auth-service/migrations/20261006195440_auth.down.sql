-- Migration: auth (rollback)
-- Created: 2026-10-06T19:54:40Z

BEGIN;

-- Add your rollback SQL here

DROP TABLE IF EXISTS auth.users;

DROP TABLE IF EXISTS auth.homes;

DROP TABLE IF EXISTS auth.home_members

DROP INDEX IF EXISTS idx_user_email;

DROP SCHEMA IF EXISTS auth CASCADE;

COMMIT;
