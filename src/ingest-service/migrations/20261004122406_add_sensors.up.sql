-- Migration: add_sensors
-- Created: 2026-10-04T12:24:06Z
-- Description: This is not the correct schema for sensors, only for test

BEGIN;

CREATE SCHEMA IF NOT EXISTS ingest;

CREATE TABLE IF NOT EXISTS ingest.sensor_data (
    id          BIGSERIAL PRIMARY KEY,
    entityID    VARCHAR(255) NOT NULL,
    event_type  VARCHAR(100) NOT NULL,
    time_fired  TIMESTAMPTZ NOT NULL
);


COMMIT;
