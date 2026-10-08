# 0001: Kafka as the internal event bus, MQTT at the device edge

- **Status:** Accepted
- **Date:** 2026-10-05

## Context

The architecture design presented in the status meeting and the week 37
retrospective places a Kafka event bus between the IoT gateway and the
microservices. Devices reach the gateway over protocol-specific links (MQTT
first).

The branch `feat/database-schemas` instead stated that services stay
consistent by publishing MQTT events, which would make Mosquitto both the
device broker and the service bus. The group had to choose one model for the
prototype and the report.

Relevant requirements and research questions:

- FR-DI-01/02/03: ingest from all devices, timestamps and deduplication, and
  buffering during connectivity loss.
- FR-DT-02: keep the twin synchronised through events.
- RQ1: handling inconsistent, stale or missing data.
- RQ4: trade-offs and benefits of event-driven synchronisation.
- Architectural keypoints: low coupling and support for different protocols.

## Decision

Use both, each where it fits:

- **MQTT (Mosquitto)** is used only at the edge, between gateways or devices
  and `ingest-service`. Device-specific protocols and payloads never go further.
- **Kafka** is the bus between services. `ingest-service` validates and
  normalises readings (see `docs/architecture/mqtt-envelope.md`) and publishes
  them to Kafka. Commands travel the other way: from services over Kafka to
  `ingest-service`, then over MQTT to the gateway
  (`docs/architecture/gateway-api.md`).
- Messages are keyed by `gateway_id` + `external_entity_id`, so all events for
  one entity land in one partition and stay in order.
- Each service consumes with its own consumer group, so copies of a service
  share the load.

## Consequences

**Positive**

- **Events are kept and can be replayed.** A service that was down catches up
  from its saved consumer group position. A brand new consumer group starts at
  the newest records by default (`ConsumeResetOffset(AtEnd)` in
  `src/shared/kafka`); rebuilding state from full history, for example a
  rebuilt `twin_state`, means deliberately starting at the oldest offset.
  This supports FR-DI-03, RQ1 and RQ4.
- **Services scale on their own.** Consumer groups split partitions between
  copies, so we don't need MQTT shared subscriptions on the service side.
- **The edge is decoupled from the core.** Adding a gateway type or protocol
  changes only `ingest-service`, not the services behind it.

**Negative**

- **One more stateful component to run and secure.** In the SkyHiGh setup Kafka
  runs as a single broker on `data-vm` with replication factor 1, so it has no
  broker failover. That is acceptable for the prototype and worth discussing in
  the report.
- **More resources.** Kafka is JVM-based and adds about 1 GB of RAM.
- **`docs/architecture/data-model.md` rule 4** ("consistency is eventual, via
  events") must name Kafka instead of MQTT.

## Alternatives considered

- **MQTT for everything:** simpler, with one broker. It has no replayable log,
  so delivery to an offline consumer depends on persistent sessions and QoS,
  and scaling consumers depends on shared subscriptions.
- **Direct REST calls between services:** tighter coupling and synchronous
  failure chains. Rejected for the read path; REST remains for the
  dashboard-facing API.
