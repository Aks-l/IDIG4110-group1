import { mockAutomations, mockRooms } from './mock-data';
import type { Automation, AutomationDraft, Device } from './types';

/**
 * In-memory state behind the placeholder routes in src/app/api. Device and
 * automation mutations are kept for the lifetime of the server process so
 * the placeholder API behaves like a real backend; restarting the server
 * resets the state.
 *
 * Device ids follow the `${roomId}-${deviceId}` convention used by the mock
 * devices adapter, so every device id is unique and identifies its room.
 */

const deviceOnOverrides = new Map<string, boolean>();

function devicesFromFixtures(): Device[] {
  return Object.values(mockRooms).flatMap((room) =>
    room.devices.map((device) => ({
      ...device,
      id: `${room.id}-${device.id}`,
      roomId: room.id,
    })),
  );
}

export function listPlaceholderDevices(): Device[] {
  return devicesFromFixtures().map((device) => {
    const on = deviceOnOverrides.get(device.id);
    return on === undefined ? device : { ...device, on };
  });
}

export function findPlaceholderDevice(deviceId: string): Device | undefined {
  return listPlaceholderDevices().find((device) => device.id === deviceId);
}

export function setPlaceholderDeviceState(
  deviceId: string,
  on: boolean,
): Device | undefined {
  const device = findPlaceholderDevice(deviceId);
  if (!device) return undefined;
  deviceOnOverrides.set(deviceId, on);
  return { ...device, on };
}

let placeholderAutomations: Automation[] = structuredClone(mockAutomations);
let nextAutomationId =
  placeholderAutomations.reduce(
    (max, automation) => Math.max(max, Number(automation.id) || 0),
    0,
  ) + 1;

export function listPlaceholderAutomations(): Automation[] {
  return placeholderAutomations.map((automation) => ({ ...automation }));
}

export function setPlaceholderAutomationEnabled(
  id: string,
  enabled: boolean,
): Automation | undefined {
  const automation = placeholderAutomations.find(
    (candidate) => candidate.id === id,
  );
  if (!automation) return undefined;
  automation.enabled = enabled;
  return { ...automation };
}

export function createPlaceholderAutomation(draft: AutomationDraft): Automation {
  const automation: Automation = {
    ...draft,
    id: String(nextAutomationId++),
    enabled: true,
    runCount: 0,
  };
  placeholderAutomations = [automation, ...placeholderAutomations];
  return { ...automation };
}
