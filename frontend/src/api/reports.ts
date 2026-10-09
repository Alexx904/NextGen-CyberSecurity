import { apiClient } from './client';
import { ApiResponse } from '../types/auth';
import { ReportFilterParams, ReportListResponseData, ReportFormat, ReportTemplate } from '../types/report';
import { downloadReportFile } from '../utils/download';

export const reportsApi = {
  /**
   * Recupera l'elenco paginato dei report aggregati con filtri opzionali
   */
  list: async (params?: ReportFilterParams): Promise<ReportListResponseData> => {
    const response = await apiClient.get<ApiResponse<ReportListResponseData>>('/reports', {
      params,
    });
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Impossibile recuperare lo storico dei report');
    }
    return response.data.data;
  },

  /**
   * Esegue il download del report nel formato (pdf|csv) e template (executive|technical) specificato
   */
  download: async (
    scanId: string,
    format: ReportFormat = 'pdf',
    template: ReportTemplate = 'executive',
    callbacks?: { onStart?: () => void; onSuccess?: () => void; onError?: (err: Error) => void }
  ): Promise<void> => {
    const ext = format === 'pdf' ? 'pdf' : 'csv';
    const defaultFilename = `report_${template}_scan_${scanId.slice(0, 8)}.${ext}`;
    return downloadReportFile({
      url: `/reports/${scanId}/download?type=${format}&template=${template}`,
      defaultFilename,
      onStart: callbacks?.onStart,
      onSuccess: callbacks?.onSuccess,
      onError: callbacks?.onError,
    });
  },
};

