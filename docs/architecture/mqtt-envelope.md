# MQTT message envelope (read path)

This document defines the normalized message format that the Device Integration Gateway publishes to MQTT after translating gateway-native data. Ingest-service consumes this shape and writes it to the `readings` hypertable. No service other than the gateway adapter ever sees a raw gateway payload.

The envelope is gateway-agnostic by design. It is split into a **required core** that every gateway must be able to supply, and **extended data** that a gateway may or may not have. Fields outside the core are nullable, so a gateway that cannot supply them simply leaves them null. Extended data that has no dedicated column goes in `attributes`, unmodified, so nothing the gateway knows is silently dropped.

## Topic

```
twin/{gateway_id}/readings
```

- `gateway_id` is the UUID of the row in `gateways` for the source gateway.
- One topic per gateway keeps routing and access control simple.

## Envelope

```json
{
  "gateway_id": "11111111-1111-1111-1111-111111111111",
  "external_entity_id": "sensor.living_room_temperature",
  "timestamp": "2026-10-01T14:32:10.123456+00:00",
  "value_num": 21.5,
  "value_text": null,
  "device_class": "temperature",
  "unit": "°C",
  "attributes": {
    "friendly_name": "Living room temperature"
  }
}
```

### Required core

Every gateway adapter must supply these. A message missing a core field is rejected by ingest.

| Field | Type | Maps to readings column | Notes |
|---|---|---|---|
| `gateway_id` | uuid | `gateway_id` | which gateway produced this reading |
| `external_entity_id` | string | `external_entity_id` | the gateway-native identifier for the thing being read, verbatim |
| `timestamp` | RFC 3339 | `time` | true event time at the gateway, never "now" |
| `value_num` or `value_text` | number / string | `value_num` / `value_text` | exactly one must be non-null (see Value below) |

### Extended data (optional, nullable)

A gateway supplies these only if it has them. Null means "this gateway did not provide it", not "no value".

| Field | Type | Maps to readings column | Notes |
|---|---|---|---|
| `value_num` | number | `value_num` | set when the reading is numeric |
| `value_text` | string | `value_text` | set when the reading is a state, mode, or any non-numeric value |
| `device_class` | string | `device_class` | what kind of quantity/state this is (temperature, motion, ...); null if the gateway has no such concept |
| `unit` | string | `unit` | unit of measurement for numeric readings; null if none |
| `attributes` | object | `attributes` | any extra gateway-specific data, preserved verbatim |

`recorded_at` is not in the envelope. Ingest stamps it at consumption time, so the gap between `timestamp` (event time at the gateway) and `recorded_at` (ingest time) measures delivery and buffering lag.

## Value

A reading carries one value, expressed in one of two columns:

- Numeric readings go in `value_num`, with `value_text` null. If the gateway supplies a unit, it goes in `unit`.
- Non-numeric readings (states, modes, on/off, open/closed) go in `value_text`, with `value_num` null.

Exactly one of the two must be non-null. This is enforced by a check constraint on `readings`. Adapters must not invent a numeric value for a non-numeric state, and must not drop a reading just because it is non-numeric.

## Attributes

`attributes` is the escape hatch that guarantees no information loss. Any gateway-native field that does not map to a dedicated column goes here, unmodified. Consumers that need gateway-specific detail read it from `attributes`; the dedicated columns are conveniences for the common query paths, not a ceiling on what is stored.

## Worked example: Home Assistant

This is one concrete application of the rules above. Other gateways follow the same contract with their own translation.

HA emits a `state_changed` WebSocket event with `data.new_state`. The adapter translates it like this:

- `external_entity_id` = `new_state.entity_id` verbatim.
- `timestamp` = event `time_fired` (fall back to `new_state.last_changed` only if `time_fired` is absent).
- Value: HA `new_state.state` is always a string. Try to parse it as a finite float. If it parses and is not a sentinel state, put it in `value_num`. Otherwise put the raw string in `value_text`.
- Sentinel states such as `unavailable`, `unknown`, and `none` must go to `value_text`, never parsed to a number or dropped. They carry meaning (device offline, no data yet) that the twin and rules need.
- `unit` = `new_state.attributes.unit_of_measurement` when present, else null.
- `device_class` = `new_state.attributes.device_class` when present, else null.
- `attributes` = the full HA `attributes` object, unmodified.

HA event (trimmed):

```json
{
  "event_type": "state_changed",
  "time_fired": "2026-10-01T14:32:10.123456+00:00",
  "data": {
    "entity_id": "sensor.living_room_temperature",
    "new_state": {
      "entity_id": "sensor.living_room_temperature",
      "state": "21.5",
      "attributes": {
        "device_class": "temperature",
        "unit_of_measurement": "°C",
        "friendly_name": "Living room temperature"
      }
    }
  }
}
```

This produces the envelope example at the top: `state: "21.5"` becomes `value_num: 21.5`, `device_class` and `unit` are lifted to columns, and the full `attributes` object is preserved. A binary sensor (`state: "on"`, no unit) becomes `value_text: "on"`, `value_num: null`, `unit: null`.

## Guarantees

- At-least-once delivery: QoS 1 on publish and subscribe. Consumers must tolerate duplicates, since gateways can re-send states on reconnect.
- Retained messages: not used. State recovery comes from a fresh full-state sync at adapter startup, not from retained MQTT messages.
- Ordering is not guaranteed across entities. Per-entity ordering is preserved well enough for a last-write-wins twin, but consumers keying on order must use `timestamp`, not arrival order.

## Adding a gateway

A new gateway adapter produces this same envelope; the contract is fixed, only the translation differs. When adding one, document which extended fields it can and cannot supply. Anything it cannot supply is left null; anything extra it knows goes in `attributes`. Do not shrink the core and do not drop gateway data to fit the columns.
