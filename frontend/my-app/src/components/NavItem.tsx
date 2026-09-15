'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

type NavItemProps = {
  href: string;
  label: string;
};

export function NavItem({ href, label }: NavItemProps) {
  const pathname = usePathname();
  const isActive = pathname === href;

  return (
    <Link
        href={href}
        className={`block rounded px-4 py-3 text-base ${
            isActive ? 'bg-black text-white' : 'text-gray-700 hover:bg-gray-100'
        }`}
        >
        {label}
    </Link>
  );
}
//<aside className="flex w-96 flex-col gap-1 overflow-y-auto border-r bg-white p-3"></aside>