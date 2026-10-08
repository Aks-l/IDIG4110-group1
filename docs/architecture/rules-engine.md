# Rules engine

The rules engine (`src/rules-engine`) watches the reading stream and reacts to
it. When a rule fires, it can send device commands, raise an incident, and
schedule notifications. It covers FR-WA-01 (user-defined rules), FR-WA-02
(smoke response) and FR-WA-03 (water leak response), and gives a basis for
SRQ1 (naive triggers compared with context-aware rules).

## Where it sits

```
gateway ─MQTT─▶ ingest-service ─Kafka twin.readings─▶ rules-engine
                      ▲                                   │
                      └──────── Kafka twin.commands ◀─────┤
                                                          ├─▶ rules_db (rules, incidents, notifications)
                                                          ├─Kafka twin.incidents─▶ notification service, UI
                                                          └─ REST /api/v1 (through the API gateway)
```

| Topic | Producer | Consumer | Key | Payload |
|---|---|---|---|---|
| `twin.readings` | ingest-service | rules-engine (group `rules-engine`), twin-core (group `twin-core`) | `{gateway_id}/{external_entity_id}` | normalized reading, [mqtt-envelope.md](mqtt-envelope.md) |
| `twin.commands` | rules-engine | ingest-service (group `ingest-service-commands`) | `{gateway_id}/{external_entity_id}` | command, [gateway-api.md](gateway-api.md) |
| `twin.incidents` | rules-engine | notification service and UI (planned) | `home_id` | `IncidentEvent` in `src/shared/dto` |

## Rule format

A rule is stored in the `rules` table. `trigger_config` holds the trigger,
conditions and cooldown; `action_config` holds the actions. The API uses the
same fields directly on the rule.

```json
{
  "home_id": "00000000-0000-0000-0000-000000000001",
  "name": "Smoke detected",
  "priority": 100,
  "trigger":    { "entity": "binary_sensor.kitchen_smoke", "operator": "eq", "value": true, "for_seconds": 0 },
  "conditions": [ { "entity": "binary_sensor.home_occupied", "operator": "eq", "value": "on" } ],
  "cooldown_seconds": 300,
  "actions": [
    { "type": "command",  "entity": "lock.front_door", "command": "unlock" },
    { "type": "incident", "severity": "critical", "message": "Smoke detected ({{entity}})" },
    { "type": "notify",   "channels": ["ui", "push"] }
  ]
}
```

**Trigger and conditions** each compare the latest reading of one entity:

- `operator` is `eq`, `ne`, `gt`, `gte`, `lt` or `lte`. The `gt`, `gte`, `lt`
  and `lte` operators need a numeric `value`.
- A numeric `value` compares numerically. This also works when the reading
  carries the number as text.
- A string `value` compares case-insensitively with the reading's text value.
- A boolean `value` stands for the `on`/`off` states that binary sensors report.
- `gateway_id` is optional. Without it, the entity matches on any gateway.

**Actions** run in the order listed:

| Type | Fields | Effect |
|---|---|---|
| `command` | `entity`, `command`, optional `parameters`, optional `gateway_id` | Publishes a command on `twin.commands`. The gateway defaults to the gateway of the triggering reading. `expires_at` is the issue time plus `COMMAND_TTL_SECONDS` (default 30). |
| `incident` | `severity` (`info`, `warning`, `critical`), `message` | Stores an incident and publishes it on `twin.incidents`. `{{entity}}` and `{{value}}` in the message are replaced. At most one per rule. |
| `notify` | `channels` (`ui`, `email`, `push`) | Creates one pending notification per channel for the rule's incident. A notify action requires an incident action, because notifications belong to an incident. |

## Evaluation

- **State.** The engine keeps the latest reading of every entity in memory. A
  rule holds when the trigger and every condition match. An entity with no
  reading yet does not match.
- **Edge-triggered.** A rule fires when it goes from not holding to holding. It
  does not fire again on each reading while it keeps holding. It can fire again
  once it has stopped holding and starts again.
- **`for_seconds`.** The rule fires only after it has held for that long. If it
  stops holding in the meantime, the wait is cancelled. A background tick
  (`TICK_MILLIS`, default 1000) does the check.
- **Cooldown.** After firing, a rule does not fire again for
  `cooldown_seconds`. A firing suppressed this way is skipped, not queued.
- **Priority.** Rules that fire on the same reading run in descending
  `priority` order.
- **Out-of-order data.** A reading older than the one already known for its
  entity is ignored, so late or replayed data never overwrites newer state
  (RQ1).
- **Rule changes.** Changes through the API apply immediately. Changes made
  directly in the database are picked up by a periodic reload
  (`RULE_REFRESH_SECONDS`, default 30). An edited rule starts again as not
  holding, and its cooldown still applies.
- **Statistics.** Each firing increments `fire_count` and sets `last_fired_at`
  on the rule, for the automations page.

The engine runs as **one replica**. Its state is in memory and conditions can
span entities, so splitting the stream across copies would give wrong
answers. Running several copies would need shared state; see Limitations.

A new consumer group starts at the newest readings. After a restart, the engine
resumes from its committed offset, but its in-memory state is empty until each
entity reports again.

## API

All paths are under `/api/v1`. Bodies are JSON.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/rules?home_id=` | List rules, highest priority first |
| `POST` | `/rules` | Create a rule. Returns `422` with every validation problem listed |
| `GET` | `/rules/{id}` | One rule |
| `PUT` | `/rules/{id}` | Replace a rule |
| `PATCH` | `/rules/{id}` | `{"enabled": true\|false}`, the automations page toggle |
| `DELETE` | `/rules/{id}` | Delete a rule. Returns `409` if it has incidents; disable it instead, so the history stays intact |
| `GET` | `/incidents?home_id=&status=&limit=` | Newest first, at most 100 by default |
| `POST` | `/incidents/{id}/acknowledge` | `open` to `acknowledged` |
| `POST` | `/incidents/{id}/resolve` | To `resolved`. Returns `409` if it is already resolved |
| `GET` | `/healthz` | Liveness: the process is up |
| `GET` | `/readyz` | Readiness: rules_db answers a ping |

## Built-in rules

The migration `20261006120001_seed_safety_rules` adds two rules for the demo
home `00000000-0000-0000-0000-000000000001`:

- **Smoke detected** (FR-WA-02): unlocks `lock.front_door` and `lock.back_door`,
  turns on `siren.alarm`, raises a critical incident, and notifies ui, push and
  email.
- **Water leak detected** (FR-WA-03): closes `valve.main_water`, raises a
  critical incident, and notifies ui and push.

Their entity ids follow Home Assistant naming. Point them at real devices with
`PUT /rules/{id}`.

## Configuration

These are environment variables. On k3s they come from the `data-tier` Secret;
see `infra/k8s/rules-engine.yaml`.

| Variable | Default |
|---|---|
| `HTTP_PORT` | `8080` |
| `DB_URL`, or `DB_HOST` `DB_PORT` `DB_USER` `DB_PASSWORD` `DB_NAME` `DB_SSLMODE` | `rules-db:5432`, `rules`/`rules`, `rules_db`, `disable` |
| `KAFKA_BROKERS` | `kafka:29092` |
| `KAFKA_GROUP` | `rules-engine` |
| `MIGRATIONS_DIR` | `migrations` |
| `COMMAND_TTL_SECONDS` | `30` |
| `TICK_MILLIS` | `1000` |
| `RULE_REFRESH_SECONDS` | `30` |

## Running and deploying

- **Locally:** run `docker compose up`. The rules engine runs on
  http://localhost:8083 and reloads on code changes. Readings flow from
  `device-simulator` through Mosquitto and `ingest-service` to Kafka.
- **Tests:** in `src/rules-engine`, run `go test ./...` to test the engine, rule
  validation and the executor. `ingest-service` has tests for the reading
  conversion and the command forwarding.
- **Image:** `infra/prod/rules-engine.prod.Dockerfile` builds a small distroless
  image from the repository root, with the migrations included. The service
  applies its migrations at startup.
- **CI:** `.github/workflows/rules-engine.yml` runs `go vet` and `go test` on
  pull requests. On `main` it also publishes
  `ghcr.io/aks-l/idig4110-group1/rules-engine:latest`.
- **SkyHiGh:** `./infra/skyhigh/deploy.sh up k8s` applies
  `infra/k8s/rules-engine.yaml`. Credentials come from the `data-tier` Secret.
  The package must be public once it exists; see
  [infra/skyhigh/README.md](../../infra/skyhigh/README.md#services-on-k3s).
  `ingest-service` is not deployed to the cluster yet, so on SkyHiGh the engine
  receives no readings until it is.

## Delivery semantics

The engine consumes `twin.readings` with liveness over completeness: one
skipped reading is cheaper than a consumer stuck on a poison record.

- A handler error is logged and the record is skipped; it is not retried.
- A crash replays up to about 5 seconds of records (the autocommit interval),
  so a firing can run twice. Incidents are deduplicated by a unique
  constraint on `(rule_id, triggered_at)`; commands are deduplicated by the
  ingest adapter by command id.
- After a restart the in-memory state is empty until each entity reports
  again, and a brand new consumer group starts at the newest records
  (`ConsumeResetOffset(AtEnd)` in `src/shared/kafka`); replaying full history
  is a deliberate operation, not the default.
- A reading is stored in ingest_db before it is published, so a failed
  publish leaves it stored but unpublished and this engine misses it. The
  latest-wins state model self heals on the next report of the entity.

## Limitations and next steps

- **No age limit on readings.** Conditions do not check how old a reading is
  yet. A `max_age_seconds` on matches would let rules treat stale sensors as
  unknown (RQ1).
- **One replica only.** State is in memory. Scaling out would need shared state,
  or partitioning rules by home together with keying readings by home.
- **Notifications are only recorded.** They are stored as `pending`; the
  notification service that delivers them and marks them `sent` does not
  exist yet.
- **Commands are not recorded in the database.** ingest-service forwards
  commands but does not record them in `ingest.commands` yet.
- **No authorization.** `home_id` comes from the caller. Once the API gateway
  and authentication exist, home access should be checked there.
