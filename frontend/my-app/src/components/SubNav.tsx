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
      className="rounded-md border bg-white px-3 py-1.5 text-sm"
    >
      {views.map((v) => (
        <option key={v} value={v}>
          {v}
        </option>
      ))}
    </select>
  );
}