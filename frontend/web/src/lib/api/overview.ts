import { apiClient } from './client';
import type { OverviewData } from './types';

export function getOverview() {
  return apiClient<OverviewData>('/overview');
}