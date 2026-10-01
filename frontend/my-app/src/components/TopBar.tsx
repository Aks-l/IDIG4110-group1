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
    <header className="grid min-h-20 grid-cols-[1fr_auto_1fr] items-center border-b border-[#dbe4dc] bg-[#fbfcfa] px-5 sm:px-8">
      <div>
        <h1 className="text-lg font-bold tracking-[-0.03em] text-[#17221d]">My Home</h1>
        <p className="text-[11px] font-semibold uppercase tracking-[0.16em] text-[#68766d]">Smart living, simply</p>
      </div>
      <div className="text-center">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-[#1f6f5b]">Workspace</p>
        <span className="font-semibold text-[#17221d]">{pageTitle}</span>
      </div>
      <div className="justify-self-end text-right">
        <p className="text-[11px] font-semibold uppercase tracking-[0.16em] text-[#68766d]">System status</p>
        <span className="inline-flex items-center gap-2 text-sm font-semibold text-[#1f6f5b]"><span className="h-2 w-2 rounded-full bg-[#39a56f] shadow-[0_0_0_4px_#dff1e6]" />Live {time}</span>
      </div>
    </header>
  );
}