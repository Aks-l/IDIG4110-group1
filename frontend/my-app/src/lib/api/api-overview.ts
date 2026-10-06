import { apiClient } from './api-client';
import type { OverviewData } from './types';

export function getOverview() {
  return apiClient<OverviewData>('/overview');
}