'use client';

import { useEffect, useState } from 'react';
import { InfoBox } from '@/components/InfoBox';
import { loadEnergy } from '@/lib/api/energy.data';
import type {
  EnergyPageData,
  EnergyPeriod,
  EnergyGranularity,
  EnergyPoint,
} from '@/lib/api/types';

const PERIODS: EnergyPeriod[] = ['today', 'week', 'month', 'year'];

function formatLabel(timestamp: string, granularity: EnergyGranularity): string {
  const date = new Date(timestamp);
  switch (granularity) {
    case 'hour':
      return date.toLocaleTimeString('en-GB', {
        hour: '2-digit',
        minute: '2-digit',
        timeZone: 'UTC',
      });
    case 'day':
      return date.toLocaleDateString('en-GB', {
        day: 'numeric',
        month: 'short',
        timeZone: 'UTC',
      });
    case 'month':
      return date.toLocaleDateString('en-GB', {
        month: 'short',
        timeZone: 'UTC',
      });
  }
}

function formatChange(ratio: number): string {
  const pct = (ratio * 100).toFixed(0);
  return ratio >= 0 ? `+${pct}%` : `${pct}%`;
}

// ---------------------------------------------------------------------------
// Chart
// ---------------------------------------------------------------------------

type ConsumptionChartProps = {
  points: EnergyPoint[];
  granularity: EnergyGranularity;
  period: EnergyPeriod;
};

function ConsumptionChart({ points, granularity, period }: ConsumptionChartProps) {
  const width = 760;
  const height = 250;
  const chartTop = 24;
  const chartBottom = 198;

  const values = points.map((p) => p.kwh);
  const labels = points.map((p) => formatLabel(p.timestamp, granularity));
  const maxValue = Math.max(...values, 1) * 1.15;

  // Show labels only at intervals — otherwise they overlap on 24/30-point charts.
  const labelStep =
    points.length <= 12 ? 1 :
    points.length <= 24 ? 4 :
    5;

  const coords = values.map((value, i) => ({
    x: (i / Math.max(values.length - 1, 1)) * width,
    y: chartBottom - (value / maxValue) * (chartBottom - chartTop),
  }));

  const line = coords.map(({ x, y }) => `${x},${y}`).join(' ');
  const area = `0,${chartBottom} ${line} ${width},${chartBottom}`;

  return (
    <div className="panel p-5 sm:p-6">
      <div className="mb-5 flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Consumption</p>
          <h2 className="mt-1 text-lg font-semibold text-[#17221d]">Energy usage</h2>
        </div>
        <span className="rounded-full bg-[#e5f2e9] px-3 py-1 text-xs font-semibold text-[#1f6f5b]">
          {period}
        </span>
      </div>

      <div className="overflow-hidden">
        <svg
          viewBox={`0 0 ${width} ${height}`}
          className="h-auto w-full min-w-[560px]"
          role="img"
          aria-label={`${period} energy consumption chart`}
        >
          <defs>
            <linearGradient id="energy-fill" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stopColor="#77b99c" stopOpacity="0.42" />
              <stop offset="100%" stopColor="#dcefe3" stopOpacity="0.1" />
            </linearGradient>
          </defs>

          {[0, 1, 2, 3].map((gridLine) => {
            const y = chartTop + (gridLine / 3) * (chartBottom - chartTop);
            return (
              <line key={gridLine} x1="0" x2={width} y1={y} y2={y} stroke="#e3ebe4" strokeWidth="1" />
            );
          })}

          <polygon points={area} fill="url(#energy-fill)" />
          <polyline
            points={line}
            fill="none"
            stroke="#1f6f5b"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="4"
          />

          {coords.map(({ x }, i) => {
            const isFirst = i === 0;
            const isLast = i === coords.length - 1;
            if (!isFirst && !isLast && i % labelStep !== 0) return null;

            return (
              <text
                key={labels[i]}
                x={x}
                y="226"
                fill="#68766d"
                fontSize="12"
                textAnchor={isFirst ? 'start' : isLast ? 'end' : 'middle'}
              >
                {labels[i]}
              </text>
            );
          })}
        </svg>
      </div>

      <div className="mt-1 flex items-center justify-between text-xs text-[#68766d]">
        <span>Power draw</span>
        <span>kWh</span>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

export default function EnergyPage() {
  const [period, setPeriod] = useState<EnergyPeriod>('today');
  const [data, setData] = useState<EnergyPageData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
  loadEnergy(period)
    .then(setData)
    .catch(() => setError('Unable to load energy data'));
  }, [period]);

  if (error) return <p className="text-red-600">{error}</p>;
  if (!data) return <div className="panel p-6 text-sm text-[#68766d]">Loading energy data…</div>;

  const { summary, series, byRoom } = data;

  return (
    <div className="mx-auto max-w-6xl space-y-8">
      {/* Header */}
      <div>
        <p className="text-xs font-bold uppercase tracking-[0.2em] text-[#1f6f5b]">Resource overview</p>
        <h1 className="page-heading mt-2 text-3xl font-bold text-[#17221d] sm:text-4xl">Energy</h1>
        <p className="mt-2 text-sm text-[#68766d]">
          Understand where your energy goes and spot opportunities to save.
        </p>
      </div>

      {/* Summary stats */}
      <section className="grid gap-3 sm:grid-cols-3">
        <InfoBox label="Total energy used" value={`${summary.totalKwh.toFixed(2)} kWh`} />
        <InfoBox
          label="Estimated cost"
          value={`${summary.currency}${summary.estimatedCost.toFixed(2)}`}
        />
        <InfoBox
          label="Compared to previous"
          value={formatChange(summary.comparedToPrevious)}
        />
      </section>

      {/* Period selector + chart */}
      <section className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-sm font-semibold text-[#17221d]">Usage over time</p>
            <p className="text-xs text-[#68766d]">Compare your consumption across different periods.</p>
          </div>
          <div
            className="flex rounded-xl border border-[#dbe4dc] bg-white p-1 shadow-sm"
            role="group"
            aria-label="Energy period"
          >
            {PERIODS.map((option) => (
              <button
                key={option}
                type="button"
                onClick={() => {
                setData(null);
                setError(null);
                setPeriod(option);
              }}
                aria-pressed={period === option}
                className={`rounded-lg px-3 py-1.5 text-xs font-semibold capitalize transition-colors sm:px-4 ${
                  period === option
                    ? 'bg-[#1f6f5b] text-white shadow-sm'
                    : 'text-[#68766d] hover:bg-[#eef3ef] hover:text-[#17221d]'
                }`}
              >
                {option}
              </button>
            ))}
          </div>
        </div>

        <ConsumptionChart points={series.points} granularity={series.granularity} period={period} />
      </section>

      {/* By room */}
      <section className="panel p-5">
        <div className="mb-5 flex items-center justify-between">
          <div>
            <h2 className="mt-1 text-lg font-semibold text-[#17221d]">Enrgy Usage: By room</h2>
          </div>
          <span className="text-xs font-medium capitalize text-[#68766d]">{period}</span>
        </div>

        <ul className="space-y-4">
          {byRoom.map((room) => (
            <li key={room.roomId}>
              <div className="mb-1.5 flex items-center justify-between text-sm">
                <span className="font-medium text-[#31453a]">{room.roomName}</span>
                <span className="font-semibold text-[#17221d]">
                  {room.kwh.toFixed(2)} kWh · {summary.currency}
                  {room.estimatedCost.toFixed(2)}
                </span>
              </div>
              <div className="h-2 overflow-hidden rounded-full bg-[#e7eee8]">
                <div
                  className="h-full rounded-full bg-[#4c9a7c]"
                  style={{ width: `${room.percentage * 100}%` }}
                />
              </div>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}