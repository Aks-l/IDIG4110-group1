import type { Automation, Event, RoomData, Projection } from './types';

export const mockAutomations: Automation[] = [
  {
    id: '1',
    name: 'Hallway motion light',
    description: 'Turn on the hallway light when someone walks in at night',
    category: 'Comfort',
    enabled: true,
    lastRun: '2026-10-07T22:30:00Z',
    runCount: 128,
    conditions: [
      { kind: 'motion', location: 'Hallway' },
      { kind: 'time', from: '18:00', to: '06:00' },
    ],
    actions: [
      { kind: 'device', deviceName: 'Hallway Light', command: 'set', value: '40%' },
    ],
  },
  {
    id: '2',
    name: 'Good Morning',
    description: 'Wake the house up on weekday mornings',
    category: 'Comfort',
    enabled: true,
    lastRun: '2026-10-07T07:00:00Z',
    runCount: 62,
    conditions: [
      { kind: 'time', from: '07:00', to: '07:15' },
      { kind: 'day', days: ['mon', 'tue', 'wed', 'thu', 'fri'] },
      { kind: 'presence', state: 'home' },
    ],
    actions: [
      { kind: 'device', deviceName: 'Living Room Lights', command: 'turn_on' },
      { kind: 'device', deviceName: 'Thermostat', command: 'set', value: '22°C' },
    ],
  },
  {
    id: '3',
    name: 'Away Mode',
    description: 'Secure the house when everyone leaves',
    category: 'Security',
    enabled: false,
    lastRun: '2026-10-06T18:12:00Z',
    runCount: 12,
    conditions: [
      { kind: 'presence', state: 'away' },
    ],
    actions: [
      { kind: 'device', deviceName: 'Front Door', command: 'set', value: 'locked' },
      { kind: 'device', deviceName: 'All Lights', command: 'turn_off' },
      { kind: 'notify', message: 'Away mode activated' },
    ],
  },
  {
    id: '4',
    name: 'Energy saver',
    description: 'Turn off lights when nobody is around',
    category: 'Energy',
    enabled: true,
    lastRun: '2026-10-07T13:45:00Z',
    runCount: 47,
    conditions: [
      { kind: 'motion', location: 'Living Room' },
      { kind: 'time', from: '09:00', to: '17:00' },
    ],
    actions: [
      { kind: 'delay', minutes: 30 },
      { kind: 'device', deviceName: 'Living Room Lights', command: 'turn_off' },
    ],
  },
  {
    id: '5',
    name: 'Cold weather alert',
    description: 'Heads-up when it gets cold outside',
    category: 'Notification',
    enabled: true,
    lastRun: '2026-10-07T06:00:00Z',
    runCount: 5,
    conditions: [
      { kind: 'temperature', op: 'lt', value: 15 },
    ],
    actions: [
      { kind: 'notify', message: 'Outside temperature dropped below 15°C' },
    ],
  },
];
export const mockRooms: Record<string, RoomData> = {
  'living-room': {
    id: 'living-room',
    name: 'Living Room',
    metrics: {
      temperature: { value: 23, change: 1 },
      humidity: { value: 48, change: 2 },
      co2: { value: 620, change: 10 },
      occupancy: { value: 2, change: 1 },
    },
    devices: [
      { id: '1', name: 'Ceiling Light', type: 'Light', on: true },
      { id: '2', name: 'Floor Lamp', type: 'Light', on: false },
      { id: '3', name: 'Thermostat', type: 'Climate', on: true },
      { id: '4', name: 'Smart TV', type: 'Media', on: false },
      { id: '5', name: 'Window Sensor', type: 'Sensor', on: true },
    ],
    activity: [
      { id: '1', timestamp: '2026-09-15T14:32:00', description: 'Window closed' },
      { id: '2', timestamp: '2026-09-15T14:20:00', description: 'Light turned on' },
      { id: '3', timestamp: '2026-09-15T13:58:00', description: 'Thermostat adjusted' },
    ],
  },
  kitchen: {
    id: 'kitchen',
    name: 'Kitchen',
    metrics: {
      temperature: { value: 24, change: 2 },
      humidity: { value: 55, change: 3 },
      co2: { value: 700, change: 15 },
      occupancy: { value: 1, change: -1 },
    },
    devices: [
      { id: '1', name: 'Oven', type: 'Appliance', on: true },
      { id: '2', name: 'Fridge', type: 'Appliance', on: true },
      { id: '3', name: 'Toaster', type: 'Appliance', on: false },
      { id: '5', name: 'Ceiling Light', type: 'Light', on: true },
      { id: '6', name: 'Airfryer', type: 'Appliance', on: true },
      { id: '7', name: 'Microwave', type: 'Appliance', on: true },
      { id: '8', name: 'Door Sensor', type: 'Sensor', on: true },
    ],
    activity: [
      { id: '1', timestamp: '2026-09-15T14:40:00', description: 'Oven turned on' },
      { id: '2', timestamp: '2026-09-15T14:12:00', description: 'Smoke detected' },
      { id: '3', timestamp: '2026-09-15T13:55:00', description: 'Fridge door opened' },
    ],
  },
  bedroom: {
    id: 'bedroom',
    name: 'Bedroom',
    metrics: {
      temperature: { value: 21, change: -1 },
      humidity: { value: 45, change: 0 },
      co2: { value: 500, change: -5 },
      occupancy: { value: 0, change: -2 },
    },
    devices: [
      { id: '1', name: 'Bedside Lamp', type: 'Light', on: false },
      { id: '2', name: 'Ceiling Fan', type: 'Fan', on: false },
      { id: '3', name: 'Window Sensor', type: 'Sensor', on: true },
    ],
    activity: [
      { id: '1', timestamp: '2026-09-15T13:30:00', description: 'Window closed' },
      { id: '2', timestamp: '2026-09-15T11:15:00', description: 'Lamp turned off' },
    ],
  },
  bathroom: {
    id: 'bathroom',
    name: 'Bathroom',
    metrics: {
      temperature: { value: 25, change: 3 },
      humidity: { value: 72, change: 8 },
      co2: { value: 450, change: 0 },
      occupancy: { value: 1, change: 1 },
    },
    devices: [
      { id: '1', name: 'Mirror Light', type: 'Light', on: true },
      { id: '2', name: 'Exhaust Fan', type: 'Fan', on: true },
      { id: '3', name: 'Leak Sensor', type: 'Sensor', on: true },
    ],
    activity: [
      { id: '1', timestamp: '2026-09-15T14:50:00', description: 'Exhaust fan on' },
      { id: '2', timestamp: '2026-09-15T14:45:00', description: 'Door opened' },
    ],
  },
};


export const mockEvents: Event[] = [
  { id: '1', timestamp: '2026-09-15T14:32:00', title: 'Front door opened', location: 'Entrance', severity: 'info' },
  { id: '2', timestamp: '2026-09-15T14:20:00', title: 'Bedroom window closed', location: 'Bedroom', severity: 'info' },
  { id: '3', timestamp: '2026-09-15T13:58:00', title: 'Oven door seal broken', location: 'Kitchen', severity: 'critical' },
  { id: '4', timestamp: '2026-09-15T13:41:00', title: 'Motion detected', location: 'Hallway', severity: 'info' },
  { id: '5', timestamp: '2026-09-15T13:12:00', title: 'Smoke level rising', location: 'Kitchen', severity: 'warning' },
  { id: '6', timestamp: '2026-09-15T12:55:00', title: 'Living room light turned on', location: 'Living Room', severity: 'info' },
];



export const mockProjections: Projection[] = [
  {
    id: 'p1',
    timestamp: '2026-10-07T08:15:00',
    title: 'Pipe burst risk',
    description: 'Pressure anomalies detected in the bathroom supply line over the past 3 days.',
    location: 'Bathroom',
    category: 'maintenance',
    severity: 'high',
    confidence: 0.78,
    horizon: '5–10 days',
    recommendedAction: 'Inspect shut-off valve and pressure regulator.',
  },
  {
    id: 'p2',
    timestamp: '2026-10-07T07:50:00',
    title: 'Fire risk — unattended cooking',
    description: 'Cooktop left on with no motion detected in the kitchen for 18 minutes.',
    location: 'Kitchen',
    category: 'safety',
    severity: 'critical',
    confidence: 0.92,
    horizon: 'immediate',
    recommendedAction: 'Turn off cooktop or confirm someone is present.',
  },
  {
    id: 'p3',
    timestamp: '2026-10-06T21:30:00',
    title: 'Refrigerator efficiency dropping',
    description: 'Power draw has increased 22% over the past two weeks while internal temperature held steady.',
    location: 'Kitchen',
    category: 'maintenance',
    severity: 'medium',
    confidence: 0.68,
    horizon: '2–3 weeks',
    recommendedAction: 'Check door seals and condenser coils.',
  },
  {
    id: 'p4',
    timestamp: '2026-10-06T18:12:00',
    title: 'Entry risk tonight',
    description: 'Front door left unlocked after 22:00 — three nights in a row this week.',
    location: 'Entrance',
    category: 'security',
    severity: 'high',
    confidence: 0.81,
    horizon: 'tonight',
    recommendedAction: 'Enable auto-lock schedule for the front door.',
  },
  {
    id: 'p5',
    timestamp: '2026-10-06T14:00:00',
    title: 'Energy spike forecast',
    description: 'Cold snap forecast — expect 18% higher heating load tomorrow evening.',
    category: 'energy',
    severity: 'low',
    confidence: 0.85,
    horizon: 'tomorrow',
    recommendedAction: 'Pre-heat earlier or shift high-draw tasks to off-peak hours.',
  },
];