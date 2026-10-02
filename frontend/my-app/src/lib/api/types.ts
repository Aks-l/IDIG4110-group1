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
  location: string;
  severity: Severity;
};