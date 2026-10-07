import { getRoomLayout } from './3d-model';
import { dataSource } from './data-source';
import type { RoomLayout } from './types';

export const mockRoomLayout: RoomLayout[] = [
  { roomId: 'living-room', position: [-2.85, 0.3, 1.75], size: [4.6, 0.6, 3.4] },
  { roomId: 'kitchen', position: [2.3, 0.3, 1.75], size: [3.2, 0.6, 2.8] },
  { roomId: 'bedroom', position: [-2.85, 0.3, -2.2], size: [4.2, 0.6, 3.0] },
  { roomId: 'bathroom', position: [3.9, 0.3, -2.2], size: [2.9, 0.6, 2.5] },
];

export function loadRoomLayout(): Promise<RoomLayout[]> {
  if (dataSource === 'mock') return Promise.resolve(mockRoomLayout);
  return getRoomLayout();
}