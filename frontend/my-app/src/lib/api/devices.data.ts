import { getDevices } from './api-devices';
import { dataSource } from './data-source';
import { mockRooms } from './mock-data';
import type { Device } from './types';

const deviceOffsets: [number, number][] = [
  [-1.15, -0.65],
  [0.95, -0.45],
  [-0.75, 0.65],
  [0.95, 0.7],
  [0, 0],
];

export function loadDevices(): Promise<Device[]> {
  if (dataSource === 'api') return getDevices();

  return Promise.resolve(
    Object.values(mockRooms).flatMap((room) =>
      room.devices.map((device, index) => {
        const roomCenter: [number, number, number] = [
          room.id === 'living-room' || room.id === 'bedroom' ? -2.8 : room.id === 'bathroom' ? 3.9 : 2.3,
          0.9,
          room.id === 'living-room' || room.id === 'kitchen' ? 1.7 : -2.2,
        ];
        const [offsetX, offsetZ] = deviceOffsets[index % deviceOffsets.length];

        return {
        ...device,
        id: `${room.id}-${device.id}`,
        roomId: room.id,
          position: [roomCenter[0] + offsetX, roomCenter[1], roomCenter[2] + offsetZ],
        };
      }),
    ),
  );
}