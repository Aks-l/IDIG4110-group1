export type Automation = {
  id: string;
  name: string;
  description: string;
  category: 'Comfort' | 'Security' | 'Energy' | 'Notification';
  enabled: boolean;
  lastRun?: string;
  runCount: number;
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
  location: string;
  severity: Severity;
};