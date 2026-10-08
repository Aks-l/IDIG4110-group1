import type { NextRequest } from 'next/server';

import { mockEnergy } from '@/lib/api/energy.data';
import {
  handlePlaceholderRequest,
  jsonError,
  jsonOk,
} from '@/lib/api/placeholder-handler';
import { ENERGY_MODES, ENERGY_PERIODS } from '@/lib/api/types';
import type { EnergyMode, EnergyPeriod } from '@/lib/api/types';

export const dynamic = 'force-dynamic';

function isEnergyPeriod(value: string): value is EnergyPeriod {
  return (ENERGY_PERIODS as readonly string[]).includes(value);
}

function isEnergyMode(value: string): value is EnergyMode {
  return (ENERGY_MODES as readonly string[]).includes(value);
}

export async function GET(request: NextRequest) {
  return handlePlaceholderRequest(request, () => {
    const params = request.nextUrl.searchParams;
    const period = params.get('period') ?? 'today';
    const mode = params.get('mode') ?? 'current';

    if (!isEnergyPeriod(period)) {
      return jsonError(400, `period must be one of: ${ENERGY_PERIODS.join(', ')}`);
    }
    if (!isEnergyMode(mode)) {
      return jsonError(400, `mode must be one of: ${ENERGY_MODES.join(', ')}`);
    }

    return jsonOk(mockEnergy[mode][period]);
  });
}
