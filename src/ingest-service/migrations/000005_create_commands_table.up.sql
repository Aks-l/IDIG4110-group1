CREATE TABLE commands (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway_id         uuid NOT NULL REFERENCES gateways (id),
    external_entity_id text NOT NULL,
    command            text NOT NULL,
    parameters         jsonb,
    status             text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'acknowledged', 'failed', 'timeout')),
    result             jsonb,
    correlation_id     uuid NOT NULL UNIQUE,
    issued_by          text,
    issued_at          timestamptz NOT NULL DEFAULT now(),
    acked_at           timestamptz,
    expires_at         timestamptz
);

CREATE INDEX commands_gateway_issued_idx ON commands (gateway_id, issued_at DESC);
