import type { NextRequest } from 'next/server';

import { backendErrorResponse } from '@/lib/api/backend';
import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';
import { listPlaceholderDevices } from '@/lib/api/placeholder-store';
import {
  loadTwinHomeState,
  twinHomeStateToDevices,
  twinServes,
} from '@/lib/api/twin';

export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  // Hybrid: serve /devices from the twin-core scaffold when configured.
  if (twinServes('/devices')) {
    try {
      const state = await loadTwinHomeState();
      return jsonOk(state ? twinHomeStateToDevices(state) : []);
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () => jsonOk(listPlaceholderDevices()));
}
