import { apiClient } from './client';
import type { ApiResponse } from '../types/auth';
import type { AssessmentAsset, ComplianceSnapshot, IssueCommandInput } from '../types/compliance';
export const complianceApi = {
  async snapshot(company: string, department?: string, signal?: AbortSignal) {
    const result = await apiClient.get<ApiResponse<ComplianceSnapshot>>('/compliance', { params: { company_id: company, department_id: department || undefined }, signal });
    if (!result.data.data) throw new Error('Dati compliance non disponibili');
    return result.data.data;
  },
  async asset(group_id: string, name: string) { return (await apiClient.post<ApiResponse<AssessmentAsset>>('/compliance/assets', { group_id, name })).data.data; },
  async observe(asset_id: string, finding_id: string) { await apiClient.post('/compliance/observations', { asset_id, finding_id }); },
  async command(id: string, command: IssueCommandInput) { await apiClient.post(`/compliance/issues/${id}/events`, command); },
  async evidence(input: { group_id: string; control_id: string; status: string; reason: string; reference: string; source_version: string }) { await apiClient.post('/compliance/evidence', input); },
};
