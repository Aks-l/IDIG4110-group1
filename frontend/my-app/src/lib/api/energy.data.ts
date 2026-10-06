import { getEnergy } from './energy';
import { dataSource } from './data-source';
import type { EnergyData, EnergyRange } from './types';

export const mockEnergy: Record<EnergyRange, EnergyData> = {
  Today: { labels: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', 'Now'], values: [0.42, 0.28, 0.75, 1.16, 1.42, 1.08, 1.2], total: '8.4 kWh' },
  Week: { labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'], values: [11.2, 13.8, 12.6, 15.1, 14.3, 9.6, 8.4], total: '85.0 kWh' },
  Month: { labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4'], values: [82, 91, 76, 85], total: '334 kWh' },
  Year: { labels: ['Jan', 'Mar', 'May', 'Jul', 'Sep', 'Nov'], values: [310, 284, 342, 296, 334, 318], total: '3,684 kWh' },
};

export function loadEnergy(range: EnergyRange): Promise<EnergyData> {
  if (dataSource === 'mock') return Promise.resolve(mockEnergy[range]);
  return getEnergy(range);
}