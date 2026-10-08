import type { NextRequest } from 'next/server';

import {
  handlePlaceholderRequest,
  jsonError,
  jsonOk,
  readJsonObject,
} from '@/lib/api/placeholder-handler';
import { setPlaceholderAutomationEnabled } from '@/lib/api/placeholder-store';

export const dynamic = 'force-dynamic';

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ automationId: string }> },
) {
  const { automationId } = await params;
  return handlePlaceholderRequest(request, async () => {
    const body = await readJsonObject(request);
    if (typeof body.enabled !== 'boolean') {
      return jsonError(400, 'body must be {"enabled": true|false}');
    }
    const automation = setPlaceholderAutomationEnabled(automationId, body.enabled);
    return automation
      ? jsonOk(automation)
      : jsonError(404, `Unknown automation: ${automationId}`);
  });
}
