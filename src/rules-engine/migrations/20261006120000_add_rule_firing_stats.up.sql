-- How often and when each rule last fired, for the automations page.

ALTER TABLE rules
    ADD COLUMN last_fired_at timestamptz,
    ADD COLUMN fire_count    int NOT NULL DEFAULT 0;
