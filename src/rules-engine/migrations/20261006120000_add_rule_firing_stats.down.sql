ALTER TABLE rules
    DROP COLUMN IF EXISTS fire_count,
    DROP COLUMN IF EXISTS last_fired_at;
