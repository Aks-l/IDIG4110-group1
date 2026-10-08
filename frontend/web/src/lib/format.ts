import type {
  AutomationAction,
  AutomationCondition,
  DayOfWeek,
  RoomMetric,
} from '@/lib/api/types';

/**
 * Format an ISO timestamp as HH:MM in UTC.
 * Pinned to en-GB + UTC so server and client render identical strings.
 */
export function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'UTC',
  });
}

/**
 * Format an ISO timestamp as "DD Mon" in UTC.
 */
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    timeZone: 'UTC',
  });
}

const DAY_LABELS: Record<DayOfWeek, string> = {
  mon: 'Mon', tue: 'Tue', wed: 'Wed', thu: 'Thu', fri: 'Fri', sat: 'Sat', sun: 'Sun',
};

const OP_LABELS = { lt: 'below', gt: 'above', eq: 'at' } as const;

export function describeCondition(c: AutomationCondition): string {
  switch (c.kind) {
    case 'time':
      return `Between ${c.from} and ${c.to}`;
    case 'day':
      return c.days.length === 0
        ? 'Any day'
        : `On ${c.days.map((d) => DAY_LABELS[d]).join(', ')}`;
    case 'temperature':
      return `Temperature ${OP_LABELS[c.op]} ${c.value}°C`;
    case 'motion':
      return `Motion in ${c.location || '—'}`;
    case 'presence':
      return c.state === 'home' ? 'Someone is home' : 'Nobody is home';
  }
}

export function describeAction(a: AutomationAction): string {
  switch (a.kind) {
    case 'device': {
      if (a.command === 'turn_on')  return `Turn on ${a.deviceName || '—'}`;
      if (a.command === 'turn_off') return `Turn off ${a.deviceName || '—'}`;
      return `Set ${a.deviceName || '—'} to ${a.value || '—'}`;
    }
    case 'notify':
      return `Notify: "${a.message || '—'}"`;
    case 'delay':
      return `Wait ${a.minutes} min`;
  }
}

// --- room metrics ----------------------------------------------------------

/**
 * Display value for a room metric with its unit, or an em dash when there is
 * no reading: the room has no sensor for the metric, or none has arrived yet.
 */
export function formatMetricValue(metric: RoomMetric, unit = ''): string {
  return metric.value === null ? '—' : `${metric.value}${unit}`;
}

/**
 * Signed change since the previous reading, for an InfoBox change line.
 * undefined when there is no previous reading to compare against, which
 * hides the change line.
 */
export function formatMetricChange(
  metric: RoomMetric,
  unit = '',
): string | undefined {
  if (metric.change === null) return undefined;
  return `${metric.change >= 0 ? '+' : ''}${metric.change}${unit}`;
}