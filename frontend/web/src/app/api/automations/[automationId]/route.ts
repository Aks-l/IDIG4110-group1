import type { NextRequest } from 'next/server';

import {
  RULES_PATH,
  backendErrorResponse,
  backendServes,
  fetchBackendJson,
  ruleToAutomation,
} from '@/lib/api/backend';
import type { BackendRule } from '@/lib/api/backend';
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

  // Hybrid: forward the toggle to the rules-engine when configured.
  if (backendServes('/automations')) {
    const body = await readJsonObject(request);
    if (typeof body.enabled !== 'boolean') {
      return jsonError(400, 'body must be {"enabled": true|false}');
    }
    try {
      const rule = await fetchBackendJson<BackendRule>(
        `${RULES_PATH}/${automationId}`,
        {
          method: 'PATCH',
          body: JSON.stringify({ enabled: body.enabled }),
        },
      );
      return jsonOk(ruleToAutomation(rule));
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

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
