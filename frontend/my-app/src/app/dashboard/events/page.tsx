'use client';

import { useMemo, useState } from 'react';

type Severity = 'info' | 'warning' | 'critical';

type Event = {
  id: string;
  timestamp: string;
  title: string;
  location: string;
  severity: Severity;
};

type SortKey = 'timestamp' | 'location' | 'severity';

const events: Event[] = [
  { id: '1', timestamp: '2026-09-15T14:32:00', title: 'Front door opened',           location: 'Entrance',    severity: 'info'     },
  { id: '2', timestamp: '2026-09-15T14:20:00', title: 'Bedroom window closed',       location: 'Bedroom',     severity: 'info'     },
  { id: '3', timestamp: '2026-09-15T13:58:00', title: 'Oven door seal broken',       location: 'Kitchen',     severity: 'critical' },
  { id: '4', timestamp: '2026-09-15T13:41:00', title: 'Motion detected',             location: 'Hallway',     severity: 'info'     },
  { id: '5', timestamp: '2026-09-15T13:12:00', title: 'Smoke level rising',          location: 'Kitchen',     severity: 'warning'  },
  { id: '6', timestamp: '2026-09-15T12:55:00', title: 'Living room light turned on', location: 'Living Room', severity: 'info'     },
];

const severityStyles: Record<Severity, { dot: string; badge: string; label: string }> = {
  info:     { dot: 'bg-blue-500',   badge: 'bg-blue-50 text-blue-700',     label: 'Info'     },
  warning:  { dot: 'bg-yellow-500', badge: 'bg-yellow-50 text-yellow-700', label: 'Warning'  },
  critical: { dot: 'bg-red-500',    badge: 'bg-red-50 text-red-700',       label: 'Critical' },
};


const severityRank: Record<Severity, number> = {
  critical: 0,
  warning: 1,
  info: 2,
};

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString([], { day: '2-digit', month: 'short' });
}

export default function EventsPage() {
  const [sortBy, setSortBy] = useState<SortKey>('timestamp');
  const [ascending, setAscending] = useState(false);

  const sorted = useMemo(() => {
    const copy = [...events];
    const dir = ascending ? 1 : -1;

    copy.sort((a, b) => {
      if (sortBy === 'timestamp') {
        return (new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()) * dir;
      }
      if (sortBy === 'location') {
        return a.location.localeCompare(b.location) * dir;
      }
      // severity
      return (severityRank[a.severity] - severityRank[b.severity]) * dir;
    });

    return copy;
  }, [sortBy, ascending]);

  return (
    <div className="space-y-6">
      {/* Header + sort controls */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Events</h1>
          <p className="text-sm text-gray-500">{events.length} recent events</p>
        </div>

        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-500">Sort by</label>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortKey)}
            className="rounded-md border bg-white px-3 py-1.5 text-sm"
          >
            <option value="timestamp">Timestamp</option>
            <option value="location">Location</option>
            <option value="severity">Severity</option>
          </select>

          <button
            onClick={() => setAscending((v) => !v)}
            className="rounded-md border bg-white px-3 py-1.5 text-sm hover:bg-gray-50"
          >
            {ascending ? '↑ Asc' : '↓ Desc'}
          </button>
        </div>
      </div>

      {/* Event cards */}
      <ul className="flex flex-col gap-3">
        {sorted.map((e) => {
          const style = severityStyles[e.severity];
          return (
            <li
              key={e.id}
              className="flex items-center gap-4 rounded-lg border bg-white px-5 py-4 shadow-sm"
            >
              <span className={`h-2.5 w-2.5 shrink-0 rounded-full ${style.dot}`} />

              <div className="flex flex-1 flex-col">
                <span className="font-medium">{e.title}</span>
                <span className="text-sm text-gray-500">{e.location}</span>
              </div>

              <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${style.badge}`}>
                {style.label}
              </span>

              <div className="w-24 text-right text-sm text-gray-500">
                <div>{formatTime(e.timestamp)}</div>
                <div className="text-xs text-gray-400">{formatDate(e.timestamp)}</div>
              </div>
            </li>
          );
        })}
      </ul>
    </div>
  );
}