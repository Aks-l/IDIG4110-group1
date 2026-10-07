-- Migration: create_raw_messages_hypertable
-- Created: 2026-10-05T09:22:59Z
-- Description: Raw MQTT data for audit and analytics

CREATE TABLE ingest.raw_messages (
    time       timestamptz NOT NULL,
    gateway_id uuid NOT NULL,
    topic      text NOT NULL,
    payload    jsonb NOT NULL
);

SELECT create_hypertable('ingest.raw_messages', 'time');

CREATE INDEX raw_messages_gateway_time_idx ON ingest.raw_messages (gateway_id, time DESC);

SELECT add_retention_policy('ingest.raw_messages', INTERVAL '7 days');
