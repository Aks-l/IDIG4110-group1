-- Built-in safety rules for FR-WA-02 (smoke) and FR-WA-03 (water leak) on
-- the demo home. Entity ids follow Home Assistant naming; adjust them to the
-- real devices through the API. The format is described in
-- docs/architecture/rules-engine.md.

INSERT INTO rules (id, home_id, name, description, priority, trigger_config, action_config)
VALUES
(
    'a0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'Smoke detected',
    'FR-WA-02: on smoke, unlock the emergency exits, sound the alarm and notify for emergency response.',
    100,
    '{"trigger": {"entity": "binary_sensor.kitchen_smoke", "operator": "eq", "value": true},
      "conditions": [],
      "cooldown_seconds": 300}',
    '{"actions": [
        {"type": "command", "entity": "lock.front_door", "command": "unlock"},
        {"type": "command", "entity": "lock.back_door", "command": "unlock"},
        {"type": "command", "entity": "siren.alarm", "command": "turn_on"},
        {"type": "incident", "severity": "critical", "message": "Smoke detected ({{entity}}). Exits unlocked and alarm on."},
        {"type": "notify", "channels": ["ui", "push", "email"]}
    ]}'
),
(
    'a0000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'Water leak detected',
    'FR-WA-03: on a water leak, shut off the main water valve.',
    90,
    '{"trigger": {"entity": "binary_sensor.bathroom_leak", "operator": "eq", "value": true},
      "conditions": [],
      "cooldown_seconds": 300}',
    '{"actions": [
        {"type": "command", "entity": "valve.main_water", "command": "close"},
        {"type": "incident", "severity": "critical", "message": "Water leak detected ({{entity}}). Main water valve closed."},
        {"type": "notify", "channels": ["ui", "push"]}
    ]}'
)
ON CONFLICT (id) DO NOTHING;
