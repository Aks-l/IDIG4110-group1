import { SideBar } from "@/components/SideBar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex">
      <SideBar />
      <main className="flex-1 p-8">{children}</main>
    </div>
  );
}