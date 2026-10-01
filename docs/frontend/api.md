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

## File Structure

```text
frontend/my-app/src/lib/api/
├── api-client.ts       Shared HTTP client and error handling
├── api-rooms.ts        Room endpoints
├── api-devices.ts      Device endpoints
├── api-events.ts       Event endpoints
├── api-automations.ts  Automation endpoints
├── types.ts            Shared response types
├── mock-data.ts        Centralized development fixtures
├── data-source.ts      Mock/API source selection
├── rooms.data.ts       Room data-source adapter
├── events.data.ts      Event data-source adapter
└── automations.data.ts Automation data-source adapter
```

## Configuration

Create `frontend/my-app/.env.local` for local development:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_DATA_SOURCE=mock
```

`NEXT_PUBLIC_DATA_SOURCE` supports:

| Value | Behavior |
| --- | --- |
| `mock` | Uses fixtures from `mock-data.ts`; no backend is required |
| `api` | Sends requests to `NEXT_PUBLIC_API_URL` |

`.env.example` documents the required variables and may be committed. `.env.local`
contains machine-specific settings and must not contain secrets intended for
client-side use. Restart the Next.js development server after changing variables.

## Shared Client

`api-client.ts` is responsible for transport concerns only:

- Prefixing endpoints with `NEXT_PUBLIC_API_URL`
- Sending JSON request and response headers
- Parsing successful JSON responses
- Supporting `204 No Content`
- Throwing `ApiError` for non-2xx responses

Feature modules should use this wrapper instead of calling `fetch` directly.

```ts
import { apiClient } from '@/lib/api/api-client';
import type { Event } from '@/lib/api/types';

const events = await apiClient<Event[]>('/events');
```

## Endpoint Modules

All paths below are relative to `NEXT_PUBLIC_API_URL`.

### Rooms

Implemented in `api-rooms.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/rooms` | `RoomData[]` |
| `GET` | `/rooms/:roomId` | `RoomData` |

### Devices

Implemented in `api-devices.ts`:

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

### Events

Implemented in `api-events.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/events` | `Event[]` |

### Automations

Implemented in `api-automations.ts`:

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/automations` | `Automation[]` |
| `PATCH` | `/automations/:automationId` | `Automation` |

The automation update currently sends:

```json
{
	"enabled": true
}
```

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

Devices currently have endpoint functions but no separate data adapter. Add a
`devices.data.ts` adapter when devices also need mock/API switching or local
fallback behavior.

## Error Handling

`ApiError` is defined and exported from `src/lib/api/api-client.ts`:

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
import { ApiError } from '@/lib/api/api-client';

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

After agreement, update the API modules, types, and this document together.
