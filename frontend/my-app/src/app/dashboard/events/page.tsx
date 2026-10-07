'use client';

import { useEffect, useMemo, useState } from 'react';
import { loadEvents, loadProjections } from '@/lib/api/events.data';
import { formatTime, formatDate } from '@/lib/format';
import type {
  ActivityItem,
  Event,
  Projection,
  ProjectionCategory,
  ProjectionSeverity,
  Severity,
} from '@/lib/api/types';

type SortKey = 'timestamp' | 'location' | 'severity';
type KindFilter = 'all' | 'event' | 'projection';

// --- Event styling ---

const eventSeverityStyles: Record<Severity, { dot: string; badge: string; label: string }> = {
  info:     { dot: 'bg-blue-500',   badge: 'bg-blue-50 text-blue-700',     label: 'Info'     },
  warning:  { dot: 'bg-yellow-500', badge: 'bg-yellow-50 text-yellow-700', label: 'Warning'  },
  critical: { dot: 'bg-red-500',    badge: 'bg-red-50 text-red-700',       label: 'Critical' },
};

const eventSeverityFallback = {
  dot: 'bg-gray-400',
  badge: 'bg-gray-100 text-gray-600',
  label: 'Unknown',
};

// --- Projection styling ---

const projectionSeverityStyles: Record<
  ProjectionSeverity,
  { dot: string; badge: string; label: string }
> = {
  low:      { dot: 'bg-slate-400',  badge: 'bg-slate-100 text-slate-600',   label: 'Low'      },
  medium:   { dot: 'bg-amber-500',  badge: 'bg-amber-50 text-amber-700',    label: 'Medium'   },
  high:     { dot: 'bg-orange-500', badge: 'bg-orange-50 text-orange-700',  label: 'High'     },
  critical: { dot: 'bg-red-500',    badge: 'bg-red-50 text-red-700',        label: 'Critical' },
};

const projectionCategoryLabels: Record<ProjectionCategory, string> = {
  maintenance: 'Maintenance',
  safety:      'Safety',
  security:    'Security',
  energy:      'Energy',
};

// --- Severity rank (normalises events and projections onto one scale) ---

function rankEvent(s: Severity): number {
  return { critical: 0, warning: 1, info: 2 }[s] ?? 99;
}

function rankProjection(s: ProjectionSeverity): number {
  return { critical: 0, high: 1, medium: 2, low: 3 }[s] ?? 99;
}

function rankItem(item: ActivityItem): number {
  return item.kind === 'event' ? rankEvent(item.severity) : rankProjection(item.severity);
}

function countBySeverity<T extends string>(items: { severity: T }[]) {
  const counts: Partial<Record<T, number>> = {};
  for (const item of items) {
    counts[item.severity] = (counts[item.severity] ?? 0) + 1;
  }
  return counts;
}

// --- Row components ---

function EventRow({ event }: { event: Event }) {
  const style = eventSeverityStyles[event.severity] ?? eventSeverityFallback;

  return (
    <li className="panel flex items-start gap-4 px-5 py-4">
      <span className={`mt-2 h-2.5 w-2.5 shrink-0 rounded-full ${style.dot}`} />

      <div className="flex flex-1 flex-col">
        <span className="font-medium text-[#17221d]">{event.title}</span>
        {event.location && (
          <span className="text-sm text-[#68766d]">{event.location}</span>
        )}
      </div>

      <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${style.badge}`}>
        {style.label}
      </span>

      <div className="w-24 text-right text-sm text-[#68766d]">
        <div>{formatTime(event.timestamp)}</div>
        <div className="text-xs text-[#9aa89f]">{formatDate(event.timestamp)}</div>
      </div>
    </li>
  );
}

function ProjectionRow({ projection }: { projection: Projection }) {
  const style = projectionSeverityStyles[projection.severity];
  const confidencePct = Math.round(projection.confidence * 100);
  const meta = [projection.location, projectionCategoryLabels[projection.category]]
    .filter(Boolean)
    .join(' · ');

  return (
    <li className="panel flex items-start gap-4 px-5 py-4">
      <span className={`mt-2 h-2.5 w-2.5 shrink-0 rounded-full ${style.dot}`} />

      <div className="flex flex-1 flex-col gap-1.5">
        {/* Title line — same structure as events */}
        <div className="flex items-center gap-2">
          <span className="font-medium text-[#17221d]">{projection.title}</span>
          <span className="rounded-full bg-amber-50 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-amber-700">
            Predicted
          </span>
        </div>

        {/* Secondary line — same style as event location */}
        {meta && <span className="text-sm text-[#68766d]">{meta}</span>}

        {/* Sub-row: confidence + horizon — muted, small */}
        <div className="mt-0.5 flex flex-wrap items-center gap-2 text-xs text-[#68766d]">
          <span className="inline-block h-1 w-16 overflow-hidden rounded-full bg-[#e7eee8]">
            <span
              className="block h-full rounded-full bg-[#1f6f5b]"
              style={{ width: `${confidencePct}%` }}
            />
          </span>
          <span className="font-medium text-[#17221d]">{confidencePct}%</span>
          <span className="text-[#9aa89f]">·</span>
          <span>Expected in {projection.horizon}</span>
        </div>

        {/* Recommended action — muted, inline */}
        {projection.recommendedAction && (
          <p className="text-xs text-[#68766d]">
            <span className="text-[#1f6f5b]">→ </span>
            {projection.recommendedAction}
          </p>
        )}
      </div>

      {/* Badge — same position as events */}
      <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${style.badge}`}>
        {style.label}
      </span>

      {/* Timestamp — same position as events */}
      <div className="w-24 text-right text-sm text-[#68766d]">
        <div>{formatTime(projection.timestamp)}</div>
        <div className="text-xs text-[#9aa89f]">{formatDate(projection.timestamp)}</div>
      </div>
    </li>
  );
}

// --- Page ---

export default function EventsPage() {
  const [sortBy, setSortBy] = useState<SortKey>('timestamp');
  const [ascending, setAscending] = useState(false);
  const [kindFilter, setKindFilter] = useState<KindFilter>('all');

  const [events, setEvents] = useState<Event[]>([]);
  const [projections, setProjections] = useState<Projection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([loadEvents(), loadProjections()])
      .then(([e, p]) => {
        setEvents(e);
        setProjections(p);
      })
      .catch(() => setError('Unable to load activity'))
      .finally(() => setLoading(false));
  }, []);

  // Merge into one list, tagging each item with its kind
  const combined: ActivityItem[] = useMemo(
    () => [
      ...events.map((e) => ({ kind: 'event' as const, ...e })),
      ...projections.map((p) => ({ kind: 'projection' as const, ...p })),
    ],
    [events, projections]
  );

  // Apply the tab filter
  const filtered = useMemo(() => {
    if (kindFilter === 'all') return combined;
    return combined.filter((i) => i.kind === kindFilter);
  }, [combined, kindFilter]);

  // Sort
  const sorted = useMemo(() => {
    const copy = [...filtered];
    const dir = ascending ? 1 : -1;

    copy.sort((a, b) => {
      if (sortBy === 'timestamp') {
        return (new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()) * dir;
      }
      if (sortBy === 'location') {
        return (a.location ?? '').localeCompare(b.location ?? '') * dir;
      }
      return (rankItem(a) - rankItem(b)) * dir;
    });

    return copy;
  }, [filtered, sortBy, ascending]);

  // Severity breakdowns for the summary panels
  const eventCounts = countBySeverity(events);
  const projectionCounts = countBySeverity(projections);

  const eventSeverityBreakdown = (['critical', 'warning', 'info'] as Severity[])
    .filter((s) => (eventCounts[s] ?? 0) > 0)
    .map((s) => ({ key: s, count: eventCounts[s] ?? 0 }));

  const projectionSeverityBreakdown = (
    ['critical', 'high', 'medium', 'low'] as ProjectionSeverity[]
  )
    .filter((s) => (projectionCounts[s] ?? 0) > 0)
    .map((s) => ({ key: s, count: projectionCounts[s] ?? 0 }));

  if (loading) {
    return <div className="panel p-6 text-sm text-[#68766d]">Loading activity…</div>;
  }

  if (error) {
    return <p className="text-red-600">{error}</p>;
  }

  const tabs: { key: KindFilter; label: string }[] = [
    { key: 'all', label: 'All' },
    { key: 'event', label: 'Events' },
    { key: 'projection', label: 'Predictions' },
  ];

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-heading text-3xl font-bold text-[#17221d]">
            Events &amp; Predictions
          </h1>
          <p className="text-sm text-[#68766d]">
            Activity and forecasts from your home
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

      {/* Summary */}
      <section className="grid gap-3 sm:grid-cols-2">
        {/* Events summary */}
        <div className="panel p-4">
          <div className="flex items-baseline">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">
              Events
            </p>
          </div>
          <p className="text-2xl font-semibold text-[#17221d]">{events.length}</p>
          <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1.5 text-xs">
            {eventSeverityBreakdown.length === 0 ? (
              <span className="text-[#9aa89f]">No events recorded</span>
            ) : (
              eventSeverityBreakdown.map(({ key, count }) => {
                const style = eventSeverityStyles[key] ?? eventSeverityFallback;
                return (
                  <span key={key} className="flex items-center gap-1.5 text-[#68766d]">
                    <span className={`h-2 w-2 rounded-full ${style.dot}`} />
                    <span className="font-medium text-[#17221d]">{count}</span>
                    <span>{style.label}</span>
                  </span>
                );
              })
            )}
          </div>
        </div>

        {/* Predictions summary */}
        <div className="panel p-4">
          <div className="flex items-baseline justify-between">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">
              Predictions
            </p>
          </div>
          <p className="text-2xl font-semibold text-[#17221d]">{projections.length}</p>
          <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1.5 text-xs">
            {projectionSeverityBreakdown.length === 0 ? (
              <span className="text-[#9aa89f]">Nothing predicted</span>
            ) : (
              projectionSeverityBreakdown.map(({ key, count }) => {
                const style = projectionSeverityStyles[key];
                return (
                  <span key={key} className="flex items-center gap-1.5 text-[#68766d]">
                    <span className={`h-2 w-2 rounded-full ${style.dot}`} />
                    <span className="font-medium text-[#17221d]">{count}</span>
                    <span>{style.label}</span>
                  </span>
                );
              })
            )}
          </div>
        </div>
      </section>

      {/* Kind filter tabs */}
      <div
        className="flex w-fit rounded-xl border border-[#dbe4dc] bg-white p-1 shadow-sm"
        role="group"
        aria-label="Activity kind filter"
      >
        {tabs.map((t) => (
          <button
            key={t.key}
            type="button"
            onClick={() => setKindFilter(t.key)}
            aria-pressed={kindFilter === t.key}
            className={`min-w-[96px] rounded-lg px-3 py-1.5 text-center text-xs font-semibold transition-colors ${
              kindFilter === t.key
                ? 'bg-[#1f6f5b] text-white shadow-sm'
                : 'text-[#68766d] hover:bg-[#eef3ef] hover:text-[#17221d]'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Empty state */}
      {sorted.length === 0 ? (
        <div className="panel flex flex-col items-center gap-2 px-6 py-16 text-center">
          <p className="text-sm font-semibold text-[#17221d]">
            {kindFilter === 'projection'
              ? 'No predictions'
              : kindFilter === 'event'
              ? 'No events yet'
              : 'No activity yet'}
          </p>
          <p className="max-w-md text-sm text-[#68766d]">
            {kindFilter === 'projection'
              ? 'The twin has not flagged anything to watch yet.'
              : 'Activity from your home will appear here as it happens.'}
          </p>
        </div>
      ) : (
        <ul className="flex flex-col gap-3">
          {sorted.map((item) =>
            item.kind === 'event' ? (
              <EventRow key={item.id} event={item} />
            ) : (
              <ProjectionRow key={item.id} projection={item} />
            )
          )}
        </ul>
      )}
    </div>
  );
}