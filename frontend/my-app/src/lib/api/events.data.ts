import { getEvents } from './events';
import { dataSource } from './data-source';
import { mockEvents } from './mock-data';
import type { Event } from './types';

export function loadEvents(): Promise<Event[]> {
  if (dataSource === 'mock') return Promise.resolve(mockEvents);
  return getEvents();
}