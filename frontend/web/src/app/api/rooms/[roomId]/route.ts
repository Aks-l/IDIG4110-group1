import type { NextRequest } from 'next/server';

import { mockRooms } from '@/lib/api/mock-data';
import {
  handlePlaceholderRequest,
  jsonError,
  jsonOk,
} from '@/lib/api/placeholder-handler';

export const dynamic = 'force-dynamic';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ roomId: string }> },
) {
  const { roomId } = await params;
  return handlePlaceholderRequest(request, () => {
    const room = mockRooms[roomId];
    return room ? jsonOk(room) : jsonError(404, `Unknown room: ${roomId}`);
  });
}
