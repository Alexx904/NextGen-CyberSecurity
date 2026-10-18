import type { DiscoveryResult, DiscoveredDevice } from '../types/discovery';

function csvCell(value: unknown): string {
  let text = String(value ?? '');
  // Quoting alone does not prevent formula execution by spreadsheet programs.
  if (/^[\s\u0000-\u001f]*[=+@-]/.test(text)) text = `'${text}`;
  return `"${text.replace(/"/g, '""')}"`;
}

export function discoveryCSV(devices: DiscoveredDevice[]): string {
  const rows: unknown[][] = [['ip', 'status', 'hostname', 'mac', 'vendor', 'latency_ms', 'open_ports', 'scanned_at']];
  devices.forEach(device => rows.push([device.ip, device.status, device.hostname, device.mac, device.vendor, device.latency_ms, device.open_ports?.join(';'), device.scanned_at]));
  return rows.map(row => row.map(csvCell).join(',')).join('\r\n') + '\r\n';
}

export function discoveryJSON(result: DiscoveryResult, devices: DiscoveredDevice[]): string {
  return JSON.stringify({ ...result, snapshot_partial: !result.completed_at || result.completed_at.startsWith('0001-'), exported_count: devices.length, devices }, null, 2);
}

export function saveDiscoveryExport(result: DiscoveryResult, devices: DiscoveredDevice[], format: 'csv' | 'json'): void {
  const blob = new Blob([format === 'csv' ? discoveryCSV(devices) : discoveryJSON(result, devices)], { type: format === 'csv' ? 'text/csv;charset=utf-8' : 'application/json' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `discovery-${new Date().toISOString().replace(/[:.]/g, '-')}.${format}`;
  document.body.appendChild(link);
  link.click();
  link.remove();
  // Allow the browser to consume the download before revoking its URL.
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
