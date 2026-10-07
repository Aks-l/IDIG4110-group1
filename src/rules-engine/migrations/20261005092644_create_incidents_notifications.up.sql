CREATE TABLE incidents (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id         uuid NOT NULL REFERENCES rules (id),
    home_id         uuid NOT NULL,
    severity        varchar(20) NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
    status          varchar(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'acknowledged', 'resolved')),
    message         text NOT NULL,
    context         jsonb,
    triggered_at    timestamptz NOT NULL DEFAULT now(),
    acknowledged_at timestamptz,
    resolved_at     timestamptz
);

CREATE INDEX incidents_home_status_idx ON incidents (home_id, status);

CREATE TABLE notifications (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    channel     varchar(20) NOT NULL CHECK (channel IN ('ui', 'email', 'push')),
    status      varchar(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    sent_at     timestamptz
);

CREATE INDEX notifications_incident_idx ON notifications (incident_id);
