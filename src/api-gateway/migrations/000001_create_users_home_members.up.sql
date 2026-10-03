CREATE TABLE users (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email        text NOT NULL,
    display_name text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE home_members (
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    home_id    uuid NOT NULL,
    role       text NOT NULL CHECK (role IN ('owner', 'member', 'viewer')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, home_id)
);
