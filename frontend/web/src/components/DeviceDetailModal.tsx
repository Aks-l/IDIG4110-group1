'use client';

import { Modal } from '@/components/Modal';
import { Toggle } from '@/components/Toggle';
import { formatDate } from '@/lib/format';
import type { Device } from '@/lib/api/types';

type DeviceWithRoom = Device & { room: string; warning: boolean };

type DeviceDetailModalProps = {
  device: DeviceWithRoom | null;
  onClose: () => void;
  onToggle: (id: string, on: boolean) => void;
};

const deviceIcons: Record<string, string> = {
  Appliance: '◉',
  Climate: '◌',
  Fan: '◒',
  Light: '✦',
  Media: '▶',
  Sensor: '◈',
};

function statusLabel(d: DeviceWithRoom) {
  if (d.warning) return { label: 'Warning', className: 'bg-[#fff4d9] text-[#956b13]' };
  if (d.on)      return { label: 'On',      className: 'bg-[#e5f2e9] text-[#1f6f5b]' };
  return                { label: 'Off',     className: 'bg-[#eef1ee] text-[#68766d]' };
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-baseline justify-between border-b border-[#eef1ee] py-2 last:border-b-0">
      <span className="text-xs font-medium uppercase tracking-wide text-[#9aa89f]">{label}</span>
      <span className="text-sm font-medium text-[#17221d]">{value}</span>
    </div>
  );
}

export function DeviceDetailModal({ device, onClose, onToggle }: DeviceDetailModalProps) {
  if (!device) return null;
  const status = statusLabel(device);

  return (
    <Modal
      open={!!device}
      onClose={onClose}
      title="Device details"
      heightClass="h-auto max-h-[90vh]"
    >
      <div className="space-y-5">
        {/* Header block */}
        <div className="flex items-center gap-4">
          <span className="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-[#e5f2e9] text-2xl text-[#1f6f5b]">
            {deviceIcons[device.type] ?? '•'}
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-lg font-semibold text-[#17221d]">{device.name}</p>
            <p className="truncate text-sm text-[#68766d]">
              {device.type} <span className="px-1">·</span> {device.room}
            </p>
          </div>
          <span className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold ${status.className}`}>
            {status.label}
          </span>
        </div>

        {/* Toggle control */}
        <div className="flex items-center justify-between rounded-xl border border-[#dbe4dc] bg-[#fbfcfa] px-4 py-3">
          <span className="text-sm font-medium text-[#17221d]">
            {device.on ? 'Powered on' : 'Powered off'}
          </span>
          <Toggle
            checked={device.on}
            onChange={(on) => onToggle(device.id, on)}
            label={`${device.on ? 'Turn off' : 'Turn on'} ${device.name}`}
          />
        </div>

        {/* Info grid */}
        <div className="rounded-xl border border-[#dbe4dc] bg-white px-4 py-1">
          <Row label="Type" value={device.type} />
          <Row label="Room" value={device.room} />
          {device.powerWatts !== undefined && (
            <Row label="Power draw" value={`${device.powerWatts} W`} />
          )}
          {device.signal !== undefined && (
            <Row label="Signal" value={`${device.signal}%`} />
          )}
          {device.battery !== undefined && (
            <Row label="Battery" value={`${device.battery}%`} />
          )}
          {device.firmware && <Row label="Firmware" value={device.firmware} />}
          {device.lastSeen && <Row label="Last seen" value={formatDate(device.lastSeen)} />}
          {device.installedAt && <Row label="Installed" value={formatDate(device.installedAt)} />}
        </div>
      </div>
    </Modal>
  );
}