

'use client';

import { useEffect, useMemo, useState } from 'react';
import { InfoBox } from '@/components/InfoBox';
import { Toggle } from '@/components/Toggle';
import { loadDevices, setDeviceState } from '@/lib/api/devices.data';
import { loadRooms } from '@/lib/api/rooms.data';
import type { Device, RoomData } from '@/lib/api/types';

type DeviceWithRoom = Device & { room: string; warning: boolean };
type FilterValue = 'All' | string;

const deviceIcons: Record<string, string> = {
  Appliance: '◉',
  Climate: '◌',
  Fan: '◒',
  Light: '✦',
  Media: '▶',
  Sensor: '◈',
};

function isWarningDevice(device: Device) {
  return device.type === 'Sensor' && /smoke|leak/i.test(device.name);
}

export default function DevicePage() {
  const [devices, setDevices] = useState<DeviceWithRoom[]>([]);
  const [search, setSearch] = useState('');
  const [type, setType] = useState<FilterValue>('All');
  const [status, setStatus] = useState<FilterValue>('All');
  const [room, setRoom] = useState<FilterValue>('All');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([loadDevices(), loadRooms()])
      .then(([loadedDevices, loadedRooms]) => {
        const roomNames = new Map(loadedRooms.map((loadedRoom) => [loadedRoom.id, loadedRoom.name]));
        setDevices(loadedDevices.map((device) => ({
          ...device,
          room: roomNames.get(device.roomId ?? '') ?? 'Unassigned',
          warning: isWarningDevice(device),
        })));
      })
      .catch(() => setError('Unable to load devices'));
  }, []);

  const types = useMemo(() => [...new Set(devices.map((device) => device.type))].sort(), [devices]);
  const rooms = useMemo(() => [...new Set(devices.map((device) => device.room))].sort(), [devices]);
  const onlineCount = devices.filter((device) => device.on).length;
  const warningCount = devices.filter((device) => device.warning).length;

  const handleToggle = async (deviceId: string, on: boolean) => {
    const previous = devices.find((device) => device.id === deviceId)?.on;
    setDevices((current) => current.map((device) => device.id === deviceId ? { ...device, on } : device));

    try {
      await setDeviceState(deviceId, on);
    } catch {
      if (previous !== undefined) {
        setDevices((current) => current.map((device) => device.id === deviceId ? { ...device, on: previous } : device));
      }
    }
  };

  const filteredDevices = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();

    return devices.filter((device) => {
      const matchesSearch = !normalizedSearch || `${device.name} ${device.type} ${device.room}`.toLowerCase().includes(normalizedSearch);
      const matchesType = type === 'All' || device.type === type;
      const matchesRoom = room === 'All' || device.room === room;
      const matchesStatus = status === 'All'
        || (status === 'Online' && device.on)
        || (status === 'Offline' && !device.on)
        || (status === 'Warnings' && device.warning);

      return matchesSearch && matchesType && matchesRoom && matchesStatus;
    });
  }, [devices, room, search, status, type]);

  if (error) return <p className="text-red-600">{error}</p>;

  return (
    <div className="mx-auto max-w-6xl space-y-8">
      <div>
        <p className="text-xs font-bold uppercase tracking-[0.2em] text-[#1f6f5b]">Connected home</p>
        <h1 className="page-heading mt-2 text-3xl font-bold text-[#17221d] sm:text-4xl">Devices</h1>
        <p className="mt-2 text-sm text-[#68766d]">Keep track of every connected device across your home.</p>
      </div>

      <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <InfoBox label="Total" value={String(devices.length)} />
        <InfoBox label="Online" value={String(onlineCount)} change="Live" />
        <InfoBox label="Offline" value={String(devices.length - onlineCount)} />
        <InfoBox label="Warnings" value={String(warningCount)} change={warningCount ? 'Review' : 'Clear'} />
      </section>

      <section className="panel p-4 sm:p-5">
        <div className="flex flex-col gap-3 lg:flex-row">
          <label className="relative flex-1">
            <span className="sr-only">Search devices</span>
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[#68766d]">⌕</span>
            <input
              type="search"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search devices..."
              className="w-full rounded-xl border border-[#dbe4dc] bg-[#fbfcfa] py-2.5 pl-9 pr-3 text-sm text-[#17221d] outline-none transition focus:border-[#1f6f5b]"
            />
          </label>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3 lg:w-[520px]">
            <FilterSelect label="Type" value={type} options={types} onChange={setType} />
            <FilterSelect label="Status" value={status} options={['Online', 'Offline', 'Warnings']} onChange={setStatus} />
            <FilterSelect label="Room" value={room} options={rooms} onChange={setRoom} />
          </div>
        </div>
      </section>

      <section className="panel overflow-hidden">
        <div className="flex items-center justify-between border-b border-[#dbe4dc] px-5 py-4">
          <div>
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Inventory</p>
            <h2 className="mt-1 text-lg font-semibold text-[#17221d]">All devices</h2>
          </div>
          <span className="text-xs font-medium text-[#68766d]">{filteredDevices.length} shown</span>
        </div>

        {filteredDevices.length > 0 ? (
          <ul>
            {filteredDevices.map((device) => (
              <li key={device.id} className="flex items-center gap-3 border-b border-[#e6ece7] px-5 py-4 last:border-b-0 hover:bg-[#f8faf8] sm:gap-4">
                <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#e5f2e9] text-lg text-[#1f6f5b]" aria-hidden="true">
                  {deviceIcons[device.type] ?? '•'}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-semibold text-[#17221d]">{device.name}</p>
                  <p className="truncate text-xs text-[#68766d]">{device.type} <span className="px-1">·</span> {device.room}</p>
                </div>
                <span className={`hidden rounded-full px-2.5 py-1 text-xs font-semibold sm:inline-flex ${device.warning ? 'bg-[#fff4d9] text-[#956b13]' : device.on ? 'bg-[#e5f2e9] text-[#1f6f5b]' : 'bg-[#eef1ee] text-[#68766d]'}`}>
                  {device.warning ? 'Warning' : device.on ? 'On' : 'Off'}
                </span>
                <Toggle
                  checked={device.on}
                  onChange={(on) => handleToggle(device.id, on)}
                  label={`${device.on ? 'Turn off' : 'Turn on'} ${device.name}`}
                />
                <span className={`h-2.5 w-2.5 shrink-0 rounded-full ${device.warning ? 'bg-[#e5a93d]' : device.on ? 'bg-[#39a56f] shadow-[0_0_0_4px_#dff1e6]' : 'bg-[#a8b4ab]'}`} title={device.warning ? 'Warning' : device.on ? 'Online' : 'Offline'} />
              </li>
            ))}
          </ul>
        ) : (
          <div className="px-5 py-12 text-center">
            <p className="font-semibold text-[#17221d]">No devices found</p>
            <p className="mt-1 text-sm text-[#68766d]">Try adjusting your search or filters.</p>
          </div>
        )}
      </section>
    </div>
  );
}

function FilterSelect({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: FilterValue;
  options: string[];
  onChange: (value: string) => void;
}) {
  const [open, setOpen] = useState(false);

  return (
    <div className="relative min-w-0">
      <div className="flex min-w-0 items-center gap-3 rounded-xl border border-[#dbe4dc] bg-[#fbfcfa] px-3 py-2">
      <span className="flex-1 text-xs font-semibold text-[#68766d]">{label}</span>
        <button
          type="button"
          aria-expanded={open}
          onClick={() => setOpen((isOpen) => !isOpen)}
          className="flex min-w-0 max-w-[58%] items-center justify-end gap-2 text-right text-sm font-medium text-[#31453a]"
        >
          <span className="truncate">{value}</span>
          <span className="shrink-0 text-xs text-[#68766d]" aria-hidden="true">⌄</span>
        </button>
      </div>
      {open && (
        <div className="absolute left-0 right-0 top-[calc(100%+0.35rem)] z-20 overflow-hidden rounded-xl border border-[#dbe4dc] bg-white p-1 shadow-[0_12px_28px_rgba(35,64,48,0.15)]">
          {['All', ...options].map((option) => (
            <button
              key={option}
              type="button"
              onClick={() => {
                onChange(option);
                setOpen(false);
              }}
              className={`block w-full truncate rounded-lg px-3 py-2 text-left text-sm transition-colors ${value === option ? 'bg-[#e5f2e9] font-semibold text-[#1f6f5b]' : 'text-[#31453a] hover:bg-[#f1f5f1]'}`}
            >
              {option}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}