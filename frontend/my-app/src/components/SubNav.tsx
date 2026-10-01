'use client';

export type RoomView = 'Room Overview' | 'Devices' | 'History';

const views: RoomView[] = ['Room Overview', 'Devices', 'History'];

type SubNavProps = {
  view: RoomView;
  onViewChange: (view: RoomView) => void;
};

export function SubNav({ view, onViewChange }: SubNavProps) {
  return (
    <select
      value={view}
      onChange={(e) => onViewChange(e.target.value as RoomView)}
      className="rounded-xl border border-[#dbe4dc] bg-white px-3 py-2 text-sm font-medium text-[#31453a] shadow-sm"
    >
      {views.map((v) => (
        <option key={v} value={v}>
          {v}
        </option>
      ))}
    </select>
  );
}