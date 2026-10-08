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
import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';

export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  // Hybrid: serve /events from the rules-engine when configured.
  if (backendServes('/events')) {
    try {
      const incidents = await fetchBackendJson<BackendIncident[]>(INCIDENTS_PATH);
      return jsonOk(incidents.map(incidentToEvent));
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () => jsonOk(mockEvents));
}
