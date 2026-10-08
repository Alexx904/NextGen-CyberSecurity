import { SeverityLevel } from './scan';

export interface FindingItem {
  id: string;
  scan_id: string;
  tenant_id: string;
  host_ip: string;
  hostname?: string;
  port: number;
  protocol: string;
  service_name?: string;
  service_version?: string;
  cve_id?: string;
  cvss_score?: number;
  severity: SeverityLevel;
  raw_banner?: string;
  remediation?: string;
  created_at: string;
}

export interface FindingListResponseData {
  items: FindingItem[];
  total: number;
  limit: number;
  offset: number;
}

