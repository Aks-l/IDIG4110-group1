// Data formats for the stats and devices
type Stat = {
  label: string;
  value: string;
};

type Device = {
  connected: number;
  warning: number;
  events: number;
};

async function getStats(): Promise<Stat[]> {
  return [
    { label: 'Temperature', value: '23°C' },
    { label: 'Humidity',    value: '48%' },
    { label: 'Energy',      value: '1.2 kW' },
    { label: 'Air Quality', value: 'Good' },
    { label: 'Occupancy',   value: '3 people' },
  ];
}

async function getDevices(): Promise<Device> {
  return {
    connected: 24,
    warning: 3,
    events: 7,
  };
}

export default async function OverviewPage() {
  const [stats, devices] = await Promise.all([getStats(), getDevices()]);

  const deviceStats = [
    { label: 'Connected', value: devices.connected },
    { label: 'Warning',   value: devices.warning },
    { label: 'Events',    value: devices.events },
  ];

  return (
    <div className="space-y-10">
      {/* Section 1: Stats */}
      <section className="flex flex-col gap-6 text-xl">
        {stats.map((s) => (
          <div key={s.label} className="flex justify-between border-b pb-3">
            <span className="text-gray-500">{s.label}:</span>
            <span className="font-semibold">{s.value}</span>
          </div>
        ))}
      </section>

      {/* Section 2: Devices */}
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-gray-700">Devices</h2>

        <div className="flex flex-col gap-6 text-xl">
          {deviceStats.map((d) => (
            <div key={d.label} className="flex justify-between border-b pb-3">
              <span className="text-gray-500">{d.label}:</span>
              <span className="font-semibold">{d.value}</span>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}