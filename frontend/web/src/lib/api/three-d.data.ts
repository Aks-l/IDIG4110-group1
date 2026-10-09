import { getThreeDRooms } from './3d-model';
import { dataSource } from './data-source';
import type { RoomConnection, RoomLayout, ThreeDRoomsData } from './types';

export const mockRoomLayout: RoomLayout[] = [
  { roomId: 'living-room', position: [-2.85, 0.3, 1.75], size: [4.6, 0.6, 3.4] },
  { roomId: 'kitchen', position: [2.3, 0.3, 1.75], size: [3.2, 0.6, 2.8] },
  { roomId: 'bedroom', position: [-2.85, 0.3, -2.2], size: [4.2, 0.6, 3.0] },
  { roomId: 'bathroom', position: [3.9, 0.3, -2.2], size: [2.9, 0.6, 2.5] },
];

// Passages between the mock rooms — the three blocks the scene used to
// hard-code, now part of the layout payload like the connections the twin
// hybrid derives from the home's connects_to relations.
export const mockConnections: RoomConnection[] = [
  { fromRoomId: 'living-room', toRoomId: 'kitchen', label: 'open plan', position: [0.075, 0.3, 1.75], size: [1.37, 0.6, 0.85] },
  { fromRoomId: 'living-room', toRoomId: 'bedroom', label: 'archway', position: [-2.85, 0.3, -0.325], size: [1.05, 0.6, 0.87] },
  { fromRoomId: 'kitchen', toRoomId: 'bathroom', label: 'doorway', position: [3.2, 0.3, -0.3], size: [0.95, 0.6, 1.42] },
];

export function loadThreeDRooms(): Promise<ThreeDRoomsData> {
  if (dataSource === 'mock') {
    return Promise.resolve({ rooms: mockRoomLayout, connections: mockConnections });
  }
  return getThreeDRooms();
}