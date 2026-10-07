'use client';

import { useEffect, useState } from 'react';
import { InfoBox } from '@/components/InfoBox';
import { loadEnergy } from '@/lib/api/energy.data';
import { ENERGY_PERIODS, ENERGY_MODES } from '@/lib/api/types';
import type {
  EnergyPageData,
  EnergyPeriod,
  EnergyGranularity,
  EnergyPoint,
  EnergyMode,
} from '@/lib/api/types';

const PERIOD_LABELS: Record<EnergyMode, Record<EnergyPeriod, string>> = {
  current: {
    today: 'Today',
    week: 'This Week',
    month: 'This Month',
    year: 'This Year',
  },
  projected: {
    today: 'Tomorrow',
    week: 'Next Week',
    month: 'Next Month',
    year: 'Next Year',
  },
};

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
  comparisonPoints?: EnergyPoint[];
  granularity: EnergyGranularity;
  period: EnergyPeriod;
  mode: EnergyMode;
};

function ConsumptionChart({
  points,
  comparisonPoints,
  granularity,
  period,
  mode,
}: ConsumptionChartProps) {
  if (points.length === 0) {
    return (
      <div className="panel flex items-center justify-center px-6 py-16">
        <p className="text-sm text-[#68766d]">
          No consumption data for {PERIOD_LABELS[mode][period].toLowerCase()}.
        </p>
      </div>
    );
  }

  const width = 760;
  const height = 260;
  const chartLeft = 48;    // room for Y-axis label
  const chartRight = 760;
  const chartTop = 24;
  const chartBottom = 198;
  const chartWidth = chartRight - chartLeft;
  const chartHeight = chartBottom - chartTop;

  const values = points.map((p) => p.kwh);
  const labels = points.map((p) => formatLabel(p.timestamp, granularity));

  const comparisonValues = comparisonPoints?.map((p) => p.kwh) ?? [];
  const showComparison = comparisonValues.length === values.length && values.length > 0;

  const allValues = [...values, ...comparisonValues];
  const maxValue = Math.max(...allValues, 1) * 1.15;

  // Show labels only at intervals — otherwise they overlap on 24/30-point charts.
  const labelStep =
    points.length <= 12 ? 1 :
    points.length <= 24 ? 4 :
    5;

  const coords = values.map((value, i) => ({
    x: chartLeft + (i / Math.max(values.length - 1, 1)) * chartWidth,
    y: chartBottom - (value / maxValue) * chartHeight,
  }));

  const comparisonCoords = showComparison
    ? comparisonValues.map((value, i) => ({
        x: chartLeft + (i / Math.max(comparisonValues.length - 1, 1)) * chartWidth,
        y: chartBottom - (value / maxValue) * chartHeight,
      }))
    : [];

  const line = coords.map(({ x, y }) => `${x},${y}`).join(' ');
  const area = `${chartLeft},${chartBottom} ${line} ${chartRight},${chartBottom}`;
  const comparisonLine = comparisonCoords.map(({ x, y }) => `${x},${y}`).join(' ');

  const isProjected = mode === 'projected';
  const accentColor = isProjected ? '#8a5cf6' : '#1f6f5b';
  const fillTop = isProjected ? '#b79cff' : '#77b99c';
  const fillBottom = isProjected ? '#ece4ff' : '#dcefe3';

  // Y-axis tick values (4 gridlines)
  const yTicks = [0, 1, 2, 3].map((i) => {
    const value = maxValue - (i / 3) * maxValue;
    const y = chartTop + (i / 3) * chartHeight;
    return { value, y };
  });

  return (
    <div className="panel p-5 sm:p-6">
      <div className="mb-5 flex items-start justify-between gap-4">
        <div>
          <p
            className="text-xs font-bold uppercase tracking-[0.16em]"
            style={{ color: accentColor }}
          >
            {isProjected ? 'Forecast' : 'Consumption'}
          </p>
          <h2 className="mt-1 text-lg font-semibold text-[#17221d]">
            {isProjected ? 'Projected vs. current' : 'Energy usage'}
          </h2>
        </div>
        <span
          className="rounded-full px-3 py-1 text-xs font-semibold"
          style={{
            backgroundColor: isProjected ? '#ede7ff' : '#e5f2e9',
            color: accentColor,
          }}
        >
          {PERIOD_LABELS[mode][period]}
        </span>
      </div>

      <div className="overflow-hidden">
        <svg
          viewBox={`0 0 ${width} ${height}`}
          className="h-auto w-full min-w-[560px]"
          role="img"
          aria-label={`${PERIOD_LABELS[mode][period]} energy ${isProjected ? 'forecast' : 'consumption'} chart`}
        >
          <defs>
            <linearGradient id="energy-fill" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stopColor={fillTop} stopOpacity="0.42" />
              <stop offset="100%" stopColor={fillBottom} stopOpacity="0.1" />
            </linearGradient>
          </defs>

          {/* Y-axis label (rotated) */}
          <text
            x={14}
            y={chartTop + chartHeight / 2}
            fill="#68766d"
            fontSize="12"
            textAnchor="middle"
            transform={`rotate(-90, 14, ${chartTop + chartHeight / 2})`}
          >
            kWh
          </text>

          {/* Y-axis tick labels */}
          {yTicks.map(({ value, y }, i) => (
            <text
              key={i}
              x={chartLeft - 8}
              y={y + 4}
              fill="#68766d"
              fontSize="11"
              textAnchor="end"
            >
              {value.toFixed(1)}
            </text>
          ))}

          {/* Horizontal grid lines */}
          {yTicks.map(({ y }, i) => (
            <line
              key={i}
              x1={chartLeft}
              x2={chartRight}
              y1={y}
              y2={y}
              stroke="#e3ebe4"
              strokeWidth="1"
            />
          ))}

          {/* Y-axis line */}
          <line
            x1={chartLeft}
            x2={chartLeft}
            y1={chartTop}
            y2={chartBottom}
            stroke="#c9d4cb"
            strokeWidth="1"
          />

          {/* Comparison line (current period) — muted, drawn behind */}
          {showComparison && (
            <polyline
              points={comparisonLine}
              fill="none"
              stroke="#9fb3a7"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              strokeDasharray="4 4"
              opacity="0.7"
            />
          )}

          {/* Primary area + line */}
          <polygon points={area} fill="url(#energy-fill)" />
          <polyline
            points={line}
            fill="none"
            stroke={accentColor}
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="4"
            strokeDasharray={isProjected ? '8 6' : undefined}
          />

          {/* X-axis labels */}
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

      {/* Legend + axis footer */}
      <div className="mt-2 flex items-center justify-between text-xs text-[#68766d]">
        <div className="flex items-center gap-4">
          {showComparison && (
            <span className="flex items-center gap-1.5">
              <span
                className="inline-block h-0.5 w-4 rounded-full bg-[#9fb3a7]"
                style={{ borderTop: '2px dashed #9fb3a7', height: 0 }}
              />
              Current
            </span>
          )}
          <span className="flex items-center gap-1.5">
            <span
              className="inline-block h-0.5 w-4 rounded-full"
              style={{ backgroundColor: accentColor }}
            />
            {isProjected ? 'Projected' : 'Usage'}
          </span>
        </div>
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
  const [mode, setMode] = useState<EnergyMode>('current');
  const [data, setData] = useState<EnergyPageData | null>(null);
  const [comparisonData, setComparisonData] = useState<EnergyPageData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const main = await loadEnergy(period, mode);
        if (cancelled) return;
        setData(main);

        if (mode === 'projected') {
          const comp = await loadEnergy(period, 'current');
          if (cancelled) return;
          setComparisonData(comp);
        } else {
          setComparisonData(null);
        }
      } catch {
        if (!cancelled) setError('Unable to load energy data');
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [period, mode]);

  const reset = () => {
    setData(null);
    setComparisonData(null);
    setError(null);
  };

  if (error) return <p className="text-red-600">{error}</p>;
  if (!data) return <div className="panel p-6 text-sm text-[#68766d]">Loading energy data…</div>;

  const { summary, series, byRoom } = data;
  const isProjected = mode === 'projected';

  return (
    <div className="mx-auto max-w-6xl space-y-8">
      {/* Header */}
      <div>
        <p className="text-xs font-bold uppercase tracking-[0.2em] text-[#1f6f5b]">Resource overview</p>
        <h1 className="page-heading mt-2 text-3xl font-bold text-[#17221d] sm:text-4xl">Energy</h1>
        <p className="mt-2 text-sm text-[#68766d]">
          {isProjected
            ? 'Forecast of upcoming energy usage and costs, compared to current period.'
            : 'Current energy usage and costs.'}
        </p>
      </div>

      {/* Summary stats */}
      <section className="grid gap-3 sm:grid-cols-3">
        <InfoBox
          label={isProjected ? 'Projected energy' : 'Total energy used'}
          value={
            series.points.length === 0
              ? 'N/A'
              : `${summary.totalKwh.toFixed(2)} kWh`
          }
        />
        <InfoBox
          label={isProjected ? 'Projected cost' : 'Estimated cost'}
          value={
            series.points.length === 0
              ? 'N/A'
              : `${summary.currency}${summary.estimatedCost.toFixed(2)}`
          }
        />
        <InfoBox
          label="Compared to previous"
          value={
            summary.comparedToPrevious === null
              ? 'N/A'
              : formatChange(summary.comparedToPrevious)
          }
        />
      </section>

      {/* Period + mode selector + chart */}
      <section className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-sm font-semibold text-[#17221d]">
              {isProjected ? 'Forecast over time' : 'Usage over time'}
            </p>
            <p className="text-xs text-[#68766d]">
              {isProjected
                ? 'Projected consumption with current period shown for comparison.'
                : 'Compare your consumption across different periods.'}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {/* Mode toggle */}
            <div
              className="flex rounded-xl border border-[#dbe4dc] bg-white p-1 shadow-sm"
              role="group"
              aria-label="Energy view mode"
            >
              {ENERGY_MODES.map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => {
                    if (mode === option) return;
                    reset();
                    setMode(option);
                  }}
                  aria-pressed={mode === option}
                  className={`min-w-[80px] rounded-lg px-3 py-1.5 text-center text-xs font-semibold capitalize transition-colors sm:px-4 ${
                    mode === option
                      ? 'bg-[#17221d] text-white shadow-sm'
                      : 'text-[#68766d] hover:bg-[#eef3ef] hover:text-[#17221d]'
                  }`}
                >
                  {option}
                </button>
              ))}
            </div>

            {/* Period toggle */}
            <div
              className="flex rounded-xl border border-[#dbe4dc] bg-white p-1 shadow-sm"
              role="group"
              aria-label="Energy period"
            >
              {ENERGY_PERIODS.map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => {
                    if (period === option) return;
                    reset();
                    setPeriod(option);
                  }}
                  aria-pressed={period === option}
                  className={`min-w-[96px] rounded-lg px-3 py-1.5 text-center text-xs font-semibold transition-colors sm:px-4 ${
                    period === option
                      ? 'bg-[#1f6f5b] text-white shadow-sm'
                      : 'text-[#68766d] hover:bg-[#eef3ef] hover:text-[#17221d]'
                  }`}
                >
                  {PERIOD_LABELS[mode][option]}
                </button>
              ))}
            </div>
          </div>
        </div>

        <ConsumptionChart
          points={series.points}
          comparisonPoints={comparisonData?.series.points}
          granularity={series.granularity}
          period={period}
          mode={mode}
        />
      </section>

      {/* By room */}
      <section className="panel p-5">
        <div className="mb-5 flex items-center justify-between">
          <div>
            <h2 className="mt-1 text-lg font-semibold text-[#17221d]">
              {isProjected ? 'Projected usage: By room' : 'Energy usage: By room'}
            </h2>
          </div>
          <span className="text-xs font-medium text-[#68766d]">
            {PERIOD_LABELS[mode][period]}
          </span>
        </div>

        {byRoom.length === 0 ? (
          <p className="py-6 text-center text-sm text-[#68766d]">
            No room-level data for {PERIOD_LABELS[mode][period].toLowerCase()} yet.
          </p>
        ) : (
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
                    className="h-full rounded-full"
                    style={{
                      width: `${room.percentage * 100}%`,
                      backgroundColor: isProjected ? '#8a5cf6' : '#4c9a7c',
                    }}
                  />
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}