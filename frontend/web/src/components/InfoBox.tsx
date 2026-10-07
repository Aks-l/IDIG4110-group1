type InfoBoxProps = {
  label: string;
  value: string;
  change?: string;
};

export function InfoBox({ label, value, change }: InfoBoxProps) {
  return (
    <div className="panel px-4 py-4">
      <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[#68766d]">{label}</p>
      <p className="mt-2 text-2xl font-bold tracking-[-0.04em] text-[#17221d]">{value}</p>
      {change && <p className="mt-1 text-xs font-semibold text-[#1f6f5b]">{change} from last reading</p>}
    </div>
  );
}