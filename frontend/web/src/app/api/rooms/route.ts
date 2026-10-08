import type { NextRequest } from 'next/server';

import { backendErrorResponse } from '@/lib/api/backend';
import { mockRooms } from '@/lib/api/mock-data';
import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';
import {
  loadTwinHomeState,
  twinHomeStateToRooms,
  twinServes,
} from '@/lib/api/twin';

// Placeholder route implementing the contract in docs/frontend/api.md.
// Serves fixture data, forwards to PLACEHOLDER_API_TARGET when set, or serves
// twin-core (hybrid) when configured.
export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  // Hybrid: serve /rooms from the twin-core scaffold when configured.
  if (twinServes('/rooms')) {
    try {
      const state = await loadTwinHomeState();
      return jsonOk(state ? twinHomeStateToRooms(state) : []);
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () => jsonOk(Object.values(mockRooms)));
}
