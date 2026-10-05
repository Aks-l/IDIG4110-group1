-- Migration: create_gateways_table
-- Created: 2026-10-05T09:22:58Z
-- Description: Gateway registry

CREATE TABLE ingest.gateways (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name             varchar(255) NOT NULL,
    gateway_type     varchar(50) NOT NULL CHECK (gateway_type IN ('home_assistant')),
    connection_config jsonb,
    status           varchar(20) NOT NULL DEFAULT 'unknown' CHECK (status IN ('online', 'offline', 'error', 'unknown')),
    last_seen_at     timestamptz,
    protocol_version varchar(20),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (name, gateway_type)
);
