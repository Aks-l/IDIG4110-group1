import type { NextRequest } from 'next/server';

import { backendErrorResponse } from '@/lib/api/backend';
import { mockOverview } from '@/lib/api/overview.data';
import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';
import {
  loadTwinHomeState,
  twinHomeStateToOverview,
  twinServes,
} from '@/lib/api/twin';

// Placeholder route implementing the contract in docs/frontend/api.md.
// Serves fixture data, forwards to PLACEHOLDER_API_TARGET when set, or serves
// twin-core (hybrid) when configured.
export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  // Hybrid: serve /overview from the twin-core scaffold when configured.
  if (twinServes('/overview')) {
    try {
      const state = await loadTwinHomeState();
      return jsonOk(
        state
          ? twinHomeStateToOverview(state)
          : { stats: [], devices: { connected: 0, warning: 0, events: 0 } },
      );
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () => jsonOk(mockOverview));
}
