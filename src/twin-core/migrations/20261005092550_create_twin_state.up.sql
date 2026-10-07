CREATE TABLE twin_state (
    entity_id  uuid PRIMARY KEY REFERENCES entities (id) ON DELETE CASCADE,
    value_num  double precision,
    value_text text,
    attributes jsonb,
    updated_at timestamptz NOT NULL,
    previous_value_num  double precision,
    previous_value_text text,
    previous_attributes jsonb,
    previous_updated_at timestamptz,
    CHECK ((value_num IS NOT NULL) <> (value_text IS NOT NULL))
);
