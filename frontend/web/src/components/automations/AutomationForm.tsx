'use client';

import { useState } from 'react';
import type {
  ActionKind,
  AutomationAction,
  AutomationCategory,
  AutomationCondition,
  AutomationDraft,
  ConditionKind,
  DayOfWeek,
  TemperatureOp,
} from '@/lib/api/types';

type AutomationFormProps = {
  onSubmit: (draft: AutomationDraft) => Promise<void>;
  onCancel: () => void;
};

const categories: AutomationCategory[] = ['Comfort', 'Security', 'Energy', 'Notification'];
const conditionKinds: { kind: ConditionKind; label: string }[] = [
  { kind: 'time',        label: 'Time range' },
  { kind: 'day',         label: 'Day of week' },
  { kind: 'temperature', label: 'Temperature' },
  { kind: 'motion',      label: 'Motion' },
  { kind: 'presence',    label: 'Presence' },
];
const actionKinds: { kind: ActionKind; label: string }[] = [
  { kind: 'device', label: 'Control device' },
  { kind: 'notify', label: 'Send notification' },
  { kind: 'delay',  label: 'Wait / delay' },
];

const ALL_DAYS: DayOfWeek[] = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'];
const DAY_LABELS: Record<DayOfWeek, string> = {
  mon: 'Mon', tue: 'Tue', wed: 'Wed', thu: 'Thu', fri: 'Fri', sat: 'Sat', sun: 'Sun',
};

// --- Defaults for new rows ---

function defaultCondition(kind: ConditionKind): AutomationCondition {
  switch (kind) {
    case 'time':        return { kind: 'time', from: '08:00', to: '18:00' };
    case 'day':         return { kind: 'day', days: ['mon'] };
    case 'temperature': return { kind: 'temperature', op: 'lt', value: 15 };
    case 'motion':      return { kind: 'motion', location: '' };
    case 'presence':    return { kind: 'presence', state: 'home' };
  }
}

function defaultAction(kind: ActionKind): AutomationAction {
  switch (kind) {
    case 'device': return { kind: 'device', deviceName: '', command: 'turn_on' };
    case 'notify': return { kind: 'notify', message: '' };
    case 'delay':  return { kind: 'delay', minutes: 5 };
  }
}

const empty: AutomationDraft = {
  name: '',
  description: '',
  category: 'Comfort',
  conditions: [],
  actions: [],
};

// --- Sub-components ---

function ConditionRow({
  value,
  onChange,
  onRemove,
}: {
  value: AutomationCondition;
  onChange: (c: AutomationCondition) => void;
  onRemove: () => void;
}) {
  const changeKind = (kind: ConditionKind) => onChange(defaultCondition(kind));

  return (
    <div className="rounded-lg border bg-[#fafcfa] p-3">
      <div className="mb-2 flex items-center gap-2">
        <select
          value={value.kind}
          onChange={(e) => changeKind(e.target.value as ConditionKind)}
          className="rounded-md border bg-white px-2 py-1 text-xs font-medium"
        >
          {conditionKinds.map((k) => (
            <option key={k.kind} value={k.kind}>{k.label}</option>
          ))}
        </select>
        <button
          type="button"
          onClick={onRemove}
          className="ml-auto rounded-md px-2 py-1 text-xs text-[#9aa89f] hover:bg-[#f1f5f1] hover:text-red-600"
        >
          Remove
        </button>
      </div>

      {/* Time */}
      {value.kind === 'time' && (
        <div className="flex items-center gap-2">
          <input
            type="time"
            value={value.from}
            onChange={(e) => onChange({ ...value, from: e.target.value })}
            className="rounded-md border px-2 py-1 text-sm"
          />
          <span className="text-sm text-[#68766d]">to</span>
          <input
            type="time"
            value={value.to}
            onChange={(e) => onChange({ ...value, to: e.target.value })}
            className="rounded-md border px-2 py-1 text-sm"
          />
        </div>
      )}

      {/* Day of week */}
      {value.kind === 'day' && (
        <div className="flex flex-wrap gap-1">
          {ALL_DAYS.map((d) => {
            const on = value.days.includes(d);
            return (
              <button
                key={d}
                type="button"
                onClick={() =>
                  onChange({
                    ...value,
                    days: on ? value.days.filter((x) => x !== d) : [...value.days, d],
                  })
                }
                className={`rounded-full px-3 py-1 text-xs font-medium transition ${
                  on
                    ? 'bg-[#1f6f5b] text-white'
                    : 'bg-white text-[#68766d] ring-1 ring-[#dbe4dc] hover:bg-[#eef3ef]'
                }`}
              >
                {DAY_LABELS[d]}
              </button>
            );
          })}
        </div>
      )}

      {/* Temperature */}
      {value.kind === 'temperature' && (
        <div className="flex items-center gap-2">
          <select
            value={value.op}
            onChange={(e) => onChange({ ...value, op: e.target.value as TemperatureOp })}
            className="rounded-md border bg-white px-2 py-1 text-sm"
          >
            <option value="lt">below</option>
            <option value="gt">above</option>
            <option value="eq">at</option>
          </select>
          <input
            type="number"
            value={value.value}
            onChange={(e) => onChange({ ...value, value: Number(e.target.value) })}
            className="w-24 rounded-md border px-2 py-1 text-sm"
          />
          <span className="text-sm text-[#68766d]">°C</span>
        </div>
      )}

      {/* Motion */}
      {value.kind === 'motion' && (
        <input
          value={value.location}
          onChange={(e) => onChange({ ...value, location: e.target.value })}
          placeholder="Location (e.g. Hallway)"
          className="w-full rounded-md border px-2 py-1 text-sm"
        />
      )}

      {/* Presence */}
      {value.kind === 'presence' && (
        <select
          value={value.state}
          onChange={(e) => onChange({ ...value, state: e.target.value as 'home' | 'away' })}
          className="rounded-md border bg-white px-2 py-1 text-sm"
        >
          <option value="home">Someone is home</option>
          <option value="away">Nobody is home</option>
        </select>
      )}
    </div>
  );
}

function ActionRow({
  value,
  onChange,
  onRemove,
}: {
  value: AutomationAction;
  onChange: (a: AutomationAction) => void;
  onRemove: () => void;
}) {
  const changeKind = (kind: ActionKind) => onChange(defaultAction(kind));

  return (
    <div className="rounded-lg border bg-[#fafcfa] p-3">
      <div className="mb-2 flex items-center gap-2">
        <select
          value={value.kind}
          onChange={(e) => changeKind(e.target.value as ActionKind)}
          className="rounded-md border bg-white px-2 py-1 text-xs font-medium"
        >
          {actionKinds.map((k) => (
            <option key={k.kind} value={k.kind}>{k.label}</option>
          ))}
        </select>
        <button
          type="button"
          onClick={onRemove}
          className="ml-auto rounded-md px-2 py-1 text-xs text-[#9aa89f] hover:bg-[#f1f5f1] hover:text-red-600"
        >
          Remove
        </button>
      </div>

      {/* Device */}
      {value.kind === 'device' && (
        <div className="flex flex-wrap items-center gap-2">
          <select
            value={value.command}
            onChange={(e) =>
              onChange({ ...value, command: e.target.value as 'turn_on' | 'turn_off' | 'set' })
            }
            className="rounded-md border bg-white px-2 py-1 text-sm"
          >
            <option value="turn_on">Turn on</option>
            <option value="turn_off">Turn off</option>
            <option value="set">Set to</option>
          </select>
          <input
            value={value.deviceName}
            onChange={(e) => onChange({ ...value, deviceName: e.target.value })}
            placeholder="Device (e.g. Hallway Light)"
            className="flex-1 rounded-md border px-2 py-1 text-sm"
          />
          {value.command === 'set' && (
            <input
              value={value.value ?? ''}
              onChange={(e) => onChange({ ...value, value: e.target.value })}
              placeholder="Value (e.g. 40%)"
              className="w-32 rounded-md border px-2 py-1 text-sm"
            />
          )}
        </div>
      )}

      {/* Notify */}
      {value.kind === 'notify' && (
        <input
          value={value.message}
          onChange={(e) => onChange({ ...value, message: e.target.value })}
          placeholder="Message (e.g. Away mode activated)"
          className="w-full rounded-md border px-2 py-1 text-sm"
        />
      )}

      {/* Delay */}
      {value.kind === 'delay' && (
        <div className="flex items-center gap-2">
          <input
            type="number"
            min={1}
            value={value.minutes}
            onChange={(e) => onChange({ ...value, minutes: Number(e.target.value) })}
            className="w-24 rounded-md border px-2 py-1 text-sm"
          />
          <span className="text-sm text-[#68766d]">minutes</span>
        </div>
      )}
    </div>
  );
}

// --- Main form ---

export function AutomationForm({ onSubmit, onCancel }: AutomationFormProps) {
  const [draft, setDraft] = useState<AutomationDraft>(empty);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const setField = <K extends keyof AutomationDraft>(key: K, value: AutomationDraft[K]) =>
    setDraft((d) => ({ ...d, [key]: value }));

  const addCondition = () =>
    setDraft((d) => ({ ...d, conditions: [...d.conditions, defaultCondition('time')] }));
  const updateCondition = (i: number, c: AutomationCondition) =>
    setDraft((d) => ({
      ...d,
      conditions: d.conditions.map((x, idx) => (idx === i ? c : x)),
    }));
  const removeCondition = (i: number) =>
    setDraft((d) => ({ ...d, conditions: d.conditions.filter((_, idx) => idx !== i) }));

  const addAction = () =>
    setDraft((d) => ({ ...d, actions: [...d.actions, defaultAction('device')] }));
  const updateAction = (i: number, a: AutomationAction) =>
    setDraft((d) => ({ ...d, actions: d.actions.map((x, idx) => (idx === i ? a : x)) }));
  const removeAction = (i: number) =>
    setDraft((d) => ({ ...d, actions: d.actions.filter((_, idx) => idx !== i) }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!draft.name.trim()) return setError('Name is required');
    if (draft.conditions.length === 0) return setError('Add at least one condition');
    if (draft.actions.length === 0) return setError('Add at least one action');

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
          onChange={(e) => setField('name', e.target.value)}
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
          onChange={(e) => setField('description', e.target.value)}
          placeholder="Short summary of what this automation does"
          className="w-full rounded-md border px-3 py-2 text-sm"
        />
      </div>

      {/* Category */}
      <div>
        <label className="mb-1 block text-sm font-medium">Category</label>
        <select
          value={draft.category}
          onChange={(e) => setField('category', e.target.value as AutomationCategory)}
          className="w-full rounded-md border bg-white px-3 py-2 text-sm"
        >
          {categories.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </div>

      {/* IF — conditions */}
      <fieldset className="space-y-2 rounded-md border p-3">
        <legend className="px-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          IF — all must be true
        </legend>
        {draft.conditions.length === 0 && (
          <p className="py-2 text-xs text-[#9aa89f]">No conditions yet — add at least one.</p>
        )}
        {draft.conditions.map((c, i) => (
          <ConditionRow
            key={i}
            value={c}
            onChange={(next) => updateCondition(i, next)}
            onRemove={() => removeCondition(i)}
          />
        ))}
        <button
          type="button"
          onClick={addCondition}
          className="rounded-md border border-dashed border-[#c9d4cb] px-3 py-1.5 text-xs font-medium text-[#1f6f5b] hover:bg-[#eef3ef]"
        >
          + Add condition
        </button>
      </fieldset>

      {/* DO — actions */}
      <fieldset className="space-y-2 rounded-md border p-3">
        <legend className="px-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          DO
        </legend>
        {draft.actions.length === 0 && (
          <p className="py-2 text-xs text-[#9aa89f]">No actions yet — add at least one.</p>
        )}
        {draft.actions.map((a, i) => (
          <ActionRow
            key={i}
            value={a}
            onChange={(next) => updateAction(i, next)}
            onRemove={() => removeAction(i)}
          />
        ))}
        <button
          type="button"
          onClick={addAction}
          className="rounded-md border border-dashed border-[#c9d4cb] px-3 py-1.5 text-xs font-medium text-[#1f6f5b] hover:bg-[#eef3ef]"
        >
          + Add action
        </button>
      </fieldset>

      {error && <p className="text-sm text-red-600">{error}</p>}

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
          className="rounded-md bg-[#1f6f5b] px-4 py-2 text-sm font-semibold text-white hover:bg-[#145442] disabled:opacity-50"
        >
          {submitting ? 'Creating…' : 'Create'}
        </button>
      </div>
    </form>
  );
}