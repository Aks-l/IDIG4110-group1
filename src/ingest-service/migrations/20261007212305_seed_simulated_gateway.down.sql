-- Migration: seed_simulated_gateway (rollback)
-- Created: 2026-10-07T21:23:05Z
-- Description: Remove the seeded simulated gateway

BEGIN;

DELETE FROM ingest.gateways
WHERE id = '00000000-0000-0000-0000-000000000001';

COMMIT;
