export type Automation = {
  id: string;
  name: string;
  description: string;
  category: 'Comfort' | 'Security' | 'Energy' | 'Notification';
  enabled: boolean;
  lastRun?: string;
  runCount: number;
};

export type AutomationDraft = {
  name: string;
  description: string;
  category: Automation['category'];
  triggerType: string;
  triggerDetail: string;
  actionCommand: string;
  actionTarget: string;
};

export type OverviewData = {
  stats: { label: string; value: string }[];
  devices: { connected: number; warning: number; events: number };
};

// ENERGY TYPES

export const ENERGY_PERIODS = ['today', 'week', 'month', 'year'] as const;
export type EnergyPeriod = (typeof ENERGY_PERIODS)[number];

export const ENERGY_MODES = ['current', 'projected'] as const;
export type EnergyMode = (typeof ENERGY_MODES)[number];

export type EnergyGranularity = 'hour' | 'day' | 'month';

export type EnergySummary = {
  period: EnergyPeriod;
  totalKwh: number;              // total consumption in the period
  estimatedCost: number;         // computed from totalKwh × tariff
  currency: string;              // 'NOK'
  comparedToPrevious: number | null;    // -0.12 = 12% less than previous period
};

export type EnergyPoint = {
  timestamp: string;   // ISO 8601, start of the bucket
  kwh: number;         // consumption for this bucket
  cost: number;        // estimated cost for this bucket
};

export type EnergySeries = {
  period: EnergyPeriod;
  granularity: EnergyGranularity;
  points: EnergyPoint[];
};

export type RoomEnergy = {
  roomId: string;
  roomName: string;
  kwh: number;            // consumption in selected period
  percentage: number;     // 0..1, share of total (for bar widths)
  estimatedCost: number;  // cost for this room
};

export type EnergyPageData = {
  summary: EnergySummary;
  series: EnergySeries;
  byRoom: RoomEnergy[];
};

export type RoomMetric = {
  value: number;
  change: number;
};

export type RoomMetrics = {
  temperature: RoomMetric;
  humidity: RoomMetric;
  co2: RoomMetric;
  occupancy: RoomMetric;
};

export type Device = {
  id: string;
  name: string;
  type: string;
  on: boolean;
  roomId?: string;
  position?: [number, number, number];
};

export type RoomLayout = {
  roomId: string;
  position: [number, number, number];
  size: [number, number, number];
};

export type Activity = {
  id: string;
  timestamp: string;
  description: string;
};

export type RoomData = {
  id: string;
  name: string;
  metrics: RoomMetrics;
  devices: Device[];
  activity: Activity[];
};

export type Severity = 'info' | 'warning' | 'critical';

export type Event = {
  id: string;
  timestamp: string;
  title: string;
  location?: string;
  severity: Severity;
};