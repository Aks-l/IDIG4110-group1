export type RoomType = "living-room" | "kitchen" | "bedroom" | "bathroom" | "office" | "hallway";
export type RoomStatus = "normal" | "warning" | "critical";
export type SensorStatus = "normal" | "warning" | "critical" | "offline";
export type SensorType = "temperature" | "humidity" | "temperature-humidity" | "co2" | "air-quality" | "motion" | "light" | "door" | "window" | "smoke" | "water-leak" | "power" | "energy" | "noise" | "occupancy";
export type AlertSeverity = "critical" | "warning" | "info" | "resolved";
export type AlertStatus = "open" | "acknowledged" | "resolved";

export interface Room {
  id: string;
  name: string;
  type: RoomType;
  floor: number;
  temperature: number;
  humidity: number;
  airQuality: number;
  co2: number;
  motion: boolean;
  lightsOn: number;
  lightCount: number;
  status: RoomStatus;
  position: [number, number, number];
  size: [number, number, number];
}

export interface SensorThresholds {
  warningMin?: number;
  warningMax?: number;
  criticalMin?: number;
  criticalMax?: number;
}

export interface Sensor {
  id: string;
  name: string;
  type: SensorType;
  roomId: string;
  description: string;
  position: [number, number, number];
  value: number;
  unit: string;
  status: SensorStatus;
  lastUpdated: string;
  average: number;
  min: number;
  max: number;
  thresholds: SensorThresholds;
}

export interface Alert {
  id: string;
  severity: AlertSeverity;
  status: AlertStatus;
  roomId: string;
  sensorId?: string;
  timestamp: string;
  description: string;
}

export interface TimePoint {
  time: string;
  temperature: number;
  energy: number;
  humidity: number;
  co2: number;
}
