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
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX areas_home_idx ON areas (home_id);

-- Routing: readings and bare-id API calls resolve their home before the
-- home tables, whose row level security hides everything outside
-- app.home_id (WithHome in internal/db). The registries hold routing data
-- only; which user may access which home is the auth service's concern.

CREATE TABLE device_registry (
    gateway_id  uuid        NOT NULL,
    external_id varchar(255) NOT NULL,
    home_id     uuid        NOT NULL REFERENCES homes (id) ON DELETE CASCADE,
    device_id   uuid        NOT NULL,
    PRIMARY KEY (gateway_id, external_id)
);

CREATE TABLE node_registry (
    id      uuid        NOT NULL,
    kind    varchar(20) NOT NULL CHECK (kind IN ('area', 'device', 'entity', 'relation')),
    home_id uuid        NOT NULL REFERENCES homes (id) ON DELETE CASCADE,
    PRIMARY KEY (id, kind)
);

-- Row level security: every statement on home tables runs with app.home_id
-- set, an unset or foreign home sees no rows and cannot write. FORCE makes
-- the policy apply to the table owner too.
ALTER TABLE areas ENABLE ROW LEVEL SECURITY;
ALTER TABLE areas FORCE ROW LEVEL SECURITY;
CREATE POLICY home_isolation ON areas
    USING (home_id = current_setting('app.home_id', true)::uuid);

-- Superusers bypass row level security and PG17 refuses to demote the
-- bootstrap superuser the postgres image creates. The runtime drops into
-- this no-login role on connect (db.Init) so the home_isolation policies
-- bind to every query.
CREATE ROLE twin_app NOLOGIN;
GRANT USAGE ON SCHEMA public TO twin_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO twin_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO twin_app;

-- Let whichever user ran the migrations switch into the runtime role
DO $$
BEGIN
    EXECUTE format('GRANT twin_app TO %I', current_user);
END
$$;
