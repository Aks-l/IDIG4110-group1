CREATE TABLE twin_relations (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    home_id       uuid NOT NULL REFERENCES homes (id) ON DELETE CASCADE,
    from_kind     varchar(20) NOT NULL CHECK (from_kind IN ('area', 'device', 'entity')),
    from_id       uuid NOT NULL,
    to_kind       varchar(20) NOT NULL CHECK (to_kind IN ('area', 'device', 'entity')),
    to_id         uuid NOT NULL,
    relation_type varchar(50) NOT NULL CHECK (relation_type IN ('connects_to', 'contains', 'monitors', 'controls', 'same_physical_device', 'depends_on')),
    bidirectional boolean NOT NULL DEFAULT true,
    label         varchar(255),
    properties    jsonb,
    valid_from    timestamptz,
    valid_to      timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX twin_relations_from_idx ON twin_relations (home_id, from_kind, from_id);
CREATE INDEX twin_relations_to_idx ON twin_relations (home_id, to_kind, to_id);

-- One edge per (home, from, to, relation_type): re-registering the same
-- relation conflicts instead of duplicating the edge in the graph view.
CREATE UNIQUE INDEX twin_relations_edge_unique
    ON twin_relations (home_id, from_kind, from_id, to_kind, to_id, relation_type);

ALTER TABLE twin_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE twin_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY home_isolation ON twin_relations
    USING (home_id = current_setting('app.home_id', true)::uuid);
