import type { Device } from './types';

/**
 * Demo heuristic for devices that carry a warning badge: smoke and leak
 * sensors. twin-core has no device health concept yet, so this one function
 * is how the devices page and the overview agree on the warning count.
 * The type check is case-insensitive because the twin maps `device_type`
 * verbatim (for example "sensor") while the fixtures use "Sensor".
 */
export function isWarningDevice(device: Device): boolean {
  return (
    device.type.toLowerCase() === 'sensor' && /smoke|leak/i.test(device.name)
  );
}