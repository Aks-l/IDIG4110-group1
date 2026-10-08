import type { NextRequest } from 'next/server';

import { mockProjections } from '@/lib/api/mock-data';
import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';

export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  return handlePlaceholderRequest(request, () => jsonOk(mockProjections));
}
