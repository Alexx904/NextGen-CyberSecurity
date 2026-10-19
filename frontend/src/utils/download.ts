import { apiClient } from '../api/client';

export interface DownloadFileOptions {
  url: string;
  defaultFilename: string;
  method?: 'get' | 'post';
  data?: unknown;
  onStart?: () => void;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

/**
 * downloadReportFile esegue una richiesta autenticata via apiClient (Axios),
 * gestendo il token JWT Bearer, il recupero del filename dall'header Content-Disposition
 * e il salvataggio nativo nel browser tramite Blob URL.
 */
export async function downloadReportFile({
  url,
  defaultFilename,
  method = 'get',
  data,
  onStart,
  onSuccess,
  onError,
}: DownloadFileOptions): Promise<void> {
  try {
    onStart?.();

    const response = method === 'post' ? await apiClient.post(url, data, { responseType: 'blob' }) : await apiClient.get(url, {
      responseType: 'blob',
    });

    // Estrazione del filename dall'header Content-Disposition se presente
    let filename = defaultFilename;
    const disposition = response.headers['content-disposition'] || response.headers['Content-Disposition'];
    if (disposition && typeof disposition === 'string') {
      const match = disposition.match(/filename\*?=['"]?(?:UTF-\d['"])?([^;\r\n"']*)['"]?/i);
      if (match && match[1]) {
        filename = decodeURIComponent(match[1].trim());
      }
    }

    const contentType = String(
      response.headers['content-type'] ||
      response.headers['Content-Type'] ||
      'application/octet-stream'
    );

    const blob = new Blob([response.data], { type: contentType });
    const downloadUrl = window.URL.createObjectURL(blob);

    const link = document.createElement('a');
    link.href = downloadUrl;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();

    // Pulizia del DOM e della memoria del browser
    document.body.removeChild(link);
    window.URL.revokeObjectURL(downloadUrl);

    onSuccess?.();
  } catch (err: any) {
    let errorMessage = 'Impossibile scaricare il file richiesto';

    // Se l'errore è un Blob contenente un payload JSON di errore dal server
    if (err.response?.data instanceof Blob && err.response.data.type?.includes('application/json')) {
      try {
        const text = await err.response.data.text();
        const parsed = JSON.parse(text);
        if (parsed.error?.message) {
          errorMessage = parsed.error.message;
        }
      } catch {
        // Fallback al messaggio di default
      }
    } else if (err.response?.data?.error?.message) {
      errorMessage = err.response.data.error.message;
    } else if (err.message) {
      errorMessage = err.message;
    }

    const error = new Error(errorMessage);
    onError?.(error);
    throw error;
  }
}
