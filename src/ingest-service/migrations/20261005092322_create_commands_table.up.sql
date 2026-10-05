-- Migration: create_commands_table
-- Created: 2026-10-05T09:23:22Z
-- Description: Device command registry

CREATE TABLE ingest.commands (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway_id         uuid NOT NULL REFERENCES ingest.gateways (id),
    external_entity_id varchar(255) NOT NULL,
    command            varchar(100) NOT NULL,
    parameters         jsonb,
    status             varchar(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'acknowledged', 'failed', 'timeout')),
    result             jsonb,
    correlation_id     uuid NOT NULL UNIQUE,
    issued_by          varchar(255),
    issued_at          timestamptz NOT NULL DEFAULT now(),
    acked_at           timestamptz,
    expires_at         timestamptz
);

CREATE INDEX commands_gateway_issued_idx ON ingest.commands (gateway_id, issued_at DESC);
