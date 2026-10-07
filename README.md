# IDIG4110 Group 1: Smart Home Digital Twin

Platform for reading data from multiple smart home gateways (Home Assistant Green first, Milesight and Wattsense later) and collating it into a digital twin. Microservice architecture in Go, orchestrated with Docker Compose.

## Architecture

See [docs/architecture](docs/architecture) for the data model and messaging contracts.

## Development

```sh
docker compose up
```

### Migrations

Migrations live in `src/<service>/migrations`; each service has its own `cmd/migrate`. Ingest uses `make migrate-up`; twin-core: `docker compose exec twin-core go run ./cmd/migrate up` (the compose service presets `DB_URL`).

### Services

| Service | Host port | Purpose |
|---|---|---|
| ingest-service | 8081 | Ingests gateway telemetry over MQTT into `ingest_db` |
| device-simulator | 8082 | Publishes mock Home Assistant state_changed events for development |
| twin-core | 8083 | Receives normalized readings from ingest over HTTP, tracks current state per entity (`twin_state`), and serves it to the frontend |

### Databases

Database-per-service: each microservice owns its own Postgres container. All use the TimescaleDB image; the `timescaledb` extension is only enabled where needed.

| Container    | Database      | User     | Password | Host port |
|--------------|---------------|----------|----------|-----------|
| ingest-db    | ingest_db     | ingest   | ingest   | 5432      |
| twin-db      | twin_db       | twin     | twin     | 5433      |
| rules-db     | rules_db      | rules    | rules    | 5434      |
| identity-db  | identity_db   | identity | identity | 5435      |

Dev credentials are intentionally simple. Production credentials come from environment variables (see .env.example).

### Message broker

Mosquitto (MQTT) on port 1883, anonymous access in dev.
