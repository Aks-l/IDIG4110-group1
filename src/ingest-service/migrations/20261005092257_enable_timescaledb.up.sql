-- Migration: enable_timescaledb
-- Created: 2026-10-05T09:22:57Z
-- Description: Enable the TimescaleDB

CREATE EXTENSION IF NOT EXISTS timescaledb;
