import type { NextRequest } from 'next/server';

import { handlePlaceholderRequest, jsonOk } from '@/lib/api/placeholder-handler';
import { listPlaceholderDevices } from '@/lib/api/placeholder-store';

export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  return handlePlaceholderRequest(request, () => jsonOk(listPlaceholderDevices()));
}
