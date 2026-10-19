import type { SeverityLevel } from './scan';

export type FindingConfidence = 'CONFIRMED' | 'PROBABLE' | 'UNVERIFIED';
export type DetectionMethod = 'nse' | 'banner' | 'nvd_keyword' | 'port_state';

export interface RegulatoryContext {
  rule_version: string;
  confidence: FindingConfidence | '';
  detection_method: DetectionMethod | '';
  implications: { kind: string; observation: string; meaning: string; action: string; references: { id: string; title: string; source: string }[] }[];
  limit: string;
}

export interface FindingItem {
  regulatory_context?: RegulatoryContext;
  id: string;
  scan_id: string;
  tenant_id: string;
  host_ip: string;
  hostname?: string;
  port: number;
  protocol: string;
  port_state?: string;
  service_name?: string;
  service_version?: string;
  cve_id?: string;
  cvss_score?: number;
  severity: SeverityLevel;
  confidence?: FindingConfidence;
  detection_method?: DetectionMethod;
  cpe?: string;
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

