import type { NextRequest } from 'next/server';

import { mockRooms } from '@/lib/api/mock-data';
import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';

// Placeholder route implementing the contract in docs/frontend/api.md.
// Serves fixture data, or forwards to PLACEHOLDER_API_TARGET when set.
export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  return handlePlaceholderRequest(request, () => jsonOk(Object.values(mockRooms)));
}
