'use client';

import { useState } from 'react';

const views = ['Room Overview', 'Devices', 'History', 'Settings'];

export function SubNav() {
  const [view, setView] = useState('Room Overview');

  return (
    <select
      value={view}
      onChange={(e) => setView(e.target.value)}
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