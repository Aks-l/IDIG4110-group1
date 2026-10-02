'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';

type NavItemProps = {
  href: string;
  label: string;
  icon?: ReactNode;
};

export function NavItem({ href, label, icon }: NavItemProps) {
  const pathname = usePathname();
  const isActive = pathname === href;

  return (
    <Link
        href={href}
        className={`mb-1 block rounded-xl px-3 py-3 text-sm font-medium transition-colors max-md:px-2 max-md:text-center ${
          isActive ? 'bg-[#1f6f5b] text-white shadow-[0_8px_18px_rgba(31,111,91,0.2)]' : 'text-[#53635a] hover:bg-[#edf3ee] hover:text-[#17221d]'
        }`}
        >
        <span className="flex items-center justify-center gap-2 md:hidden">{icon ?? label.slice(0, 1)}</span>
        <span className="flex items-center gap-2 max-md:hidden">{icon}{label}</span>
    </Link>
  );
}
//<aside className="flex w-96 flex-col gap-1 overflow-y-auto border-r bg-white p-3"></aside>