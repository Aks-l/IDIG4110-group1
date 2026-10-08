import { fetchBaseJson } from './backend';
import type {
  Device,
  RoomData,
  RoomLayout,
  RoomMetrics,
} from './types';

/**
 * Hybrid mode for the twin-core scaffold (branch
 * 47-scaffold-twin-core-service, documented in
 * docs/architecture/twin-state-api.md): serve the placeholder paths it can
 * back, translated onto the frontend contract.
 *
 * Enabled with two environment variables (read at request time):
 *
 *   PLACEHOLDER_TWIN_URL=http://localhost:8084
 *   PLACEHOLDER_TWIN_PATHS=/rooms,/devices,/3d/rooms
 *
 * Everything reads one dashboard view, GET /api/v1/homes/{home_id}/state:
 * areas and devices with their entities and current state. The home is
 * PLACEHOLDER_TWIN_HOME_ID when set, otherwise the first home from
 * GET /api/v1/homes. The mappings below are approximations where the
 * models differ; see docs/frontend/api.md.
 */

const DEFAULT_TWIN_HOME_ID = '00000000-0000-0000-0000-000000000001';

/** Twin-core paths used by the hybrid. */
export const TWIN_HOMES_PATH = '/api/v1/homes';

export function twinUrl(): string {
  return (process.env.PLACEHOLDER_TWIN_URL ?? '').trim().replace(/\/+$/, '');
}

export function twinPaths(): string[] {
  return (process.env.PLACEHOLDER_TWIN_PATHS ?? '')
    .split(',')
    .map((entry) => entry.trim())
    .filter((entry) => entry !== '');
}

/** True when the path (for example "/rooms") is served from twin-core. */
export function twinServes(path: string): boolean {
  return twinUrl() !== '' && twinPaths().includes(path);
}

/**
 * Resolves the home whose state the hybrid serves: the configured
 * PLACEHOLDER_TWIN_HOME_ID, or the first home twin-core knows about.
 */
export async function resolveTwinHomeId(): Promise<string | null> {
  const configured = (process.env.PLACEHOLDER_TWIN_HOME_ID ?? '').trim();
  if (configured !== '') return configured;

  const homes = await fetchBaseJson<TwinHome[]>(twinUrl(), TWIN_HOMES_PATH);
  return homes[0]?.id ?? null;
}

/**
 * Loads the dashboard view of the configured home. Returns null when the
 * backend has no homes yet, so callers can serve empty results.
 */
export async function loadTwinHomeState(): Promise<TwinHomeState | null> {
  const homeId = await resolveTwinHomeId();
  if (homeId === null) return null;
  return fetchBaseJson<TwinHomeState>(twinUrl(), `${TWIN_HOMES_PATH}/${homeId}/state`);
}

// --- twin-core response shapes ----------------------------------------------

/** Latest known value of an entity; exactly one of value_num/value_text. */
export type TwinStateValue = {
  value_num?: number | null;
  value_text?: string | null;
  attributes?: Record<string, unknown> | null;
  updated_at: string;
};

/** Entity joined with its latest state (state/previous null until read). */
export type TwinEntityState = {
  entity_id: string;
  external_entity_id: string;
  name: string;
  domain: string;
  controllable: boolean;
  device_class?: string;
  unit?: string;
  area_id?: string | null;
  device_id: string;
  home_id: string;
  gateway_id: string;
  state?: TwinStateValue | null;
  previous?: TwinStateValue | null;
};

/** Room or zone within a home; geometry is free-form (any in the backend). */
export type TwinArea = {
  id: string;
  home_id: string;
  name: string;
  floor?: number | null;
  area_type?: string | null;
  geometry?: {
    position?: number[];
    size?: number[];
  } | null;
};

/** Normalized device model; area_id is null until it is assigned to a room. */
export type TwinDevice = {
  id: string;
  home_id: string;
  area_id?: string | null;
  gateway_id: string;
  external_id: string;
  name: string;
  manufacturer?: string | null;
  model?: string | null;
  sw_version?: string | null;
  device_type?: string | null;
};

export type TwinHome = {
  id: string;
  name: string;
  address?: string | null;
  timezone: string;
};

/** Area with its registered entities. */
export type TwinAreaState = TwinArea & { entities: TwinEntityState[] };

/** Device with its entities. */
export type TwinDeviceState = TwinDevice & { entities: TwinEntityState[] };

/** Dashboard view of one home. */
export type TwinHomeState = {
  home: TwinHome;
  areas: TwinAreaState[];
  devices: TwinDeviceState[];
};

// --- twin -> frontend mapping ------------------------------------------------

/** Numeric value of a state; value_text counts when it parses as a number. */
function stateValue(state?: TwinStateValue | null): number | undefined {
  if (!state) return undefined;
  if (typeof state.value_num === 'number') return state.value_num;
  if (typeof state.value_text === 'string' && state.value_text.trim() !== '') {
    const parsed = Number(state.value_text);
    if (Number.isFinite(parsed)) return parsed;
  }
  return undefined;
}

function attributeNumber(
  entity: TwinEntityState,
  keys: string[],
): number | undefined {
  const attributes = entity.state?.attributes;
  if (!attributes) return undefined;
  for (const key of keys) {
    const value = attributes[key];
    if (typeof value === 'number') return value;
    if (typeof value === 'string') {
      const parsed = Number(value);
      if (Number.isFinite(parsed)) return parsed;
    }
  }
  return undefined;
}

const METRIC_CLASSES: Record<keyof RoomMetrics, string[]> = {
  temperature: ['temperature'],
  humidity: ['humidity'],
  co2: ['carbon_dioxide', 'co2'],
  occupancy: ['occupancy', 'presence'],
};

function matchesMetric(entity: TwinEntityState, classes: string[]): boolean {
  const haystack =
    `${entity.device_class ?? ''} ${entity.name} ${entity.external_entity_id}`.toLowerCase();
  return classes.some((entityClass) => haystack.includes(entityClass));
}

/**
 * Room metrics from the area's entities: the current value of the first
 * sensor matching each metric, plus the delta against its previous reading.
 * Metrics without a sensor read 0 until the backend reports them.
 */
function roomMetrics(entities: TwinEntityState[]): RoomMetrics {
  const metrics: RoomMetrics = {
    temperature: { value: 0, change: 0 },
    humidity: { value: 0, change: 0 },
    co2: { value: 0, change: 0 },
    occupancy: { value: 0, change: 0 },
  };
  for (const key of Object.keys(metrics) as (keyof RoomMetrics)[]) {
    const entity = entities.find((candidate) =>
      matchesMetric(candidate, METRIC_CLASSES[key]),
    );
    const value = entity ? stateValue(entity?.state) : undefined;
    if (value === undefined) continue;
    metrics[key].value = value;
    metrics[key].change = value - (stateValue(entity?.previous) ?? value);
  }
  return metrics;
}

function isOn(entity: TwinEntityState): boolean {
  const state = entity.state;
  if (!state) return false;
  if (typeof state.value_num === 'number') return state.value_num > 0;
  const text = state.value_text?.toLowerCase();
  return text === 'on' || text === 'true' || text === '1' || text === 'open';
}

/**
 * twin device with its entities -> frontend Device. Fields without a
 * backend source (installedAt, position) are omitted; devices without a
 * controllable entity read as off.
 */
function twinDeviceToDevice(device: TwinDeviceState): Device {
  const controlEntity = device.entities.find((entity) => entity.controllable);
  const powerEntity = device.entities.find((entity) =>
    matchesMetric(entity, ['power']),
  );
  const batteryEntity = device.entities.find((entity) =>
    matchesMetric(entity, ['battery']),
  );

  const lastSeen = device.entities
    .map((entity) => entity.state?.updated_at)
    .filter((value): value is string => typeof value === 'string')
    .sort()
    .pop();

  const signal = device.entities
    .map((entity) =>
      attributeNumber(entity, ['signal', 'rssi', 'linkquality']),
    )
    .find((value) => value !== undefined);

  const domain = device.entities[0]?.domain;
  const type = device.device_type
    ? device.device_type
    : domain
      ? domain
          .split('_')
          .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
          .join(' ')
      : 'Device';

  return {
    id: device.id,
    name: device.name,
    type,
    on: controlEntity ? isOn(controlEntity) : false,
    roomId: device.area_id ?? undefined,
    powerWatts: powerEntity ? stateValue(powerEntity.state) : undefined,
    signal,
    battery: batteryEntity
      ? stateValue(batteryEntity.state) ??
        attributeNumber(batteryEntity, ['battery'])
      : undefined,
    firmware: device.sw_version ?? undefined,
    lastSeen,
  };
}

/**
 * Rooms from the home state areas. Activity stays empty: twin-core keeps no
 * activity history, only current and previous state.
 */
export function twinHomeStateToRooms(state: TwinHomeState): RoomData[] {
  return state.areas.map((area) => ({
    id: area.id,
    name: area.name,
    metrics: roomMetrics(area.entities),
    devices: state.devices
      .filter((device) => device.area_id === area.id)
      .map(twinDeviceToDevice),
    activity: [],
  }));
}

export function twinHomeStateToRoom(
  state: TwinHomeState,
  roomId: string,
): RoomData | null {
  return (
    twinHomeStateToRooms(state).find((room) => room.id === roomId) ?? null
  );
}

export function twinHomeStateToDevices(state: TwinHomeState): Device[] {
  return state.devices.map(twinDeviceToDevice);
}

export function twinHomeStateToDevice(
  state: TwinHomeState,
  deviceId: string,
): Device | null {
  const device = state.devices.find((candidate) => candidate.id === deviceId);
  return device ? twinDeviceToDevice(device) : null;
}

function asTriple(value: unknown): [number, number, number] | null {
  if (
    Array.isArray(value) &&
    value.length === 3 &&
    value.every((entry) => typeof entry === 'number')
  ) {
    return [value[0], value[1], value[2]];
  }
  return null;
}

/**
 * Room layouts from the areas' geometry when it carries position and size
 * ({ position: [x, y, z], size: [w, h, d] }); otherwise a deterministic
 * grid so the 3D scene has something to place.
 */
export function twinHomeStateToRoomLayouts(state: TwinHomeState): RoomLayout[] {
  return state.areas.map((area, index) => ({
    roomId: area.id,
    position: asTriple(area.geometry?.position) ?? [
      (index % 2) * 5,
      0,
      Math.floor(index / 2) * 5,
    ],
    size: asTriple(area.geometry?.size) ?? [3.5, 2.6, 3.5],
  }));
}