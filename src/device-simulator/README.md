# Device simulator

The simulator creates one configured house and gateway. The house definition is
loaded from [`config/house.json`](config/house.json); adding another instance
of an existing device type only requires adding another device object there.
Each device can also define a `simulation` block in that file. This controls
how its readings change without requiring code changes.

## Architecture

- `internal/config` loads and validates the MQTT YAML and house JSON.
- `internal/simulator/device.go` contains the reusable capability-based device
  implementation.
- `internal/simulator/gateway.go` owns the configured devices and routes
  simulation ticks and commands.
- `internal/simulator/storage.go` defines the storage boundary and provides an
  in-memory implementation for now.
- `internal/simulator/simulator.go` adapts gateway readings to the existing
  MQTT raw-message envelope.

Each device advertises capabilities such as `turn_on`, `set_temperature`, or
`read_motion`. Commands unsupported by a device are rejected. Readings are
published as `sensors/{device_id}/{property}/state`, and the state event's
entity ID is `{device_id}.{property}`.

The simulator listens for commands on
`twin/{gateway_id}/commands`, so the existing rules-engine command path can
control configured devices. The gateway keeps device state in memory and
persists readings through the `Storage` interface. A database-backed storage
implementation can replace `MemoryStorage` without changing device logic.

## Adding devices

For an existing type, add a device to a room in `config/house.json`:

```json
{
  "id": "bedroom_thermostat",
  "type": "thermostat",
  "initial_state": {
    "current_temperature": 20.0,
    "target_temperature": 21.0
  }
}
```

The `initial_state` values are the starting values. The optional `simulation`
values control later readings:

```json
{
  "id": "bedroom_temperature",
  "type": "temperature_sensor",
  "initial_state": { "temperature": 20.0 },
  "simulation": {
    "enabled": true,
    "temperature": {
      "min": 18.0,
      "max": 24.0,
      "max_change": 0.1
    }
  }
}
```

Useful settings are:

- `temperature.min`, `temperature.max`, `temperature.max_change`
- `humidity.min`, `humidity.max`, `humidity.max_change`
- `motion_probability` and `open_probability`, from `0.0` to `1.0`
- `power.min` and `power.max` for smart plugs
- `energy_per_tick` for energy sensors
- `temperature_step` for air conditioners, ovens, and thermostats

Set a probability to `1.0` to make a motion or door sensor always active, or
`0.0` to keep it inactive. This is useful for testing rule effects. Set
`enabled` to `false` to stop changing the device automatically; its current
state is still published once and the device remains available for commands.

For a new type, add its capabilities and simulation/command behavior in
`internal/simulator/device.go`, then add a configuration entry. The gateway
and MQTT layer do not need to change.

The IDs already include the house and gateway concepts in configuration. The
gateway is currently created once by the process, but `NewGateway` and
`Storage` do not use global state, so a future manager can create one gateway
per house or simulation.
