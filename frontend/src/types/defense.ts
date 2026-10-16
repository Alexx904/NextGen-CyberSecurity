import type { FirewallStatus, AddFirewallRulePayload } from './ssh';

export interface RemediationRequest {
  server_id: string;
}

export interface RemediationResult {
  node_id: string;
  server_id: string;
  rule: Omit<AddFirewallRulePayload, 'with_safety_timer' | 'safety_timeout_seconds'>;
  command: string;
  safety_timeout_seconds: number;
  firewall?: FirewallStatus;
}
