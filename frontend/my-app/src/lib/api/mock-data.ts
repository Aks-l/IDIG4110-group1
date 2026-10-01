import type { Automation, Event, RoomData } from './types';

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
      { id: '3', name: 'Range Hood', type: 'Appliance', on: false },
      { id: '4', name: 'Smoke Sensor', type: 'Sensor', on: true },
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

export const mockAutomations: Automation[] = [
  { id: '1', name: 'Hallway motion light', description: 'Motion in Hallway: turn on Hallway Light at 40%', category: 'Comfort', enabled: true, lastRun: '2026-09-15T14:30:00', runCount: 128 },
  { id: '2', name: 'Good Morning', description: '07:00 weekdays: lights on, thermostat 22°C', category: 'Comfort', enabled: true, lastRun: '2026-09-15T07:00:00', runCount: 62 },
  { id: '3', name: 'Away Mode', description: 'Nobody home: lock doors, cameras on, lights off', category: 'Security', enabled: false, lastRun: '2026-09-14T18:12:00', runCount: 12 },
  { id: '4', name: 'Energy saver', description: 'No motion for 30 min: turn off all lights', category: 'Energy', enabled: true, lastRun: '2026-09-15T13:45:00', runCount: 47 },
  { id: '5', name: 'Faucet leak alert', description: 'Faucet running for 10 min: send notification', category: 'Notification', enabled: true, lastRun: '2026-09-12T11:22:00', runCount: 23 },
  { id: '6', name: 'Window open alert', description: 'Any window opens while away -> send notification', category: 'Notification', enabled: true, lastRun: '2026-09-15T11:22:00', runCount: 5 },
];

export const mockEvents: Event[] = [
  { id: '1', timestamp: '2026-09-15T14:32:00', title: 'Front door opened', location: 'Entrance', severity: 'info' },
  { id: '2', timestamp: '2026-09-15T14:20:00', title: 'Bedroom window closed', location: 'Bedroom', severity: 'info' },
  { id: '3', timestamp: '2026-09-15T13:58:00', title: 'Oven door seal broken', location: 'Kitchen', severity: 'critical' },
  { id: '4', timestamp: '2026-09-15T13:41:00', title: 'Motion detected', location: 'Hallway', severity: 'info' },
  { id: '5', timestamp: '2026-09-15T13:12:00', title: 'Smoke level rising', location: 'Kitchen', severity: 'warning' },
  { id: '6', timestamp: '2026-09-15T12:55:00', title: 'Living room light turned on', location: 'Living Room', severity: 'info' },
];