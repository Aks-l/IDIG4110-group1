'use client';

import { useEffect, useState } from 'react';
import { usePathname } from 'next/navigation';

export function TopBar() {
  const pathname = usePathname();
  const segment = pathname.split('/').filter(Boolean).pop() ?? '';
  const pageTitle = segment.charAt(0).toUpperCase() + segment.slice(1);
  const [time, setTime] = useState('');
  useEffect(() => {
    const id = setInterval(() => {
      setTime(new Date().toLocaleTimeString());
    }, 1000);
    return () => clearInterval(id);
  }, []);

  return (
    <header className="grid h-14 grid-cols-3 items-center border-b bg-white px-6">
      <h1 className="text-lg font-bold">My Home</h1>
      <span className="text-center font-bold">{pageTitle}</span>
      <span className="text-right font-bold">Live: {time}</span>
    </header>
  );
}