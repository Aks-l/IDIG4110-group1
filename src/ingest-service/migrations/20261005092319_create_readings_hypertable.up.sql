-- Migration: create_readings_hypertable
-- Created: 2026-10-05T09:23:19Z
-- Description: Normalized readings

CREATE TABLE ingest.readings (
    time               timestamptz NOT NULL,
    recorded_at        timestamptz NOT NULL DEFAULT now(),
    gateway_id         uuid NOT NULL,
    external_entity_id text NOT NULL,
    event_id           uuid,
    device_class       text,
    value_num          double precision,
    value_text         text,
    unit               text,
    attributes         jsonb,
    CHECK (value_num IS NOT NULL OR value_text IS NOT NULL)
);

SELECT create_hypertable('ingest.readings', 'time');

CREATE INDEX readings_entity_time_idx ON ingest.readings (external_entity_id, time DESC);
