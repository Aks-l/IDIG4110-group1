import { apiClient } from './api-client';
import type { Automation } from './types';

export function getAutomations() {
  return apiClient<Automation[]>('/automations');
}

export function setAutomationState(automationId: string, enabled: boolean) {
  return apiClient<Automation>(`/automations/${automationId}`, {
    method: 'PATCH',
    body: JSON.stringify({ enabled }),
  });
}