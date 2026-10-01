import { getRoom, getRooms } from './api-rooms';
import { dataSource } from './data-source';
import { mockRooms } from './mock-data';
import type { RoomData } from './types';

export function loadRooms(): Promise<RoomData[]> {
  if (dataSource === 'mock') return Promise.resolve(Object.values(mockRooms));
  return getRooms();
}

export function loadRoom(roomId: string): Promise<RoomData> {
  if (dataSource === 'mock') {
    const room = mockRooms[roomId];
    if (!room) return Promise.reject(new Error(`Unknown mock room: ${roomId}`));
    return Promise.resolve(room);
  }
  return getRoom(roomId);
}