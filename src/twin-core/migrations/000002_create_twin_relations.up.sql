CREATE TABLE twin_relations (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    home_id       uuid NOT NULL REFERENCES homes (id) ON DELETE CASCADE,
    from_kind     text NOT NULL CHECK (from_kind IN ('area', 'device', 'entity')),
    from_id       uuid NOT NULL,
    to_kind       text NOT NULL CHECK (to_kind IN ('area', 'device', 'entity')),
    to_id         uuid NOT NULL,
    relation_type text NOT NULL CHECK (relation_type IN ('connects_to', 'contains', 'monitors', 'controls', 'same_physical_device', 'depends_on')),
    bidirectional boolean NOT NULL DEFAULT true,
    label         text,
    properties    jsonb,
    valid_from    timestamptz,
    valid_to      timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX twin_relations_from_idx ON twin_relations (home_id, from_kind, from_id);
CREATE INDEX twin_relations_to_idx ON twin_relations (home_id, to_kind, to_id);
