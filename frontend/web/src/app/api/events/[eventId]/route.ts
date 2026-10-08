import type { NextRequest } from 'next/server';

import { mockEvents } from '@/lib/api/mock-data';
import {
  handlePlaceholderRequest,
  jsonError,
  jsonOk,
} from '@/lib/api/placeholder-handler';

export const dynamic = 'force-dynamic';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ eventId: string }> },
) {
  const { eventId } = await params;
  return handlePlaceholderRequest(request, () => {
    const event = mockEvents.find((candidate) => candidate.id === eventId);
    return event ? jsonOk(event) : jsonError(404, `Unknown event: ${eventId}`);
  });
}
