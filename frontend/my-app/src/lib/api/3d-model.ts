import { apiClient } from './client';
import type { RoomLayout } from './types';

export function getRoomLayout() {
  return apiClient<RoomLayout[]>('/3d/rooms');
}