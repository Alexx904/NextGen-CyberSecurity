export type SSHAuthType = 'KEY' | 'PASSWORD';
export type SSHServerStatus = 'ONLINE' | 'OFFLINE' | 'UNCHECKED';

export interface SSHServerItem {
  id: string;
  tenant_id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  auth_type: SSHAuthType;
  public_key_fingerprint?: string | null;
  status: SSHServerStatus;
  last_connected_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateSSHServerPayload {
  name: string;
  host: string;
  port: number;
  username: string;
  auth_type: SSHAuthType;
  raw_secret: string;
}

export interface UpdateSSHServerPayload {
  name: string;
  host: string;
  port: number;
  username: string;
  auth_type: SSHAuthType;
  raw_secret?: string;
}

export interface SSHConnectionTestResult {
  online: boolean;
  latency_ms: number;
  os_info: string;
  is_root: boolean;
  error?: string;
}

export type ChainName = 'INPUT' | 'OUTPUT' | 'FORWARD';
export type RuleAction = 'ACCEPT' | 'DROP' | 'REJECT' | 'LOG';

export interface FirewallRule {
  line_number: number;
  packets: number;
  bytes: number;
  target: RuleAction;
  protocol: string;
  opt: string;
  in_interface: string;
  out_interface: string;
  source: string;
  destination: string;
  options: string;
}

export interface FirewallChain {
  name: ChainName;
  default_policy: RuleAction;
  packets: number;
  bytes: number;
  rules: FirewallRule[];
}

export interface FirewallStatus {
  server_id: string;
  chains: Record<ChainName, FirewallChain>;
  active_safety_timer: boolean;
  safety_timer_expires_at?: string | null;
  safety_timeout_seconds: number;
}

export interface AddFirewallRulePayload {
  chain: ChainName;
  action: RuleAction;
  protocol: string;
  source: string;
  destination: string;
  port: string;
  in_interface?: string;
  out_interface?: string;
  comment?: string;
  position?: number;
  with_safety_timer: boolean;
  safety_timeout_seconds: number;
}

