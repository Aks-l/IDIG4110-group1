# Test Plan: Smart Home Digital Twin Prototype

branch: main
documentation: `README.md`, `docs/architecture/mqtt-envelope.md`, `docs/architecture/rules-engine.md`, `docs/architecture/twin-state-api.md`, `docs/decisions/`

## 1. The system

Devices publish MQTT state events; ingest-service normalizes them and republishes on Kafka; rules-engine evaluates automations; twin-core keeps current state per entity; the gateway fronts the public API; the dashboard is the UI. Expected behavior is defined by the documentation above, not restated here.

| Service / area | Description |
|---|---|
| ingest-service | MQTT ingestion, normalization, ingest_db storage, Kafka publication, read API |
| rules-engine | Rule evaluation, incidents, commands, rules API, readyz |
| twin-core | Twin_state upserts, auto-provisioning, structure and state APIs |
| api-gateway | Public routing of the /api/v1 prefixes, CORS, health aggregation |
| web dashboard | Next.js app: devices, rooms, events, automations, overview, 3D |
| infrastructure | Mosquitto, kafka, four Postgres databases (dev compose stack) |

## 2. How to run

| Step | Command |
|---|---|
| clone | `git clone <repo> && cd IDIG4110-group1` |
| backend stack | `docker compose up -d --build` |
| ingest migrations | `make migrate-up` |
| twin-core migrations | `docker compose exec twin-core go run ./cmd/migrate up` |
| dashboard | `cd frontend/web && docker compose up -d --build` (port 3000) |
| reset | `docker compose down -v` |

Healthy when `:8080/healthz` answers 200 with all upstreams true, the two simulated sensors show growing reading counts, and `scripts/smoke_test.py` passes.

| Port | Service |
|---|---|
| 8080 | api-gateway |
| 8081 | ingest-service |
| 8083 | rules-engine |
| 8084 | twin-core |
| 3000 | dashboard |
| 1883 | mosquitto (MQTT) |
| 9092 | kafka |
| 5432 to 5435 | ingest, twin, rules, identity databases |

## 3. Scope

| Tested | Description |
|---|---|
| component contracts | Each service against its contract in `docs/architecture/` |
| integration | Boundaries: MQTT to ingest, ingest to Kafka, Kafka to consumers, gateway and dashboard to services |
| system | end-to-end flows |
| recovery | Restarts, outages, from-scratch redeploy, stability run |
| usability | Task sessions with users from outside the team |
| non-functional | Latency, burst, exposure, log hygiene |

| Not tested | Reason |
|---|---|
| device simulator | Traffic source only |
| authentication | Not implemented; all endpoints open |
| unit and quality gates | Run by the developers |
| k8s / SkyHiGh deployment | Dev compose stack only |
| performance benchmarks | Burst and stability run suffice |

## 4. Approach

Levels run in order; each gates the next.

| Level | Techniques | Notes |
|---|---|---|
| smoke | Health endpoints, `scripts/smoke_test.py` | Gate for everything else; a failure is a blocker |
| component | Equivalence partitioning, boundary values, negative testing, state transitions | MQTT publishing and twin-core's manual readings endpoint give controlled inputs independent of the simulator |
| integration | Interface testing, fault injection at each boundary | Delivery semantics under at-least-once: Duplicates, replay, out-of-order and lost messages must end in consistent state |
| system | Scenario charters, exploratory sessions | Candidates: New sensor onboarding; smoke event to one incident end to end; UI automation that fires and surfaces; structure change in the dashboard; history window |
| reliability | Fault injection (restarts, database outage, redeploy), a stability run of at least an hour | Correct behavior per the platform model (at-least-once, last-write-wins, incident dedup); known gap: A failed publish leaves a reading in ingest but not in the twin |
| usability | Task-based sessions with users from outside the team | live stack with data |
| regression | Smoke suite plus previously failing areas | After any fix and before sign-off |

## 5. Additional tests

The test group is free to write additional tests: smoke tests, integration tests, system tests or similar, automated or manual. Anything beyond manual probing is written and run in a local development environment and agreed with the dev team first, so placement, frameworks and naming match the repositories.

## 6. Entry and exit criteria

| Entry | Exit |
|---|---|
| fresh clone builds and starts | All levels executed |
| migrations applied | Findings reported per the rubric |
| smoke suite passes | No open blocker or critical defects |
| dashboard shows live data | |

## 7. Deliverables

A list of findings, reported per the rubric in section 8.

## 8. Reporting deviations

Deviations from expected behavior are reported with a severity, a description and, where possible, how to replicate them.

| Severity | Meaning |
|---|---|
| blocker | Core flow is broken; testing or the demo cannot continue |
| critical | A feature is broken with no acceptable workaround |
| major | A feature fails or deviates during normal use; a workaround exists |
| minor | Edge case or rarely used path misbehaving |
| nit | Cosmetic or wording issue; nothing functionally wrong |

| Field | Content |
|---|---|
| severity | One of the levels above |
| description | Expected versus actual, and where: Service, page or endpoint |
| reproduction | Optional: The steps, request or payload that triggers it |
| evidence | Output, screenshot or log excerpt |

## 9. Known limitations

| Area | Detail |
|---|---|
| authentication | None; all endpoints open |
| entity ids | Simulator ids are uuids, so twin-core's domain field degrades for them |
| frontend wiring | Twin-core and rules-engine called directly; overview and energy serve fixtures; device toggles rejected until commands exist |
| gateway 404s | HTML, not the JSON error shape |
| duplicates | Expected unter at-least-once; only incorrect handling is a defect |
