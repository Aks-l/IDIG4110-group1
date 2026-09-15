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
      className={`block rounded px-3 py-2 text-sm ${
        isActive ? 'bg-black text-white' : 'text-gray-700 hover:bg-gray-100'
      }`}
    >
      {label}
    </Link>
  );
}