import type { FindingItem } from './finding';

export type ScanStatus = 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'CANCELLED';
export type ScanTriggerType = 'MANUAL' | 'SCHEDULED';
export type SeverityLevel = 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'INFO';

export interface ScanSummaryStats {
  total_hosts: number;
  up_hosts: number;
  down_hosts?: number;
  total_open_ports: number;
  critical_count: number;
  high_count: number;
  medium_count: number;
  low_count: number;
  info_count: number;
  duration_seconds: number;
}

export interface ScanItem {
  progress?: ScanProgressPayload;
  id: string;
  tenant_id: string;
  target_id: string;
  profile_id: string;
  trigger_type: ScanTriggerType;
  status: ScanStatus;
  started_at?: string;
  completed_at?: string;
  summary_stats: ScanSummaryStats;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface ScanListResponseData {
  items: ScanItem[];
  total: number;
  limit: number;
  offset: number;
}

// Payload Server-Sent Events (SSE) per Vulnerability Scans
export interface ScanPhasePayload {
  scan_id: string;
  phase: 'HOST_DISCOVERY' | 'PORT_SCAN' | 'SERVICE_DETECTION' | 'CVE_CORRELATION' | string;
  message: string;
  percentage: number;
}

export interface ScanProgressPayload {
  observed_at?: string;
  phase?: string;
  phase_percentage?: number;
  remaining_seconds?: number;
  scan_id: string;
  percentage: number;
  message: string;
}

export interface ScanFindingPayload {
  scan_id: string;
  finding: FindingItem;
}

export interface ScanStatusPayload {
  scan_id: string;
  status: ScanStatus;
  summary_stats: ScanSummaryStats;
}

export interface ScanErrorPayload {
  scan_id: string;
  message: string;
}

export type ScanProfileCategory =
  | 'ALL_IN_ONE'
  | 'PEN_TEST'
  | 'SYSTEM_FLAWS'
  | 'WEB_APP'
  | 'DISCOVERY'
  | 'COMPLIANCE'
  | 'CUSTOM';

export type InvasivenessLevel = 'SAFE' | 'INTRUSIVE' | 'DANGEROUS';

export interface ScanProfileItem {
  id: string;
  name: string;
  nmap_args: string;
  description?: string;
  timeout_seconds: number;
  category: ScanProfileCategory;
  invasiveness?: InvasivenessLevel;
  is_custom: boolean;
  tenant_id?: string;
  created_at?: string;
  updated_at?: string;
}

export type ScanDiffStatus = 'NEW' | 'RESOLVED' | 'PERSISTENT';

export interface FindingDiffItem {
  finding_key: string;
  status: ScanDiffStatus;
  host_ip: string;
  port: number;
  protocol: string;
  port_state?: string;
  service_name?: string;
  cve_id?: string;
  cvss_score?: number;
  severity: SeverityLevel;
  confidence?: string;
  detection_method?: string;
  remediation?: string;
  current_scan_id: string;
  previous_scan_id?: string;
}

export interface ScanDiffResult {
  current_scan_id: string;
  previous_scan_id?: string;
  target_id: string;
  target_value: string;
  generated_at: string;
  total_new: number;
  total_resolved: number;
  total_persistent: number;
  items: FindingDiffItem[];
}

