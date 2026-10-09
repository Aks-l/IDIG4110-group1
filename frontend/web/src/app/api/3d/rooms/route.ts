import type { NextRequest } from 'next/server';

import { backendErrorResponse } from '@/lib/api/backend';
import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';
import {
  loadTwinHomeState,
  loadTwinRelations,
  twinConnections,
  twinHomeStateToRoomLayouts,
  twinServes,
} from '@/lib/api/twin';
import { mockConnections, mockRoomLayout } from '@/lib/api/three-d.data';

export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  // Hybrid: serve /3d/rooms from the twin-core scaffold when configured.
  if (twinServes('/3d/rooms')) {
    try {
      const [state, relations] = await Promise.all([
        loadTwinHomeState(),
        loadTwinRelations(),
      ]);
      if (!state) return jsonOk({ rooms: [], connections: [] });
      const rooms = twinHomeStateToRoomLayouts(state);
      return jsonOk({ rooms, connections: twinConnections(rooms, relations) });
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () =>
    jsonOk({ rooms: mockRoomLayout, connections: mockConnections }),
  );
}
