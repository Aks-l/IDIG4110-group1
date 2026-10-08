# IDIG4110 Group 1: Smart Home Digital Twin

Platform for reading data from multiple smart home gateways (Home Assistant Green first, Milesight and Wattsense later) and collating it into a digital twin. Microservice architecture in Go, orchestrated with Docker Compose.

## Architecture

See [docs/architecture](docs/architecture) for the data model and messaging contracts.

## Development

```sh
docker compose up
```

### Migrations

Migrations live in `src/<service>/migrations`; each service has its own `cmd/migrate`. Ingest uses `make migrate-up`; twin-core: `docker compose exec twin-core go run ./cmd/migrate up` (the compose service presets `DB_URL`) — though the twin-core and rules-engine servers also apply pending migrations at startup. The twin-core migration `20261008200000_seed_demo_home` seeds the demo home (`00000000-0000-0000-0000-000000000001`) with areas, devices, entities and state, including the entity ids the seeded safety rules act on.

### Services

| Service | Host port | Purpose |
|---|---|---|
| ingest-service | 8081 | Ingests gateway telemetry over MQTT into `ingest_db` |
| device-simulator | 8082 | Publishes mock Home Assistant state_changed events for development |
| twin-core | 8084 | Receives normalized readings from ingest over HTTP, tracks current state per entity (`twin_state`), and serves it to the frontend |

### Databases

Database-per-service: each microservice owns its own Postgres container. All use the TimescaleDB image; the `timescaledb` extension is only enabled where needed.

| Container    | Database      | User     | Password | Host port |
|--------------|---------------|----------|----------|-----------|
| ingest-db    | ingest_db     | ingest   | ingest   | 5432      |
| twin-db      | twin_db       | twin     | twin     | 5433      |
| rules-db     | rules_db      | rules    | rules    | 5434      |
| identity-db  | identity_db   | identity | identity | 5435      |

Dev credentials are intentionally simple. Production credentials come from environment variables (see .env.example).

### Services

| Service | Host port | What it does | Docs |
|---|---|---|---|
| ingest-service | 8081 | Receives gateway data over MQTT, stores it in `ingest_db`, publishes normalized readings on Kafka and forwards device commands from Kafka to MQTT | [mqtt-envelope.md](docs/architecture/mqtt-envelope.md), [gateway-api.md](docs/architecture/gateway-api.md) |
| rules-engine | 8083 | Evaluates automation rules against the readings and sends commands, raises incidents and queues notifications. REST API under `/api/v1` | [rules-engine.md](docs/architecture/rules-engine.md) |
| device-simulator | 8082 | Publishes simulated sensor states to Mosquitto | |

### Message brokers

- **Kafka** is the event bus between services. Containers use `kafka:29092`; tools on the host use `localhost:9092`. Topics: `twin.readings`, `twin.commands` and `twin.incidents`. See [decision 0001](docs/decisions/0001-kafka-event-bus.md).
- **Mosquitto (MQTT)** listens on port 1883 for gateways and devices, with anonymous access in dev.

### Deployment

The SkyHiGh (OpenStack) environment and its deploy script are described in [infra/skyhigh/README.md](infra/skyhigh/README.md).
