import { apiClient } from './client';
import type { ThreeDRoomsData } from './types';

export function getThreeDRooms() {
  return apiClient<ThreeDRoomsData>('/3d/rooms');
}