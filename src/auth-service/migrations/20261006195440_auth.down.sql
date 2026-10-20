-- Migration: auth (rollback)
-- Created: 2026-10-06T19:54:40Z

BEGIN;

-- Add your rollback SQL here

DROP TABLE IF EXISTS auth.home_members

DROP TABLE IF EXISTS auth.users;

DROP TABLE IF EXISTS auth.homes;

DROP SCHEMA IF EXISTS auth CASCADE;

COMMIT;
