export interface DiscoveredDevice {
  ip: string;
  mac?: string;
  vendor?: string;
  hostname?: string;
  latency_ms: number;
  status: string;
  open_ports?: number[];
  scanned_at: string;
}

export interface DiscoveryResult {
  id: string;
  tenant_id: string;
  subnet: string;
  devices: DiscoveredDevice[];
  total_discovered: number;
  duration_seconds: number;
  started_at: string;
  completed_at: string;
}

export interface DiscoveryScanParams {
  subnet: string;
  include_common_ports?: boolean;
  session_id?: string;
}

// Payload Server-Sent Events (SSE) per LAN Discovery
export interface DeviceDiscoveredPayload {
  session_id: string;
  device: DiscoveredDevice;
}

export interface DiscoveryProgressPayload {
  session_id: string;
  scanned_hosts: number;
  total_hosts: number;
  percentage: number;
  current_ip: string;
}

export interface DiscoveryCompletePayload {
  session_id: string;
  total_discovered: number;
  duration_seconds: number;
}

export interface DiscoveryErrorPayload {
  session_id: string;
  message: string;
}

export interface ImportTargetDeviceInput {
  ip: string;
  hostname?: string;
  vendor?: string;
  description?: string;
}

export interface ImportTargetsRequestPayload {
  group_id?: string | null;
  devices: ImportTargetDeviceInput[];
}

export interface ImportItemError {
  ip: string;
  reason: string;
}

export interface ImportTargetsResponseData {
  group_id?: string | null;
  total_imported: number;
  total_skipped: number;
  imported: Array<{
    id: string;
    value: string;
    type: string;
    group_id?: string;
    description?: string;
    created_at: string;
  }>;
  errors?: ImportItemError[];
}
