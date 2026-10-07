import { apiClient } from './client';
import type { RoomData } from './types';

export function getRooms() {
  return apiClient<RoomData[]>('/rooms');
}

export function getRoom(roomId: string) {
  return apiClient<RoomData>(`/rooms/${roomId}`);
}