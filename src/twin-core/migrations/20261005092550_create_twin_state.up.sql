CREATE TABLE twin_state (
    entity_id  uuid PRIMARY KEY REFERENCES entities (id) ON DELETE CASCADE,
    value_num  double precision,
    value_text text,
    attributes jsonb,
    updated_at timestamptz NOT NULL,
    CHECK ((value_num IS NOT NULL) <> (value_text IS NOT NULL))
);
