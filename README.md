# IDIG4110 Group 1: Smart Home Digital Twin

Platform for reading data from multiple smart home gateways (Home Assistant Green first, Milesight and Wattsense later) and collating it into a digital twin. Microservice architecture in Go, orchestrated with Docker Compose.

## Architecture

See [docs/architecture](docs/architecture) for the data model and messaging contracts.

## Development

```sh
docker compose up
```

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
