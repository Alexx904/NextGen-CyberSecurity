import { apiClient, getStoredAccessToken } from './client';
import { ApiResponse } from '../types/auth';
import { ScanItem } from '../types/scan';
import { FindingListResponseData } from '../types/finding';

export const getScanStreamUrl = (scanId: string): string => {
  const token = getStoredAccessToken();
  const base = `/api/v1/scans/${encodeURIComponent(scanId)}/stream`;
  return token ? `${base}?token=${encodeURIComponent(token)}` : base;
};

export const scanApi = {
  getStreamUrl: getScanStreamUrl,

  /**
   * Recupera i dettagli completi di una scansione dato il suo ID
   */
  getScan: async (id: string): Promise<ScanItem> => {
    const response = await apiClient.get<ApiResponse<ScanItem>>(`/scans/${id}`);
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Impossibile recuperare i dettagli della scansione');
    }
    return response.data.data;
  },

  /**
   * Recupera la lista dei finding associati ad una scansione
   */
  getFindings: async (scanId: string, limit = 100): Promise<FindingListResponseData> => {
    const response = await apiClient.get<ApiResponse<FindingListResponseData>>(
      `/scans/${scanId}/findings?limit=${limit}`
    );
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Impossibile recuperare i finding della scansione');
    }
    return response.data.data;
  },

  /**
   * Richiede la cancellazione / interruzione graceful di una scansione in corso
   */
  cancelScan: async (scanId: string): Promise<void> => {
    const response = await apiClient.post<ApiResponse<void>>(`/scans/${scanId}/cancel`);
    if (!response.data.success) {
      throw new Error(response.data.error?.message || 'Impossibile annullare la scansione');
    }
  },
};

