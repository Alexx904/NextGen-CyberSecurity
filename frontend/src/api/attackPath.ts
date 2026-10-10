import { apiClient } from './client';
import { ApiResponse } from '../types/auth';
import { AttackGraph, AttackGraphFilterParams, Chokepoint } from '../types/attackPath';

export const attackPathApi = {
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

