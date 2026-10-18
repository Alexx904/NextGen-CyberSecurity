import { apiClient } from './client';
import { ApiResponse } from '../types/auth';
import { AttackGraph, AttackGraphFilterParams, Chokepoint, ChokepointRemediationResult } from '../types/attackPath';

export const attackPathApi = {
  previewRemediation: async (nodeId: string, serverId: string): Promise<ChokepointRemediationResult> => {
    const res = await apiClient.post<ApiResponse<ChokepointRemediationResult>>(`/attack-paths/chokepoints/${encodeURIComponent(nodeId)}/preview`, { server_id: serverId });
    if (!res.data.success || !res.data.data) throw new Error(res.data.error?.message || 'Preview non disponibile');
    return res.data.data;
  },
  remediate: async (nodeId: string, serverId: string): Promise<ChokepointRemediationResult> => {
    const res = await apiClient.post<ApiResponse<ChokepointRemediationResult>>(`/attack-paths/chokepoints/${encodeURIComponent(nodeId)}/remediate`, { server_id: serverId });
    if (!res.data.success || !res.data.data) throw new Error(res.data.error?.message || 'Applicazione non confermata; verifica il firewall');
    return res.data.data;
  },
  /**
   * Recupera la topologia completa del grafo delle minacce e dei percorsi di attacco.
   * Accetta opzionalmente min_cvss per filtrare vulnerabilità a bassa criticità.
   */
  getGraph: async (params?: AttackGraphFilterParams): Promise<AttackGraph> => {
    const response = await apiClient.get<ApiResponse<AttackGraph>>('/attack-paths/graph', {
      params,
    });
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Impossibile recuperare il grafo di attacco');
    }
    return response.data.data;
  },

  /**
   * Recupera l'elenco prioritizzato dei chokepoint di remediation per il tenant.
   */
  getChokepoints: async (): Promise<Chokepoint[]> => {
    const response = await apiClient.get<ApiResponse<Chokepoint[]>>('/attack-paths/chokepoints');
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Impossibile recuperare i chokepoint');
    }
    return response.data.data;
  },
};

