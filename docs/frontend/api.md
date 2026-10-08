# Frontend API Integration

## Purpose

The frontend uses a small API client layer so pages and components do not call
`fetch` directly. API communication is organized by feature:

```text
Page or hook
	-> feature data loader
		-> feature API module
			-> apiClient
				-> backend API
```

The current endpoint paths and response types are provisional. They provide the
frontend infrastructure until the backend API contract is agreed and should be
updated together with the backend documentation when that contract exists.
Until the backend exposes these endpoints, the placeholder routes in
`frontend/web/src/app/api` serve the same contract with fixture data; see
[Placeholder endpoints](#placeholder-endpoints).

## File Structure

```text
frontend/web/src/lib/api/
├── client.ts             Shared HTTP client and error handling
├── rooms.ts              Room endpoints
├── devices.ts            Device endpoints
├── events.ts             Event endpoints
├── automations.ts        Automation endpoints
├── overview.ts           Overview endpoints
├── energy.ts             Energy endpoints
├── 3d-model.ts           3D layout endpoints
├── types.ts              Shared response types
├── mock-data.ts          Centralized development fixtures
├── data-source.ts        Mock/API source selection
├── rooms.data.ts         Room data-source adapter
├── events.data.ts       Event data-source adapter
├── automations.data.ts   Automation data-source adapter
├── devices.data.ts       Device data-source adapter and mutations
├── three-d.data.ts       3D layout data-source adapter
├── overview.data.ts      Overview data-source adapter
├── energy.data.ts        Energy data-source adapter
├── placeholder-handler.ts  Placeholder route plumbing: fixtures or proxy
├── placeholder-store.ts    In-memory state behind the placeholder routes
├── backend.ts              Rules hybrid: backend calls and Rule/Incident mapping
└── twin.ts                 Twin hybrid: twin-core state calls and mapping

frontend/web/src/app/api/
└── Placeholder route handlers, one folder per endpoint (rooms, devices,
    overview, events, projections, energy, automations, 3d)
```

## Configuration

Create `frontend/web/.env.local` for local development:

```env
NEXT_PUBLIC_API_URL=http://localhost:3000/api
NEXT_PUBLIC_DATA_SOURCE=api
```

`NEXT_PUBLIC_DATA_SOURCE` supports:

| Value | Behavior |
| --- | --- |
| `mock` | Uses fixtures from `mock-data.ts`; no backend is required |
| `api` | Sends requests to `NEXT_PUBLIC_API_URL` |

`.env.example` documents the required variables and may be committed. `.env.local`
contains machine-specific settings and must not contain secrets intended for
client-side use. Restart the Next.js development server after changing variables.

`PLACEHOLDER_API_TARGET` is a server-side variable for the placeholder routes
(see below): when set, they forward every request to that base URL instead of
serving fixtures. It is read at request time, so it can change without a
rebuild. The hybrid variables `PLACEHOLDER_BACKEND_URL` and
`PLACEHOLDER_BACKEND_PATHS` (also server-side, also read at request time) are
described in the next section.

## Shared Client

`client.ts` is responsible for transport concerns only:

- Prefixing endpoints with `NEXT_PUBLIC_API_URL`
- Sending JSON request and response headers
- Parsing successful JSON responses
- Supporting `204 No Content`
- Throwing `ApiError` for non-2xx responses

Feature modules should use this wrapper instead of calling `fetch` directly.

```ts
import { apiClient } from '@/lib/api/client';
import type { Event } from '@/lib/api/types';

const events = await apiClient<Event[]>('/events');
```

## Placeholder endpoints

The backend does not serve this contract yet, so `frontend/web/src/app/api`
contains Next.js route handlers that implement each endpoint below with
fixture data. They exist so the frontend exercises its real HTTP path —
`apiClient`, `ApiError`, loading and error states — without a backend, and so
the response shapes are verifiable end to end.

- `GET` requests serve the fixtures; `PATCH` and `POST` keep their changes in
  memory for the lifetime of the server process.
- The handlers import the same fixtures as mock mode, so `mock` and `api`
  modes show the same data.

Swapping in the real backend is configuration-only, with two options:

1. Point the browser at the backend: set `NEXT_PUBLIC_API_URL` to its base
   URL (for example `http://localhost:8080/api`). The placeholder routes are
   bypassed and can be deleted once every endpoint is real.
2. Keep the browser on the placeholder routes: set `PLACEHOLDER_API_TARGET`
   to the backend base URL. Every placeholder route then forwards its request
   there unchanged, so endpoints can migrate one by one as the backend
   implements them.

The real backend must serve these paths and shapes under the base URL. The
merged api-gateway (host port `8080`, see
`src/api-gateway/config/config.yaml`) fronts `/api/v1/ingest/*`,
`/api/v1/rules*`, `/api/v1/incidents*` and twin-core's public API
(`/api/v1/homes*`, `/api/v1/areas*`, `/api/v1/devices*`,
`/api/v1/entities*`, `/api/v1/relations*`, `/api/v1/state*`), and it
registers the bare collection prefixes too, fixing the earlier
307-redirect quirk on `/api/v1/rules` and `/api/v1/incidents`.

### Hybrid: backend-backed paths

The paths the backend already implements can be served from it while the rest
keep serving fixtures — hybrid mode. Set two server-side variables:

```env
PLACEHOLDER_BACKEND_URL=http://localhost:8083
PLACEHOLDER_BACKEND_PATHS=/automations,/events
```

`PLACEHOLDER_BACKEND_URL` is the backend base URL — the rules-engine directly
(host port `8083` in the development compose setup). The merged api-gateway
(host port `8080`) also fronts the rules paths and now registers the bare
collection prefix as well (`mux.Handle(u.Prefix, p)` in
`src/api-gateway/internal/server/router.go`), fixing the 307-redirect quirk
that used to make `/api/v1/rules` 404; either base works, the rules-engine
just skips the hop. When a path is listed in
`PLACEHOLDER_BACKEND_PATHS`, its placeholder route fetches from the backend,
translates the response onto the frontend contract, and serves that; backend
errors surface as-is, there is no silent fallback to fixtures. Unlisted paths
are unaffected, and backend-backed paths take precedence over
`PLACEHOLDER_API_TARGET` forwarding.

Docker Compose runs this way out of the box: `frontend/web/docker-compose.yml`
defaults to `NEXT_PUBLIC_DATA_SOURCE=api` and these hybrid variables, with
`host.docker.internal` in place of `localhost` — the container reaches the
stack's published ports through the Docker host. See
`frontend/web/README.md` for the Docker workflow.

| Placeholder path | Gateway path | Translation |
| --- | --- | --- |
| `GET /automations` | `GET /api/v1/rules` | `Rule` → `Automation` |
| `POST /automations` | `POST /api/v1/rules` | draft → rule input, see below |
| `PATCH /automations/:id` | `PATCH /api/v1/rules/{id}` | `{enabled}` passes through, `Rule` → `Automation` |
| `GET /events` | `GET /api/v1/incidents` | `Incident` → `Event` |
| `GET /events/:id` | `GET /api/v1/incidents` | found in the newest incidents (the list caps at 100) |

The mapping lives in `src/lib/api/backend.ts` and is lossy where the models
differ:

- `fire_count` and `last_fired_at` become `runCount` and `lastRun`. The
  category is derived from the actions: rules that raise incidents are
  `Security`, notify-only rules are `Notification`, the rest `Comfort`.
- Backend matches render as the closest condition kind: numeric matches as
  temperature comparisons, presence sensors as presence, other binary
  matches as motion in the entity's location (for example smoke).
- Creating an automation needs a rule trigger, so the first mappable
  condition becomes the trigger and the rest become conditions. Time and day
  conditions, delay actions and multiple notify actions cannot be expressed
  by the rules backend and are rejected with a `422` and a message instead of
  being silently dropped.
- The frontend has no home concept, so rules are created in
  `PLACEHOLDER_BACKEND_HOME_ID` (default `00000000-0000-0000-0000-000000000001`,
  the seeded demo home). Temperature conditions use the entity in
  `PLACEHOLDER_BACKEND_TEMPERATURE_ENTITY` (default: the first simulated
  sensor). Device actions become `domain.entity` ids with the domain guessed
  from the device name; motion conditions become
  `binary_sensor.<location>_motion`.

### Hybrid: twin-core-backed paths

twin-core (documented in `docs/architecture/twin-state-api.md`) serves the
structural model (homes, areas, devices, entities, relations) and
`twin_state`, the latest known value per entity. A second hybrid mode serves
the placeholder paths it can back, again config-only:

```env
PLACEHOLDER_TWIN_URL=http://localhost:8084
PLACEHOLDER_TWIN_PATHS=/rooms,/devices,/3d/rooms,/overview
```

Everything reads one dashboard view, `GET /api/v1/homes/{home_id}/state`:
areas and devices with their entities and current state. The home is
`PLACEHOLDER_TWIN_HOME_ID` when set, otherwise the first home from
`GET /api/v1/homes` — in practice the seeded demo home
`00000000-0000-0000-0000-000000000001` (`20261008200000_seed_demo_home`),
where the device-simulator's readings land and auto-provisioning still
drops unknown devices.
Backend errors surface as-is, like the rules hybrid. Both this hybrid and
the rules hybrid are enabled by default: `npm run dev` reads them from
`.env.local` (`localhost` URLs) and Docker Compose defaults to the
`host.docker.internal` equivalents in `frontend/web/docker-compose.yml`.

| Placeholder path | Twin-core endpoint | Translation |
| --- | --- | --- |
| `GET /rooms`, `GET /rooms/:id` | `GET /api/v1/homes/{home_id}/state` | areas → `RoomData` |
| `GET /devices`, `GET /devices/:id` | `GET /api/v1/homes/{home_id}/state` | devices + entities → `Device` |
| `GET /3d/rooms` | `GET /api/v1/homes/{home_id}/state` | area geometry → `RoomLayout` |
| `GET /overview` | `GET /api/v1/homes/{home_id}/state` | home aggregate → `OverviewData` |

The mapping lives in `src/lib/api/twin.ts` and is honest about what
twin-core lacks:

- Room metrics take the first entity matching each metric (`device_class` or
  name: temperature, humidity, carbon dioxide/co2, occupancy/presence); the
  change is the delta against the entity's previous reading. Values and
  changes are rounded to two decimals — the raw subtraction carries
  floating point noise. Metrics without a sensor — or without a reading
  yet — are null and render as no data, and room activity is always empty —
  twin-core keeps current and previous state only, no history.
- `GET /overview` aggregates the home: temperature and humidity are averages
  over the rooms with a reading, air quality comes from the worst CO₂,
  occupancy is the total, and stats without data are omitted entirely.
  `connected` is the twin device count, so it matches the devices page, and
  `warning` reuses the devices page's smoke/leak heuristic
  (`src/lib/api/device-status.ts`); `events` is 0, since twin-core keeps no
  event history.
- Device toggles (`PATCH /devices/:id`) are rejected with `501`: twin-core
  has no command endpoint, and `POST /api/v1/readings` is the ingest path, not
  a device command.
- Devices without a controllable entity read as off; `installedAt` and
  `position` have no backend source and are omitted. Battery can come from a
  battery entity or the `battery` attribute; signal from the `signal`,
  `rssi`, or `linkquality` attributes; `lastSeen` is the newest
  `state.updated_at` of the device's entities.
- `GET /3d/rooms` reads `position` and `size` triples from the area's
  free-form `geometry` when present, and otherwise falls back to a
  deterministic grid.

twin-core is part of the stack: the development compose publishes it on
host port `8084` (the api-gateway moved to host port `8080` and also routes
twin-core's public API), so the hybrid works against the running backend
out of the box. twin-core applies its migrations at startup; they seed a
demo home (`20261008200000_seed_demo_home`) with rooms, devices and state,
so there is data to show without creating anything by hand.

## Endpoint Modules

All paths below are relative to `NEXT_PUBLIC_API_URL`.

### Rooms

Implemented in `rooms.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/rooms` | `RoomData[]` |
| `GET` | `/rooms/:roomId` | `RoomData` |

### Devices

Implemented in `devices.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/devices` | `Device[]` |
| `GET` | `/devices/:deviceId` | `Device` |
| `PATCH` | `/devices/:deviceId` | `Device` |

The device update currently sends:

```json
{
	"on": true
}
```

### Overview

Implemented in `overview.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/overview` | `OverviewData` |

### Events

Implemented in `events.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/events` | `Event[]` |
| `GET` | `/events/:eventId` | `Event` |

### Projections

Implemented in `events.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/projections` | `Projection[]` |

### Energy

Implemented in `energy.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/energy?period=:period&mode=:mode` | `EnergyPageData` |

`period` is `today`, `week`, `month` or `year`; `mode` is `current` or
`projected`. Both fall back to `today` and `current` when omitted.

### Automations

Implemented in `automations.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/automations` | `Automation[]` |
| `POST` | `/automations` | `Automation` |
| `PATCH` | `/automations/:automationId` | `Automation` |

The automation update currently sends:

```json
{
	"enabled": true
}
```

The automation create sends an `AutomationDraft`:

```json
{
	"name": "Hallway motion light",
	"description": "Turn on the hallway light when someone walks in at night",
	"category": "Comfort",
	"conditions": [{ "kind": "motion", "location": "Hallway" }],
	"actions": [{ "kind": "device", "deviceName": "Hallway Light", "command": "set", "value": "40%" }]
}
```

### 3D layout

Implemented in `3d-model.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/3d/rooms` | `RoomLayout[]` |

## Response Types

The current TypeScript types are defined in `src/lib/api/types.ts`.

```ts
type Device = {
	id: string;
	name: string;
	type: string;
	on: boolean;
};

type RoomMetric = {
	value: number;
	change: number;
};

type RoomMetrics = {
	temperature: RoomMetric;
	humidity: RoomMetric;
	co2: RoomMetric;
	occupancy: RoomMetric;
};

type Activity = {
	id: string;
	timestamp: string;
	description: string;
};

type RoomData = {
	id: string;
	name: string;
	metrics: RoomMetrics;
	devices: Device[];
	activity: Activity[];
};
```

These types currently match the mock data and are a provisional frontend
contract. When the backend schema is finalized, decide whether to:

1. Update these types to match the backend response directly; or
2. Add separate API response types and map them to frontend types.

The second option is preferable when API values differ from display values.

## Mock/API Data Adapters

Pages should use data adapters such as `loadRooms()`, `loadRoom()`, and
`loadEvents()` rather than choosing between mock data and API calls themselves.

```ts
import { loadRoom, loadRooms } from '@/lib/api/rooms.data';

const rooms = await loadRooms();
const room = await loadRoom('living-room');
```

This keeps the source-selection decision in one place and allows the UI to work
with the same return type in both modes. API failures in API mode are surfaced
to the page as errors; the application does not silently replace failed
production data with mock data.

The current adapters are:

| Adapter | Source covered |
| --- | --- |
| `rooms.data.ts` | Rooms and individual room details |
| `events.data.ts` | Events |
| `automations.data.ts` | Automations |
| `devices.data.ts` | Devices and device mutations |
| `three-d.data.ts` | 3D room layout |
| `overview.data.ts` | Overview dashboard data |
| `energy.data.ts` | Energy range data |

The placeholder routes serve the same fixtures over HTTP, so `mock` mode and
`api` mode show the same data until the real backend is connected.

## Error Handling

`ApiError` is defined and exported from `src/lib/api/client.ts`:

```ts
export class ApiError extends Error {
	constructor(
		message: string,
		public readonly status: number,
	) {
		super(message);
		this.name = 'ApiError';
	}
}
```

Non-2xx responses throw `ApiError` with the actual HTTP status. Network
failures, such as an unavailable server, DNS failure, or failed CORS preflight,
also throw `ApiError`, but with `status: 0` because no HTTP response was
received.

```ts
import { ApiError } from '@/lib/api/client';

try {
	const devices = await getDevices();
} catch (error) {
	if (error instanceof ApiError) {
		if (error.status === 0) {
			console.error('Network failure:', error.message);
		} else {
			console.error('HTTP failure:', error.status);
		}
	}
}
```

Pages should provide loading and error states. Authentication, authorization,
validation, and server errors remain distinguishable by their HTTP status code;
`status: 0` identifies a transport-level failure. Errors from parsing an
unexpected response body may still be thrown separately and should be handled
as unexpected API responses when response validation is added.

## Backend Contract Checklist

Before switching shared environments to API mode, confirm:

- Base URL and environment-specific URLs
- Endpoint paths and HTTP methods
- Room and device identifier format
- JSON property names and nullability
- Timestamp format and timezone rules
- Device and automation update semantics
- Authentication and CORS requirements
- Error response format
- Pagination or filtering requirements for events and history

Until each item is confirmed, the placeholder endpoints in
`frontend/web/src/app/api` stand in for the backend and can be swapped out with
`NEXT_PUBLIC_API_URL` or `PLACEHOLDER_API_TARGET` as described above. After
agreement, update the API modules, types, and this document together.
