CREATE TABLE gateways (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name             text NOT NULL,
    gateway_type     text NOT NULL CHECK (gateway_type IN ('home_assistant')),
    connection_config jsonb,
    status           text NOT NULL DEFAULT 'unknown' CHECK (status IN ('online', 'offline', 'error', 'unknown')),
    last_seen_at     timestamptz,
    protocol_version text,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (name, gateway_type)
);
