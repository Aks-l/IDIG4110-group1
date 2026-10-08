import type { NextRequest } from 'next/server';

import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';
import { mockRoomLayout } from '@/lib/api/three-d.data';

export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  return handlePlaceholderRequest(request, () => jsonOk(mockRoomLayout));
}
