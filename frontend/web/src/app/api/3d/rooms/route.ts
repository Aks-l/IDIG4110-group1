import type { NextRequest } from 'next/server';

import { backendErrorResponse } from '@/lib/api/backend';
import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';
import {
  loadTwinHomeState,
  twinHomeStateToRoomLayouts,
  twinServes,
} from '@/lib/api/twin';
import { mockRoomLayout } from '@/lib/api/three-d.data';

export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  // Hybrid: serve /3d/rooms from the twin-core scaffold when configured.
  if (twinServes('/3d/rooms')) {
    try {
      const state = await loadTwinHomeState();
      return jsonOk(state ? twinHomeStateToRoomLayouts(state) : []);
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () => jsonOk(mockRoomLayout));
}
