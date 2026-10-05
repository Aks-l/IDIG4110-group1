CREATE TABLE devices (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    home_id      uuid NOT NULL REFERENCES homes (id) ON DELETE CASCADE,
    area_id      uuid REFERENCES areas (id) ON DELETE SET NULL,
    gateway_id   uuid NOT NULL,
    external_id  varchar(255) NOT NULL,
    name         varchar(255) NOT NULL,
    manufacturer varchar(255),
    model        varchar(255),
    sw_version   varchar(100),
    device_type  varchar(50),
    status       varchar(20) NOT NULL DEFAULT 'unknown',
    last_seen_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (gateway_id, external_id)
);

CREATE TABLE entities (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id          uuid NOT NULL REFERENCES devices (id) ON DELETE CASCADE,
    area_id            uuid REFERENCES areas (id) ON DELETE SET NULL,
    external_entity_id varchar(255) NOT NULL,
    name               varchar(255),
    domain             varchar(50) NOT NULL,
    device_class       varchar(100),
    unit               varchar(50),
    controllable       boolean NOT NULL DEFAULT false,
    command_map        jsonb,
    state_ttl_seconds  int,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    UNIQUE (device_id, external_entity_id)
);

CREATE INDEX entities_area_idx ON entities (area_id);

-- Event lookup hot path: readings arrive with the source's entity id.
-- Resolve (device -> gateway, external_entity_id) without a sequential scan.
CREATE INDEX entities_external_entity_idx ON entities (external_entity_id);
