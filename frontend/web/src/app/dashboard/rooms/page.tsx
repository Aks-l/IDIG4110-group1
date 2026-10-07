'use client';

import { useEffect, useState } from 'react';
import { RoomView, SubNav } from '@/components/SubNav';
import { InfoBox } from '@/components/InfoBox';
import { loadRoom, loadRooms } from '@/lib/api/rooms.data';
import { formatTime } from '@/lib/format';
import type { RoomData } from '@/lib/api/types';


// --- Page ---

export default function RoomsPage() {
  const [rooms, setRooms] = useState<RoomData[]>([]);
  const [currentRoom, setCurrentRoom] = useState('');
  const [view, setView] = useState<RoomView>('Room Overview');
  const [data, setData] = useState<RoomData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadRooms()
      .then((loadedRooms) => {
        setRooms(loadedRooms);
        setCurrentRoom((selectedRoom) =>
          loadedRooms.some((room) => room.id === selectedRoom)
            ? selectedRoom
            : loadedRooms[0]?.id ?? '',
        );
      })
      .catch(() => setError('Unable to load rooms'));
  }, []);

  useEffect(() => {
    if (!currentRoom) return;

    loadRoom(currentRoom)
      .then(setData)
      .catch(() => setError('Unable to load room data'));
  }, [currentRoom]);

  if (error) return <p className="text-red-600">{error}</p>;
  if (!data) return <p>Loading room data...</p>;

  const deviceList = (
    <section className="panel flex max-h-[420px] flex-col p-4">
      <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-gray-500">
        Devices
      </h2>
      <ul className="flex-1 space-y-2 overflow-y-auto pr-1">
        {data.devices.map((d) => (
          <li
            key={d.id}
            className="flex items-center justify-between rounded-xl bg-[#f1f5f1] px-4 py-3"
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
  );

  const activityList = (
    <section className="panel flex max-h-[420px] flex-col p-4">
      <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-gray-500">
        Recent Activity
      </h2>
      <ul className="flex-1 space-y-2 overflow-y-auto pr-1">
        {data.activity.map((a) => (
          <li
            key={a.id}
            className="flex items-center justify-between rounded-xl bg-[#f1f5f1] px-4 py-3"
          >
            <span className="text-sm">{a.description}</span>
            <span className="text-xs text-gray-500">{formatTime(a.timestamp)}</span>
          </li>
        ))}
      </ul>
    </section>
  );

  return (
    <div className="space-y-6">
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <SubNav view={view} onViewChange={setView} />

        <select
          value={currentRoom}
          onChange={(e) => setCurrentRoom(e.target.value)}
          className="rounded-md border bg-white px-2 py-1 text-sm font-semibold"
        >
          {rooms.map((room) => (
            <option key={room.id} value={room.id}>
              {room.name}
            </option>
          ))}
        </select>
      </div>

      <hr className="soft-divider" />

      {view === 'Room Overview' && (
        <>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <InfoBox
              label="Temperature"
              value={`${data.metrics.temperature.value}°C`}
              change={`${data.metrics.temperature.change >= 0 ? '+' : ''}${data.metrics.temperature.change}°C`}
            />
            <InfoBox
              label="Humidity"
              value={`${data.metrics.humidity.value}%`}
              change={`${data.metrics.humidity.change >= 0 ? '+' : ''}${data.metrics.humidity.change}%`}
            />
            <InfoBox
              label="CO₂"
              value={`${data.metrics.co2.value} ppm`}
              change={`${data.metrics.co2.change >= 0 ? '+' : ''}${data.metrics.co2.change} ppm`}
            />
            <InfoBox
              label="Occupancy"
              value={`${data.metrics.occupancy.value} ${data.metrics.occupancy.value === 1 ? 'person' : 'people'}`}
              change={`${data.metrics.occupancy.change >= 0 ? '+' : ''}${data.metrics.occupancy.change}`}
            />
          </div>

          <hr className="soft-divider" />

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {deviceList}
            {activityList}
          </div>
        </>
      )}

      {view === 'Devices' && deviceList}
      {view === 'History' && activityList}
    </div>
  );
}