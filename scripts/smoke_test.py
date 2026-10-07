#!/usr/bin/env python3
"""Smoke test for the twin_state path of twin-core.

Starts twin-core (and twin-db via depends_on), applies its migrations,
then checks live: a reading in, current + previous state out, and the
error contract.

Usage: python scripts/smoke_test.py
Requires: docker compose (see docker-compose.yml), Python 3.7+
"""

import json
import subprocess
import sys
import time
import urllib.error
import urllib.request

API = "http://localhost:8083"
GATEWAY = "11111111-1111-1111-1111-111111111111"
ENTITY = f"sensor.smoke_{int(time.time())}"

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


def req(method, path, body=None):
    """HTTP request -> (status, decoded json body)."""
    data = json.dumps(body).encode() if body is not None else None
    request = urllib.request.Request(API + path, data=data, method=method)
    if data:
        request.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(request) as resp:
            return resp.status, json.loads(resp.read() or b"null")
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read() or b"null")
    except OSError as e:  # refused, reset, timeout: service down
        return 0, {"error": str(e)}


def reading(ts, value):
    return {
        "gateway_id": GATEWAY,
        "external_entity_id": ENTITY,
        "timestamp": ts,
        "value_num": value,
        "value_text": None,
    }


def main():
    subprocess.run(["docker", "compose", "up", "-d", "twin-core"],
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(60):  # wait for twin-core to come up
        if req("GET", "/healthz")[0] == 200:
            break
        time.sleep(2)
    subprocess.run(["docker", "compose", "exec", "-T", "twin-core",
                    "go", "run", "./cmd/migrate", "up"], capture_output=True)

    print("twin-core")
    status, _ = req("GET", "/healthz")
    check("GET /healthz", status == 200, f"status {status}")

    status, _ = req("POST", "/api/v1/readings", reading("2026-10-07T14:00:00Z", 21.5))
    check("POST /api/v1/readings", status == 201, f"status {status}")

    status, page = req("GET", "/api/v1/state?limit=500")
    item = next((i for i in page.get("items", [])
                 if i.get("external_entity_id") == ENTITY), None)
    check("entity auto-provisioned and listed", item is not None)
    entity_id = item["entity_id"] if item else ""

    status, ent = req("GET", f"/api/v1/entities/{entity_id}/state")
    state = ent.get("state") or {}
    check("current state is 21.5",
          status == 200 and state.get("value_num") == 21.5, f"status {status}")

    status, _ = req("POST", "/api/v1/readings", reading("2026-10-07T14:10:00Z", 22.5))
    check("POST newer reading", status == 201, f"status {status}")
    status, ent = req("GET", f"/api/v1/entities/{entity_id}/state")
    state = ent.get("state") or {}
    previous = ent.get("previous") or {}
    check("newer reading moves 21.5 to previous",
          status == 200 and state.get("value_num") == 22.5
          and previous.get("value_num") == 21.5, f"status {status}")

    bad = reading("2026-10-07T15:00:00Z", 1)
    bad["value_text"] = "both values set"
    status, _ = req("POST", "/api/v1/readings", bad)
    check("invalid reading -> 400", status == 400, f"status {status}")

    status, _ = req("GET", "/api/v1/entities/00000000-0000-0000-0000-000000000009/state")
    check("unknown entity -> 404", status == 404, f"status {status}")

    print(f"\npassed: {passed}  failed: {failed}")
    sys.exit(1 if failed else 0)


if __name__ == "__main__":
    main()
