-- Demo home seed. Populates the fixed default home
-- 00000000-0000-0000-0000-000000000001 with a coherent smart home, so the
-- stack demos out of the box once migrations have run:
--
--   * five areas with floor-plan geometry for the 3D layout
--   * devices and entities for them, including every entity id the
--     rules-engine safety rules (20261006120001_seed_safety_rules) act on:
--     binary_sensor.kitchen_smoke, lock.front_door, lock.back_door,
--     siren.alarm, valve.main_water, binary_sensor.bathroom_leak
--   * twin_state rows for the entities without a live feed, seeded with
--     past timestamps so real readings always supersede them
--   * node_registry rows for every node (bare-id API calls resolve through
--     it) and device_registry routes for the two simulated temperature
--     sensors, so their live readings update the Living Room and Kitchen
--     entities instead of auto-provisioning placeholders
--   * connects_to relations between the areas for the graph view
--
-- Idempotent: ON CONFLICT DO NOTHING everywhere, so it can also run on a
-- database where auto-provisioning already created rows. The simulated
-- sensors are only pre-registered while nothing has claimed their routes
-- yet; existing placeholders keep their route and state.

-- The fixed default home. twin-core creates it lazily as 'Unassigned' for
-- auto-provisioning (ensureDefaultHomeQuery); the seed turns it into the
-- demo home it now is. Readings for unknown ids still land here.
INSERT INTO homes (id, name, address, timezone)
VALUES ('00000000-0000-0000-0000-000000000001', 'Demo Home', 'Storgata 1, 0155 Oslo', 'Europe/Oslo')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    address = EXCLUDED.address,
    timezone = EXCLUDED.timezone,
    updated_at = now();

-- Areas, a single-storey floor plan. Geometry follows the convention the
-- frontend reads: { "position": [x, y, z], "size": [w, h, d] }.
INSERT INTO areas (id, home_id, name, floor, area_type, geometry) VALUES
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Living Room', 1, 'room', '{"position": [2, 0, 2], "size": [4, 2.6, 4]}'),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Kitchen', 1, 'room', '{"position": [7, 0, 2], "size": [3.5, 2.6, 3.5]}'),
    ('10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'Hallway', 1, 'room', '{"position": [4.5, 0, 5.75], "size": [8.5, 2.6, 2]}'),
    ('10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'Bedroom', 1, 'room', '{"position": [2, 0, 8.75], "size": [4, 2.6, 3.5]}'),
    ('10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'Bathroom', 1, 'room', '{"position": [7, 0, 8.75], "size": [3.5, 2.6, 2.5]}')
ON CONFLICT DO NOTHING;

INSERT INTO node_registry (id, kind, home_id) VALUES
    ('10000000-0000-0000-0000-000000000001', 'area', '00000000-0000-0000-0000-000000000001'),
    ('10000000-0000-0000-0000-000000000002', 'area', '00000000-0000-0000-0000-000000000001'),
    ('10000000-0000-0000-0000-000000000003', 'area', '00000000-0000-0000-0000-000000000001'),
    ('10000000-0000-0000-0000-000000000004', 'area', '00000000-0000-0000-0000-000000000001'),
    ('10000000-0000-0000-0000-000000000005', 'area', '00000000-0000-0000-0000-000000000001')
ON CONFLICT DO NOTHING;

-- Devices, all behind the demo gateway 00000000-0000-0000-0000-000000000001
-- (the id the device-simulator publishes under). The two simulated
-- temperature sensors are registered in the block further down instead.
INSERT INTO devices (id, home_id, area_id, gateway_id, external_id, name, manufacturer, model, sw_version, device_type, status, last_seen_at) VALUES
    ('20000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'multi.living_room', 'Living Room Climate Sensor', 'Aqara', 'TH Sensor T1', '1.0.0', 'sensor', 'online', now()),
    ('20000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'light.living_room', 'Living Room Light', 'Philips Hue', 'White A60', '1.0.0', 'light', 'online', now()),
    ('20000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'binary_sensor.kitchen_smoke', 'Kitchen Smoke Detector', 'FireAngel', 'SW1-PF', '1.0.0', 'sensor', 'online', now()),
    ('20000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'light.kitchen', 'Kitchen Light', 'Philips Hue', 'White A60', '1.0.0', 'light', 'online', now()),
    ('20000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'valve.main_water', 'Main Water Valve', 'Netro', 'Sprite', '1.0.0', 'valve', 'online', now()),
    ('20000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'binary_sensor.bathroom_leak', 'Bathroom Leak Sensor', 'Aqara', 'Leak Sensor', '1.0.0', 'sensor', 'online', now()),
    ('20000000-0000-0000-0000-000000000009', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'sensor.bathroom_humidity', 'Bathroom Climate Sensor', 'Aqara', 'TH Sensor T1', '1.0.0', 'sensor', 'online', now()),
    ('20000000-0000-0000-0000-00000000000a', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'sensor.bedroom_climate', 'Bedroom Climate Sensor', 'Aqara', 'TH Sensor T1', '1.0.0', 'sensor', 'online', now()),
    ('20000000-0000-0000-0000-00000000000b', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'light.bedroom', 'Bedroom Light', 'Philips Hue', 'White A60', '1.0.0', 'light', 'online', now()),
    ('20000000-0000-0000-0000-00000000000c', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'fan.bedroom', 'Bedroom Fan', 'Dyson', 'AM07', '1.0.0', 'fan', 'online', now()),
    ('20000000-0000-0000-0000-00000000000d', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'lock.front_door', 'Front Door Lock', 'Yale', 'Doorman V2N', '1.0.0', 'lock', 'online', now()),
    ('20000000-0000-0000-0000-00000000000e', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'lock.back_door', 'Back Door Lock', 'Yale', 'Doorman V2N', '1.0.0', 'lock', 'online', now()),
    ('20000000-0000-0000-0000-00000000000f', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'siren.alarm', 'Alarm Siren', 'Aeotec', 'Siren 6', '1.0.0', 'siren', 'online', now()),
    ('20000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'binary_sensor.hallway_motion', 'Hallway Motion Sensor', 'Aqara', 'Motion P1', '1.0.0', 'sensor', 'online', now())
ON CONFLICT DO NOTHING;
INSERT INTO node_registry (id, kind, home_id) VALUES
    ('20000000-0000-0000-0000-000000000002', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000003', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000005', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000006', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000007', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000008', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000009', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-00000000000a', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-00000000000b', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-00000000000c', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-00000000000d', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-00000000000e', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-00000000000f', 'device', '00000000-0000-0000-0000-000000000001'),
    ('20000000-0000-0000-0000-000000000010', 'device', '00000000-0000-0000-0000-000000000001')
ON CONFLICT DO NOTHING;
-- The two simulated temperature sensors: device-simulator publishes
-- external entity ids 00000000-0000-0000-0000-000000000001/-0002 under
-- gateway 00000000-0000-0000-0000-000000000001. Registering them here (plus
-- device_registry routes) directs the live readings into the demo home's
-- Living Room and Kitchen temperature entities instead of letting
-- auto-provisioning create placeholder devices. Only while nothing has
-- claimed their routes: a database that already provisioned placeholders
-- keeps them and their state.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM device_registry
        WHERE gateway_id = '00000000-0000-0000-0000-000000000001'
          AND external_id IN ('00000000-0000-0000-0000-000000000001',
                              '00000000-0000-0000-0000-000000000002')
    ) THEN
        INSERT INTO devices (id, home_id, area_id, gateway_id, external_id, name, manufacturer, model, sw_version, device_type, status, last_seen_at) VALUES
            ('20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Living Room Temperature Sensor', 'Simulated', 'TempProbe', '1.0.0', 'sensor', 'online', now()),
            ('20000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', 'Kitchen Temperature Sensor', 'Simulated', 'TempProbe', '1.0.0', 'sensor', 'online', now());
        INSERT INTO device_registry (gateway_id, external_id, home_id, device_id) VALUES
            ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001'),
            ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000004');
        INSERT INTO entities (id, device_id, area_id, home_id, external_entity_id, name, domain, device_class, unit, controllable) VALUES
            ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Living Room Temperature', 'sensor', 'temperature', '°C', false),
            ('30000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', 'Kitchen Temperature', 'sensor', 'temperature', '°C', false);
        INSERT INTO node_registry (id, kind, home_id) VALUES
            ('20000000-0000-0000-0000-000000000001', 'device', '00000000-0000-0000-0000-000000000001'),
            ('20000000-0000-0000-0000-000000000004', 'device', '00000000-0000-0000-0000-000000000001'),
            ('30000000-0000-0000-0000-000000000001', 'entity', '00000000-0000-0000-0000-000000000001'),
            ('30000000-0000-0000-0000-000000000005', 'entity', '00000000-0000-0000-0000-000000000001');
    END IF;
END $$;
-- Entities on the seeded devices. Names follow Home Assistant ids where
-- the safety rules act on them, so rule targets resolve to real entities.
INSERT INTO entities (id, device_id, area_id, home_id, external_entity_id, name, domain, device_class, unit, controllable) VALUES
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'humidity.living_room', 'Living Room Humidity', 'sensor', 'humidity', '%', false),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'co2.living_room', 'Living Room CO2', 'sensor', 'co2', 'ppm', false),
    ('30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'light.living_room', 'Living Room Light', 'light', NULL, NULL, true),
    ('30000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'binary_sensor.kitchen_smoke', 'Kitchen Smoke', 'binary_sensor', 'smoke', NULL, false),
    ('30000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'light.kitchen', 'Kitchen Light', 'light', NULL, NULL, true),
    ('30000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'valve.main_water', 'Main Water Valve', 'valve', NULL, NULL, true),
    ('30000000-0000-0000-0000-000000000009', '20000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'binary_sensor.bathroom_leak', 'Bathroom Leak', 'binary_sensor', 'moisture', NULL, false),
    ('30000000-0000-0000-0000-00000000000a', '20000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'sensor.bathroom_humidity', 'Bathroom Humidity', 'sensor', 'humidity', '%', false),
    ('30000000-0000-0000-0000-00000000000b', '20000000-0000-0000-0000-00000000000a', '10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'temperature.bedroom', 'Bedroom Temperature', 'sensor', 'temperature', '°C', false),
    ('30000000-0000-0000-0000-00000000000c', '20000000-0000-0000-0000-00000000000a', '10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'humidity.bedroom', 'Bedroom Humidity', 'sensor', 'humidity', '%', false),
    ('30000000-0000-0000-0000-00000000000d', '20000000-0000-0000-0000-00000000000b', '10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'light.bedroom', 'Bedroom Light', 'light', NULL, NULL, true),
    ('30000000-0000-0000-0000-00000000000e', '20000000-0000-0000-0000-00000000000c', '10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'fan.bedroom', 'Bedroom Fan', 'fan', NULL, NULL, true),
    ('30000000-0000-0000-0000-00000000000f', '20000000-0000-0000-0000-00000000000d', '10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'lock.front_door', 'Front Door Lock', 'lock', NULL, NULL, true),
    ('30000000-0000-0000-0000-000000000010', '20000000-0000-0000-0000-00000000000e', '10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'lock.back_door', 'Back Door Lock', 'lock', NULL, NULL, true),
    ('30000000-0000-0000-0000-000000000011', '20000000-0000-0000-0000-00000000000f', '10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'siren.alarm', 'Alarm Siren', 'siren', NULL, NULL, true),
    ('30000000-0000-0000-0000-000000000012', '20000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'binary_sensor.hallway_motion', 'Hallway Motion', 'binary_sensor', 'motion', NULL, false)
ON CONFLICT DO NOTHING;
INSERT INTO node_registry (id, kind, home_id) VALUES
    ('30000000-0000-0000-0000-000000000002', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000003', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000004', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000006', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000007', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000008', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000009', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-00000000000a', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-00000000000b', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-00000000000c', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-00000000000d', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-00000000000e', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-00000000000f', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000010', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000011', 'entity', '00000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000012', 'entity', '00000000-0000-0000-0000-000000000001')
ON CONFLICT DO NOTHING;
-- State for the entities without a live feed. Timestamps are in the past
-- on purpose: the last-write-wins upsert lets any real reading supersede
-- them. Previous values give change detection something to show.
INSERT INTO twin_state (entity_id, home_id, value_num, value_text, attributes, updated_at, previous_value_num, previous_value_text, previous_updated_at) VALUES
    ('30000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 45, NULL, '{"friendly_name": "Living Room Humidity"}', '2026-10-01 09:00:00+00', 44, NULL, '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 620, NULL, '{"friendly_name": "Living Room CO2"}', '2026-10-01 09:00:00+00', 645, NULL, '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', NULL, 'on', '{"friendly_name": "Living Room Light"}', '2026-10-01 09:00:00+00', NULL, 'off', '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', 0, NULL, '{"friendly_name": "Kitchen Smoke"}', '2026-10-01 09:00:00+00', 0, NULL, '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', NULL, 'off', '{"friendly_name": "Kitchen Light"}', '2026-10-01 09:00:00+00', NULL, 'on', '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000001', NULL, 'open', '{"friendly_name": "Main Water Valve"}', '2026-10-01 09:00:00+00', NULL, 'closed', '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000009', '00000000-0000-0000-0000-000000000001', 0, NULL, '{"friendly_name": "Bathroom Leak"}', '2026-10-01 09:00:00+00', 0, NULL, '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-00000000000a', '00000000-0000-0000-0000-000000000001', 55, NULL, '{"friendly_name": "Bathroom Humidity"}', '2026-10-01 09:00:00+00', 56, NULL, '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-00000000000b', '00000000-0000-0000-0000-000000000001', 20.0, NULL, '{"friendly_name": "Bedroom Temperature"}', '2026-10-01 09:00:00+00', 20.2, NULL, '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-00000000000c', '00000000-0000-0000-0000-000000000001', 41, NULL, '{"friendly_name": "Bedroom Humidity"}', '2026-10-01 09:00:00+00', 42, NULL, '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-00000000000d', '00000000-0000-0000-0000-000000000001', NULL, 'off', '{"friendly_name": "Bedroom Light"}', '2026-10-01 09:00:00+00', NULL, 'on', '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-00000000000e', '00000000-0000-0000-0000-000000000001', NULL, 'off', '{"friendly_name": "Bedroom Fan"}', '2026-10-01 09:00:00+00', NULL, NULL, NULL),
    ('30000000-0000-0000-0000-00000000000f', '00000000-0000-0000-0000-000000000001', NULL, 'locked', '{"friendly_name": "Front Door Lock"}', '2026-10-01 09:00:00+00', NULL, 'locked', '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', NULL, 'locked', '{"friendly_name": "Back Door Lock"}', '2026-10-01 09:00:00+00', NULL, 'locked', '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001', NULL, 'off', '{"friendly_name": "Alarm Siren"}', '2026-10-01 09:00:00+00', NULL, 'off', '2026-10-01 08:45:00+00'),
    ('30000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000001', 0, NULL, '{"friendly_name": "Hallway Motion"}', '2026-10-01 09:00:00+00', 1, NULL, '2026-10-01 08:45:00+00')
ON CONFLICT DO NOTHING;
-- The rooms connect through the hallway; the graph view composes these
-- edges with the home state.
INSERT INTO twin_relations (id, home_id, from_kind, from_id, to_kind, to_id, relation_type, bidirectional, label) VALUES
    ('40000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'area', '10000000-0000-0000-0000-000000000001', 'area', '10000000-0000-0000-0000-000000000002', 'connects_to', true, 'open plan'),
    ('40000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'area', '10000000-0000-0000-0000-000000000003', 'area', '10000000-0000-0000-0000-000000000001', 'connects_to', true, 'archway'),
    ('40000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 'area', '10000000-0000-0000-0000-000000000003', 'area', '10000000-0000-0000-0000-000000000002', 'connects_to', true, 'doorway'),
    ('40000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', 'area', '10000000-0000-0000-0000-000000000003', 'area', '10000000-0000-0000-0000-000000000004', 'connects_to', true, 'doorway'),
    ('40000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', 'area', '10000000-0000-0000-0000-000000000003', 'area', '10000000-0000-0000-0000-000000000005', 'connects_to', true, 'doorway')
ON CONFLICT DO NOTHING;

INSERT INTO node_registry (id, kind, home_id) VALUES
    ('40000000-0000-0000-0000-000000000001', 'relation', '00000000-0000-0000-0000-000000000001'),
    ('40000000-0000-0000-0000-000000000002', 'relation', '00000000-0000-0000-0000-000000000001'),
    ('40000000-0000-0000-0000-000000000003', 'relation', '00000000-0000-0000-0000-000000000001'),
    ('40000000-0000-0000-0000-000000000004', 'relation', '00000000-0000-0000-0000-000000000001'),
    ('40000000-0000-0000-0000-000000000005', 'relation', '00000000-0000-0000-0000-000000000001')
ON CONFLICT DO NOTHING;
