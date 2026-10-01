'use client';

import { useEffect, useState } from 'react';
import { InfoBox } from '@/components/InfoBox';
import { Toggle } from '@/components/Toggle';
import { loadAutomations } from '@/lib/api/automations.data';
import type { Automation } from '@/lib/api/types';

// --- Helper ---

// helper function to format ISO timestamp used on automation tasks 
function timeAgo(iso: string) {
  const diff = (Date.now() - new Date(iso).getTime()) / 1000;
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)} min ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} h ago`;
  return `${Math.floor(diff / 86400)} d ago`;
}

const categoryStyles: Record<Automation['category'], string> = {
  Comfort:      'bg-blue-50 text-blue-700',
  Security:     'bg-red-50 text-red-700',
  Energy:       'bg-green-50 text-green-700',
  Notification: 'bg-purple-50 text-purple-700',
};

// --- Page ---

export default function AutomationsPage() {
  const [automations, setAutomations] = useState<Automation[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadAutomations()
      .then(setAutomations)
      .catch(() => setError('Unable to load automations'));
  }, []);

  const toggle = (id: string) =>
    setAutomations((list) =>
      list.map((a) => (a.id === id ? { ...a, enabled: !a.enabled } : a))
    );

  // Summary numbers
  const activeCount = automations.filter((a) => a.enabled).length;
  const disabledCount = automations.filter((a) => !a.enabled).length;
  const totalRuns = automations.reduce((sum, a) => sum + a.runCount, 0);
  const todayRuns = automations.filter(
    (a) => a.lastRun && new Date(a.lastRun).toDateString() === new Date().toDateString()
  ).length;

  if (error) return <p className="text-red-600">{error}</p>;

  return (
    <div className="space-y-6">
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Automations</h1>
        <button className="rounded-md bg-black px-3 py-1.5 text-sm font-medium text-white hover:bg-gray-800">
          + New Automation
        </button>
      </div>

      <hr className="border-gray-200" />

      {/* Summary stats */}
      <div className="grid grid-cols-4 gap-3">
        <InfoBox label="Active"   value={String(activeCount)}   />
        <InfoBox label="Disabled" value={String(disabledCount)} />
        <InfoBox label="Total runs" value={String(totalRuns)}   />
        <InfoBox label="Ran today"  value={String(todayRuns)}   />
      </div>

      <hr className="border-gray-200" />

      {/* Automation list */}
      <section className="space-y-3">
        <ul className="space-y-2">
          {automations.map((a) => (
            <li
              key={a.id}
              className="flex items-center gap-4 rounded-lg border bg-white px-4 py-3 shadow-sm"
            >
              {/* Left: name, description, meta */}
              <div className="flex flex-1 flex-col">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{a.name}</span>
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                      categoryStyles[a.category]
                    }`}
                  >
                    {a.category}
                  </span>
                </div>
                <span className="text-sm text-gray-500">{a.description}</span>
                <span className="mt-1 text-xs text-gray-400">
                  {a.lastRun ? `Last run: ${timeAgo(a.lastRun)}` : 'Never run'}
                  {' · '}
                  {a.runCount} runs
                </span>
              </div>

              {/* Right: toggle */}
              {/* Right: toggle */}
              <Toggle
                checked={a.enabled}
                onChange={() => toggle(a.id)}
                label={a.enabled ? 'Disable automation' : 'Enable automation'}
              />
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}