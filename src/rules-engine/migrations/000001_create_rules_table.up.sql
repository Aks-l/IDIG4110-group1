CREATE TABLE rules (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    home_id        uuid NOT NULL,
    name           text NOT NULL,
    description    text,
    enabled        boolean NOT NULL DEFAULT true,
    priority       int NOT NULL DEFAULT 0,
    trigger_config jsonb NOT NULL,
    action_config  jsonb NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX rules_home_idx ON rules (home_id);
