-- Migration: add_sensors (rollback)
-- Created: 2026-10-04T12:24:06Z

BEGIN;

DROP TABLE IF EXISTS ingest.sensor_data;
DROP SCHEMA IF EXISTS ingest;

COMMIT;
