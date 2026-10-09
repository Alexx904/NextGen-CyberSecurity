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


