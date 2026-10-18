import { apiClient } from './client';
import type { ApiResponse } from '../types/auth';

export interface VirtualPatchFile { engine: string; filename: string; content: string }
export interface VirtualPatchBundle {
  finding_id: string;
  cve_id: string;
  catalog_revision: string;
  supported: boolean;
  product: string;
  strategy: string;
  operational_impact: string;
  covered_cves: string[];
  reason: string;
  limitations: string[];
  sources: string[];
  files: VirtualPatchFile[];
}

export async function getVirtualPatches(id: string): Promise<VirtualPatchBundle> {
  const response = await apiClient.get<ApiResponse<VirtualPatchBundle>>(`/findings/${encodeURIComponent(id)}/virtual-patches`);
  if (!response.data.success || !response.data.data) throw new Error(response.data.error?.message || 'Template non disponibili');
  if (response.data.data.finding_id !== id) throw new Error('Template non coerenti con il finding selezionato');
  return response.data.data;
}
