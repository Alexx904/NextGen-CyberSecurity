import { apiClient } from './client';
import { ApiResponse } from '../types/auth';
import {
  SSHServerItem,
  CreateSSHServerPayload,
  UpdateSSHServerPayload,
  SSHConnectionTestResult,
  FirewallStatus,
  AddFirewallRulePayload,
  ChainName,
} from '../types/ssh';

export const sshApi = {
  listServers: async (): Promise<SSHServerItem[]> => {
    const res = await apiClient.get<ApiResponse<SSHServerItem[]>>('/ssh/servers');
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Impossibile recuperare i server SSH');
    }
    return res.data.data;
  },

  getServer: async (id: string): Promise<SSHServerItem> => {
    const res = await apiClient.get<ApiResponse<SSHServerItem>>(`/ssh/servers/${id}`);
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Server SSH non trovato');
    }
    return res.data.data;
  },

  createServer: async (payload: CreateSSHServerPayload): Promise<SSHServerItem> => {
    const res = await apiClient.post<ApiResponse<SSHServerItem>>('/ssh/servers', payload);
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Errore durante la creazione del server SSH');
    }
    return res.data.data;
  },

  updateServer: async (id: string, payload: UpdateSSHServerPayload): Promise<SSHServerItem> => {
    const res = await apiClient.put<ApiResponse<SSHServerItem>>(`/ssh/servers/${id}`, payload);
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Errore durante l\'aggiornamento del server SSH');
    }
    return res.data.data;
  },

  deleteServer: async (id: string): Promise<void> => {
    const res = await apiClient.delete<ApiResponse<{ message: string }>>(`/ssh/servers/${id}`);
    if (!res.data.success) {
      throw new Error(res.data.error?.message || 'Errore durante l\'eliminazione del server SSH');
    }
  },

  testConnection: async (id: string): Promise<SSHConnectionTestResult> => {
    const res = await apiClient.post<ApiResponse<SSHConnectionTestResult>>(`/ssh/servers/${id}/test-connection`);
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Test di connettività fallito');
    }
    return res.data.data;
  },

  getFirewallStatus: async (serverId: string): Promise<FirewallStatus> => {
    const res = await apiClient.get<ApiResponse<FirewallStatus>>(`/ssh/servers/${serverId}/firewall`);
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Impossibile recuperare lo stato del firewall');
    }
    return res.data.data;
  },

  addFirewallRule: async (serverId: string, payload: AddFirewallRulePayload): Promise<FirewallStatus> => {
    const res = await apiClient.post<ApiResponse<FirewallStatus>>(`/ssh/servers/${serverId}/firewall/rules`, payload);
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Errore durante l\'inserimento della regola firewall');
    }
    return res.data.data;
  },

  deleteFirewallRule: async (
    serverId: string,
    lineNum: number,
    chain: ChainName = 'INPUT',
    withSafetyTimer: boolean = true,
    timeout: number = 60
  ): Promise<FirewallStatus> => {
    const res = await apiClient.delete<ApiResponse<FirewallStatus>>(
      `/ssh/servers/${serverId}/firewall/rules/${lineNum}?chain=${chain}&safety_timer=${withSafetyTimer}&timeout=${timeout}`
    );
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Errore durante la cancellazione della regola firewall');
    }
    return res.data.data;
  },

  applyWithSafetyTimer: async (serverId: string, timeoutSec: number = 60): Promise<FirewallStatus> => {
    const res = await apiClient.post<ApiResponse<FirewallStatus>>(`/ssh/servers/${serverId}/firewall/apply`, {
      safety_timeout_seconds: timeoutSec,
    });
    if (!res.data.success || !res.data.data) {
      throw new Error(res.data.error?.message || 'Errore durante l\'attivazione del safety timer');
    }
    return res.data.data;
  },

  confirmFirewall: async (serverId: string): Promise<void> => {
    const res = await apiClient.post<ApiResponse<{ message: string }>>(`/ssh/servers/${serverId}/firewall/confirm`);
    if (!res.data.success) {
      throw new Error(res.data.error?.message || 'Errore durante la conferma della configurazione firewall');
    }
  },

  rollbackFirewall: async (serverId: string, backupId?: string): Promise<void> => {
    const res = await apiClient.post<ApiResponse<{ message: string }>>(`/ssh/servers/${serverId}/firewall/rollback`, {
      backup_id: backupId || undefined,
    });
    if (!res.data.success) {
      throw new Error(res.data.error?.message || 'Errore durante il rollback delle regole firewall');
    }
  },
};

