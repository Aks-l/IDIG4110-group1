import { apiClient } from './client';
import type { EnergyPageData, EnergyPeriod } from './types';

export function getEnergy(period: EnergyPeriod) {
  return apiClient<EnergyPageData>(`/energy?period=${period}`);
}