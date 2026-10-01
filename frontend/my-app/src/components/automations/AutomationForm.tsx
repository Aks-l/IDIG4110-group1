'use client';

import { useState } from 'react';
import type { AutomationDraft } from '@/lib/api/types';

type AutomationFormProps = {
  onSubmit: (draft: AutomationDraft) => Promise<void>;
  onCancel: () => void;
};

const categories = ['Comfort', 'Security', 'Energy', 'Notification'] as const;
const triggerTypes = ['Motion detected', 'Door opened', 'Time of day', 'Temperature above'];
const actionCommands = ['Turn on', 'Turn off', 'Set to', 'Send notification'];

const empty: AutomationDraft = {
  name: '',
  description: '',
  category: 'Comfort',
  triggerType: 'Motion detected',
  triggerDetail: '',
  actionCommand: 'Turn on',
  actionTarget: '',
};

export function AutomationForm({ onSubmit, onCancel }: AutomationFormProps) {
  const [draft, setDraft] = useState<AutomationDraft>(empty);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const set = <K extends keyof AutomationDraft>(key: K, value: AutomationDraft[K]) =>
    setDraft((d) => ({ ...d, [key]: value }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!draft.name.trim()) {
      setError('Name is required');
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      await onSubmit(draft);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {/* Name */}
      <div>
        <label className="mb-1 block text-sm font-medium">Name</label>
        <input
          value={draft.name}
          onChange={(e) => set('name', e.target.value)}
          placeholder="e.g. Hallway motion light"
          className="w-full rounded-md border px-3 py-2 text-sm"
          autoFocus
        />
      </div>

      {/* Description */}
      <div>
        <label className="mb-1 block text-sm font-medium">Description</label>
        <input
          value={draft.description}
          onChange={(e) => set('description', e.target.value)}
          placeholder="e.g. Motion in Hallway → turn on Hallway Light at 40%"
          className="w-full rounded-md border px-3 py-2 text-sm"
        />
      </div>

      {/* Category */}
      <div>
        <label className="mb-1 block text-sm font-medium">Category</label>
        <select
          value={draft.category}
          onChange={(e) => set('category', e.target.value as AutomationDraft['category'])}
          className="w-full rounded-md border bg-white px-3 py-2 text-sm"
        >
          {categories.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </div>

      {/* WHEN */}
      <fieldset className="space-y-2 rounded-md border p-3">
        <legend className="px-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          When
        </legend>
        <select
          value={draft.triggerType}
          onChange={(e) => set('triggerType', e.target.value)}
          className="w-full rounded-md border bg-white px-3 py-2 text-sm"
        >
          {triggerTypes.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
        <input
          value={draft.triggerDetail}
          onChange={(e) => set('triggerDetail', e.target.value)}
          placeholder="Detail (e.g. Hallway, 07:00, 25°C)"
          className="w-full rounded-md border px-3 py-2 text-sm"
        />
      </fieldset>

      {/* THEN */}
      <fieldset className="space-y-2 rounded-md border p-3">
        <legend className="px-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Then
        </legend>
        <select
          value={draft.actionCommand}
          onChange={(e) => set('actionCommand', e.target.value)}
          className="w-full rounded-md border bg-white px-3 py-2 text-sm"
        >
          {actionCommands.map((a) => (
            <option key={a} value={a}>{a}</option>
          ))}
        </select>
        <input
          value={draft.actionTarget}
          onChange={(e) => set('actionTarget', e.target.value)}
          placeholder="Target (e.g. Hallway Light, or 40%)"
          className="w-full rounded-md border px-3 py-2 text-sm"
        />
      </fieldset>

      {error && (
        <p className="text-sm text-red-600">{error}</p>
      )}

      <div className="flex justify-end gap-2 border-t pt-4">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md border px-4 py-2 text-sm hover:bg-gray-50"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={submitting}
          className="rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
        >
          {submitting ? 'Creating…' : 'Create'}
        </button>
      </div>
    </form>
  );
}