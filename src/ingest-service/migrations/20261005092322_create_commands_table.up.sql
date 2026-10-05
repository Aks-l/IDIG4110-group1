-- Migration: create_commands_table
-- Created: 2026-10-05T09:23:22Z
-- Description: Device command registry

CREATE TABLE ingest.commands (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway_id         uuid NOT NULL REFERENCES ingest.gateways (id),
    external_entity_id text NOT NULL,
    command            text NOT NULL,
    parameters         jsonb,
    status             text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'acknowledged', 'failed', 'timeout')),
    result             jsonb,
    correlation_id     uuid NOT NULL UNIQUE,
    issued_by          text,
    issued_at          timestamptz NOT NULL DEFAULT now(),
    acked_at           timestamptz,
    expires_at         timestamptz
);

CREATE INDEX commands_gateway_issued_idx ON ingest.commands (gateway_id, issued_at DESC);
