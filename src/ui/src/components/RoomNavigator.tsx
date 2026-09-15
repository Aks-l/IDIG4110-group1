"use client";

import { ChevronLeft, ChevronRight, Home } from "lucide-react";
import { Room } from "@/types";

export default function RoomNavigator({ rooms, selectedRoomId, focusRoom, onSelect, onShowAll }: { rooms: Room[]; selectedRoomId: string; focusRoom: boolean; onSelect: (id: string) => void; onShowAll: () => void }) {
  const selectedIndex = Math.max(0, rooms.findIndex((room) => room.id === selectedRoomId));
  const previous = rooms[(selectedIndex - 1 + rooms.length) % rooms.length];
  const next = rooms[(selectedIndex + 1) % rooms.length];

  return <div className="flex flex-wrap items-center justify-end gap-1">
    <button onClick={onShowAll} className={`flex items-center gap-1 border px-2 py-2 text-[10px] ${!focusRoom ? "border-cyan-400/50 bg-cyan-400/10 text-cyan-200" : "border-[#304754] bg-[#0b151c]/90 text-slate-400 hover:text-slate-100"}`} title="Show the full home"><Home size={13} /> Full house</button>
    <button onClick={() => onSelect(previous.id)} className="border border-[#304754] bg-[#0b151c]/90 p-2 text-slate-400 hover:text-slate-100" title={`Previous room: ${previous.name}`}><ChevronLeft size={14} /></button>
    <select value={selectedRoomId} onChange={(event) => onSelect(event.target.value)} className="max-w-36 border border-[#304754] bg-[#0b151c] px-2 py-2 text-[10px] text-slate-200 outline-none" aria-label="Choose room">
      {rooms.map((room) => <option key={room.id} value={room.id}>{room.name}</option>)}
    </select>
    <button onClick={() => onSelect(next.id)} className="border border-[#304754] bg-[#0b151c]/90 p-2 text-slate-400 hover:text-slate-100" title={`Next room: ${next.name}`}><ChevronRight size={14} /></button>
  </div>;
}
