import { NavItem } from "./NavItem";
import { Box } from 'lucide-react';

const prefix = "/dashboard";
const menuItems = [
    { href: prefix + '/overview', label: 'Overview' },
    { href: prefix + '/rooms', label: 'Rooms' },
    { href: prefix + '/devices', label: 'Devices' },
    { href: prefix + '/events', label: 'Events' },
    { href: prefix + '/energy', label: 'Energy' },
    { href: prefix + '/automation', label: 'Automation' },
    { href: prefix + '/3d-model', label: '3D Model', icon: <Box size={16} strokeWidth={2} /> },
  // { href: prefix + '/settings', label: 'Settings' },
];

export function SideBar() {
  return (
    <aside className="flex w-64 shrink-0 flex-col overflow-y-auto border-r border-[#dbe4dc] bg-[#fbfcfa] p-4 max-md:w-20 max-md:px-2">
      <div className="mb-7 px-3 max-md:px-0 max-md:text-center">
        <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-[#1f6f5b]">Control center</p>
        <p className="mt-1 text-sm font-semibold text-[#68766d] max-md:hidden">Your home, at a glance</p>
      </div>
      {menuItems.map((item) => (
        <NavItem key={item.href} href={item.href} label={item.label} icon={item.icon} />
      ))}
    </aside>
  );
}
