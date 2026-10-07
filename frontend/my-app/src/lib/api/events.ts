import { apiClient } from './client';
import type { Event, Projection } from './types';

export function getEvents() {
  return apiClient<Event[]>('/events');
}

export function getEvent(eventId: string) {
  return apiClient<Event>(`/events/${eventId}`);
}

export function getProjections() {
  return apiClient<Projection[]>('/projections');
}