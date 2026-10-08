import type { NextRequest } from 'next/server';

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

export const dynamic = 'force-dynamic';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ deviceId: string }> },
) {
  const { deviceId } = await params;
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
  return handlePlaceholderRequest(request, async () => {
    const body = await readJsonObject(request);
    if (typeof body.on !== 'boolean') {
      return jsonError(400, 'body must be {"on": true|false}');
    }
    const device = setPlaceholderDeviceState(deviceId, body.on);
    return device ? jsonOk(device) : jsonError(404, `Unknown device: ${deviceId}`);
  });
}
