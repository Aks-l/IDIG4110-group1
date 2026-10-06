import { apiClient } from './client';
import type { EnergyData, EnergyRange } from './types';

export function getEnergy(range: EnergyRange) {
  return apiClient<EnergyData>(`/energy?range=${range.toLowerCase()}`);
}