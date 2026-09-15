"use client";

import { useTwinStore } from "@/stores/useTwinStore";

export default function RoomDeviceList({ roomId }: { roomId: string }) {
  const sensors = useTwinStore((state) => state.sensors);
  const devices = sensors.filter((sensor) => sensor.roomId === roomId);
  const selectedSensorId = useTwinStore((state) => state.selectedSensorId);
  const setSelectedSensor = useTwinStore((state) => state.setSelectedSensor);

  return <div>
    <div className="mb-2 flex items-center justify-between">
      <p className="text-[9px] uppercase tracking-wider text-slate-500">Devices in this room</p>
      <span className="text-[9px] text-slate-600">{devices.length} connected</span>
    </div>
    {devices.length ? <div className="space-y-2">
      {devices.map((device) => <button key={device.id} onClick={() => setSelectedSensor(device.id)} className={`flex w-full items-center justify-between border px-3 py-2 text-left transition-colors ${selectedSensorId === device.id ? "border-cyan-400/60 bg-cyan-400/10" : "border-[#253744] hover:border-cyan-400/40 hover:bg-[#14232e]"}`}>
        <span><span className="block text-[10px] text-slate-200">{device.name}</span><span className="mt-1 block text-[9px] text-slate-500">{device.type} · {device.id}</span></span>
        <span className="text-right"><span className={`block text-[10px] ${device.status === "normal" ? "text-emerald-300" : device.status === "warning" ? "text-amber-300" : "text-red-300"}`}>{device.value} {device.unit}</span><span className="mt-1 block text-[9px] uppercase text-slate-500">{selectedSensorId === device.id ? "Selected" : device.status}</span></span>
      </button>)}
    </div> : <p className="border border-dashed border-[#304754] p-3 text-[10px] text-slate-500">No devices assigned to this room.</p>}
  </div>;
}
