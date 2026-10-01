CREATE TABLE readings (
    time               timestamptz NOT NULL,
    recorded_at        timestamptz NOT NULL DEFAULT now(),
    gateway_id         uuid NOT NULL,
    external_entity_id text NOT NULL,
    device_class       text,
    value_num          double precision,
    value_text         text,
    unit               text,
    attributes         jsonb,
    CHECK (value_num IS NOT NULL OR value_text IS NOT NULL)
);

SELECT create_hypertable('readings', 'time');

CREATE INDEX readings_entity_time_idx ON readings (external_entity_id, time DESC);
