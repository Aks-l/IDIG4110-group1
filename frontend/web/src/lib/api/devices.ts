import { apiClient } from './client';
import type { Device } from './types';

export function getDevices() {
  return apiClient<Device[]>('/devices');
}

export function getDevice(deviceId: string) {
  return apiClient<Device>(`/devices/${deviceId}`);
}

export function setDeviceState(deviceId: string, on: boolean) {
  return apiClient<Device>(`/devices/${deviceId}`, {
    method: 'PATCH',
    body: JSON.stringify({ on }),
  });
}