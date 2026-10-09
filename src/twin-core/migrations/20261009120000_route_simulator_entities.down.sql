DELETE FROM device_registry
WHERE gateway_id = '00000000-0000-0000-0000-000000000001'
  AND external_id IN (
    'humidity.living_room',
    'co2.living_room',
    'light.living_room',
    'binary_sensor.kitchen_smoke',
    'light.kitchen',
    'valve.main_water',
    'binary_sensor.bathroom_leak',
    'sensor.bathroom_humidity',
    'temperature.bedroom',
    'humidity.bedroom',
    'light.bedroom',
    'fan.bedroom',
    'lock.front_door',
    'lock.back_door',
    'siren.alarm',
    'binary_sensor.hallway_motion'
  );