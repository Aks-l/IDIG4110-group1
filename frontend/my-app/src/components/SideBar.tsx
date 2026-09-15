import { NavItem } from "./NavItem";

const menuItems = [
    { href: '/overview', label: 'Overview' },
    { href: '/rooms', label: 'Rooms' },
    { href: '/devices', label: 'Devices' },
    { href: '/energy', label: 'Energy' },
    { href: '/automation', label: 'Automation' },
    { href: '/3d-model', label: '3D Model' },
    { href: '/settings', label: 'Settings' },
];

export function SideBar() {
  return (
    <aside className="flex h-screen w-56 flex-col gap-1 border-r bg-white p-3">
      <h2 className="mb-4 px-3 text-lg font-bold">My App</h2>
      {menuItems.map((item) => (
        <NavItem key={item.href} href={item.href} label={item.label} />
      ))}
    </aside>
  );
}