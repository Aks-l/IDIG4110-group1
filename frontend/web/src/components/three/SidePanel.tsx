'use client';

import { Toggle } from '@/components/Toggle';
import type { Device, RoomData } from '@/lib/api/types';

type SidePanelProps = {
  device: Device | null;
  room: RoomData | null;
  rooms: RoomData[];
  onToggleDevice: (deviceId: string, on: boolean) => void;
};

export function SidePanel({ device, room, rooms, onToggleDevice }: SidePanelProps) {
  if (device) {
    const deviceRoom = rooms.find((candidate) => candidate.id === device.roomId);
    return (
      <aside className="panel h-full p-5">
        <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Selected device</p>
        <h2 className="mt-2 text-xl font-bold text-[#17221d]">{device.name}</h2>
        <p className="mt-1 text-sm text-[#68766d]">{device.type}</p>

        <div className="mt-6 space-y-4">
          <DetailRow label="Room" value={deviceRoom?.name ?? device.roomId ?? 'Unassigned'} />
          <DetailRow label="Connection" value="Online" valueClass="text-[#1f6f5b]" />
          <div className="flex items-center justify-between border-t border-[#dbe4dc] pt-4">
            <span className="text-sm font-medium text-[#31453a]">Power state</span>
            <Toggle checked={device.on} onChange={(on) => onToggleDevice(device.id, on)} label={`${device.on ? 'Turn off' : 'Turn on'} ${device.name}`} />
          </div>
        </div>
      </aside>
    );
  }

  if (room) {
    return (
      <aside className="panel h-full p-5">
        <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Selected room</p>
        <h2 className="mt-2 text-xl font-bold text-[#17221d]">{room.name}</h2>
        <div className="mt-6">
          <p className="mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-[#68766d]">Room snapshot</p>
          <div className="grid grid-cols-2 gap-2">
            <Metric label="Temperature" value={`${room.metrics.temperature.value}°C`} />
            <Metric label="Humidity" value={`${room.metrics.humidity.value}%`} />
            <Metric label="CO₂" value={`${room.metrics.co2.value} ppm`} />
            <Metric label="Occupancy" value={String(room.metrics.occupancy.value)} />
          </div>
        </div>
        <div className="mt-6">
          <p className="mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-[#68766d]">Devices in room</p>
          <ul className="space-y-2">
            {room.devices.map((roomDevice) => (
              <li key={roomDevice.id} className="flex items-center justify-between rounded-lg bg-[#f1f5f1] px-3 py-2 text-sm">
                <span className="text-[#31453a]">{roomDevice.name}</span>
                <span className={`h-2 w-2 rounded-full ${roomDevice.on ? 'bg-[#39a56f]' : 'bg-[#a8b4ab]'}`} />
              </li>
            ))}
          </ul>
        </div>
      </aside>
    );
  }

  return (
    <aside className="panel flex h-full min-h-52 flex-col items-center justify-center p-5 text-center">
      <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[#e5f2e9] text-xl text-[#1f6f5b]">⌁</div>
      <h2 className="mt-4 font-semibold text-[#17221d]">Explore your home</h2>
      <p className="mt-2 max-w-xs text-sm leading-6 text-[#68766d]">Click a room or device in the model to see its details here.</p>
    </aside>
  );
}

function DetailRow({ label, value, valueClass = 'text-[#17221d]' }: { label: string; value: string; valueClass?: string }) {
  return <div className="flex items-center justify-between border-b border-[#e6ece7] pb-3 text-sm"><span className="text-[#68766d]">{label}</span><span className={`font-semibold ${valueClass}`}>{value}</span></div>;
}

function Metric({ label, value }: { label: string; value: string }) {
  return <div className="rounded-lg bg-[#f1f5f1] p-3"><p className="text-[10px] font-semibold uppercase tracking-wide text-[#68766d]">{label}</p><p className="mt-1 font-semibold text-[#17221d]">{value}</p></div>;
}