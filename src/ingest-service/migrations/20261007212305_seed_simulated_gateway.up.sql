-- Migration: seed_simulated_gateway
-- Created: 2026-10-07T21:23:05Z
-- Description: Seed the simulated gateway used by the device-simulator

BEGIN;

INSERT INTO ingest.gateways (id, name, gateway_type, status, last_seen_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'simulated-gateway',
    'home_assistant',
    'online',
    now()
) ON CONFLICT (name, gateway_type) DO NOTHING;

COMMIT;
