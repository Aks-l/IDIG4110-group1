'use client';

import { useState } from 'react';
import { SubNav } from '@/components/SubNav';
import { InfoBox } from '@/components/InfoBox';

// --- Types ---

type Stat = { label: string; value: string; change: string };
type Device = { id: string; name: string; type: string; on: boolean };
type Activity = { id: string; timestamp: string; description: string };

type RoomData = {
  stats: Stat[];
  devices: Device[];
  activity: Activity[];
};

// TEMPORARY HARDCODED DATA

const ROOMS = ['Living Room', 'Kitchen', 'Bedroom', 'Bathroom'] as const;
type Room = (typeof ROOMS)[number];

const ROOM_DATA: Record<Room, RoomData> = {
  'Living Room': {
    stats: [
      { label: 'Temperature', value: '23°C',     change: '+1°C'    },
      { label: 'Humidity',    value: '48%',      change: '+2%'     },
      { label: 'CO₂',         value: '620 ppm',  change: '+10 ppm' },
      { label: 'Occupancy',   value: '2 people', change: '+1'      },
    ],
    devices: [
      { id: '1', name: 'Ceiling Light', type: 'Light',   on: true  },
      { id: '2', name: 'Floor Lamp',    type: 'Light',   on: false },
      { id: '3', name: 'Thermostat',    type: 'Climate', on: true  },
      { id: '4', name: 'Smart TV',      type: 'Media',   on: false },
      { id: '5', name: 'Window Sensor', type: 'Sensor',  on: true  },
    ],
    activity: [
      { id: '1', timestamp: '2026-09-15T14:32:00', description: 'Window closed'       },
      { id: '2', timestamp: '2026-09-15T14:20:00', description: 'Light turned on'     },
      { id: '3', timestamp: '2026-09-15T13:58:00', description: 'Thermostat adjusted' },
    ],
  },

  'Kitchen': {
    stats: [
      { label: 'Temperature', value: '24°C',     change: '+2°C'    },
      { label: 'Humidity',    value: '55%',      change: '+3%'     },
      { label: 'CO₂',         value: '700 ppm',  change: '+15 ppm' },
      { label: 'Occupancy',   value: '1 person', change: '-1'      },
    ],
    devices: [
      { id: '1', name: 'Oven',         type: 'Appliance', on: true  },
      { id: '2', name: 'Fridge',       type: 'Appliance', on: true  },
      { id: '3', name: 'Range Hood',   type: 'Appliance', on: false },
      { id: '4', name: 'Smoke Sensor', type: 'Sensor',    on: true  },
    ],
    activity: [
      { id: '1', timestamp: '2026-09-15T14:40:00', description: 'Oven turned on'  },
      { id: '2', timestamp: '2026-09-15T14:12:00', description: 'Smoke detected'  },
      { id: '3', timestamp: '2026-09-15T13:55:00', description: 'Fridge door opened' },
    ],
  },

  'Bedroom': {
    stats: [
      { label: 'Temperature', value: '21°C',     change: '-1°C'   },
      { label: 'Humidity',    value: '45%',      change: '0%'     },
      { label: 'CO₂',         value: '500 ppm',  change: '-5 ppm' },
      { label: 'Occupancy',   value: '0 people', change: '-2'     },
    ],
    devices: [
      { id: '1', name: 'Bedside Lamp', type: 'Light',  on: false },
      { id: '2', name: 'Ceiling Fan',  type: 'Fan',    on: false },
      { id: '3', name: 'Window Sensor',type: 'Sensor', on: true  },
    ],
    activity: [
      { id: '1', timestamp: '2026-09-15T13:30:00', description: 'Window closed'  },
      { id: '2', timestamp: '2026-09-15T11:15:00', description: 'Lamp turned off' },
    ],
  },

  'Bathroom': {
    stats: [
      { label: 'Temperature', value: '25°C',     change: '+3°C'   },
      { label: 'Humidity',    value: '72%',      change: '+8%'    },
      { label: 'CO₂',         value: '450 ppm',  change: '0 ppm'  },
      { label: 'Occupancy',   value: '1 person', change: '+1'     },
    ],
    devices: [
      { id: '1', name: 'Mirror Light',    type: 'Light',  on: true  },
      { id: '2', name: 'Exhaust Fan',     type: 'Fan',    on: true  },
      { id: '3', name: 'Leak Sensor',     type: 'Sensor', on: true  },
    ],
    activity: [
      { id: '1', timestamp: '2026-09-15T14:50:00', description: 'Exhaust fan on' },
      { id: '2', timestamp: '2026-09-15T14:45:00', description: 'Door opened'    },
    ],
  },
};

// --- Helper ---

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

// --- Page ---

export default function RoomsPage() {
  const [currentRoom, setCurrentRoom] = useState<Room>('Living Room');
  const data = ROOM_DATA[currentRoom];

  return (
    <div className="space-y-6">
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <SubNav />

        <select
          value={currentRoom}
          onChange={(e) => setCurrentRoom(e.target.value as Room)}
          className="rounded-md border bg-white px-2 py-1 text-sm font-semibold"
        >
          {ROOMS.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </select>
      </div>

      <hr className="border-gray-200" />

      {/* Info boxes */}
      <div className="grid grid-cols-4 gap-3">
        {data.stats.map((s) => (
          <InfoBox key={s.label} label={s.label} value={s.value} change={s.change} />
        ))}
      </div>

      <hr className="border-gray-200" />

      {/* Two-column bottom section */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Devices */}
        <section className="rounded-lg border bg-white p-4 shadow-sm">
          <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-gray-500">
            Devices
          </h2>
          <ul className="space-y-2">
            {data.devices.map((d) => (
              <li
                key={d.id}
                className="flex items-center justify-between rounded-md bg-gray-50 px-4 py-3"
              >
                <div className="flex flex-col">
                  <span className="font-medium">{d.name}</span>
                  <span className="text-xs text-gray-500">{d.type}</span>
                </div>
                <span
                  className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    d.on ? 'bg-green-100 text-green-700' : 'bg-gray-200 text-gray-600'
                  }`}
                >
                  {d.on ? 'On' : 'Off'}
                </span>
              </li>
            ))}
          </ul>
        </section>

        {/* Recent activity */}
        <section className="rounded-lg border bg-white p-4 shadow-sm">
          <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-gray-500">
            Recent Activity
          </h2>
          <ul className="space-y-2">
            {data.activity.map((a) => (
              <li
                key={a.id}
                className="flex items-center justify-between rounded-md bg-gray-50 px-4 py-3"
              >
                <span className="text-sm">{a.description}</span>
                <span className="text-xs text-gray-500">{formatTime(a.timestamp)}</span>
              </li>
            ))}
          </ul>
        </section>
      </div>
    </div>
  );
}