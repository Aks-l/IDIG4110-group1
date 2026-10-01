import { getAutomations } from './api-automations';
import { dataSource } from './data-source';
import { mockAutomations } from './mock-data';
import type { Automation } from './types';

export function loadAutomations(): Promise<Automation[]> {
  if (dataSource === 'mock') return Promise.resolve(mockAutomations);
  return getAutomations();
}