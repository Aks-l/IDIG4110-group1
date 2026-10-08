import type { NextRequest } from 'next/server';

import { backendErrorResponse } from '@/lib/api/backend';
import { mockRooms } from '@/lib/api/mock-data';
import {
  handlePlaceholderRequest,
  jsonError,
  jsonOk,
} from '@/lib/api/placeholder-handler';
import {
  loadTwinHomeState,
  twinHomeStateToRoom,
  twinServes,
} from '@/lib/api/twin';

export const dynamic = 'force-dynamic';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ roomId: string }> },
) {
  const { roomId } = await params;

  // Hybrid: serve /rooms/:id from the twin-core scaffold when configured.
  if (twinServes('/rooms')) {
    try {
      const state = await loadTwinHomeState();
      const room = state ? twinHomeStateToRoom(state, roomId) : null;
      return room ? jsonOk(room) : jsonError(404, `Unknown room: ${roomId}`);
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () => {
    const room = mockRooms[roomId];
    return room ? jsonOk(room) : jsonError(404, `Unknown room: ${roomId}`);
  });
}
