import { getOverview } from './overview';
import { dataSource } from './data-source';
import type { OverviewData } from './types';

export const mockOverview: OverviewData = {
  stats: [
    { label: 'Temperature', value: '23°C' },
    { label: 'Humidity', value: '48%' },
    { label: 'Energy', value: '1.2 kW' },
    { label: 'Air Quality', value: 'Good' },
    { label: 'Occupancy', value: '3 people' },
  ],
  devices: { connected: 24, warning: 3, events: 7 },
};

export function loadOverview(): Promise<OverviewData> {
  if (dataSource === 'mock') return Promise.resolve(mockOverview);
  return getOverview();
}