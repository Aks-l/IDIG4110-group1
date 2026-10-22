#!/usr/bin/env python3
"""Smoke test for the dev compose stack (see docs/architecture/, docker-compose.yml).

Covers the smoke level of the test plan: the stack is up, every service is
healthy, the gateway aggregates upstream health and proxies the /api/v1
prefixes with CORS, the simulated sensors produce a growing stream of
readings, and one manually published MQTT reading flows through ingest's
read API (the core path that is on main; twin-core has no HTTP API yet).

Usage: python scripts/smoke_test.py
Requires: the dev stack running (docker compose up -d --build, make migrate-up),
Python 3.7+, and the mosquitto container for publishing (docker compose exec).
"""

import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.request

INGEST = os.environ.get("SMOKE_INGEST_URL", "http://localhost:8081")
RULES = os.environ.get("SMOKE_RULES_URL", "http://localhost:8083")
GATEWAY = os.environ.get("SMOKE_GATEWAY_URL", "http://localhost:8084")

# The gateway seeded by migrations and used by the device-simulator.
SIMULATED_GATEWAY = "00000000-0000-0000-0000-000000000001"

ENTITY = f"smoke_test_{int(time.time())}"
MQTT_TOPIC = f"sensors/{ENTITY}/value/state"

passed = 0
failed = 0


def check(name, ok, detail=""):
    global passed, failed
    if ok:
        passed += 1
        print(f"  PASS {name}")
    else:
        failed += 1
        print(f"  FAIL {name} ({detail})")


def req(base, path, method="GET", body=None, headers=None):
    """HTTP request -> (status, decoded json body, response headers)."""
    data = json.dumps(body).encode() if body is not None else None
    request = urllib.request.Request(base + path, data=data, method=method)
    request.add_header("Accept", "application/json")
    if data:
        request.add_header("Content-Type", "application/json")
    for key, value in (headers or {}).items():
        request.add_header(key, value)
    try:
        with urllib.request.urlopen(request, timeout=10) as resp:
            raw = resp.read()
            return resp.status, json.loads(raw) if raw else None, dict(resp.headers)
    except urllib.error.HTTPError as e:
        raw = e.read()
        try:
            parsed = json.loads(raw) if raw else None
        except ValueError:
            parsed = raw.decode("utf-8", "replace")
        return e.code, parsed, dict(e.headers)
    except OSError as e:  # refused, reset, timeout: service down
        return 0, {"error": str(e)}, {}


def wait_for(base, path="/healthz", timeout=120):
    deadline = time.time() + timeout
    while time.time() < deadline:
        if req(base, path)[0] == 200:
            return True
        time.sleep(2)
    return False


def mqtt_publish(topic, payload):
    """Publish via the mosquitto container's own client."""
    result = subprocess.run(
        ["docker", "compose", "exec", "-T", "mosquitto",
         "mosquitto_pub", "-t", topic, "-q", "1", "-s"],
        input=payload, capture_output=True, text=True)
    return result.returncode == 0


def sensors(base):
    """GET the sensor list -> (status, {external_entity_id: item})."""
    status, body, _ = req(base, "/api/v1/ingest/sensors")
    items = body if isinstance(body, list) else []
    return status, {i.get("external_entity_id"): i for i in items
                    if isinstance(i, dict)}


def main():
    print("waiting for the stack (docker compose up -d --build, make migrate-up)")
    if not wait_for(GATEWAY) or not wait_for(INGEST) or not wait_for(RULES):
        print("\nstack not reachable: start it first with")
        print("  docker compose up -d --build && make migrate-up")
        sys.exit(2)

    # Idempotent: a no-op when the migrations are already applied.
    subprocess.run(["docker", "compose", "exec", "-T", "ingest-service",
                    "go", "run", "./cmd/migrate", "up"], capture_output=True)

    print("health")
    status, _, _ = req(INGEST, "/healthz")
    check("ingest GET /healthz", status == 200, f"status {status}")

    status, _, _ = req(RULES, "/healthz")
    check("rules-engine GET /healthz", status == 200, f"status {status}")

    status, body, _ = req(GATEWAY, "/healthz")
    check("gateway GET /healthz", status == 200, f"status {status}")
    check("gateway reports status ok",
          body is not None and body.get("status") == "ok", f"body {body}")
    upstreams = (body or {}).get("upstreams") or {}
    check("gateway sees all upstreams true",
          upstreams and all(upstreams.values()),
          f"upstreams {upstreams}")
    check("gateway aggregates ingest, rules and incidents",
          {"ingest", "rules", "incidents"} <= set(upstreams),
          f"upstreams {sorted(upstreams)}")

    print("gateway routing")
    status, body, headers = req(GATEWAY, "/api/v1/ingest/sensors")
    check("gateway proxies /api/v1/ingest", status == 200, f"status {status}")
    check("gateway sends CORS header",
          "Access-Control-Allow-Origin" in headers, f"headers {headers}")

    status, _, _ = req(GATEWAY, "/api/v1/rules")
    check("gateway proxies /api/v1/rules", status == 200, f"status {status}")

    status, _, _ = req(GATEWAY, "/api/v1/incidents")
    check("gateway proxies /api/v1/incidents", status == 200, f"status {status}")

    print("ingest read API")
    status, by_entity = sensors(INGEST)
    check("GET /api/v1/ingest/sensors", status == 200, f"status {status}")
    required = {"sensor_id", "external_entity_id", "device_class",
                "unit", "last_seen", "reading_count"}
    check("sensor items carry the contract fields",
          by_entity and all(required <= set(i) for i in by_entity.values()),
          f"fields {sorted(next(iter(by_entity.values()), {}))}")

    status, _, _ = req(INGEST, "/api/v1/ingest/sensor-data?from=not-a-timestamp")
    check("invalid from -> 400", status == 400, f"status {status}")

    print("simulated sensors")
    before = by_entity
    check("at least two sensors are streaming",
          sum(1 for i in before.values() if i.get("reading_count", 0) > 0) >= 2,
          f"{[k for k, v in before.items() if v.get('reading_count', 0) > 0]}")

    deadline = time.time() + 45
    grew = False
    while time.time() < deadline and not grew:
        time.sleep(5)
        _, after = sensors(INGEST)
        grew = any(after.get(k, {}).get("reading_count", 0) > v.get("reading_count", 0)
                   for k, v in before.items() if v.get("reading_count", 0) > 0)
    check("reading counts grow while watching", grew,
          f"before {sorted((k, v.get('reading_count')) for k, v in before.items())}")

    print("MQTT reading through ingest")
    event = {
        "event_type": "state_changed",
        "time_fired": time.strftime("%Y-%m-%dT%H:%M:%S", time.gmtime()) + "Z",
        "entity_id": ENTITY,
        "new_state": {
            "state": "21.5",
            "attributes": {"device_class": "temperature",
                           "unit_of_measurement": "°C"},
        },
    }
    envelope = {
        "time": event["time_fired"],
        "gateway_id": SIMULATED_GATEWAY,
        "topic": MQTT_TOPIC,
        "payload": event,
    }
    ok = mqtt_publish(MQTT_TOPIC, json.dumps(envelope))
    check("publish to " + MQTT_TOPIC, ok, "mosquitto_pub failed")

    seen = False
    reading = None
    deadline = time.time() + 15
    while time.time() < deadline and not seen:
        status, body, _ = req(INGEST, "/api/v1/ingest/sensor-data?entity_id=" + ENTITY)
        rows = body if isinstance(body, list) else []
        reading = next((r for r in rows
                        if r.get("value_num") == 21.5), None)
        seen = reading is not None
        if not seen:
            time.sleep(2)
    check("reading reaches the read API", seen, f"status {status}, body {body}")
    if reading:
        check("reading keeps its unit",
              reading.get("unit") == "°C", f"unit {reading.get('unit')!r}")
        check("reading keeps its device class",
              reading.get("device_class") == "temperature",
              f"device_class {reading.get('device_class')!r}")

    _, after = sensors(INGEST)
    check("published entity is listed as a sensor",
          after.get(ENTITY, {}).get("reading_count", 0) >= 1,
          f"sensors {sorted(after)}")

    print(f"\npassed: {passed}  failed: {failed}")
    sys.exit(1 if failed else 0)


if __name__ == "__main__":
    main()
