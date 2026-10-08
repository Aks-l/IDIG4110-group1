import { getAutomations, createAutomation as apiCreateAutomation, setAutomationState as apiSetAutomationState } from './automations';
import { dataSource } from './data-source';
import { mockAutomations } from './mock-data';
import type { Automation, AutomationDraft } from './types';

export function loadAutomations(): Promise<Automation[]> {
  if (dataSource === 'mock') return Promise.resolve(mockAutomations);
  return getAutomations();
}

export function createAutomation(draft: AutomationDraft): Promise<Automation> {
  if (dataSource === 'mock') {
    // Fabricate an Automation from the draft, keeping its conditions and
    // actions so the automations page can render the summary.
    return Promise.resolve({
      ...draft,
      id: `mock-${Date.now()}`,
      enabled: true,
      runCount: 0,
    });
  }
  return apiCreateAutomation(draft);
}

export function setAutomationState(automationId: string, enabled: boolean): Promise<Automation> {
  if (dataSource === 'api') return apiSetAutomationState(automationId, enabled);

  const automation = mockAutomations.find((candidate) => candidate.id === automationId);
  return Promise.resolve({
    ...(automation ?? {
      id: automationId,
      name: 'Automation',
      description: '',
      category: 'Comfort' as const,
      conditions: [],
      actions: [],
      runCount: 0,
    }),
    enabled,
  });
}