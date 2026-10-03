CREATE TABLE homes (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    address    text,
    timezone   text NOT NULL DEFAULT 'UTC',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE areas (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    home_id    uuid NOT NULL REFERENCES homes (id) ON DELETE CASCADE,
    name       text NOT NULL,
    floor      int,
    area_type  text,
    geometry   jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX areas_home_idx ON areas (home_id);
