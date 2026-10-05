-- Migration: create_readings_hypertable
-- Created: 2026-10-05T09:23:19Z
-- Description: Normalized readings

CREATE TABLE ingest.readings (
    time               timestamptz NOT NULL,
    recorded_at        timestamptz NOT NULL DEFAULT now(),
    gateway_id         uuid NOT NULL,
    external_entity_id varchar(255) NOT NULL,
    event_id           uuid,
    device_class       varchar(100),
    value_num          double precision,
    value_text         text,
    unit               varchar(50),
    attributes         jsonb,
    CHECK (value_num IS NOT NULL OR value_text IS NOT NULL)
);

SELECT create_hypertable('ingest.readings', 'time');

CREATE INDEX readings_entity_time_idx ON ingest.readings (external_entity_id, time DESC);
