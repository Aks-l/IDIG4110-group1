import { Alert, Room, Sensor, SensorType, TimePoint } from "@/types";

export const rooms: Room[] = [
  { id: "living-room", name: "Living Room", type: "living-room", floor: 1, temperature: 21.4, humidity: 42, airQuality: 96, co2: 684, motion: true, lightsOn: 3, lightCount: 4, status: "normal", position: [-2.6, 0.08, 1.05], size: [4.8, 0.18, 3.3] },
  { id: "kitchen", name: "Kitchen", type: "kitchen", floor: 1, temperature: 22.1, humidity: 48, airQuality: 91, co2: 812, motion: true, lightsOn: 4, lightCount: 5, status: "warning", position: [2.25, 0.08, 1.05], size: [4.2, 0.18, 3.3] },
  { id: "hallway", name: "Hallway", type: "hallway", floor: 1, temperature: 20.7, humidity: 41, airQuality: 98, co2: 590, motion: true, lightsOn: 2, lightCount: 3, status: "normal", position: [0, 0.08, -0.95], size: [1.2, 0.18, 2.8] },
  { id: "bedroom", name: "Bedroom", type: "bedroom", floor: 1, temperature: 20.2, humidity: 45, airQuality: 94, co2: 734, motion: false, lightsOn: 1, lightCount: 2, status: "normal", position: [-2.6, 0.08, -2.05], size: [4.8, 0.18, 1.5] },
  { id: "bathroom", name: "Bathroom", type: "bathroom", floor: 1, temperature: 23.4, humidity: 58, airQuality: 88, co2: 768, motion: false, lightsOn: 1, lightCount: 2, status: "normal", position: [0.85, 0.08, -2.05], size: [1.6, 0.18, 1.5] },
  { id: "office", name: "Office", type: "office", floor: 1, temperature: 21.1, humidity: 40, airQuality: 97, co2: 642, motion: false, lightsOn: 1, lightCount: 2, status: "normal", position: [3.05, 0.08, -2.05], size: [2.4, 0.18, 1.5] },
];

export const sensors: Sensor[] = [
  { id: "S-TEMP-LIVING", name: "Living Room Temperature", type: "temperature", roomId: "living-room", description: "Ambient temperature sensor", position: [-2.6, 1.05, 1.05], value: 21.4, unit: "°C", status: "normal", lastUpdated: "10 sec ago", average: 21.1, min: 19.8, max: 22.7, thresholds: { warningMin: 18, warningMax: 26, criticalMin: 15, criticalMax: 30 } },
  { id: "S-HUM-LIVING", name: "Living Room Humidity", type: "humidity", roomId: "living-room", description: "Relative humidity sensor", position: [-2.1, 1.05, 1.05], value: 42, unit: "%", status: "normal", lastUpdated: "10 sec ago", average: 43, min: 37, max: 49, thresholds: { warningMin: 30, warningMax: 60, criticalMin: 20, criticalMax: 75 } },
  { id: "S-CO2-KITCHEN", name: "Kitchen CO2 Monitor", type: "co2", roomId: "kitchen", description: "Indoor air quality and CO2 sensor", position: [2.25, 1.05, 1.05], value: 812, unit: "ppm", status: "warning", lastUpdated: "15 sec ago", average: 746, min: 560, max: 1020, thresholds: { warningMax: 800, criticalMax: 1200 } },
  { id: "S-MOTION-HALL", name: "Hallway Motion", type: "motion", roomId: "hallway", description: "Passive infrared motion sensor", position: [0, 1.2, -0.95], value: 1, unit: "detected", status: "normal", lastUpdated: "4 sec ago", average: 0, min: 0, max: 1, thresholds: {} },
  { id: "S-TEMP-BED", name: "Bedroom Temperature", type: "temperature", roomId: "bedroom", description: "Bedroom climate sensor", position: [-2.6, 1.05, -2.05], value: 20.2, unit: "°C", status: "normal", lastUpdated: "10 sec ago", average: 20.4, min: 19.2, max: 21.3, thresholds: { warningMin: 18, warningMax: 26, criticalMin: 15, criticalMax: 30 } },
  { id: "S-WATER-BATH", name: "Bathroom Leak Guard", type: "water-leak", roomId: "bathroom", description: "Under-sink water leak detector", position: [0.85, 0.3, -2.05], value: 0, unit: "clear", status: "normal", lastUpdated: "22 sec ago", average: 0, min: 0, max: 0, thresholds: {} },
];

export const alerts: Alert[] = [
  { id: "ALT-1041", severity: "warning", status: "open", roomId: "kitchen", sensorId: "S-CO2-KITCHEN", timestamp: "12 min ago", description: "Kitchen CO2 is above the preferred indoor air threshold" },
  { id: "ALT-1038", severity: "info", status: "open", roomId: "hallway", sensorId: "S-MOTION-HALL", timestamp: "21 min ago", description: "Motion detected in hallway" },
  { id: "ALT-1031", severity: "resolved", status: "resolved", roomId: "bathroom", sensorId: "S-WATER-BATH", timestamp: "2 hr ago", description: "Bathroom water leak check returned clear" },
];

export const timeSeries: TimePoint[] = [
  { time: "06:00", temperature: 19.8, energy: 1.8, humidity: 44, co2: 604 }, { time: "08:00", temperature: 20.4, energy: 2.4, humidity: 43, co2: 682 }, { time: "10:00", temperature: 21.1, energy: 3.2, humidity: 42, co2: 744 }, { time: "12:00", temperature: 21.8, energy: 4.2, humidity: 41, co2: 812 }, { time: "14:00", temperature: 21.6, energy: 3.8, humidity: 42, co2: 768 }, { time: "16:00", temperature: 21.4, energy: 4.6, humidity: 43, co2: 684 }, { time: "18:00", temperature: 22, energy: 5.1, humidity: 44, co2: 726 }, { time: "20:00", temperature: 21.5, energy: 3.7, humidity: 45, co2: 642 },
];

export const sensorTypeLabels: Record<SensorType, string> = { temperature: "Temperature", humidity: "Humidity", "temperature-humidity": "Temperature & Humidity", co2: "CO2", "air-quality": "Air Quality", motion: "Motion", light: "Light", door: "Door", window: "Window", smoke: "Smoke", "water-leak": "Water Leak", power: "Power", energy: "Energy", noise: "Noise", occupancy: "Occupancy" };
export const sensorTypeOptions = Object.entries(sensorTypeLabels).map(([value, label]) => ({ value: value as SensorType, label }));
export const getRoom = (id: string) => rooms.find((room) => room.id === id) ?? rooms[0];
