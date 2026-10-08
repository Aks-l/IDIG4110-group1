# Twin state API (current state read model)

twin-core keeps `twin_state`, the latest known value per entity, and serves it to the frontend. This document defines how state gets in, how it is stored, how the frontend reads it, and how the structural model (homes, areas, devices, entities) is managed.

## Inbound: readings from the event bus

twin-core consumes `twin.readings` (consumer group `twin-core`), so `twin_state` follows the readings stream as ingest-service publishes it ([decision 0001](../decisions/0001-kafka-event-bus.md)). The shared consumer logs handler errors and skips the record, and a brand new consumer group starts at the newest records; see [rules-engine.md, Delivery semantics](rules-engine.md#delivery-semantics) for the stream semantics.

`POST /api/v1/readings` remains available for manual pushes and tests, applying one reading through the same path:

```
POST /api/v1/readings
Content-Type: application/json
```

The body is one reading in the normalized model from [mqtt-envelope.md](mqtt-envelope.md). Responses:

- `201 Created`: reading applied to `twin_state`
- `400 Bad Request`: malformed body or contract violation, reason in the error body
- `500 Internal Server Error`: safe to retry

For each reading twin-core:

1. Routes `(gateway_id, external_entity_id)` through `device_registry` to a home and device, then resolves the `entities` row inside that home's row level security context (`app.home_id`, see [decision 0002](../decisions/0002-row-level-security.md)).
2. Upserts `twin_state` for that entity.

The upsert is last-write-wins keyed on the reading's `timestamp` (stored as `twin_state.updated_at`): an out-of-order, replayed, or retried reading never regresses newer state, so ingest can retry freely; no deduplication table is needed. Each accepted, strictly newer reading moves the superseded row into the previous-value columns: change detection, not history. A correction with the same timestamp overwrites the current value but leaves previous alone.

### Auto-provisioning of unknown entities

Readings can arrive before gateway sync has registered their devices. Instead of dropping them, twin-core provisions placeholders on first sight:

- home: the fixed `Unassigned` sentinel home (`00000000-0000-0000-0000-000000000001`)
- device: one placeholder per `external_entity_id`, keyed by `(gateway_id, external_id)`
- entity: `domain` derived from the id prefix (`sensor.x` -> `sensor`), `name` from the `friendly_name` attribute when present

The entity's first state is served immediately; gateway sync refines or replaces the placeholders later.

### Validation

Field rules run as `validate` tags on the request structs (`shared/dto`, `shared/validate`). A violation returns `400` with the reason in the error body: `gateway_id` must be a uuid, `external_entity_id` non-empty and at most 255 characters, and `timestamp` present. Exactly one of `value_num` / `value_text` must be set; this cross-field rule is checked in the service.

## Read API (frontend)

All endpoints are under `/api/v1` and return JSON. `state` is `null` for an entity that has not been read yet; `previous` is `null` until a second accepted reading arrives.

| Method | Path | Returns |
|---|---|---|
| GET | `/healthz` | liveness |
| GET | `/api/v1/homes` | homes list; admin scope: the gateway should only expose it to admins, and the frontend learns its homes from the gateway |
| GET | `/api/v1/homes/{home_id}/state` | one home: areas and devices with their entities and current state |
| GET | `/api/v1/entities/{entity_id}/state` | one entity with current state |
| GET | `/api/v1/state` | all entities with current state, filterable and paginated |

Errors use `{"code": <http status>, "message": "<reason>"}`; unknown ids return 404, malformed ids 400.

The frontend reaches all of these through the API gateway, which forwards the same paths to twin-core unchanged; the admin scoping for `GET /api/v1/homes` (and the rest of auth) lands in the gateway, not in twin-core. `/api/v1/readings` is not routed through the gateway: it stays on the internal ingest push path.

An entity with an `area_id` appears both under that area and under its device in the home state view; the frontend picks the grouping it needs.

### Filtering and pagination on `GET /api/v1/state`

The list endpoint takes optional query params; all filters combine with `AND`. Values are exact, case-sensitive matches; list params accept comma-separated values.

| Param | Type | Matches |
|---|---|---|
| `home_id` | uuid | the entity's home |
| `area_id` | uuid | the entity's area; entities without an area never match |
| `device_id` | uuid | the entity's device |
| `domain` | list, e.g. `sensor,switch` | entity domain |
| `device_class` | list | entity device class |
| `device_type` | list | the device's type |
| `controllable` | `true` / `false` | whether the entity can be commanded |
| `floor` | number | the entity's area floor; also requires an area |
| `limit` | 1 to 500, default 100 | page size |
| `offset` | >= 0, default 0 | page skip |

Unknown params, repeated params, and malformed values return `400` with the reason; typos should fail loudly, not silently drop a filter.

The response is a page envelope:

```json
{
  "items": [ { "...entity state...": "as in the entity state endpoint" } ],
  "total": 137,
  "limit": 100,
  "offset": 0
}
```

`total` counts all matching entities before paging, so the frontend can render a pager. A single home's state view is small enough to filter client-side; this endpoint is the granular one for bigger or cross-home lists, and the gateway can combine it with auth, e.g. proxy `?home_id=<the user's home>` after checking membership.

### Example

`GET /api/v1/entities/{entity_id}/state`:

```json
{
  "entity_id": "0c9f5a1e-2b7f-4fbd-9f7e-3f8f7c1d5e11",
  "external_entity_id": "sensor.living_room_temperature",
  "name": "Living room temperature",
  "domain": "sensor",
  "controllable": false,
  "device_class": "temperature",
  "unit": "°C",
  "area_id": null,
  "device_id": "6a3a41a0-0a0e-4a95-9f4f-2f5f0b6b1a77",
  "home_id": "00000000-0000-0000-0000-000000000001",
  "gateway_id": "11111111-1111-1111-1111-111111111111",
  "state": {
    "value_num": 21.5,
    "value_text": null,
    "attributes": {
      "friendly_name": "Living room temperature"
    },
    "updated_at": "2026-10-06T14:32:10.123456+00:00"
  },
  "previous": {
    "value_num": 21.1,
    "value_text": null,
    "attributes": null,
    "updated_at": "2026-10-06T14:22:10.123456+00:00"
  }
}
```

## Structure API (management)

The structural model (homes, areas, devices, entities, and the relations that connect them) is managed through create, update, and delete endpoints. The frontend reaches them through the gateway API, which proxies to twin-core. Registration never requires optional fields: a device or entity can be created before it has an area, a name, or a unit.

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/v1/homes` | create a home |
| PATCH | `/api/v1/homes/{home_id}` | update a home |
| POST | `/api/v1/areas` | create an area |
| PATCH | `/api/v1/areas/{area_id}` | update an area |
| POST | `/api/v1/devices` | create a device |
| PATCH | `/api/v1/devices/{device_id}` | update a device |
| POST | `/api/v1/entities` | create an entity |
| PATCH | `/api/v1/entities/{entity_id}` | update an entity |
| DELETE | `/api/v1/homes/{home_id}` | delete a home with everything under it |
| DELETE | `/api/v1/areas/{area_id}` | delete an area; its devices and entities are unassigned, not deleted |
| DELETE | `/api/v1/devices/{device_id}` | delete a device together with its entities and their state |
| DELETE | `/api/v1/entities/{entity_id}` | delete an entity together with its current state |
| POST | `/api/v1/relations` | create a relation (one graph edge) |
| GET | `/api/v1/homes/{home_id}/relations` | list the home's relations (the graph edges) |
| PATCH | `/api/v1/relations/{relation_id}` | update a relation's metadata |
| DELETE | `/api/v1/relations/{relation_id}` | delete a relation |

Creates return `201` with the created resource; updates return `200` with the updated resource; deletes return `204` with an empty body. Field rules for the create and update bodies run as `validate` tags on the request structs in `src/twin-core/internal/domain/requests.go`.

### Registration tolerates unknowns

Only identity fields are required on create: `name` for homes and areas, `home_id` + `gateway_id` + `external_id` for devices, `device_id` + `external_entity_id` for entities. Everything else either stays `null` or gets a default:

- `devices.area_id`, `entities.area_id`: nullable; register before assigning a room
- `devices.name`: defaults to `external_id`
- `entities.name`: defaults to `external_entity_id`
- `entities.domain`: defaults to the `external_entity_id` prefix (`sensor.x` -> `sensor`)
- `entities.controllable`: defaults to `false`
- `homes.timezone`: defaults to `UTC`

Example: register a device with only what is known, then assign it to a room once the frontend learns it:

```
POST /api/v1/devices
{ "home_id": "<uuid>", "gateway_id": "<uuid>", "external_id": "lora-abc123" }

PATCH /api/v1/devices/{device_id}
{ "name": "Living room hub", "area_id": "<uuid>" }
```

### PATCH semantics

Each field in a `PATCH` body means one of three things:

- absent: no change
- `null`: clear the value where the column is nullable; `{"area_id": null}` removes a device from its area
- set to a value: update

An empty body `{}` is a no-op that returns the current resource.

### Deletion semantics

Relation endpoints are plain uuids in `twin_relations` (an edge may connect an area, a device, or an entity), so deleting a node removes its edges explicitly rather than through foreign keys:

| Delete of | What happens |
|---|---|
| a home | everything under it (areas, devices, entities, their state, and relations) cascades |
| an area | the area and its relations are removed; devices and entities keep existing with `area_id` null (unassigned, not deleted) |
| a device | the device is deleted with its entities and their `twin_state` rows; relations touching the device or its entities are removed |
| an entity | the entity is deleted with its `twin_state` row; relations touching it are removed |

Deleting the `Unassigned` sentinel home is refused with `409`; auto-provisioning depends on it. Deletes return `204` and an empty body; unknown ids return `404`.

### Relations (graph edges)

A relation connects two nodes of one home: areas and devices for the floor-plan graph, entities when finer detail is needed. The graph view composes `GET /homes/{home_id}/state` (nodes with live state) and `GET /homes/{home_id}/relations` (edges).

```
POST /api/v1/relations
{
  "home_id": "<uuid>",
  "from_kind": "area",
  "from_id": "<uuid>",
  "to_kind": "device",
  "to_id": "<uuid>",
  "relation_type": "connects_to",
  "bidirectional": true,
  "label": "doorway",
  "properties": { "via": "north door" },
  "valid_from": "2026-10-06T00:00:00Z",
  "valid_to": null
}
```

- `from_kind` / `to_kind`: `area`, `device`, or `entity`; `relation_type`: `connects_to`, `contains`, `monitors`, `controls`, `same_physical_device`, or `depends_on`
- `home_id` is optional: it is derived from the endpoints and only cross-checked when the client sends it (`400` on mismatch)
- both endpoints must exist and belong to the same home; an unknown endpoint is `404`, a cross-home pair is `400`
- a relation cannot connect a node to itself (`400`)
- the edge `(home, from, to, relation_type)` must be new; duplicates return `409` instead of a second identical edge
- `bidirectional` defaults to `true`; direction is stored as given, and `bidirectional` tells the graph view to treat the edge as undirected
- `label`, `properties`, `valid_from`, `valid_to` are optional metadata

`PATCH /api/v1/relations/{relation_id}` changes an edge's metadata (`relation_type`, `bidirectional`, `label`, `properties`, `valid_from`, `valid_to`) with the usual absent/null/value semantics. The endpoints themselves are immutable: delete and re-create the relation to rewire it.

### Errors

- `400`: invalid request; required field missing, malformed uuid, value too long, or `null` into a required column
- `404`: unknown id, or a create referencing a home, area, or device that does not exist, including a relation endpoint
- `409`: duplicate (`devices` on `(gateway_id, external_id)`, `entities` on `(device_id, external_entity_id)`, `relations` on `(home, from, to, relation_type)`), or the attempt to delete the `Unassigned` sentinel home

Readings auto-provision placeholders under the `Unassigned` home, keyed by the same unique constraints. Explicitly registering an existing `(gateway_id, external_id)` returns `409`; a placeholder can instead be claimed with `PATCH`, for instance setting its `home_id` and `area_id`.

## Ownership

`twin_state` and the structural tables are owned by twin-core per [data-model.md](data-model.md); no other service reads or writes `twin_db` directly. History ownership is split: the full archive stays in `ingest_db.readings` with ingest-service, and a day-of-readings view should be served by the gateway proxying ingest, not duplicated here. twin-core serves the present plus the single previous reading for change detection.

### Shared module

Helpers that twin-core needed first live in `src/shared` so every service reuses one implementation:

- `shared/dto.IsValidUUID`: the canonical uuid check for every id.
- `shared/json-utils.Optional[T]`: PATCH absent/null/value semantics for partial updates.
- `shared/validate`: go-playground/validator wired for the shared types; request structs carry `validate` tags, and `validate.Struct` reports the first failure.
- `shared/pgerrors`: PostgreSQL error-code classification (unique, foreign key, not null, check, too long) for mapping onto service sentinels.
- `shared/httpclient`: JSON-over-HTTP client speaking the `{"code","message"}` error contract; `IsRetryable` separates transient failures (5xx, rate limit, transport) from permanent 4xx rejections. Intended for service-to-service REST, for example the API gateway proxying to twin-core; readings themselves flow over Kafka (see Inbound above).