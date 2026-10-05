-- Migration: create_ingest_schema
-- Description: Create ingest schema, deletes it on rollback

CREATE SCHEMA IF NOT EXISTS ingest;
