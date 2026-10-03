-- Temporary local identity, until the auth middleware takes over.
-- users holds the minimum the platform needs to function standalone.
-- Authentication and authorization (service and data access) move to the
-- middleware; these tables are then retired or reduced to a read cache.
CREATE TABLE users (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email        text NOT NULL,
    display_name text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE home_members (
    -- Soft ref, not a FK: middleware user ids can replace local user ids
    -- without a schema change.
    user_id    uuid NOT NULL,
    home_id    uuid NOT NULL,
    role       text NOT NULL CHECK (role IN ('owner', 'member', 'viewer')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, home_id)
);
