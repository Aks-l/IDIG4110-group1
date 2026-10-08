import type { NextRequest } from 'next/server';

import {
  INCIDENTS_PATH,
  backendErrorResponse,
  backendServes,
  fetchBackendJson,
  incidentToEvent,
} from '@/lib/api/backend';
import type { BackendIncident } from '@/lib/api/backend';
import { mockEvents } from '@/lib/api/mock-data';
import {
  handlePlaceholderRequest,
  jsonError,
  jsonOk,
} from '@/lib/api/placeholder-handler';

export const dynamic = 'force-dynamic';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ eventId: string }> },
) {
  const { eventId } = await params;

  // Hybrid: the backend has no single-incident endpoint, so the detail is
  // looked up in the newest incidents (the list endpoint caps at 100).
  if (backendServes('/events')) {
    try {
      const incidents = await fetchBackendJson<BackendIncident[]>(INCIDENTS_PATH);
      const incident = incidents.find((candidate) => candidate.id === eventId);
      return incident
        ? jsonOk(incidentToEvent(incident))
        : jsonError(404, `Unknown event: ${eventId}`);
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () => {
    const event = mockEvents.find((candidate) => candidate.id === eventId);
    return event ? jsonOk(event) : jsonError(404, `Unknown event: ${eventId}`);
  });
}
