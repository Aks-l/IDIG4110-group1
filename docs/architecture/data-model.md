# Data model and schema ownership

The platform uses a database-per-service layout. Each microservice owns one Postgres database and is the only service that reads or writes it directly. Other services reference its records by UUID and stay in sync through events on the Kafka bus (see [decision 0001](../decisions/0001-kafka-event-bus.md)).

This document is the contract for that model: which database exists, which service owns it, what each table is for, and the rules that keep the databases decoupled. Future branches build against this; if a change would break a rule here, change the rule deliberately in this document first, not silently in a migration.

## Databases

| Database | Owning service | Purpose |
|---|---|---|
| `ingest_db` | ingest-service | Gateway registry, normalized telemetry, raw payload audit, command audit |
| `twin_db` | twin-core | Structural model of each home: areas, devices, entities, relationships, current state |
| `rules_db` | rules-engine | Automation rules, raised incidents, notification tracking |
| `identity_db` | api-gateway (interim) | Users and home membership |

Each database runs in its own container. `ingest_db` uses the TimescaleDB image for hypertables; the others use the same image for consistency.

## Tables

### ingest_db (exists)

| Table | Purpose |
|---|---|
| `gateways` | Registry of connected gateways: type, status, last-seen, connection config |
| `readings` | Hypertable of normalized sensor readings. The core telemetry store |
| `raw_messages` | Hypertable of raw MQTT payloads before normalization, 7-day retention |
| `commands` | Audit of commands sent to devices, with lifecycle status and correlation id |

### twin_db (planned)

| Table | Purpose |
|---|---|
| `homes` | One row per home |
| `areas` | Rooms/zones within a home |
| `twin_relations` | Generic typed-edge graph: room adjacency, device-to-room, cross-gateway device links |
| `devices` | Normalized device model; each device maps to a gateway-native device |
| `entities` | Individual sensors/controls on a device; register to rooms directly |
| `twin_state` | Latest known value per entity; the fast read path for current state |

### rules_db (planned)

| Table | Purpose |
|---|---|
| `rules` | Automation/alerting rules with trigger and action config |
| `incidents` | Raised when rules trigger; severity and lifecycle status |
| `notifications` | Per-channel delivery tracking per incident |

### identity_db (exists, temporary)

| Table | Purpose |
|---|---|
| `users` | Minimal local user records. Temporary: authentication moves to the external auth middleware |
| `home_members` | Which users belong to which homes, with a role. `user_id` is a soft ref so middleware user ids can replace local ones without a schema change |

Authentication and the authorization of service and data access are owned by the external auth middleware, not by this database. When the middleware lands, `users` is retired or reduced to a read cache; `home_members` remains as the platform's home-level access data, keyed by middleware user ids.

## Rules

### 1. No cross-database foreign keys

A table in one database never has a foreign key into another database. Cross-service references are plain UUID columns (for instance `devices.gateway_id` references a row in `ingest_db.gateways`, but with no database-enforced constraint). Each database must be able to dump, restore, and migrate independently.

### 2. Ownership is exclusive

Only the owning service connects to its database. If service A needs data owned by service B, it calls B's API or consumes B's events; it never opens a connection to B's database.

### 3. UUIDs are the shared identity

Every row that another service might reference gets a UUID primary key generated at creation. That id is the stable cross-service key; it never changes and is never reused.

### 4. Consistency is eventual, via events

When service A changes something service B cares about, A publishes an event on Kafka and B updates its own copy. MQTT is used only between gateways and ingest-service. There is no distributed transaction. Schemas are designed so this is sufficient: for instance `twin_state` is rebuilt from `readings` events, not locked against them.

### 5. Vocabularies are enforced and documented

Enumerated values (gateway types, statuses, relation types, incident severities) are check constraints in SQL and listed below. Adding an enumerated value is a deliberate, reviewed migration, not a free-form string; open-ended command names are documented separately.

### 6. Extension data goes in JSONB, not new columns

Gateway-specific or experimental data goes in designated JSONB columns (`attributes`, `properties`, `trigger_config`, `command_map`). A new dedicated column is added only when a query path needs to filter or index on it directly.

## Vocabularies

| Vocabulary | Values | Where enforced |
|---|---|---|
| gateway type | `home_assistant` | `gateways.gateway_type` check |
| gateway status | `online`, `offline`, `error`, `unknown` | `gateways.status` check |
| command status | `pending`, `sent`, `acknowledged`, `failed`, `timeout` | `commands.status` check |
| relation type | `connects_to`, `contains`, `monitors`, `controls`, `same_physical_device`, `depends_on` | `twin_relations.relation_type` check (planned) |
| incident severity | `info`, `warning`, `critical` | `incidents.severity` check (planned) |
| incident status | `open`, `acknowledged`, `resolved` | `incidents.status` check (planned) |
| notification channel | `ui`, `email`, `push` | `notifications.channel` check (planned) |
| home member role | `owner`, `member`, `viewer` | `home_members.role` check (planned) |

Normalized command names (`turn_on`, `turn_off`, `set_temperature`, ...) are defined in [gateway-api.md](gateway-api.md).

## Future hooks already in the schema

These exist now so later features need no migration on populated tables:

- `readings.recorded_at` alongside `time`: separates event time from ingest time, enabling offline buffering, backfill, and sync-lag measurement.
- `commands.issued_by` and `commands.expires_at`: audit of who issued a command, and a TTL so stale commands expire instead of firing late.
- `commands.correlation_id`: links a command to its MQTT response and to any resulting state change.
- `gateways.protocol_version`: adapter contract versioning once more gateways join.
- `entities.command_map` (planned): per-entity mapping from normalized command to gateway-native action.
- `twin_relations.valid_from` / `valid_to` (planned): time-varying topology without losing history.

## Related documents

- [mqtt-envelope.md](mqtt-envelope.md): the normalized read model flowing into `readings`
- [gateway-api.md](gateway-api.md): the command path the `commands` table audits
