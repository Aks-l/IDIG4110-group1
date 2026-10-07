import { TopBar } from  '@/components/TopBar';
import { SideBar } from '@/components/SideBar';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex h-screen flex-col bg-[#f4f6f3]">
      <TopBar />
      <div className="flex flex-1 overflow-hidden">
        <SideBar />
        <main className="dashboard-main flex-1 overflow-y-auto p-5 sm:p-8">{children}</main>
      </div>
    </div>
  );
}