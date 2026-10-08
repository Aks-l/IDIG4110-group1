import { loadOverview } from '@/lib/api/overview.data';

// The overview is fetched over HTTP at request time. Without this, Next.js
// would try to prerender the page during `next build`, when no API is up.
export const dynamic = 'force-dynamic';

export default async function OverviewPage() {
  const { stats, devices } = await loadOverview();

  const deviceStats = [
    { label: 'Connected', value: devices.connected },
    { label: 'Warning',   value: devices.warning },
    { label: 'Events',    value: devices.events },
  ];

  return (
    <div className="mx-auto max-w-6xl space-y-8">
      <section className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <h2 className="page-heading mt-2 text-3xl font-bold text-[#17221d] sm:text-4xl">Residency Overview</h2>
        </div>
        <span className="w-fit rounded-full bg-[#e5f2e9] px-3 py-1.5 text-xs font-semibold text-[#1f6f5b]">All systems nominal</span>
      </section>

      <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        {stats.map((s) => (
          <div key={s.label} className="panel px-4 py-5">
            <span className="text-xs font-semibold uppercase tracking-[0.1em] text-[#68766d]">{s.label}</span>
            <p className="mt-3 text-2xl font-bold tracking-[-0.04em] text-[#17221d]">{s.value}</p>
          </div>
        ))}
      </section>

      <section className="panel overflow-hidden">
        <div className="flex items-center justify-between border-b border-[#dbe4dc] px-5 py-4">
          <div>
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-[#1f6f5b]">Network health</p>
            <h2 className="mt-1 text-lg font-semibold text-[#17221d]">Devices</h2>
          </div>
          <span className="text-xs font-medium text-[#68766d]">Updated just now</span>
        </div>

        <div className="grid gap-px bg-[#dbe4dc] sm:grid-cols-3">
          {deviceStats.map((d) => (
            <div key={d.label} className="bg-white px-5 py-5">
              <span className="text-sm text-[#68766d]">{d.label}</span>
              <p className="mt-1 text-2xl font-bold text-[#17221d]">{d.value}</p>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}