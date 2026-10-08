import type { NextRequest } from 'next/server';

import { backendErrorResponse } from '@/lib/api/backend';
import {
  handlePlaceholderRequest,
  jsonError,
  jsonOk,
  readJsonObject,
} from '@/lib/api/placeholder-handler';
import {
  findPlaceholderDevice,
  setPlaceholderDeviceState,
} from '@/lib/api/placeholder-store';
import {
  loadTwinHomeState,
  twinHomeStateToDevice,
  twinServes,
} from '@/lib/api/twin';

export const dynamic = 'force-dynamic';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ deviceId: string }> },
) {
  const { deviceId } = await params;

  // Hybrid: serve /devices/:id from the twin-core scaffold when configured.
  if (twinServes('/devices')) {
    try {
      const state = await loadTwinHomeState();
      const device = state ? twinHomeStateToDevice(state, deviceId) : null;
      return device
        ? jsonOk(device)
        : jsonError(404, `Unknown device: ${deviceId}`);
    } catch (error) {
      return backendErrorResponse(error);
    }
  }

  return handlePlaceholderRequest(request, () => {
    const device = findPlaceholderDevice(deviceId);
    return device ? jsonOk(device) : jsonError(404, `Unknown device: ${deviceId}`);
  });
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ deviceId: string }> },
) {
  const { deviceId } = await params;

  // The twin-core scaffold has no command endpoint yet, so toggles are
  // rejected instead of silently served from fixtures.
  if (twinServes('/devices')) {
    return jsonError(501, 'The backend does not implement device commands yet');
  }

  return handlePlaceholderRequest(request, async () => {
    const body = await readJsonObject(request);
    if (typeof body.on !== 'boolean') {
      return jsonError(400, 'body must be {"on": true|false}');
    }
    const device = setPlaceholderDeviceState(deviceId, body.on);
    return device ? jsonOk(device) : jsonError(404, `Unknown device: ${deviceId}`);
  });
}
