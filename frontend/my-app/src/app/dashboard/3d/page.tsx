'use client';

import dynamic from 'next/dynamic';
import { useEffect, useMemo, useState } from 'react';
import { SidePanel } from '@/components/three/SidePanel';
import { loadDevices } from '@/lib/api/devices.data';
import { loadRoomLayout } from '@/lib/api/three-d.data';
import { loadRooms } from '@/lib/api/rooms.data';
import { dataSource } from '@/lib/api/data-source';
import { setDeviceState } from '@/lib/api/api-devices';
import type { Device, RoomData, RoomLayout } from '@/lib/api/types';

const Scene = dynamic(() => import('@/components/three/Scene'), {
  ssr: false,
  loading: () => <div className="flex h-full items-center justify-center text-sm text-[#68766d]">Loading 3D model...</div>,
});

export default function ThreeDModelPage() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [rooms, setRooms] = useState<RoomData[]>([]);
  const [roomLayout, setRoomLayout] = useState<RoomLayout[]>([]);
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null);
  const [selectedRoomId, setSelectedRoomId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([loadDevices(), loadRooms(), loadRoomLayout()])
      .then(([loadedDevices, loadedRooms, loadedLayout]) => {
        const roomByDevice = new Map<string, string>(
          loadedRooms.flatMap((room) => room.devices.map((device) => [`${room.id}-${device.id}`, room.id] as const)),
        );
        setDevices(loadedDevices.map((device) => ({
          ...device,
          roomId: device.roomId ?? roomByDevice.get(device.id),
        })));
        setRooms(loadedRooms);
        setRoomLayout(loadedLayout);
      })
      .catch(() => setError('Unable to load the digital twin'))
      .finally(() => setLoading(false));
  }, []);

  const selectedDevice = devices.find((device) => device.id === selectedDeviceId) ?? null;
  const selectedRoom = rooms.find((room) => room.id === selectedRoomId) ?? null;

  const handleSelectRoom = (roomId: string) => {
    setSelectedRoomId(roomId);
    setSelectedDeviceId(null);
  };

  const handleSelectDevice = (deviceId: string) => {
    setSelectedDeviceId(deviceId);
    setSelectedRoomId(null);
  };

  const handleToggleDevice = async (deviceId: string, on: boolean) => {
    const previous = devices.find((device) => device.id === deviceId)?.on;
    setDevices((current) => current.map((device) => device.id === deviceId ? { ...device, on } : device));

    if (dataSource !== 'api') return;

    try {
      await setDeviceState(deviceId, on);
    } catch {
      if (previous !== undefined) {
        setDevices((current) => current.map((device) => device.id === deviceId ? { ...device, on: previous } : device));
      }
    }
  };

  const roomButtons = useMemo(() => rooms.map((room) => ({
    ...room,
    layout: roomLayout.find((layout) => layout.roomId === room.id),
  })), [roomLayout, rooms]);

  return (
    <div className="mx-auto flex min-h-full max-w-7xl flex-col gap-6">
      <header>
        <p className="text-xs font-bold uppercase tracking-[0.2em] text-[#1f6f5b]">Digital twin</p>
        <h1 className="page-heading mt-2 text-3xl font-bold text-[#17221d] sm:text-4xl">3D Model</h1>
        <p className="mt-2 text-sm text-[#68766d]">Explore a model of your home</p>
      </header>

      {error ? <p className="text-red-600">{error}</p> : loading ? (
        <div className="panel flex min-h-[520px] items-center justify-center text-sm text-[#68766d]">Loading your home...</div>
      ) : (
        <>
          <div className="grid min-h-[560px] flex-1 gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
            <section className="panel min-h-[520px] overflow-hidden p-1" aria-label="Interactive 3D home model">
              <Scene
                devices={devices}
                rooms={rooms}
                roomLayout={roomLayout}
                selectedDeviceId={selectedDeviceId}
                selectedRoomId={selectedRoomId}
                onSelectDevice={handleSelectDevice}
                onSelectRoom={handleSelectRoom}
                onClearSelection={() => {
                  setSelectedDeviceId(null);
                  setSelectedRoomId(null);
                }}
              />
            </section>
            <SidePanel device={selectedDevice} room={selectedRoom} rooms={rooms} onToggleDevice={handleToggleDevice} />
          </div>

          <section className="panel p-4" aria-label="Room selector">
            <div className="mb-3 flex items-center justify-between">
              <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Rooms</p>
              <span className="text-xs text-[#68766d]">Select a room to inspect it</span>
            </div>
            <div className="flex flex-wrap gap-2">
              {roomButtons.map((room) => (
                <button
                  key={room.id}
                  type="button"
                  onClick={() => handleSelectRoom(room.id)}
                  className={`rounded-xl px-4 py-2 text-sm font-semibold transition-colors ${selectedRoomId === room.id ? 'bg-[#1f6f5b] text-white' : 'bg-[#eef3ef] text-[#31453a] hover:bg-[#dcebe0]'}`}
                >
                  {room.name}
                </button>
              ))}
            </div>
          </section>
        </>
      )}
    </div>
  );
}