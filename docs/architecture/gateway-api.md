# Gateway command API (write path)

This document defines the internal command model for changing device state: how any service (rules-engine, UI via the API gateway) asks for a device to do something without knowing gateway specifics. It is the write-side counterpart to the normalized read model in [mqtt-envelope.md](mqtt-envelope.md). The `commands` table in ingest_db audits every command against this contract.

Gateway-native protocols never leave the Device Integration Gateway. Callers speak this contract; the adapter translates to the source's native API (for Home Assistant, a `call_service` WebSocket message).

## Command model

A command is a request to change one entity on one gateway:

```json
{
  "id": "55555555-5555-5555-5555-555555555555",
  "gateway_id": "11111111-1111-1111-1111-111111111111",
  "external_entity_id": "light.living_room",
  "command": "turn_on",
  "parameters": {
    "brightness": 80
  },
  "issued_at": "2026-10-03T19:40:00.000000+00:00",
  "expires_at": "2026-10-03T19:40:30.000000+00:00"
}
```

| Field | Type | Required | Meaning |
|---|---|---|---|
| `id` | uuid | yes | the command's id from the `commands` table; the dedup key |
| `gateway_id` | uuid | yes | which gateway the target entity lives on |
| `external_entity_id` | string | yes | the entity to act on, same identifier the read path uses |
| `command` | string | yes | normalized command name (see vocabulary below) |
| `parameters` | object | no | command arguments |
| `issued_at` | RFC 3339 | yes | when the command was issued |
| `expires_at` | RFC 3339 | no | TTL; after this the command must not execute |

`id` and `issued_at` ride in the message, not just the database. QoS 1 redelivers messages on reconnect, so the adapter uses `id` to detect a command it has already seen and `issued_at` with `expires_at` to refuse a command that sat queued past its TTL. Without them in the message, redelivery would re-execute commands and stale commands would fire late.

`gateway_id` is part of the command, not just the transport. An entity id is only unique within one gateway; two gateways can expose the same `external_entity_id`, so the pair `(gateway_id, external_entity_id)` is what identifies the target. The MQTT message carries both, making it self-contained and routable without depending on where it arrived.

Commands are recorded in `commands` with the same `id`, a `correlation_id` for matching the result, `issued_by` naming the calling service, and the same `expires_at` TTL.

## Command vocabulary

The vocabulary is open-ended: `command` is any string, and new commands can be introduced without changing this contract or the schema. What makes that safe is that callers never talk to devices directly, and translation is per entity, not global.

Resolution works like this:

1. The adapter looks up the command in the target entity's `entities.command_map` (for instance `{"turn_on": {"domain": "light", "service": "turn_on"}}`), populated at sync time.
2. If the entity has a mapping, the adapter translates and executes.
3. If not, the command is marked `failed` with a reason like `unsupported_command` in `result`. It is never guessed at or dropped silently.

Common commands and their Home Assistant mapping, as a starting point rather than a closed list:

| Command | Parameters | HA service call |
|---|---|---|
| `turn_on` | optional brightness | `light.turn_on` / `switch.turn_on` |
| `turn_off` | none | `light.turn_off` / `switch.turn_off` |
| `toggle` | none | `*.toggle` |
| `set_temperature` | `temperature` | `climate.set_temperature` |
| `set_brightness` | `brightness` (0-100) | `light.turn_on` with brightness |
| `lock` / `unlock` | none | `lock.lock` / `lock.unlock` |
| `open` / `close` | none | `cover.open_cover` / `cover.close_cover` |

Adding a command means adding it to relevant entities' `command_map` and, if it is meant to be portable, documenting it here. Nothing requires a migration.

## Flow

1. Caller issues a command via `POST /gateways/{gateway_id}/commands` on the Device Integration Gateway with the body above (minus `id` and `issued_at`, which the gateway generates).
2. The gateway records it in `commands` with status `pending`, generates `id` and `correlation_id`, stamps `issued_at`, and publishes the full message to `twin/{gateway_id}/commands`.
3. The adapter picks it up, checks it has not seen this `id` before, checks `now() < expires_at`, translates per `command_map`, and executes against the source (HA: `call_service`).
4. The adapter publishes the outcome to `twin/{gateway_id}/commands/{correlation_id}/result` and updates the row: `sent`, then `acknowledged` on success or `failed` with the error in `result`.
5. If no response arrives before `expires_at`, status becomes `timeout`. Expired commands never fire.

## Lifecycle states

`pending` -> `sent` -> `acknowledged` | `failed` | `timeout`. Enforced by the `commands.status` check constraint.

## HA notes

- HA `call_service` returns no device response; success means HA accepted the call. Confirmation that the device actually changed comes as a `state_changed` event on the read path, correlatable via the event's `context.id` (stored as `readings.event_id`).
- Adapter startup performs a full `get_states` sync; command state is not recovered from retained MQTT messages.

## Guarantees

- QoS 1 on command and result topics; redelivery is expected and neutralized by the `id` dedup check.
- Every command is auditable end to end: who issued it, when, what the adapter did, and what the outcome was.
- A command that cannot be delivered before its TTL is marked `timeout`, never executed late. Adapters enforce this before executing, using `issued_at`/`expires_at` from the message.

## Supporting a new source

The command envelope and lifecycle are fixed; the set of commands is not. A new adapter expresses what its devices can do through each entity's `command_map`, and reports outcomes on the result topic. Commands a device cannot execute come back as `failed` with the reason in `result`.
