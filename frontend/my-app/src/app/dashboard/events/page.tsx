'use client';

import { useEffect, useMemo, useState } from 'react';
import { loadEvents } from '@/lib/api/events.data';
import type { Event, Severity } from '@/lib/api/types';
import { formatDate, formatTime } from '@/lib/format';

type SortKey = 'timestamp' | 'location' | 'severity';

const severityStyles: Record<Severity, { dot: string; badge: string; label: string }> = {
  info:     { dot: 'bg-blue-500',   badge: 'bg-blue-50 text-blue-700',     label: 'Info'     },
  warning:  { dot: 'bg-yellow-500', badge: 'bg-yellow-50 text-yellow-700', label: 'Warning'  },
  critical: { dot: 'bg-red-500',    badge: 'bg-red-50 text-red-700',       label: 'Critical' },
};

// Fallback used when the backend returns a severity value we don't recognise.
const severityFallback = {
  dot: 'bg-gray-400',
  badge: 'bg-gray-100 text-gray-600',
  label: 'Unknown',
};

const severityRank: Record<Severity, number> = {
  critical: 0,
  warning: 1,
  info: 2,
};


export default function EventsPage() {
  const [sortBy, setSortBy] = useState<SortKey>('timestamp');
  const [ascending, setAscending] = useState(false);
  const [events, setEvents] = useState<Event[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadEvents()
      .then(setEvents)
      .catch(() => setError('Unable to load events'))
      .finally(() => setLoading(false));
  }, []);

  const sorted = useMemo(() => {
    const copy = [...events];
    const dir = ascending ? 1 : -1;

    copy.sort((a, b) => {
      if (sortBy === 'timestamp') {
        return (new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()) * dir;
      }
      if (sortBy === 'location') {
        return (a.location ?? '').localeCompare(b.location ?? '') * dir;
      }
      return ((severityRank[a.severity] ?? 99) - (severityRank[b.severity] ?? 99)) * dir;
    });

    return copy;
  }, [events, sortBy, ascending]);

  if (loading) {
    return <div className="panel p-6 text-sm text-[#68766d]">Loading events…</div>;
  }

  if (error) {
    return <p className="text-red-600">{error}</p>;
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      {/* Header + sort controls */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-heading text-3xl font-bold text-[#17221d]">Events</h1>
          <p className="text-sm text-[#68766d]">
            {events.length} recent {events.length === 1 ? 'event' : 'events'}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <label className="text-sm text-[#68766d]">Sort by</label>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortKey)}
            className="rounded-xl border border-[#dbe4dc] bg-white px-3 py-1.5 text-sm text-[#17221d] shadow-sm"
          >
            <option value="timestamp">Timestamp</option>
            <option value="location">Location</option>
            <option value="severity">Severity</option>
          </select>

          <button
            onClick={() => setAscending((v) => !v)}
            className="rounded-xl border border-[#dbe4dc] bg-white px-3 py-1.5 text-sm font-medium text-[#17221d] shadow-sm hover:bg-[#f1f5f1]"
          >
            {ascending ? '↑ Asc' : '↓ Desc'}
          </button>
        </div>
      </div>

      {/* Empty state */}
      {events.length === 0 ? (
        <div className="panel flex flex-col items-center gap-2 px-6 py-16 text-center">
          <p className="text-sm font-semibold text-[#17221d]">No events yet</p>
          <p className="max-w-md text-sm text-[#68766d]">
            No events could be found. 
          </p>
        </div>
      ) : (
        <ul className="flex flex-col gap-3">
          {sorted.map((e) => {
            const style = severityStyles[e.severity] ?? severityFallback;
            return (
              <li
                key={e.id}
                className="panel flex items-center gap-4 px-5 py-4"
              >
                <span className={`h-2.5 w-2.5 shrink-0 rounded-full ${style.dot}`} />

                <div className="flex flex-1 flex-col">
                  <span className="font-medium text-[#17221d]">{e.title}</span>
                  {e.location && (
                    <span className="text-sm text-[#68766d]">{e.location}</span>
                  )}
                </div>

                <span
                  className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${style.badge}`}
                >
                  {style.label}
                </span>

                <div className="w-24 text-right text-sm text-[#68766d]">
                  <div>{formatTime(e.timestamp)}</div>
                  <div className="text-xs text-[#9aa89f]">{formatDate(e.timestamp)}</div>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}