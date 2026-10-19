import { apiClient } from './client';
import type { ApiResponse } from '../types/auth';

export type AlertKind = 'slack' | 'teams' | 'telegram' | 'webhook' | 'smtp';
export interface AlertConfig {
  endpoint?: string; token?: string; chat_id?: string; hmac_secret?: string;
  smtp_host?: string; smtp_port?: number; smtp_username?: string; smtp_password?: string; smtp_from?: string; smtp_to?: string[];
}
export interface AlertChannel {
  id: string; tenant_id: string; company_id?: string; name: string; kind: AlertKind; enabled: boolean;
  events: string[]; created_at: string; updated_at: string;
}
export interface AlertChannelInput {
  name: string; kind: AlertKind; enabled: boolean; company_id?: string; events: string[];
  expected_updated_at?: string; config?: AlertConfig;
}

function data<T>(response: { data: ApiResponse<T> }): T {
  if (!response.data.success || response.data.data === undefined || response.data.data === null) throw new Error('Operazione canale non completata');
  return response.data.data;
}
export const alertsApi = {
  list: async (): Promise<AlertChannel[]> => data(await apiClient.get<ApiResponse<AlertChannel[]>>('/alert-channels/')),
  save: async (input: AlertChannelInput, id?: string): Promise<AlertChannel> => id
    ? data(await apiClient.put<ApiResponse<AlertChannel>>(`/alert-channels/${encodeURIComponent(id)}`, input))
    : data(await apiClient.post<ApiResponse<AlertChannel>>('/alert-channels/', input)),
  remove: async (id: string): Promise<void> => { await apiClient.delete(`/alert-channels/${encodeURIComponent(id)}`); },
  test: async (kind: AlertKind, config: AlertConfig): Promise<{ accepted: boolean; code: string; message: string }> =>
    data(await apiClient.post<ApiResponse<{ accepted: boolean; code: string; message: string }>>('/alert-channels/test', { kind, config, confirmed_external_send: true })),
};
