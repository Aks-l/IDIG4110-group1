import { fetchBaseJson } from './backend';
import { isWarningDevice } from './device-status';
import type {
  Device,
  OverviewData,
  RoomConnection,
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
 *   PLACEHOLDER_TWIN_PATHS=/rooms,/devices,/3d/rooms,/overview
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

/**
 * The home's graph edges (GET /api/v1/homes/{home_id}/relations), for the
 * same home the state view loads. The 3D layout composes the area-to-area
 * connects_to edges with the room geometry.
 */
export async function loadTwinRelations(): Promise<TwinRelation[]> {
  const homeId = await resolveTwinHomeId();
  if (homeId === null) return [];
  return fetchBaseJson<TwinRelation[]>(
    twinUrl(),
    `${TWIN_HOMES_PATH}/${homeId}/relations`,
  );
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

/** One graph edge between two nodes (areas carry the room adjacency). */
export type TwinRelation = {
  id: string;
  home_id?: string | null;
  from_kind: string;
  from_id: string;
  to_kind: string;
  to_id: string;
  relation_type: string;
  bidirectional?: boolean | null;
  label?: string | null;
};

// --- twin -> frontend mapping ------------------------------------------------

/**
 * Reading values display with at most two decimals: sensors report more
 * precision than the UI needs, and subtracting the previous reading
 * surfaces floating point noise (19.8 - 19.5 comes out as
 * 0.3000000000000007).
 */
function round2(value: number): number {
  return Math.round(value * 100) / 100;
}

/** Numeric value of a state; value_text counts when it parses as a number. */
function stateValue(state?: TwinStateValue | null): number | undefined {
  if (!state) return undefined;
  if (typeof state.value_num === 'number') return round2(state.value_num);
  if (typeof state.value_text === 'string' && state.value_text.trim() !== '') {
    const parsed = Number(state.value_text);
    if (Number.isFinite(parsed)) return round2(parsed);
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
    if (typeof value === 'number') return round2(value);
    if (typeof value === 'string') {
      const parsed = Number(value);
      if (Number.isFinite(parsed)) return round2(parsed);
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
 * Metrics without a sensor — or without a reading yet — stay null, and the
 * UI shows them as missing data instead of a misleading 0.
 */
function roomMetrics(entities: TwinEntityState[]): RoomMetrics {
  const metrics: RoomMetrics = {
    temperature: { value: null, change: null },
    humidity: { value: null, change: null },
    co2: { value: null, change: null },
    occupancy: { value: null, change: null },
  };
  for (const key of Object.keys(metrics) as (keyof RoomMetrics)[]) {
    // The first matching sensor that has a reading; matches without one are
    // skipped so a silent sensor does not mask a working one.
    const entity = entities.find(
      (candidate) =>
        matchesMetric(candidate, METRIC_CLASSES[key]) &&
        stateValue(candidate.state) !== undefined,
    );
    const value = entity ? stateValue(entity?.state) : undefined;
    if (value === undefined) continue;
    metrics[key].value = value;
    const previous = stateValue(entity?.previous);
    metrics[key].change = previous === undefined ? null : round2(value - previous);
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

/**
 * Overview data from the home state. The stats are home-level aggregates of
 * the room metrics — temperature and humidity average the rooms that have a
 * reading, air quality comes from the worst CO₂, occupancy is the total —
 * and stats without any reading are omitted, like the rooms' missing data.
 * `connected` is the twin device count and `warning` uses the devices page's
 * smoke/leak heuristic so the two pages agree. `events` is 0: twin-core
 * keeps no event history, only current and previous state.
 */
export function twinHomeStateToOverview(state: TwinHomeState): OverviewData {
  const rooms = twinHomeStateToRooms(state);
  const stats: OverviewData['stats'] = [];

  const metricValues = (key: keyof RoomMetrics) =>
    rooms
      .map((room) => room.metrics[key].value)
      .filter((value): value is number => value !== null);

  const average = (values: number[]): number =>
    round2(values.reduce((sum, value) => sum + value, 0) / values.length);

  const temperatures = metricValues('temperature');
  const humidities = metricValues('humidity');
  const co2s = metricValues('co2');
  const occupancies = metricValues('occupancy');

  if (temperatures.length > 0) {
    stats.push({ label: 'Temperature', value: `${average(temperatures)}°C` });
  }
  if (humidities.length > 0) {
    stats.push({ label: 'Humidity', value: `${average(humidities)}%` });
  }
  if (co2s.length > 0) {
    stats.push({ label: 'Air Quality', value: airQuality(Math.max(...co2s)) });
  }
  if (occupancies.length > 0) {
    const total = occupancies.reduce((sum, value) => sum + value, 0);
    stats.push({
      label: 'Occupancy',
      value: `${total} ${total === 1 ? 'person' : 'people'}`,
    });
  }

  const devices = twinHomeStateToDevices(state);
  return {
    stats,
    devices: {
      connected: devices.length,
      warning: devices.filter(isWarningDevice).length,
      events: 0,
    },
  };
}

/** Rough air-quality label from the home's worst CO₂ reading (ppm). */
function airQuality(worstCo2: number): string {
  if (worstCo2 < 800) return 'Good';
  if (worstCo2 < 1200) return 'Moderate';
  return 'Poor';
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
 * grid so the 3D scene has something to place. Geometry y is the floor
 * level (the seed keeps its single-storey rooms at 0), while the scene
 * places boxes by their center on the ground plane, so the mapping lifts
 * y by half the height.
 */
export function twinHomeStateToRoomLayouts(state: TwinHomeState): RoomLayout[] {
  return state.areas.map((area, index) => {
    const floor: [number, number, number] = asTriple(
      area.geometry?.position,
    ) ?? [(index % 2) * 5, 0, Math.floor(index / 2) * 5];
    const size: [number, number, number] = asTriple(area.geometry?.size) ?? [
      3.5,
      2.6,
      3.5,
    ];
    return {
      roomId: area.id,
      position: [floor[0], floor[1] + size[1] / 2, floor[2]],
      size,
    };
  });
}

/** Passage height and floor offset, matching the mock passage blocks. */
const CONNECTION_HEIGHT = 0.6;
const CONNECTION_Y = 0.3;
/** The bridge slightly overlaps both rooms so no hairline gap remains. */
const CONNECTION_OVERLAP = 0.1;

/** Opening width per relation label; other labels open the full shared span. */
const OPENING_WIDTHS: Record<string, number> = {
  doorway: 1.2,
  archway: 2.2,
};

/** Interval a room covers on one horizontal axis of its layout triple. */
function axisSpan(layout: RoomLayout, axis: 0 | 2): [number, number] {
  return [
    layout.position[axis] - layout.size[axis] / 2,
    layout.position[axis] + layout.size[axis] / 2,
  ];
}

/** Distance between two rooms on one horizontal axis; >= 0 when apart. */
function axisGap(from: RoomLayout, to: RoomLayout, axis: 0 | 2): number {
  const [fromMin, fromMax] = axisSpan(from, axis);
  const [toMin, toMax] = axisSpan(to, axis);
  return Math.max(fromMin - toMax, toMin - fromMax);
}

/**
 * Box for the passage between two rooms: it bridges the gap on the axis
 * the rooms face each other on, and opens the shared wall on the other,
 * sized by the relation label. Null when the rooms share no wall (apart
 * on both horizontal axes, or overlapping volumes).
 */
function passageBox(
  from: RoomLayout,
  to: RoomLayout,
  label: string | undefined,
): Pick<RoomConnection, 'position' | 'size'> | null {
  const gapX = axisGap(from, to, 0);
  const gapZ = axisGap(from, to, 2);

  let facing: 0 | 2;
  if (gapX >= 0 && gapZ >= 0) facing = gapX <= gapZ ? 0 : 2;
  else if (gapX >= 0) facing = 0;
  else if (gapZ >= 0) facing = 2;
  else return null;
  const across: 0 | 2 = facing === 0 ? 2 : 0;

  // Gap between the near faces on the facing axis.
  const gapStart = Math.min(axisSpan(from, facing)[1], axisSpan(to, facing)[1]);
  const gapEnd = Math.max(axisSpan(from, facing)[0], axisSpan(to, facing)[0]);
  // Shared wall interval on the across axis.
  const acrossStart = Math.max(
    axisSpan(from, across)[0],
    axisSpan(to, across)[0],
  );
  const acrossEnd = Math.min(
    axisSpan(from, across)[1],
    axisSpan(to, across)[1],
  );
  if (acrossEnd <= acrossStart) return null;

  const opening =
    OPENING_WIDTHS[label?.toLowerCase() ?? ''] ?? Number.POSITIVE_INFINITY;

  const position: [number, number, number] = [0, CONNECTION_Y, 0];
  const size: [number, number, number] = [0, CONNECTION_HEIGHT, 0];
  position[facing] = (gapStart + gapEnd) / 2;
  size[facing] = Math.max(gapEnd - gapStart + CONNECTION_OVERLAP, 0.3);
  position[across] = (acrossStart + acrossEnd) / 2;
  size[across] = Math.min(acrossEnd - acrossStart, opening);
  return { position, size };
}

/**
 * Passages for the 3D scene from the room layouts and the home's relations:
 * every area-to-area `connects_to` edge becomes a floor-level box bridging
 * the two rooms (doorway, archway, open plan), which is how the seed models
 * the demo home's room adjacency.
 */
export function twinConnections(
  layouts: RoomLayout[],
  relations: TwinRelation[],
): RoomConnection[] {
  const layoutById = new Map(layouts.map((layout) => [layout.roomId, layout]));
  const connections: RoomConnection[] = [];
  for (const relation of relations) {
    if (
      relation.relation_type !== 'connects_to' ||
      relation.from_kind !== 'area' ||
      relation.to_kind !== 'area'
    ) {
      continue;
    }
    const from = layoutById.get(relation.from_id);
    const to = layoutById.get(relation.to_id);
    if (!from || !to) continue;
    const box = passageBox(from, to, relation.label ?? undefined);
    if (!box) continue;
    connections.push({
      fromRoomId: from.roomId,
      toRoomId: to.roomId,
      label: relation.label ?? undefined,
      ...box,
    });
  }
  return connections;
}