import { dataSource } from './data-source';
import type { EnergyPageData, EnergyPeriod } from '@/lib/api/types';
import { getEnergy } from './energy';

export function loadEnergy(period: EnergyPeriod): Promise<EnergyPageData> {
  if (dataSource === 'mock') return Promise.resolve(mockEnergyByPeriod[period]);
  return getEnergy(period);
}

const CURRENCY = 'NOK';
const TARIFF = 1.4; // NOK per kWh — used to derive cost values

// Helper so cost always matches kwh (avoids typos)
const point = (timestamp: string, kwh: number) => ({
  timestamp,
  kwh,
  cost: Number((kwh * TARIFF).toFixed(2)),
});

// ---------------------------------------------------------------------------
// TODAY — 24 hourly buckets
// ---------------------------------------------------------------------------

const today: EnergyPageData = {
  summary: {
    period: 'today',
    totalKwh: 27.61,
    estimatedCost: 38.65,
    currency: CURRENCY,
    comparedToPrevious: -0.08, // 8% less than yesterday
  },
  series: {
    period: 'today',
    granularity: 'hour',
    points: [
      point('2026-10-07T00:00:00Z', 0.42),
      point('2026-10-07T01:00:00Z', 0.38),
      point('2026-10-07T02:00:00Z', 0.35),
      point('2026-10-07T03:00:00Z', 0.33),
      point('2026-10-07T04:00:00Z', 0.36),
      point('2026-10-07T05:00:00Z', 0.52),
      point('2026-10-07T06:00:00Z', 1.18),
      point('2026-10-07T07:00:00Z', 2.34),
      point('2026-10-07T08:00:00Z', 2.08),
      point('2026-10-07T09:00:00Z', 1.21),
      point('2026-10-07T10:00:00Z', 0.94),
      point('2026-10-07T11:00:00Z', 0.86),
      point('2026-10-07T12:00:00Z', 0.91),
      point('2026-10-07T13:00:00Z', 0.88),
      point('2026-10-07T14:00:00Z', 0.95),
      point('2026-10-07T15:00:00Z', 1.14),
      point('2026-10-07T16:00:00Z', 1.62),
      point('2026-10-07T17:00:00Z', 2.41),
      point('2026-10-07T18:00:00Z', 2.72),
      point('2026-10-07T19:00:00Z', 2.18),
      point('2026-10-07T20:00:00Z', 1.65),
      point('2026-10-07T21:00:00Z', 1.02),
      point('2026-10-07T22:00:00Z', 0.68),
      point('2026-10-07T23:00:00Z', 0.48),
    ],
  },
  byRoom: [
    { roomId: 'kitchen',     roomName: 'Kitchen',     kwh: 8.28, percentage: 0.30, estimatedCost: 11.59 },
    { roomId: 'living-room', roomName: 'Living Room', kwh: 6.90, percentage: 0.25, estimatedCost: 9.66  },
    { roomId: 'bathroom',    roomName: 'Bathroom',    kwh: 4.42, percentage: 0.16, estimatedCost: 6.19  },
    { roomId: 'bedroom',     roomName: 'Bedroom',     kwh: 3.86, percentage: 0.14, estimatedCost: 5.41  },
    { roomId: 'hallway',     roomName: 'Hallway',     kwh: 2.21, percentage: 0.08, estimatedCost: 3.09  },
    { roomId: 'office',      roomName: 'Office',      kwh: 1.93, percentage: 0.07, estimatedCost: 2.70  },
  ],
};

// ---------------------------------------------------------------------------
// WEEK — 7 daily buckets (Wed → Tue, ending today)
// ---------------------------------------------------------------------------

const week: EnergyPageData = {
  summary: {
    period: 'week',
    totalKwh: 192.4,
    estimatedCost: 269.36,
    currency: CURRENCY,
    comparedToPrevious: 0.05, // 5% more than last week
  },
  series: {
    period: 'week',
    granularity: 'day',
    points: [
      point('2026-10-01T00:00:00Z', 24.2), // Wed
      point('2026-10-02T00:00:00Z', 26.8), // Thu
      point('2026-10-03T00:00:00Z', 25.4), // Fri
      point('2026-10-04T00:00:00Z', 27.1), // Sat
      point('2026-10-05T00:00:00Z', 26.3), // Sun
      point('2026-10-06T00:00:00Z', 30.8), // Mon
      point('2026-10-07T00:00:00Z', 31.8), // Tue (today)
    ],
  },
  byRoom: [
    { roomId: 'kitchen',     roomName: 'Kitchen',     kwh: 57.72, percentage: 0.30, estimatedCost: 80.81 },
    { roomId: 'living-room', roomName: 'Living Room', kwh: 48.10, percentage: 0.25, estimatedCost: 67.34 },
    { roomId: 'bathroom',    roomName: 'Bathroom',    kwh: 30.78, percentage: 0.16, estimatedCost: 43.09 },
    { roomId: 'bedroom',     roomName: 'Bedroom',     kwh: 26.94, percentage: 0.14, estimatedCost: 37.72 },
    { roomId: 'hallway',     roomName: 'Hallway',     kwh: 15.39, percentage: 0.08, estimatedCost: 21.55 },
    { roomId: 'office',      roomName: 'Office',      kwh: 13.47, percentage: 0.07, estimatedCost: 18.86 },
  ],
};

// ---------------------------------------------------------------------------
// MONTH — 30 daily buckets (Sep 8 → Oct 7)
// ---------------------------------------------------------------------------

const month: EnergyPageData = {
  summary: {
    period: 'month',
    totalKwh: 766.3,
    estimatedCost: 1072.82,
    currency: CURRENCY,
    comparedToPrevious: -0.12, // 12% less than last month
  },
  series: {
    period: 'month',
    granularity: 'day',
    points: [
      point('2026-09-08T00:00:00Z', 21.4),
      point('2026-09-09T00:00:00Z', 22.1),
      point('2026-09-10T00:00:00Z', 20.8),
      point('2026-09-11T00:00:00Z', 23.4),
      point('2026-09-12T00:00:00Z', 24.1),
      point('2026-09-13T00:00:00Z', 28.2),
      point('2026-09-14T00:00:00Z', 27.9),
      point('2026-09-15T00:00:00Z', 22.6),
      point('2026-09-16T00:00:00Z', 21.8),
      point('2026-09-17T00:00:00Z', 22.3),
      point('2026-09-18T00:00:00Z', 23.1),
      point('2026-09-19T00:00:00Z', 24.6),
      point('2026-09-20T00:00:00Z', 29.4),
      point('2026-09-21T00:00:00Z', 28.7),
      point('2026-09-22T00:00:00Z', 22.9),
      point('2026-09-23T00:00:00Z', 23.5),
      point('2026-09-24T00:00:00Z', 22.7),
      point('2026-09-25T00:00:00Z', 24.8),
      point('2026-09-26T00:00:00Z', 25.2),
      point('2026-09-27T00:00:00Z', 30.1),
      point('2026-09-28T00:00:00Z', 29.6),
      point('2026-09-29T00:00:00Z', 24.1),
      point('2026-09-30T00:00:00Z', 25.3),
      point('2026-10-01T00:00:00Z', 26.2),
      point('2026-10-02T00:00:00Z', 27.4),
      point('2026-10-03T00:00:00Z', 25.1),
      point('2026-10-04T00:00:00Z', 31.2),
      point('2026-10-05T00:00:00Z', 32.4),
      point('2026-10-06T00:00:00Z', 27.8),
      point('2026-10-07T00:00:00Z', 27.6),
    ],
  },
  byRoom: [
    { roomId: 'kitchen',     roomName: 'Kitchen',     kwh: 229.89, percentage: 0.30, estimatedCost: 321.85 },
    { roomId: 'living-room', roomName: 'Living Room', kwh: 191.58, percentage: 0.25, estimatedCost: 268.21 },
    { roomId: 'bathroom',    roomName: 'Bathroom',    kwh: 122.61, percentage: 0.16, estimatedCost: 171.65 },
    { roomId: 'bedroom',     roomName: 'Bedroom',     kwh: 107.28, percentage: 0.14, estimatedCost: 150.19 },
    { roomId: 'hallway',     roomName: 'Hallway',     kwh: 61.30,  percentage: 0.08, estimatedCost: 85.82  },
    { roomId: 'office',      roomName: 'Office',      kwh: 53.64,  percentage: 0.07, estimatedCost: 75.10  },
  ],
};

// ---------------------------------------------------------------------------
// YEAR — 12 monthly buckets (Nov 2025 → Oct 2026)
// ---------------------------------------------------------------------------

const year: EnergyPageData = {
  summary: {
    period: 'year',
    totalKwh: 13151.7,
    estimatedCost: 18412.38,
    currency: CURRENCY,
    comparedToPrevious: 0.03, // 3% more than previous year
  },
  series: {
    period: 'year',
    granularity: 'month',
    points: [
      point('2025-11-01T00:00:00Z', 1580.4),
      point('2025-12-01T00:00:00Z', 1680.6),
      point('2026-01-01T00:00:00Z', 1780.2),
      point('2026-02-01T00:00:00Z', 1590.8),
      point('2026-03-01T00:00:00Z', 1380.4),
      point('2026-04-01T00:00:00Z', 1080.6),
      point('2026-05-01T00:00:00Z',  760.4),
      point('2026-06-01T00:00:00Z',  540.8),
      point('2026-07-01T00:00:00Z',  500.2),
      point('2026-08-01T00:00:00Z',  570.4),
      point('2026-09-01T00:00:00Z',  920.6),
      point('2026-10-01T00:00:00Z',  766.3),
    ],
  },
  byRoom: [
    { roomId: 'kitchen',     roomName: 'Kitchen',     kwh: 3945.51, percentage: 0.30, estimatedCost: 5523.71 },
    { roomId: 'living-room', roomName: 'Living Room', kwh: 3287.93, percentage: 0.25, estimatedCost: 4603.10 },
    { roomId: 'bathroom',    roomName: 'Bathroom',    kwh: 2104.27, percentage: 0.16, estimatedCost: 2945.98 },
    { roomId: 'bedroom',     roomName: 'Bedroom',     kwh: 1841.24, percentage: 0.14, estimatedCost: 2577.74 },
    { roomId: 'hallway',     roomName: 'Hallway',     kwh: 1052.14, percentage: 0.08, estimatedCost: 1473.00 },
    { roomId: 'office',      roomName: 'Office',      kwh: 920.62,  percentage: 0.07, estimatedCost: 1288.87 },
  ],
};

// ---------------------------------------------------------------------------

export const mockEnergyByPeriod: Record<EnergyPeriod, EnergyPageData> = {
  today,
  week,
  month,
  year,
};