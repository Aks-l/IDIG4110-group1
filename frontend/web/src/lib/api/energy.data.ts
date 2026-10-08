import { dataSource } from './data-source';
import type { EnergyMode, EnergyPageData, EnergyPeriod } from '@/lib/api/types';
import { getEnergy } from './energy';

const CURRENCY = 'NOK';
const TARIFF = 1.4;

const point = (timestamp: string, kwh: number) => ({
  timestamp,
  kwh,
  cost: Number((kwh * TARIFF).toFixed(2)),
});

// =============================================================================
// ACTUAL DATA — the home's real consumption for the current period
// =============================================================================

// ---------------------------------------------------------------------------
// TODAY — 24 hourly buckets
// ---------------------------------------------------------------------------

const today: EnergyPageData = {
  summary: {
    period: 'today',
    totalKwh: 27.61,
    estimatedCost: 38.65,
    currency: CURRENCY,
    comparedToPrevious: -0.08,
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
    comparedToPrevious: 0.05,
  },
  series: {
    period: 'week',
    granularity: 'day',
    points: [
      point('2026-10-01T00:00:00Z', 24.2),
      point('2026-10-02T00:00:00Z', 26.8),
      point('2026-10-03T00:00:00Z', 25.4),
      point('2026-10-04T00:00:00Z', 27.1),
      point('2026-10-05T00:00:00Z', 26.3),
      point('2026-10-06T00:00:00Z', 30.8),
      point('2026-10-07T00:00:00Z', 31.8),
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
    comparedToPrevious: -0.12,
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
    comparedToPrevious: 0.03,
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

// =============================================================================
// PROJECTED DATA — forecasts for the upcoming period
// =============================================================================

// ---------------------------------------------------------------------------
// TOMORROW — 24 hourly buckets (+14% vs today: colder forecast)
// ---------------------------------------------------------------------------

const tomorrow: EnergyPageData = {
  summary: {
    period: 'today',
    totalKwh: 31.48,
    estimatedCost: 44.07,
    currency: CURRENCY,
    comparedToPrevious: 0.14,
  },
  series: {
    period: 'today',
    granularity: 'hour',
    points: [
      point('2026-10-08T00:00:00Z', 0.52),
      point('2026-10-08T01:00:00Z', 0.48),
      point('2026-10-08T02:00:00Z', 0.44),
      point('2026-10-08T03:00:00Z', 0.42),
      point('2026-10-08T04:00:00Z', 0.46),
      point('2026-10-08T05:00:00Z', 0.68),
      point('2026-10-08T06:00:00Z', 1.42),
      point('2026-10-08T07:00:00Z', 2.68),
      point('2026-10-08T08:00:00Z', 2.32),
      point('2026-10-08T09:00:00Z', 1.38),
      point('2026-10-08T10:00:00Z', 1.08),
      point('2026-10-08T11:00:00Z', 0.98),
      point('2026-10-08T12:00:00Z', 1.02),
      point('2026-10-08T13:00:00Z', 1.00),
      point('2026-10-08T14:00:00Z', 1.08),
      point('2026-10-08T15:00:00Z', 1.28),
      point('2026-10-08T16:00:00Z', 1.82),
      point('2026-10-08T17:00:00Z', 2.68),
      point('2026-10-08T18:00:00Z', 3.02),
      point('2026-10-08T19:00:00Z', 2.42),
      point('2026-10-08T20:00:00Z', 1.82),
      point('2026-10-08T21:00:00Z', 1.12),
      point('2026-10-08T22:00:00Z', 0.78),
      point('2026-10-08T23:00:00Z', 0.58),
    ],
  },
  byRoom: [
    { roomId: 'kitchen',     roomName: 'Kitchen',     kwh: 9.44, percentage: 0.30, estimatedCost: 13.22 },
    { roomId: 'living-room', roomName: 'Living Room', kwh: 7.87, percentage: 0.25, estimatedCost: 11.02 },
    { roomId: 'bathroom',    roomName: 'Bathroom',    kwh: 5.04, percentage: 0.16, estimatedCost: 7.05  },
    { roomId: 'bedroom',     roomName: 'Bedroom',     kwh: 4.41, percentage: 0.14, estimatedCost: 6.17  },
    { roomId: 'hallway',     roomName: 'Hallway',     kwh: 2.52, percentage: 0.08, estimatedCost: 3.53  },
    { roomId: 'office',      roomName: 'Office',      kwh: 2.20, percentage: 0.07, estimatedCost: 3.08  },
  ],
};

// ---------------------------------------------------------------------------
// NEXT WEEK — 7 daily buckets (+17% vs this week)
// ---------------------------------------------------------------------------

const nextWeek: EnergyPageData = {
  summary: {
    period: 'week',
    totalKwh: 225.5,
    estimatedCost: 315.70,
    currency: CURRENCY,
    comparedToPrevious: 0.17,
  },
  series: {
    period: 'week',
    granularity: 'day',
    points: [
      point('2026-10-08T00:00:00Z', 29.4),
      point('2026-10-09T00:00:00Z', 31.2),
      point('2026-10-10T00:00:00Z', 30.8),
      point('2026-10-11T00:00:00Z', 34.6),
      point('2026-10-12T00:00:00Z', 35.2),
      point('2026-10-13T00:00:00Z', 33.4),
      point('2026-10-14T00:00:00Z', 30.9),
    ],
  },
  byRoom: [
    { roomId: 'kitchen',     roomName: 'Kitchen',     kwh: 67.65, percentage: 0.30, estimatedCost: 94.71 },
    { roomId: 'living-room', roomName: 'Living Room', kwh: 56.38, percentage: 0.25, estimatedCost: 78.93 },
    { roomId: 'bathroom',    roomName: 'Bathroom',    kwh: 36.08, percentage: 0.16, estimatedCost: 50.51 },
    { roomId: 'bedroom',     roomName: 'Bedroom',     kwh: 31.57, percentage: 0.14, estimatedCost: 44.20 },
    { roomId: 'hallway',     roomName: 'Hallway',     kwh: 18.04, percentage: 0.08, estimatedCost: 25.26 },
    { roomId: 'office',      roomName: 'Office',      kwh: 15.79, percentage: 0.07, estimatedCost: 22.11 },
  ],
};

// ---------------------------------------------------------------------------
// NEXT MONTH — 30 daily buckets (+24% vs this month, winter approaching)
// ---------------------------------------------------------------------------

const nextMonth: EnergyPageData = {
  summary: {
    period: 'month',
    totalKwh: 954.0,
    estimatedCost: 1335.60,
    currency: CURRENCY,
    comparedToPrevious: 0.24,
  },
  series: {
    period: 'month',
    granularity: 'day',
    points: [
      point('2026-10-08T00:00:00Z', 28.5),
      point('2026-10-09T00:00:00Z', 29.7),
      point('2026-10-10T00:00:00Z', 27.8),
      point('2026-10-11T00:00:00Z', 30.1),
      point('2026-10-12T00:00:00Z', 31.3),
      point('2026-10-13T00:00:00Z', 33.6),
      point('2026-10-14T00:00:00Z', 32.9),
      point('2026-10-15T00:00:00Z', 29.2),
      point('2026-10-16T00:00:00Z', 28.9),
      point('2026-10-17T00:00:00Z', 30.4),
      point('2026-10-18T00:00:00Z', 30.9),
      point('2026-10-19T00:00:00Z', 32.4),
      point('2026-10-20T00:00:00Z', 34.7),
      point('2026-10-21T00:00:00Z', 34.1),
      point('2026-10-22T00:00:00Z', 29.8),
      point('2026-10-23T00:00:00Z', 30.4),
      point('2026-10-24T00:00:00Z', 29.7),
      point('2026-10-25T00:00:00Z', 31.5),
      point('2026-10-26T00:00:00Z', 31.8),
      point('2026-10-27T00:00:00Z', 35.3),
      point('2026-10-28T00:00:00Z', 34.8),
      point('2026-10-29T00:00:00Z', 30.0),
      point('2026-10-30T00:00:00Z', 31.1),
      point('2026-10-31T00:00:00Z', 31.9),
      point('2026-11-01T00:00:00Z', 32.9),
      point('2026-11-02T00:00:00Z', 30.9),
      point('2026-11-03T00:00:00Z', 36.3),
      point('2026-11-04T00:00:00Z', 37.3),
      point('2026-11-05T00:00:00Z', 33.3),
      point('2026-11-06T00:00:00Z', 32.5),
    ],
  },
  byRoom: [
    { roomId: 'kitchen',     roomName: 'Kitchen',     kwh: 286.20, percentage: 0.30, estimatedCost: 400.68 },
    { roomId: 'living-room', roomName: 'Living Room', kwh: 238.50, percentage: 0.25, estimatedCost: 333.90 },
    { roomId: 'bathroom',    roomName: 'Bathroom',    kwh: 152.64, percentage: 0.16, estimatedCost: 213.70 },
    { roomId: 'bedroom',     roomName: 'Bedroom',     kwh: 133.56, percentage: 0.14, estimatedCost: 186.98 },
    { roomId: 'hallway',     roomName: 'Hallway',     kwh: 76.32,  percentage: 0.08, estimatedCost: 106.85 },
    { roomId: 'office',      roomName: 'Office',      kwh: 66.78,  percentage: 0.07, estimatedCost: 93.49  },
  ],
};

// ---------------------------------------------------------------------------
// NEXT YEAR — 12 monthly buckets (+3.7% vs this year)
// ---------------------------------------------------------------------------

const nextYear: EnergyPageData = {
  summary: {
    period: 'year',
    totalKwh: 13634.4,
    estimatedCost: 19088.16,
    currency: CURRENCY,
    comparedToPrevious: 0.04,
  },
  series: {
    period: 'year',
    granularity: 'month',
    points: [
      point('2026-11-01T00:00:00Z', 1596.4),
      point('2026-12-01T00:00:00Z', 1691.6),
      point('2027-01-01T00:00:00Z', 1786.2),
      point('2027-02-01T00:00:00Z', 1606.3),
      point('2027-03-01T00:00:00Z', 1406.4),
      point('2027-04-01T00:00:00Z', 1121.6),
      point('2027-05-01T00:00:00Z',  817.4),
      point('2027-06-01T00:00:00Z',  608.8),
      point('2027-07-01T00:00:00Z',  570.2),
      point('2027-08-01T00:00:00Z',  636.9),
      point('2027-09-01T00:00:00Z',  969.6),
      point('2027-10-01T00:00:00Z',  823.0),
    ],
  },
  byRoom: [
    { roomId: 'kitchen',     roomName: 'Kitchen',     kwh: 4090.32, percentage: 0.30, estimatedCost: 5726.45 },
    { roomId: 'living-room', roomName: 'Living Room', kwh: 3408.60, percentage: 0.25, estimatedCost: 4772.04 },
    { roomId: 'bathroom',    roomName: 'Bathroom',    kwh: 2181.50, percentage: 0.16, estimatedCost: 3054.10 },
    { roomId: 'bedroom',     roomName: 'Bedroom',     kwh: 1908.82, percentage: 0.14, estimatedCost: 2672.35 },
    { roomId: 'hallway',     roomName: 'Hallway',     kwh: 1090.75, percentage: 0.08, estimatedCost: 1527.05 },
    { roomId: 'office',      roomName: 'Office',      kwh:  954.41, percentage: 0.07, estimatedCost: 1336.17 },
  ],
};

// =============================================================================
// Lookup maps + public adapter
// =============================================================================

const actualByPeriod: Record<EnergyPeriod, EnergyPageData> = {
  today,
  week,
  month,
  year,
};

const projectedByPeriod: Record<EnergyPeriod, EnergyPageData> = {
  today: tomorrow,
  week: nextWeek,
  month: nextMonth,
  year: nextYear,
};

export function loadEnergy(
  period: EnergyPeriod,
  mode: EnergyMode,
): Promise<EnergyPageData> {
  if (dataSource === 'mock') {
    const source = mode === 'projected' ? projectedByPeriod : actualByPeriod;
    return Promise.resolve(source[period]);
  }
  return getEnergy(period, mode);
}

// Shared with the /api/energy placeholder route (src/app/api/energy/route.ts),
// which serves these fixtures until the real backend provides the endpoint.
export const mockEnergy: Record<EnergyMode, Record<EnergyPeriod, EnergyPageData>> = {
  current: actualByPeriod,
  projected: projectedByPeriod,
};