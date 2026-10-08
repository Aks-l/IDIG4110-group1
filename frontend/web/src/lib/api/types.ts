export type DayOfWeek = 'mon' | 'tue' | 'wed' | 'thu' | 'fri' | 'sat' | 'sun';

export type TemperatureOp = 'lt' | 'gt' | 'eq';

// A condition that must hold for the automation to fire.
export type AutomationCondition =
  | { kind: 'time';        from: string; to: string }            // "Between 18:00 and 06:00"
  | { kind: 'day';         days: DayOfWeek[] }                    // "On Sat, Sun"
  | { kind: 'temperature'; op: TemperatureOp; value: number }     // "Temperature below 15°C"
  | { kind: 'motion';      location: string }                     // "Motion in Hallway"
  | { kind: 'presence';    state: 'home' | 'away' };              // "Nobody is home"

export type ConditionKind = AutomationCondition['kind'];

// An action to perform when all conditions are met.
export type AutomationAction =
  | { kind: 'device'; deviceName: string; command: 'turn_on' | 'turn_off' | 'set'; value?: string }
  | { kind: 'notify'; message: string }
  | { kind: 'delay';  minutes: number };

export type ActionKind = AutomationAction['kind'];

export type AutomationCategory = 'Comfort' | 'Security' | 'Energy' | 'Notification';

export type AutomationDraft = {
  name: string;
  description: string;
  category: AutomationCategory;
  conditions: AutomationCondition[];
  actions: AutomationAction[];
};

export type Automation = AutomationDraft & {
  id: string;
  enabled: boolean;
  lastRun?: string;
  runCount: number;
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

  // Detail fields (optional — not every device has all of them)
  powerWatts?: number;     // current draw
  signal?: number;         // 0..100 (wifi/zigbee strength)
  battery?: number;        // 0..100 (only battery-powered devices)
  firmware?: string;       // "1.4.2"
  lastSeen?: string;       // ISO
  installedAt?: string;    // ISO
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


// EVENTS & PROJECTIONS
export type Severity = 'info' | 'warning' | 'critical';

export type Event = {
  id: string;
  timestamp: string;
  title: string;
  location?: string;
  severity: Severity;
};

export type ProjectionCategory = 'maintenance' | 'safety' | 'security' | 'energy';
export type ProjectionSeverity = 'low' | 'medium' | 'high' | 'critical';

export type Projection = {
  id: string;
  timestamp: string;            // when the twin generated the prediction
  title: string;
  description: string;
  location?: string;
  category: ProjectionCategory;
  severity: ProjectionSeverity;
  confidence: number;           // 0..1
  horizon: string;              // "5–10 days", "tonight"
  recommendedAction?: string;
};

// A unified item for the Events page — either a past event or a future prediction.
export type ActivityItem =
  | ({ kind: 'event' } & Event)
  | ({ kind: 'projection' } & Projection);