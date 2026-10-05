-- Migration: create_raw_messages_hypertable (down)
-- Created: 2026-10-05T09:22:59Z

SELECT remove_retention_policy('ingest.raw_messages', if_exists => true);

DROP TABLE IF EXISTS ingest.raw_messages;
