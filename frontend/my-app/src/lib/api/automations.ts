import { apiClient } from './client';
import type { Automation , AutomationDraft} from './types';

export function getAutomations() {
  return apiClient<Automation[]>('/automations');
}

export function setAutomationState(automationId: string, enabled: boolean) {
  return apiClient<Automation>(`/automations/${automationId}`, {
    method: 'PATCH',
    body: JSON.stringify({ enabled }),
  });
}

export async function createAutomation(draft: AutomationDraft): Promise<Automation> {
  return apiClient<Automation>('/automations', {
    method: 'POST',
    body: JSON.stringify(draft),
  });
}

export async function setAutomationEnabled(id: string, enabled: boolean): Promise<Automation> {
  return apiClient<Automation>(`/automations/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ enabled }),
  });
}