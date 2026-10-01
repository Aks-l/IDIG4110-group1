CREATE TABLE raw_messages (
    time       timestamptz NOT NULL,
    gateway_id uuid NOT NULL,
    topic      text NOT NULL,
    payload    jsonb NOT NULL
);

SELECT create_hypertable('raw_messages', 'time');

CREATE INDEX raw_messages_gateway_time_idx ON raw_messages (gateway_id, time DESC);

SELECT add_retention_policy('raw_messages', INTERVAL '7 days');
