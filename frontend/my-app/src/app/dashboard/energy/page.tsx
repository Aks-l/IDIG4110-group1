
'use client';

import { useState } from 'react';
import { InfoBox } from '@/components/InfoBox';

type EnergyRange = 'Today' | 'Week' | 'Month' | 'Year';

const rangeData: Record<EnergyRange, { labels: string[]; values: number[]; total: string }> = {
  Today: {
    labels: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', 'Now'],
    values: [0.42, 0.28, 0.75, 1.16, 1.42, 1.08, 1.2],
    total: '8.4 kWh',
  },
  Week: {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    values: [11.2, 13.8, 12.6, 15.1, 14.3, 9.6, 8.4],
    total: '85.0 kWh',
  },
  Month: {
    labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4'],
    values: [82, 91, 76, 85],
    total: '334 kWh',
  },
  Year: {
    labels: ['Jan', 'Mar', 'May', 'Jul', 'Sep', 'Nov'],
    values: [310, 284, 342, 296, 334, 318],
    total: '3,684 kWh',
  },
};

const rooms = [
  { name: 'Kitchen', value: '420 W', share: 42 },
  { name: 'Living room', value: '310 W', share: 31 },
  { name: 'Bedroom', value: '180 W', share: 18 },
  { name: 'Other', value: '90 W', share: 9 },
];

const consumers = [
  { name: 'Oven', value: '2.1 kWh', note: 'Kitchen' },
  { name: 'EV charger', value: '1.8 kWh', note: 'Garage' },
  { name: 'Heat pump', value: '1.4 kWh', note: 'Living room' },
];

function ConsumptionChart({ range }: { range: EnergyRange }) {
  const { labels, values } = rangeData[range];
  const width = 760;
  const height = 250;
  const chartTop = 24;
  const chartBottom = 198;
  const maxValue = Math.max(...values) * 1.15;
  const points = values.map((value, index) => {
    const x = (index / (values.length - 1)) * width;
    const y = chartBottom - (value / maxValue) * (chartBottom - chartTop);
    return { x, y };
  });
  const line = points.map(({ x, y }) => `${x},${y}`).join(' ');
  const area = `0,${chartBottom} ${line} ${width},${chartBottom}`;

  return (
    <div className="panel p-5 sm:p-6">
      <div className="mb-5 flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Consumption</p>
          <h2 className="mt-1 text-lg font-semibold text-[#17221d]">Energy usage</h2>
        </div>
        <span className="rounded-full bg-[#e5f2e9] px-3 py-1 text-xs font-semibold text-[#1f6f5b]">Live data</span>
      </div>

      <div className="overflow-hidden">
        <svg viewBox={`0 0 ${width} ${height}`} className="h-auto min-w-[560px] w-full" role="img" aria-label={`${range} energy consumption chart`}>
          <defs>
            <linearGradient id="energy-fill" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stopColor="#77b99c" stopOpacity="0.42" />
              <stop offset="100%" stopColor="#dcefe3" stopOpacity="0.1" />
            </linearGradient>
          </defs>
          {[0, 1, 2, 3].map((gridLine) => {
            const y = chartTop + (gridLine / 3) * (chartBottom - chartTop);
            return <line key={gridLine} x1="0" x2={width} y1={y} y2={y} stroke="#e3ebe4" strokeWidth="1" />;
          })}
          <polygon points={area} fill="url(#energy-fill)" />
          <polyline points={line} fill="none" stroke="#1f6f5b" strokeLinecap="round" strokeLinejoin="round" strokeWidth="4" />
          {points.map(({ x, y }, index) => (
            <circle key={labels[index]} cx={x} cy={y} r="5" fill="#ffffff" stroke="#1f6f5b" strokeWidth="3" />
          ))}
          {labels.map((label, index) => (
            <text key={label} x={points[index].x} y="226" fill="#68766d" fontSize="12" textAnchor={index === 0 ? 'start' : index === labels.length - 1 ? 'end' : 'middle'}>{label}</text>
          ))}
        </svg>
      </div>
      <div className="mt-1 flex items-center justify-between text-xs text-[#68766d]">
        <span>Power draw</span>
        <span>kWh</span>
      </div>
    </div>
  );
}

export default function EnergyPage() {
  const [range, setRange] = useState<EnergyRange>('Today');
  const rangeInfo = rangeData[range];

  return (
    <div className="mx-auto max-w-6xl space-y-8">
      <div>
        <p className="text-xs font-bold uppercase tracking-[0.2em] text-[#1f6f5b]">Resource overview</p>
        <h1 className="page-heading mt-2 text-3xl font-bold text-[#17221d] sm:text-4xl">Energy</h1>
        <p className="mt-2 text-sm text-[#68766d]">Understand where your energy goes and spot opportunities to save.</p>
      </div>

      <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <InfoBox label="Current" value="1.2 kW" change="-8%" />
        <InfoBox label="Today" value={rangeInfo.total} change="-4%" />
        <InfoBox label="This month" value="334 kWh" change="+2%" />
        <InfoBox label="Estimated cost" value="$42.80" change="-$3.20" />
      </section>

      <section className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-sm font-semibold text-[#17221d]">Usage over time</p>
            <p className="text-xs text-[#68766d]">Compare your consumption across different periods.</p>
          </div>
          <div className="flex rounded-xl border border-[#dbe4dc] bg-white p-1 shadow-sm" role="group" aria-label="Energy range">
            {(Object.keys(rangeData) as EnergyRange[]).map((option) => (
              <button
                key={option}
                type="button"
                onClick={() => setRange(option)}
                aria-pressed={range === option}
                className={`rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors sm:px-4 ${range === option ? 'bg-[#1f6f5b] text-white shadow-sm' : 'text-[#68766d] hover:bg-[#eef3ef] hover:text-[#17221d]'}`}
              >
                {option}
              </button>
            ))}
          </div>
        </div>
        <ConsumptionChart range={range} />
      </section>

      <section className="grid gap-6 lg:grid-cols-2">
        <div className="panel p-5">
          <div className="mb-5 flex items-center justify-between">
            <div>
              <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Where it happens</p>
              <h2 className="mt-1 text-lg font-semibold text-[#17221d]">By room</h2>
            </div>
            <span className="text-xs font-medium text-[#68766d]">1.0 kW total</span>
          </div>
          <ul className="space-y-4">
            {rooms.map((room) => (
              <li key={room.name}>
                <div className="mb-1.5 flex items-center justify-between text-sm">
                  <span className="font-medium text-[#31453a]">{room.name}</span>
                  <span className="font-semibold text-[#17221d]">{room.value}</span>
                </div>
                <div className="h-2 overflow-hidden rounded-full bg-[#e7eee8]">
                  <div className="h-full rounded-full bg-[#4c9a7c]" style={{ width: `${room.share}%` }} />
                </div>
              </li>
            ))}
          </ul>
        </div>

        <div className="panel p-5">
          <div className="mb-5 flex items-center justify-between">
            <div>
              <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Highest impact</p>
              <h2 className="mt-1 text-lg font-semibold text-[#17221d]">Top consumers</h2>
            </div>
            <span className="text-xs font-medium text-[#68766d]">Today</span>
          </div>
          <ol className="space-y-1">
            {consumers.map((consumer, index) => (
              <li key={consumer.name} className="flex items-center gap-3 rounded-xl px-2 py-3 hover:bg-[#f1f5f1]">
                <span className="flex h-8 w-8 items-center justify-center rounded-full bg-[#e5f2e9] text-sm font-bold text-[#1f6f5b]">{index + 1}</span>
                <span className="flex-1">
                  <span className="block text-sm font-semibold text-[#17221d]">{consumer.name}</span>
                  <span className="block text-xs text-[#68766d]">{consumer.note}</span>
                </span>
                <span className="text-sm font-bold text-[#17221d]">{consumer.value}</span>
              </li>
            ))}
          </ol>
        </div>
      </section>
    </div>
  );
}