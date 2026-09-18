type InfoBoxProps = {
  label: string;
  value: string;
  change?: string;
};

export function InfoBox({ label, value, change }: InfoBoxProps) {
  return (
    <div className="rounded-lg border bg-white px-3 py-2 shadow-sm">
      <p className="text-xs font-semibold">{label}</p>
      <p className="text-lg font-semibold">{value}</p>
      {change && <p className="text-xs font-semibold">{change}</p>}
    </div>
  );
}