import { apiClient, getStoredAccessToken } from './client';
import { ApiResponse } from '../types/auth';
import {
  DiscoveryResult,
  DiscoveryScanParams,
  ImportTargetsRequestPayload,
  ImportTargetsResponseData,
} from '../types/discovery';

export const getDiscoveryStreamUrl = (sessionId: string): string => {
  const token = getStoredAccessToken();
  const base = `/api/v1/network-discovery/stream?session_id=${encodeURIComponent(sessionId)}`;
  return token ? `${base}&token=${encodeURIComponent(token)}` : base;
};

export const discoveryApi = {
  getStreamUrl: getDiscoveryStreamUrl,

  /**
   * Avvia una scansione Ping/ARP sweep veloce su subnet o range IP
   */
  scan: async (params: DiscoveryScanParams): Promise<DiscoveryResult> => {
    const response = await apiClient.post<ApiResponse<DiscoveryResult>>('/network-discovery/scan', params);
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Errore durante la scansione della rete locale');
    }
    return response.data.data;
  },

  /**
   * Recupera gli ultimi risultati di ricognizione censiti per il tenant corrente
   */
  getLastResults: async (): Promise<DiscoveryResult | null> => {
    try {
      const response = await apiClient.get<ApiResponse<DiscoveryResult>>('/network-discovery/results');
      if (response.data.success && response.data.data) {
        return response.data.data;
      }
      throw new Error(response.data.error?.message || 'Errore durante il recupero dei risultati discovery');
    } catch (error) {
      if (isAxiosError(error) && error.response?.status === 404) return null;
      throw error;
    }
  },

  /**
   * Importa massivamente i dispositivi selezionati nei Target ufficiali del tenant
   */
  importTargets: async (payload: ImportTargetsRequestPayload): Promise<ImportTargetsResponseData> => {
    const response = await apiClient.post<ApiResponse<ImportTargetsResponseData>>('/network-discovery/import-targets', payload);
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Errore durante l\'importazione dei target');
    }
    return response.data.data;
  },
};

import { isAxiosError } from 'axios';
