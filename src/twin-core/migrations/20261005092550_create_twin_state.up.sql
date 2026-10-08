CREATE TABLE twin_state (
    entity_id  uuid PRIMARY KEY REFERENCES entities (id) ON DELETE CASCADE,
    home_id    uuid NOT NULL REFERENCES homes (id) ON DELETE CASCADE,
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

CREATE INDEX twin_state_home_idx ON twin_state (home_id);

ALTER TABLE twin_state ENABLE ROW LEVEL SECURITY;
ALTER TABLE twin_state FORCE ROW LEVEL SECURITY;
CREATE POLICY home_isolation ON twin_state
    USING (home_id = current_setting('app.home_id', true)::uuid);
