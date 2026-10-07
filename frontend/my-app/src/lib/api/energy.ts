import { apiClient } from './client';
import type { EnergyPageData, EnergyPeriod, EnergyMode } from './types';

export function getEnergy(period: EnergyPeriod, mode: EnergyMode) {
  return apiClient<EnergyPageData>(`/energy?period=${period}&mode=${mode}`);
}