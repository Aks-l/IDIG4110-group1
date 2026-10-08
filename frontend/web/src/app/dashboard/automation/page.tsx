'use client';

import { useEffect, useState } from 'react';
import { InfoBox } from '@/components/InfoBox';
import { Toggle } from '@/components/Toggle';
import { Modal } from '@/components/Modal';
import { AutomationForm } from '@/components/automations/AutomationForm';
import { loadAutomations, createAutomation, setAutomationState } from '@/lib/api/automations.data';
import type { Automation, AutomationDraft } from '@/lib/api/types';
import { describeAction, describeCondition } from '@/lib/format';

// --- Helper ---

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
  const [modalOpen, setModalOpen] = useState(false);

  useEffect(() => {
    loadAutomations()
      .then(setAutomations)
      .catch(() => setError('Unable to load automations'));
  }, []);

  const toggle = async (id: string) => {
    const current = automations.find((automation) => automation.id === id);
    if (!current) return;
    const enabled = !current.enabled;
    setAutomations((list) => list.map((automation) => automation.id === id ? { ...automation, enabled } : automation));
    try {
      const updated = await setAutomationState(id, enabled);
      setAutomations((list) => list.map((automation) => automation.id === id ? updated : automation));
    } catch {
      setAutomations((list) => list.map((automation) => automation.id === id ? { ...automation, enabled: current.enabled } : automation));
    }
  };

  const handleCreate = async (draft: AutomationDraft) => {
    const created = await createAutomation(draft);
    setAutomations((list) => [created, ...list]);
    setModalOpen(false);
  };

  // Summary numbers
  const activeCount = automations.filter((a) => a.enabled).length;
  const disabledCount = automations.filter((a) => !a.enabled).length;
  const totalRuns = automations.reduce((sum, a) => sum + a.runCount, 0);
  const todayRuns = automations.filter(
    (a) =>
      a.lastRun &&
      new Date(a.lastRun).toDateString() === new Date().toDateString()
  ).length;

  if (error) return <p className="text-red-600">{error}</p>;

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <h1 className="page-heading text-3xl font-bold text-[#17221d]">Automations</h1>
        <button
          onClick={() => setModalOpen(true)}
          className="rounded-xl bg-[#1f6f5b] px-4 py-2 text-sm font-semibold text-white shadow-[0_8px_18px_rgba(31,111,91,0.2)] hover:bg-[#145442]"
        >
          + New Automation
        </button>
      </div>

      <hr className="soft-divider" />

      {/* Summary stats */}
      <div className="grid grid-cols-4 gap-3">
        <InfoBox label="Active"     value={String(activeCount)}   />
        <InfoBox label="Disabled"   value={String(disabledCount)} />
        <InfoBox label="Total runs" value={String(totalRuns)}     />
        <InfoBox label="Ran today"  value={String(todayRuns)}     />
      </div>

      <hr className="soft-divider" />

      {/* Automation list */}
      <section className="space-y-3">
        <ul className="space-y-2">
          {automations.map((a) => (
          <li key={a.id} className="panel flex items-start gap-4 px-4 py-3">
            {/* Toggle */}
            <Toggle
              checked={a.enabled}
              onChange={() => toggle(a.id)}
              label={a.enabled ? 'Disable automation' : 'Enable automation'}
            />

            {/* Body */}
            <div className="flex flex-1 flex-col gap-2">
              <div className="flex items-center gap-2">
                <span className="font-medium text-[#17221d]">{a.name}</span>
                <span
                  className={`rounded-full px-2 py-0.5 text-xs font-medium ${categoryStyles[a.category]}`}
                >
                  {a.category}
                </span>
              </div>

              {a.description && (
                <span className="text-sm text-[#68766d]">{a.description}</span>
              )}

              {/* IF / DO summary */}
              <div className="flex flex-col gap-1 text-xs text-[#68766d]">
                {a.conditions.length > 0 && (
                  <div className="flex gap-2">
                    <span className="w-8 shrink-0 font-semibold uppercase tracking-wide text-[#1f6f5b]">
                      IF
                    </span>
                    <ul className="space-y-0.5">
                      {a.conditions.map((c, i) => (
                        <li key={i}>{describeCondition(c)}</li>
                      ))}
                    </ul>
                  </div>
                )}
                {a.actions.length > 0 && (
                  <div className="flex gap-2">
                    <span className="w-8 shrink-0 font-semibold uppercase tracking-wide text-[#1f6f5b]">
                      
                    </span>
                    <ul className="space-y-0.5">
                      {a.actions.map((act, i) => (
                        <li key={i}>{describeAction(act)}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>

              <span className="text-xs text-[#9aa89f]">
                {a.lastRun ? `Last run: ${timeAgo(a.lastRun)}` : 'Never run'}
                {' · '}
                {a.runCount} runs
              </span>
            </div>
          </li>
            ))}
        </ul>

        {automations.length === 0 && (
          <p className="rounded-lg border bg-white px-4 py-6 text-center text-sm text-gray-500">
            No automations yet. Click <strong>+ New Automation</strong> to add one.
          </p>
        )}
      </section>

      {/* Create modal */}
      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title="New Automation"
      >
        <AutomationForm
          onSubmit={handleCreate}
          onCancel={() => setModalOpen(false)}
        />
      </Modal>
    </div>
  );
}