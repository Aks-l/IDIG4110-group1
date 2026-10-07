import { getEvents, getProjections } from './events';
import { dataSource } from './data-source';
import { mockEvents, mockProjections } from './mock-data';
import type { Event, Projection } from './types';

export function loadEvents(): Promise<Event[]> {
  if (dataSource === 'mock') return Promise.resolve(mockEvents);
  return getEvents();
}

export function loadProjections(): Promise<Projection[]> {
  if (dataSource === 'mock') return Promise.resolve(mockProjections);
  return getProjections();
}