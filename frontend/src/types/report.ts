import { ScanStatus, ScanSummaryStats, ScanTriggerType, SeverityLevel } from './scan';

export type ReportTemplate = 'executive' | 'technical';
export type ReportFormat = 'pdf' | 'csv';

export interface ReportItem {
  scan_id: string;
  tenant_id: string;
  target_id: string;
  target_value: string;
  target_type: 'IP' | 'CIDR' | 'RANGE' | 'FQDN';
  group_id?: string;
  group_name?: string;
  profile_id: string;
  profile_name: string;
  status: ScanStatus;
  trigger_type: ScanTriggerType;
  started_at?: string;
  completed_at?: string;
  summary_stats: ScanSummaryStats;
  max_severity: SeverityLevel;
  posture_score: number;
  posture_grade: string;
  created_at: string;
}

export interface ReportFilterParams {
  target_id?: string;
  group_id?: string;
  severity?: string;
  target_type?: string;
  date_from?: string;
  date_to?: string;
  search?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

export interface ReportListResponseData {
  items: ReportItem[];
  total: number;
  limit: number;
  offset: number;
}

