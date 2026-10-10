import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { ApiResponse, TokenPair } from '../types/auth';
import { debugLogger } from '../utils/debugLogger';

interface DebuggableRequestConfig extends InternalAxiosRequestConfig {
  _debugReqId?: string;
  _debugStartTime?: number;
  _retry?: boolean;
}

const ACCESS_TOKEN_KEY = 'nextgen_access_token';
const REFRESH_TOKEN_KEY = 'nextgen_refresh_token';

// Helper di storage
export const getStoredAccessToken = (): string | null => {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
};

export const getStoredRefreshToken = (): string | null => {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
};

export const storeTokens = (tokens: TokenPair): void => {
  localStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token);
  localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token);
};

export const purgeStoredTokens = (): void => {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
};

// Listener per notificare AuthContext in caso di logout forzato (token non rinnovabile)
type AuthExpiredListener = () => void;
const authExpiredListeners: AuthExpiredListener[] = [];

export const onAuthExpired = (listener: AuthExpiredListener): (() => void) => {
  authExpiredListeners.push(listener);
  return () => {
    const idx = authExpiredListeners.indexOf(listener);
    if (idx > -1) authExpiredListeners.splice(idx, 1);
  };
};

const notifyAuthExpired = (): void => {
  purgeStoredTokens();
  authExpiredListeners.forEach(listener => listener());
};

// Creazione istanza Axios centralizzata
export const apiClient = axios.create({
  baseURL: '/api/v1',
  timeout: 120000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptor per le richieste: traccia la chiamata su debugLogger e inietta il Bearer token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const debugConfig = config as DebuggableRequestConfig;
    debugConfig._debugStartTime = Date.now();

    // Registra la richiesta su debugLogger
    const reqId = debugLogger.logApiRequest({
      method: config.method || 'GET',
      url: config.url || '',
      data: config.data,
      headers: config.headers ? (config.headers as unknown as Record<string, unknown>) : undefined,
    });
    debugConfig._debugReqId = reqId;

    const token = getStoredAccessToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    debugLogger.logApiError({
      method: error?.config?.method || 'UNKNOWN',
      url: error?.config?.url || 'UNKNOWN',
      error,
    });
    return Promise.reject(error);
  }
);

// Gestione della coda per richieste concorrenti durante il refresh token
interface QueuedRequest {
  resolve: (token: string) => void;
  reject: (error: unknown) => void;
}

let isRefreshing = false;
let failedQueue: QueuedRequest[] = [];

const processQueue = (error: unknown, token: string | null = null): void => {
  failedQueue.forEach(prom => {
    if (token) {
      prom.resolve(token);
    } else {
      prom.reject(error);
    }
  });
  failedQueue = [];
};

// Interceptor per le risposte: logga su debugLogger, intercetta 401 e tenta il refresh automatico
apiClient.interceptors.response.use(
  (response) => {
    const config = response.config as DebuggableRequestConfig;
    const durationMs = config?._debugStartTime ? Date.now() - config._debugStartTime : undefined;
    debugLogger.logApiResponse({
      id: config?._debugReqId ? `${config._debugReqId}-res` : undefined,
      method: config?.method || 'GET',
      url: config?.url || '',
      status: response.status,
      durationMs,
      data: response.data,
    });
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as DebuggableRequestConfig;
    const durationMs = originalRequest?._debugStartTime ? Date.now() - originalRequest._debugStartTime : undefined;

    // Log dell'errore verso debugLogger
    debugLogger.logApiError({
      id: originalRequest?._debugReqId ? `${originalRequest._debugReqId}-err` : undefined,
      method: originalRequest?.method || 'UNKNOWN',
      url: originalRequest?.url || 'UNKNOWN',
      status: error.response?.status,
      durationMs,
      error,
      requestData: originalRequest?.data,
      responseData: error.response?.data,
    });

    // Se l'errore non è 401 o la richiesta non ha config, propaga l'errore
    if (!error.response || error.response.status !== 401 || !originalRequest) {
      return Promise.reject(error);
    }

    // Se l'errore 401 deriva da login o refresh stesso, non tentare ulteriore refresh
    const requestUrl = originalRequest.url || '';
    if (requestUrl.includes('/auth/login') || requestUrl.includes('/auth/refresh')) {
      return Promise.reject(error);
    }

    // Se abbiamo già tentato il refresh per questa richiesta, logout ed esci
    if (originalRequest._retry) {
      notifyAuthExpired();
      return Promise.reject(error);
    }

    const refreshToken = getStoredRefreshToken();
    if (!refreshToken) {
      notifyAuthExpired();
      return Promise.reject(error);
    }

    // Se un'altra richiesta sta già eseguendo il refresh, accoda questa richiesta
    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      })
        .then((newAccessToken) => {
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
          }
          return apiClient(originalRequest);
        })
        .catch((err) => Promise.reject(err));
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      // Invocazione diretta via axios standard per evitare loop di interceptor
      const refreshResponse = await axios.post<ApiResponse<TokenPair>>('/api/v1/auth/refresh', {
        refresh_token: refreshToken,
      });

      if (refreshResponse.data && refreshResponse.data.success && refreshResponse.data.data) {
        const newTokens = refreshResponse.data.data;
        storeTokens(newTokens);

        processQueue(null, newTokens.access_token);

        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${newTokens.access_token}`;
        }
        return apiClient(originalRequest);
      } else {
        throw new Error('Formato risposta refresh non valido');
      }
    } catch (refreshErr) {
      processQueue(refreshErr, null);
      notifyAuthExpired();
      return Promise.reject(refreshErr);
    } finally {
      isRefreshing = false;
    }
  }
);

export { reportsApi } from './reports';
export { discoveryApi } from './discovery';
export { sshApi } from './ssh';

/**
 * Richiede un ticket monouso di breve durata (TTL 30s) per connessioni real-time (SSE e WebSocket)
 * evitando l'esposizione del token JWT a lungo termine nella query string (NFR-SEC-04).
 */
export async function getStreamTicket(): Promise<string> {
  const response = await apiClient.post<{ success: boolean; data: { ticket: string; expires_in: number } }>(
    '/auth/stream-ticket'
  );
  return response.data.data.ticket;
}

