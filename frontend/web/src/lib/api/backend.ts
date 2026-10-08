import { jsonError } from './placeholder-handler';
import type {
  Automation,
  AutomationAction,
  AutomationCondition,
  AutomationDraft,
  Event,
  TemperatureOp,
} from './types';

/**
 * Hybrid mode: serve the placeholder paths that the real backend already
 * implements, translated onto the frontend contract, while the remaining
 * paths keep serving fixtures.
 *
 * Enabled with two environment variables (read at request time):
 *
 *   PLACEHOLDER_BACKEND_URL=http://localhost:8083
 *   PLACEHOLDER_BACKEND_PATHS=/automations,/events
 *
 * `/automations` maps onto the rules-engine's /api/v1/rules and `/events`
 * maps onto its /api/v1/incidents. Point PLACEHOLDER_BACKEND_URL at the
 * rules-engine directly until the api-gateway serves the collection paths
 * (it currently 307-redirects them to a trailing-slash path the
 * rules-engine 404s). The mappings below are approximations where the
 * models differ; see docs/frontend/api.md.
 */

const DEFAULT_HOME_ID = '00000000-0000-0000-0000-000000000001';
const DEFAULT_TEMPERATURE_ENTITY = '00000000-0000-0000-0000-000000000001';

/** Gateway paths of the backend endpoints behind the hybrid paths. */
export const RULES_PATH = '/api/v1/rules';
export const INCIDENTS_PATH = '/api/v1/incidents';

export function backendUrl(): string {
  return (process.env.PLACEHOLDER_BACKEND_URL ?? '').trim().replace(/\/+$/, '');
}

export function backendPaths(): string[] {
  return (process.env.PLACEHOLDER_BACKEND_PATHS ?? '')
    .split(',')
    .map((entry) => entry.trim())
    .filter((entry) => entry !== '');
}

/** True when the path (for example "/automations") is served from the backend. */
export function backendServes(path: string): boolean {
  return backendUrl() !== '' && backendPaths().includes(path);
}

/** Home rules are created in when the frontend posts an automation. */
export function backendHomeId(): string {
  return (
    (process.env.PLACEHOLDER_BACKEND_HOME_ID ?? '').trim() || DEFAULT_HOME_ID
  );
}

/**
 * Entity used for temperature conditions. The frontend condition has no
 * entity while the backend match needs one; the default is the first
 * simulated temperature sensor.
 */
export function backendTemperatureEntity(): string {
  return (
    (process.env.PLACEHOLDER_BACKEND_TEMPERATURE_ENTITY ?? '').trim() ||
    DEFAULT_TEMPERATURE_ENTITY
  );
}

// --- transport --------------------------------------------------------------

export class BackendError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'BackendError';
  }
}

/** Fetches JSON from the real backend, throwing BackendError on failure. */
export async function fetchBackendJson<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const base = backendUrl();
  let response: Response;
  try {
    response = await fetch(base + path, {
      ...init,
      headers: {
        accept: 'application/json',
        'content-type': 'application/json',
        ...init?.headers,
      },
      cache: 'no-store',
    });
  } catch {
    throw new BackendError(502, `Backend unreachable: ${base}`);
  }

  if (!response.ok) {
    throw new BackendError(response.status, await backendErrorMessage(response));
  }

  return (await response.json()) as T;
}

/** Reads the { code, message } error body the Go services return. */
async function backendErrorMessage(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { message?: unknown };
    if (typeof body?.message === 'string' && body.message.trim() !== '') {
      return body.message;
    }
  } catch {
    // not JSON; fall through to the generic message
  }
  return `Backend request failed with status ${response.status}`;
}

/** Renders a BackendError from a route handler as the shared error shape. */
export function backendErrorResponse(error: unknown): Response {
  if (error instanceof BackendError) {
    return jsonError(error.status, error.message);
  }
  return jsonError(500, 'Unexpected backend error');
}

// --- backend response shapes ------------------------------------------------

export type BackendOperator = 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte';

export type BackendMatch = {
  gateway_id?: string;
  entity: string;
  operator: BackendOperator;
  value: string | number | boolean;
};

export type BackendAction = {
  type: 'command' | 'incident' | 'notify';
  gateway_id?: string;
  entity?: string;
  command?: string;
  parameters?: Record<string, unknown>;
  severity?: 'info' | 'warning' | 'critical';
  message?: string;
  channels?: string[];
};

export type BackendRule = {
  id: string;
  home_id: string;
  name: string;
  description: string;
  enabled: boolean;
  priority: number;
  trigger: BackendMatch & { for_seconds?: number };
  conditions: BackendMatch[];
  cooldown_seconds: number;
  actions: BackendAction[];
  fire_count: number;
  last_fired_at: string | null;
};

export type BackendRuleInput = {
  home_id: string;
  name: string;
  description: string;
  enabled: boolean;
  priority: number;
  trigger: BackendMatch;
  conditions: BackendMatch[];
  cooldown_seconds: number;
  actions: BackendAction[];
};

export type BackendIncident = {
  id: string;
  rule_id: string;
  home_id: string;
  severity: 'info' | 'warning' | 'critical';
  status: 'open' | 'acknowledged' | 'resolved';
  message: string;
  context: Record<string, unknown> | null;
  triggered_at: string;
  acknowledged_at: string | null;
  resolved_at: string | null;
};

// --- backend -> frontend mapping --------------------------------------------

/**
 * binary_sensor.kitchen_smoke -> "Kitchen Smoke": the display names the
 * frontend uses for condition locations, device names and event locations.
 */
function entityToLocation(entity: string): string {
  const name = entity.split('.').pop() ?? entity;
  return name
    .split(/[_\s-]+/)
    .filter((word) => word !== '')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

function isOn(value: BackendMatch['value']): boolean {
  return value === true || value === 1 || value === 'on';
}

function operatorToOp(operator: BackendOperator): TemperatureOp {
  switch (operator) {
    case 'gt':
    case 'gte':
      return 'gt';
    case 'lt':
    case 'lte':
      return 'lt';
    default:
      return 'eq';
  }
}

/**
 * Maps a backend match onto the closest frontend condition kind. The
 * frontend kinds are a closed set, so numeric matches render as temperature
 * comparisons and other binary matches render as motion.
 */
function matchToCondition(match: BackendMatch): AutomationCondition {
  const entity = match.entity.toLowerCase();

  if (entity.includes('occup') || entity.includes('presence')) {
    return { kind: 'presence', state: isOn(match.value) ? 'home' : 'away' };
  }
  if (entity.includes('motion')) {
    return { kind: 'motion', location: entityToLocation(match.entity) };
  }
  if (typeof match.value === 'number') {
    return {
      kind: 'temperature',
      op: operatorToOp(match.operator),
      value: match.value,
    };
  }
  return { kind: 'motion', location: entityToLocation(match.entity) };
}

function formatParameters(
  command: string,
  parameters?: Record<string, unknown>,
): string {
  const entries = Object.entries(parameters ?? {});
  if (entries.length === 0) return command;
  if (entries.length === 1) return `${command} ${String(entries[0][1])}`;
  return `${command} ${JSON.stringify(parameters)}`;
}

function actionToAutomationAction(action: BackendAction): AutomationAction {
  switch (action.type) {
    case 'command': {
      const deviceName = entityToLocation(action.entity ?? '');
      if (action.command === 'turn_on') {
        return { kind: 'device', deviceName, command: 'turn_on' };
      }
      if (action.command === 'turn_off') {
        return { kind: 'device', deviceName, command: 'turn_off' };
      }
      return {
        kind: 'device',
        deviceName,
        command: 'set',
        value: formatParameters(action.command ?? 'set', action.parameters),
      };
    }
    case 'incident':
      return { kind: 'notify', message: action.message ?? 'Incident raised' };
    case 'notify':
      return {
        kind: 'notify',
        message: `channels: ${(action.channels ?? ['ui']).join(', ')}`,
      };
  }
}

/**
 * Rules that raise incidents are Security, notify-only rules Notification,
 * everything else Comfort; the backend has no category of its own.
 */
function ruleCategory(rule: BackendRule): Automation['category'] {
  if (rule.actions.some((action) => action.type === 'incident')) {
    return 'Security';
  }
  if (
    rule.actions.length > 0 &&
    rule.actions.every((action) => action.type === 'notify')
  ) {
    return 'Notification';
  }
  return 'Comfort';
}

export function ruleToAutomation(rule: BackendRule): Automation {
  return {
    id: rule.id,
    name: rule.name,
    description: rule.description,
    category: ruleCategory(rule),
    enabled: rule.enabled,
    lastRun: rule.last_fired_at ?? undefined,
    runCount: rule.fire_count,
    conditions: [
      matchToCondition(rule.trigger),
      ...rule.conditions.map(matchToCondition),
    ],
    actions: rule.actions.map(actionToAutomationAction),
  };
}

export function incidentToEvent(incident: BackendIncident): Event {
  const entity = incident.context?.entity;
  return {
    id: incident.id,
    timestamp: incident.triggered_at,
    title: incident.message,
    location:
      typeof entity === 'string' && entity !== ''
        ? entityToLocation(entity)
        : undefined,
    severity: incident.severity,
  };
}

// --- frontend -> backend mapping --------------------------------------------

function slug(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
}

/** Device name -> gateway entity, with the domain guessed from the name. */
const DEVICE_DOMAINS: [RegExp, string][] = [
  [/light|lamp/i, 'light'],
  [/lock|door/i, 'lock'],
  [/thermostat|climate|heat/i, 'climate'],
  [/siren|alarm/i, 'siren'],
  [/valve/i, 'valve'],
  [/fan/i, 'fan'],
];

function conditionToMatch(
  condition: AutomationCondition,
): { match: BackendMatch } | { error: string } {
  switch (condition.kind) {
    case 'temperature':
      return {
        match: {
          entity: backendTemperatureEntity(),
          operator: condition.op,
          value: condition.value,
        },
      };
    case 'motion': {
      const location = condition.location.trim();
      if (location === '') return { error: 'motion conditions need a location' };
      return {
        match: {
          entity: `binary_sensor.${slug(location)}_motion`,
          operator: 'eq',
          value: 'on',
        },
      };
    }
    case 'presence':
      return {
        match: {
          entity: 'binary_sensor.home_occupied',
          operator: 'eq',
          value: condition.state === 'home' ? 'on' : 'off',
        },
      };
    case 'time':
    case 'day':
      return {
        error: `the rules backend cannot express ${condition.kind} conditions yet`,
      };
  }
}

function deviceActionToBackend(
  action: Extract<AutomationAction, { kind: 'device' }>,
): { action: BackendAction } | { error: string } {
  const name = action.deviceName.trim();
  if (name === '') return { error: 'device actions need a device name' };
  const domain =
    DEVICE_DOMAINS.find(([keyword]) => keyword.test(name))?.[1] ?? 'switch';
  return {
    action: {
      type: 'command',
      entity: `${domain}.${slug(name)}`,
      command: action.command === 'set' ? 'set' : action.command,
      parameters:
        action.command === 'set' && action.value !== undefined
          ? { value: action.value }
          : undefined,
    },
  };
}

/**
 * Maps an automation draft onto the backend rule input. The frontend has no
 * trigger concept, so the first mappable condition becomes the trigger and
 * the rest become conditions. Constructs the backend cannot express are
 * rejected with an error rather than silently dropped.
 */
export function draftToRuleInput(
  draft: AutomationDraft,
): { input: BackendRuleInput } | { error: string } {
  const matches: BackendMatch[] = [];
  for (const condition of draft.conditions) {
    const mapped = conditionToMatch(condition);
    if ('error' in mapped) return { error: mapped.error };
    matches.push(mapped.match);
  }
  if (matches.length === 0) {
    return {
      error: 'at least one temperature, motion or presence condition is required',
    };
  }

  const actions: BackendAction[] = [];
  let notified = false;
  for (const action of draft.actions) {
    if (action.kind === 'delay') {
      return { error: 'the rules backend cannot express delay actions yet' };
    }
    if (action.kind === 'device') {
      const mapped = deviceActionToBackend(action);
      if ('error' in mapped) return { error: mapped.error };
      actions.push(mapped.action);
      continue;
    }
    // A notify maps to an incident plus a ui notification; the backend
    // allows at most one incident per rule.
    if (notified) {
      return {
        error: 'the rules backend supports at most one notify action per rule',
      };
    }
    notified = true;
    actions.push(
      { type: 'incident', severity: 'info', message: action.message || draft.name },
      { type: 'notify', channels: ['ui'] },
    );
  }
  if (actions.length === 0) {
    return { error: 'at least one device or notify action is required' };
  }

  const [trigger, ...conditions] = matches;
  return {
    input: {
      home_id: backendHomeId(),
      name: draft.name,
      description: draft.description,
      enabled: true,
      priority: 0,
      trigger,
      conditions,
      cooldown_seconds: 0,
      actions,
    },
  };
}
