import { NavItem } from "./NavItem";

const prefix = "/dashboard";
const menuItems = [
    { href: prefix + '/overview', label: 'Overview' },
    { href: prefix + '/rooms', label: 'Rooms' },
    { href: prefix + '/devices', label: 'Devices' },
    { href: prefix + '/events', label: 'Events' },
    { href: prefix + '/energy', label: 'Energy' },
    { href: prefix + '/automation', label: 'Automation' },
    { href: prefix + '/3d-model', label: '3D Model' },
    { href: prefix + '/settings', label: 'Settings' },
];

export function SideBar() {
  return (
    <aside className="flex w-64 flex-col gap-1 overflow-y-auto border-r bg-white p-3">
      {menuItems.map((item) => (
        <NavItem key={item.href} href={item.href} label={item.label} />
      ))}
    </aside>
  );
}
