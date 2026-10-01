export type Automation = {
  id: string;
  name: string;
  description: string;
  category: 'Comfort' | 'Security' | 'Energy' | 'Notification';
  enabled: boolean;
  lastRun?: string;
  runCount: number;
};

export type Stat = {
  label: string;
  value: string;
  change: string;
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
  stats: Stat[];
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