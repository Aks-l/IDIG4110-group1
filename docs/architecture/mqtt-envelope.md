# Normalized reading model (read path)

This document defines the internal normalized model for a single reading: a measurement or state change observed at a device, expressed in the system's own terms. Every reading in the platform takes this shape. It is what flows on MQTT, what ingest-service writes to the `readings` hypertable, and what downstream services consume. Gateway-specific payloads never leave the Device Integration Gateway; adapters translate native data into this model at the boundary.

The model is gateway-agnostic by design. It has a **core** that defines what a reading fundamentally is, and **extended data** that enriches a reading when available. Extended fields are nullable: a reading is still valid when they are absent. Extended data with no dedicated field is carried in `attributes`, unmodified, so nothing known about the reading is silently dropped.

## Topic

```
twin/{gateway_id}/readings
```

- `gateway_id` is the UUID of the row in `gateways` for the source gateway.
- One topic per gateway keeps routing and access control simple.

## Model

```json
{
  "gateway_id": "11111111-1111-1111-1111-111111111111",
  "external_entity_id": "sensor.living_room_temperature",
  "timestamp": "2026-10-01T14:32:10.123456+00:00",
  "event_id": "326ef27d19415c60c492fe330945f954",
  "value_num": 21.5,
  "value_text": null,
  "device_class": "temperature",
  "unit": "°C",
  "attributes": {
    "friendly_name": "Living room temperature"
  }
}
```

### Core

The core is the definition of a reading. A message missing a core field is not a reading and is rejected by ingest.

| Field | Type | Maps to readings column | Meaning |
|---|---|---|---|
| `gateway_id` | uuid | `gateway_id` | which source produced this reading |
| `external_entity_id` | string | `external_entity_id` | identifier for the thing being read, verbatim from the source |
| `timestamp` | RFC 3339 | `time` | when the event actually happened, never "now" |
| `value_num` or `value_text` | number / string | `value_num` / `value_text` | the observed value; exactly one is non-null (see Value below) |

### Extended data (optional, nullable)

These enrich a reading when the information exists. Null means "not known for this reading", not "no value".

| Field | Type | Maps to readings column | Meaning |
|---|---|---|---|
| `event_id` | uuid | `event_id` | the source's unique id for this event, when it has one; used for deduplication and for correlating a state change back to a command |
| `value_num` | number | `value_num` | set when the reading is numeric |
| `value_text` | string | `value_text` | set when the reading is a state, mode, or any non-numeric value |
| `device_class` | string | `device_class` | what kind of quantity/state this is (temperature, motion, ...) |
| `unit` | string | `unit` | unit of measurement for numeric readings |
| `attributes` | object | `attributes` | any extra source-specific detail, preserved verbatim |

`recorded_at` is not part of the model. Ingest stamps it at consumption time, so the gap between `timestamp` (event time at the source) and `recorded_at` (ingest time) measures delivery and buffering lag.

## Value

A reading carries one value, expressed in one of two fields:

- Numeric readings go in `value_num`, with `value_text` null. If a unit is known, it goes in `unit`.
- Non-numeric readings (states, modes, on/off, open/closed) go in `value_text`, with `value_num` null.

Exactly one of the two is non-null, enforced by a check constraint on `readings`. A non-numeric state is never forced into a number, and a reading is never dropped just because it is non-numeric.

## Attributes

`attributes` is what guarantees no information loss. Any source detail that has no dedicated field goes here, unmodified. Consumers that need source-specific detail read it from `attributes`; the dedicated fields are conveniences for the common query paths, not a ceiling on what is stored.

## Worked example: Home Assistant

This is one concrete translation into the model. Each source has its own adapter; the model does not change.

HA emits a `state_changed` WebSocket event. The adapter translates it like this:

- `external_entity_id` = `data.entity_id`, the subject of the event. Not `new_state.entity_id`: that is a snapshot field HA repeats inside the state object. The two are equal in the common case, but the event subject is the authoritative identifier, and they can diverge (for instance on entity renames, where `old_state` and `new_state` carry different ids).
- `timestamp` = event `time_fired` (fall back to `new_state.last_changed` only if `time_fired` is absent).
- Value: HA `new_state.state` is always a string. Try to parse it as a finite float. If it parses and is not a sentinel state, put it in `value_num`. Otherwise put the raw string in `value_text`.
- Sentinel states such as `unavailable`, `unknown`, and `none` must go to `value_text`, never parsed to a number or dropped. They carry meaning (device offline, no data yet) that the twin and rules need.
- `unit` = `new_state.attributes.unit_of_measurement` when present, else null.
- `device_class` = `new_state.attributes.device_class` when present, else null.
- `event_id` = the event's `context.id`. HA puts `context` beside `attributes`, not inside it, so it needs an explicit mapping. This is the deduplication key when HA re-sends states on reconnect, and it lets us correlate a state change back to a command we issued.
- `attributes` = the full HA `attributes` object, unmodified.

HA event (trimmed):

```json
{
  "event_type": "state_changed",
  "time_fired": "2026-10-01T14:32:10.123456+00:00",
  "context": {
    "id": "326ef27d19415c60c492fe330945f954",
    "parent_id": null,
    "user_id": null
  },
  "data": {
    "entity_id": "sensor.living_room_temperature",
    "new_state": {
      "entity_id": "sensor.living_room_temperature",
      "state": "21.5",
      "attributes": {
        "device_class": "temperature",
        "unit_of_measurement": "°C",
        "friendly_name": "Living room temperature"
      },
      "context": {
        "id": "326ef27d19415c60c492fe330945f954",
        "parent_id": null,
        "user_id": null
      }
    }
  }
}
```

This produces the envelope example at the top: `data.entity_id` becomes `external_entity_id`, `time_fired` becomes `timestamp`, `context.id` becomes `event_id`, `state: "21.5"` becomes `value_num: 21.5`, and the full `attributes` object is preserved. A binary sensor (`state: "on"`, no unit) becomes `value_text: "on"`, `value_num: null`, `unit: null`.

Current state: the device-simulator publishes exactly this event shape on MQTT in development, and ingest-service parses it; the translation into the normalized model above is not implemented yet.

## Guarantees

- At-least-once delivery: QoS 1 on publish and subscribe. Consumers must tolerate duplicates, since gateways can re-send states on reconnect.
- Retained messages: not used. State recovery comes from a fresh full-state sync at adapter startup, not from retained MQTT messages.
- Ordering is not guaranteed across entities. Per-entity ordering is preserved well enough for a last-write-wins twin, but consumers keying on order must use `timestamp`, not arrival order.

## Supporting a new source

The normalized model is fixed; only the translation into it differs per source. An adapter maps native data onto the core fields, fills extended fields where the information exists, and puts everything else in `attributes`. When a source has no concept that matches an extended field, that field is left null. The core is never weakened to fit a source, and source data is never dropped to fit the fields.
