import type { NextRequest } from 'next/server';

import {
  RULES_PATH,
  backendErrorResponse,
  backendServes,
  draftToRuleInput,
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
import {
  createPlaceholderAutomation,
  listPlaceholderAutomations,
} from '@/lib/api/placeholder-store';
import type {
  AutomationAction,
  AutomationCategory,
  AutomationCondition,
  AutomationDraft,
} from '@/lib/api/types';

export const dynamic = 'force-dynamic';

const categories = ['Comfort', 'Security', 'Energy', 'Notification'] as const;

export async function GET(request: NextRequest) {
  // Hybrid: serve /automations from the rules-engine when configured.
  if (backendServes('/automations')) {
    try {
      const rules = await fetchBackendJson<BackendRule[]>(RULES_PATH);
      return jsonOk(rules.map(ruleToAutomation));
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () =>
    jsonOk(listPlaceholderAutomations()),
  );
}

export async function POST(request: NextRequest) {
  if (backendServes('/automations')) {
    const draft = await readAutomationDraft(request);
    if ('error' in draft) return jsonError(400, draft.error);

    const mapped = draftToRuleInput(draft.draft);
    if ('error' in mapped) return jsonError(422, mapped.error);

    try {
      const rule = await fetchBackendJson<BackendRule>(RULES_PATH, {
        method: 'POST',
        body: JSON.stringify(mapped.input),
      });
      return jsonOk(ruleToAutomation(rule), 201);
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, async () => {
    const draft = await readAutomationDraft(request);
    if ('error' in draft) return jsonError(400, draft.error);
    return jsonOk(createPlaceholderAutomation(draft.draft), 201);
  });
}

// The placeholder trusts the client's condition and action shapes.
async function readAutomationDraft(
  request: NextRequest,
): Promise<{ draft: AutomationDraft } | { error: string }> {
  const body = await readJsonObject(request);

  const name = typeof body.name === 'string' ? body.name.trim() : '';
  if (name === '') return { error: 'body must include a non-empty "name"' };

  return {
    draft: {
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
    },
  };
}
