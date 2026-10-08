import type { NextRequest } from 'next/server';

import {
  handlePlaceholderRequest,
  jsonError,
  jsonOk,
  readJsonObject,
} from '@/lib/api/placeholder-handler';
import {
  createPlaceholderAutomation,
  listPlaceholderAutomations,
} from '@/lib/api/placeholder-store';
import type {
  AutomationAction,
  AutomationCategory,
  AutomationCondition,
} from '@/lib/api/types';

export const dynamic = 'force-dynamic';

const categories = ['Comfort', 'Security', 'Energy', 'Notification'] as const;

export async function GET(request: NextRequest) {
  return handlePlaceholderRequest(request, () =>
    jsonOk(listPlaceholderAutomations()),
  );
}

export async function POST(request: NextRequest) {
  return handlePlaceholderRequest(request, async () => {
    const body = await readJsonObject(request);

    const name = typeof body.name === 'string' ? body.name.trim() : '';
    if (name === '') {
      return jsonError(400, 'body must include a non-empty "name"');
    }

    // The placeholder trusts the client's condition and action shapes.
    const automation = createPlaceholderAutomation({
      name,
      description: typeof body.description === 'string' ? body.description : '',
      category: (categories as readonly string[]).includes(body.category as string)
        ? (body.category as AutomationCategory)
        : 'Comfort',
      conditions: (Array.isArray(body.conditions)
        ? body.conditions
        : []) as AutomationCondition[],
      actions: (Array.isArray(body.actions)
        ? body.actions
        : []) as AutomationAction[],
    });

    return jsonOk(automation, 201);
  });
}
