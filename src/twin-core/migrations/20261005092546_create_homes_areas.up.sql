CREATE TABLE homes (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       varchar(255) NOT NULL,
    address    varchar(500),
    timezone   varchar(64) NOT NULL DEFAULT 'UTC',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE areas (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    home_id    uuid NOT NULL REFERENCES homes (id) ON DELETE CASCADE,
    name       varchar(255) NOT NULL,
    floor      int,
    area_type  varchar(50),
    geometry   jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX areas_home_idx ON areas (home_id);
