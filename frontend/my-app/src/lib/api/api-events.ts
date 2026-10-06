import { apiClient } from './api-client';
import type { Event } from './types';

export function getEvents() {
  return apiClient<Event[]>('/events');
}

export function getEvent(eventId: string) {
  return apiClient<Event>(`/events/${eventId}`);
}

