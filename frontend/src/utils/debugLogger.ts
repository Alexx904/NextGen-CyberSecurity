/**
 * NextGen CyberSecurity - Debug Logger Engine
 * Ring-Buffer FIFO (Max 500 eventi), Sanitizzazione ricorsiva dei dati sensibili ([REDACTED]),
 * Event Bus Pub/Sub e helper tipizzati per API, SSE, WebSocket e Runtime errors.
 */

import {
  DebugEntry,
  DebugFilterOptions,
  DebugListener,
  DebugLogLevel,
  DebugLoggerStats,
  Unsubscribe,
} from '../types/debug';

export const DEFAULT_MAX_LOG_ENTRIES = 500;

// Pattern regex per individuare chiavi contenenti credenziali o segreti foglia
const LEAF_SECRET_KEY_REGEX =
  /(?:password|passwd|passphrase|secret|encrypted_secret|token|access_token|refresh_token|api[_-]?key|private[_-]?key|cert(?:ificate)?)/i;

// Pattern regex per container o header di autenticazione
const CONTAINER_OR_HEADER_KEY_REGEX =
  /^(?:auth|authorization|credentials)$/i;

// Pattern per Bearer token negli header HTTP
const BEARER_TOKEN_REGEX = /Bearer\s+([A-Za-z0-9\-_.~+/]+=*)/i;

// Pattern per Basic auth negli header
const BASIC_AUTH_REGEX = /Basic\s+([A-Za-z0-9+/=]+)/i;

// Pattern per JSON Web Token (JWT) inline (3 parti base64url separate da punti)
const JWT_PATTERN_REGEX = /^[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}$/;

/**
 * Genera un identificatore univoco robusto (UUID v4 nativo o fallback con timestamp/random)
 */
export function generateLogId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  const timestamp = Date.now().toString(36);
  const randomPart = Math.random().toString(36).substring(2, 10);
  return `dbg-${timestamp}-${randomPart}`;
}

/**
 * Sanitizza ricorsivamente un valore mascherando password, segreti e token con [REDACTED].
 * Include protezione nativa contro riferimenti circolari tramite WeakSet.
 * Non muta l'oggetto sorgente.
 */
export function sanitizeData(data: unknown, seen = new WeakSet<object>()): unknown {
  if (data === null || data === undefined) {
    return data;
  }

  // Se è un tipo primitivo non-stringa
  if (typeof data !== 'object' && typeof data !== 'string') {
    return data;
  }

  // Se è una stringa, controlliamo se è JSON serializzato, un Bearer/Basic token o JWT
  if (typeof data === 'string') {
    const trimmed = data.trim();
    if (
      (trimmed.startsWith('{') && trimmed.endsWith('}')) ||
      (trimmed.startsWith('[') && trimmed.endsWith(']'))
    ) {
      try {
        const parsed = JSON.parse(trimmed);
        if (typeof parsed === 'object' && parsed !== null) {
          return sanitizeData(parsed, seen);
        }
      } catch {
        // Non è JSON valido, continua come stringa
      }
    }

    if (BEARER_TOKEN_REGEX.test(data)) {
      return data.replace(BEARER_TOKEN_REGEX, 'Bearer [REDACTED]');
    }
    if (BASIC_AUTH_REGEX.test(data)) {
      return data.replace(BASIC_AUTH_REGEX, 'Basic [REDACTED]');
    }
    if (JWT_PATTERN_REGEX.test(trimmed)) {
      return '[REDACTED_JWT]';
    }
    return data;
  }

  // Gestione istanze Error
  if (data instanceof Error) {
    return {
      name: data.name,
      message: data.message,
      stack: data.stack,
    };
  }

  // Gestione Date
  if (data instanceof Date) {
    return data.toISOString();
  }

  // Protezione contro riferimenti circolari
  if (seen.has(data as object)) {
    return '[Circular]';
  }
  seen.add(data as object);

  // Se è un array
  if (Array.isArray(data)) {
    return data.map((item) => sanitizeData(item, seen));
  }

  // Se è un oggetto semplice
  const sanitizedObj: Record<string, unknown> = {};
  for (const [key, val] of Object.entries(data as Record<string, unknown>)) {
    const isObjectVal = typeof val === 'object' && val !== null && !(val instanceof Date) && !(val instanceof Error);

    if (isObjectVal) {
      if (CONTAINER_OR_HEADER_KEY_REGEX.test(key)) {
        // Se è un container tipo auth: { ... } o credentials: { ... }, naviga ricorsivamente
        sanitizedObj[key] = sanitizeData(val, seen);
      } else if (LEAF_SECRET_KEY_REGEX.test(key)) {
        // Se è una chiave foglia segreta esplicita ma è un oggetto
        sanitizedObj[key] = '[REDACTED]';
      } else {
        sanitizedObj[key] = sanitizeData(val, seen);
      }
    } else {
      // Il valore è una stringa o tipo primitivo o Date/Error
      if (typeof val === 'string') {
        if (BEARER_TOKEN_REGEX.test(val)) {
          sanitizedObj[key] = val.replace(BEARER_TOKEN_REGEX, 'Bearer [REDACTED]');
        } else if (BASIC_AUTH_REGEX.test(val)) {
          sanitizedObj[key] = val.replace(BASIC_AUTH_REGEX, 'Basic [REDACTED]');
        } else if (JWT_PATTERN_REGEX.test(val.trim())) {
          sanitizedObj[key] = '[REDACTED_JWT]';
        } else if (LEAF_SECRET_KEY_REGEX.test(key) || CONTAINER_OR_HEADER_KEY_REGEX.test(key)) {
          sanitizedObj[key] = '[REDACTED]';
        } else {
          sanitizedObj[key] = val;
        }
      } else if (LEAF_SECRET_KEY_REGEX.test(key) || CONTAINER_OR_HEADER_KEY_REGEX.test(key)) {
        sanitizedObj[key] = '[REDACTED]';
      } else {
        sanitizedObj[key] = sanitizeData(val, seen);
      }
    }
  }

  return sanitizedObj;
}

/**
 * Motore di logging client-side con Ring-Buffer e Pub/Sub Event Bus
 */
export class DebugLogger {
  private entries: DebugEntry[] = [];
  private maxEntries: number;
  private listeners: Set<DebugListener> = new Set();

  constructor(maxEntries = DEFAULT_MAX_LOG_ENTRIES) {
    this.maxEntries = maxEntries > 0 ? maxEntries : DEFAULT_MAX_LOG_ENTRIES;
  }

  /**
   * Registra una entry di debug nel ring buffer e notifica i subscriber
   */
  public log(entryInput: Omit<DebugEntry, 'id' | 'timestamp'> & { id?: string; timestamp?: string }): DebugEntry {
    const entry: DebugEntry = {
      id: entryInput.id || generateLogId(),
      timestamp: entryInput.timestamp || new Date().toISOString(),
      level: entryInput.level,
      source: entryInput.source,
      title: entryInput.title,
      url: entryInput.url,
      method: entryInput.method,
      status: entryInput.status,
      durationMs: entryInput.durationMs,
      requestData: entryInput.requestData !== undefined ? sanitizeData(entryInput.requestData) : undefined,
      responseData: entryInput.responseData !== undefined ? sanitizeData(entryInput.responseData) : undefined,
      error: entryInput.error,
      stack: entryInput.stack,
      metadata: entryInput.metadata !== undefined ? (sanitizeData(entryInput.metadata) as Record<string, unknown>) : undefined,
    };

    // Politica Ring-Buffer FIFO: se pieno, rimuove il più vecchio
    if (this.entries.length >= this.maxEntries) {
      this.entries.shift();
    }
    this.entries.push(entry);

    this.notifyListeners(entry);
    return entry;
  }

  /**
   * Helper per intercettare l'invio di una richiesta HTTP
   */
  public logApiRequest(params: {
    id?: string;
    method: string;
    url: string;
    data?: unknown;
    headers?: Record<string, unknown>;
  }): string {
    const correlationId = params.id || generateLogId();
    this.log({
      id: correlationId,
      level: 'info',
      source: 'api',
      title: `API REQ: ${params.method.toUpperCase()} ${params.url}`,
      url: params.url,
      method: params.method.toUpperCase(),
      requestData: params.data,
      metadata: params.headers ? { headers: params.headers } : undefined,
    });
    return correlationId;
  }

  /**
   * Helper per intercettare la ricezione di una risposta HTTP con successo
   */
  public logApiResponse(params: {
    id?: string;
    method: string;
    url: string;
    status: number;
    durationMs?: number;
    data?: unknown;
  }): DebugEntry {
    const level: DebugLogLevel = params.status >= 400 ? 'error' : params.status >= 300 ? 'warn' : 'info';
    return this.log({
      id: params.id,
      level,
      source: 'api',
      title: `API RES: ${params.method.toUpperCase()} ${params.url} [${params.status}]`,
      url: params.url,
      method: params.method.toUpperCase(),
      status: params.status,
      durationMs: params.durationMs,
      responseData: params.data,
    });
  }

  /**
   * Helper per intercettare un errore durante una chiamata HTTP
   */
  public logApiError(params: {
    id?: string;
    method: string;
    url: string;
    status?: number;
    durationMs?: number;
    error: unknown;
    requestData?: unknown;
    responseData?: unknown;
  }): DebugEntry {
    let errorMessage = 'Network Error';
    let stack: string | undefined;
    let responseData: unknown = params.responseData;

    if (params.error instanceof Error) {
      errorMessage = params.error.message;
      stack = params.error.stack;
    } else if (typeof params.error === 'string') {
      errorMessage = params.error;
    }

    if (params.error && typeof params.error === 'object') {
      const errObj = params.error as Record<string, unknown>;
      if (!errorMessage || errorMessage === 'Network Error') {
        if (typeof errObj.message === 'string') {
          errorMessage = errObj.message;
        }
      }
      if (responseData === undefined && errObj.response && typeof errObj.response === 'object') {
        const resp = errObj.response as Record<string, unknown>;
        if (typeof resp.status === 'number' && !params.status) {
          params.status = resp.status;
        }
        if (resp.data !== undefined) {
          responseData = resp.data;
        }
      }
    }

    return this.log({
      id: params.id,
      level: 'error',
      source: 'api',
      title: `API ERR: ${params.method.toUpperCase()} ${params.url}${params.status ? ` [${params.status}]` : ''}`,
      url: params.url,
      method: params.method.toUpperCase(),
      status: params.status,
      durationMs: params.durationMs,
      requestData: params.requestData,
      responseData,
      error: errorMessage,
      stack,
    });
  }

  /**
   * Helper per eventi Server-Sent Events (SSE)
   */
  public logSSEEvent(params: {
    eventType: string;
    data: unknown;
    url?: string;
    raw?: string;
  }): DebugEntry {
    const isError = params.eventType.toLowerCase().includes('error');
    return this.log({
      level: isError ? 'error' : 'info',
      source: 'sse',
      title: `SSE: ${params.eventType}`,
      url: params.url,
      responseData: params.data,
      metadata: params.raw ? { raw: params.raw } : undefined,
    });
  }

  /**
   * Helper per eventi o frame WebSocket
   */
  public logWSEvent(params: {
    direction: 'in' | 'out';
    data: unknown;
    url?: string;
    eventType?: string;
  }): DebugEntry {
    const title = `WS ${params.direction.toUpperCase()}: ${params.eventType || 'message'}`;
    return this.log({
      level: 'info',
      source: 'ws',
      title,
      url: params.url,
      requestData: params.direction === 'out' ? params.data : undefined,
      responseData: params.direction === 'in' ? params.data : undefined,
      metadata: { direction: params.direction, eventType: params.eventType },
    });
  }

  /**
   * Helper per catturare errori di runtime non gestiti (window.onerror, unhandledrejection)
   */
  public logRuntimeError(params: {
    error: unknown;
    message?: string;
    stack?: string;
    source?: string;
  }): DebugEntry {
    let msg = params.message || 'Unhandled Runtime Error';
    let stack = params.stack;

    if (params.error instanceof Error) {
      msg = params.message || params.error.message;
      stack = stack || params.error.stack;
    } else if (typeof params.error === 'string') {
      msg = params.message || params.error;
    }

    return this.log({
      level: 'error',
      source: 'runtime',
      title: `RUNTIME ERR: ${msg}`,
      error: msg,
      stack,
      metadata: params.source ? { source: params.source } : undefined,
    });
  }

  /**
   * Helper per log applicativi generici
   */
  public logCustom(params: {
    level: DebugLogLevel;
    title: string;
    data?: unknown;
    metadata?: Record<string, unknown>;
  }): DebugEntry {
    return this.log({
      level: params.level,
      source: 'app',
      title: params.title,
      requestData: params.data,
      metadata: params.metadata,
    });
  }

  /**
   * Restituisce una copia shallow dell'elenco completo dei log nel buffer
   */
  public getEntries(): DebugEntry[] {
    return [...this.entries];
  }

  /**
   * Restituisce le entry filtrate in base a opzioni
   */
  public getFilteredEntries(options?: DebugFilterOptions): DebugEntry[] {
    if (!options) return this.getEntries();
    const { level, source, search } = options;
    const lowerSearch = search ? search.toLowerCase().trim() : '';

    return this.entries.filter((entry) => {
      if (level && level !== 'all' && entry.level !== level) {
        return false;
      }
      if (source && source !== 'all' && entry.source !== source) {
        return false;
      }
      if (lowerSearch) {
        const titleMatch = entry.title.toLowerCase().includes(lowerSearch);
        const urlMatch = entry.url ? entry.url.toLowerCase().includes(lowerSearch) : false;
        const errorMatch = entry.error ? entry.error.toLowerCase().includes(lowerSearch) : false;
        return titleMatch || urlMatch || errorMatch;
      }
      return true;
    });
  }

  /**
   * Restituisce il numero totale di errori presenti nel buffer
   */
  public getErrorsCount(): number {
    return this.entries.filter((e) => e.level === 'error').length;
  }

  /**
   * Restituisce statistiche aggregate sul buffer di debug
   */
  public getStats(): DebugLoggerStats {
    let errors = 0;
    let warnings = 0;
    let info = 0;

    for (const entry of this.entries) {
      if (entry.level === 'error') errors++;
      else if (entry.level === 'warn') warnings++;
      else if (entry.level === 'info') info++;
    }

    return {
      total: this.entries.length,
      errors,
      warnings,
      info,
      oldestTimestamp: this.entries.length > 0 ? this.entries[0].timestamp : undefined,
      newestTimestamp: this.entries.length > 0 ? this.entries[this.entries.length - 1].timestamp : undefined,
    };
  }

  /**
   * Svuota completamente il buffer dei log e notifica i subscriber con lista vuota
   */
  public clear(): void {
    this.entries = [];
    const syntheticClearEntry: DebugEntry = {
      id: generateLogId(),
      timestamp: new Date().toISOString(),
      level: 'info',
      source: 'app',
      title: 'Console logs cleared',
    };
    this.notifyListeners(syntheticClearEntry);
  }

  /**
   * Imposta la capacità massima del buffer circolare (ridimensiona se necessario)
   */
  public setMaxEntries(max: number): void {
    this.maxEntries = max > 0 ? max : DEFAULT_MAX_LOG_ENTRIES;
    if (this.entries.length > this.maxEntries) {
      this.entries = this.entries.slice(this.entries.length - this.maxEntries);
    }
  }

  /**
   * Iscrive un listener per ricevere notifiche su ogni nuova entry
   */
  public subscribe(listener: DebugListener): Unsubscribe {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  private notifyListeners(entry: DebugEntry): void {
    const allEntries = this.getEntries();
    this.listeners.forEach((listener) => {
      try {
        listener(entry, allEntries);
      } catch (err) {
        // Isola gli errori nei listener per non interrompere il logger o l'applicazione
        if (typeof console !== 'undefined' && console.error) {
          console.error('[DebugLogger] Error in listener execution:', err);
        }
      }
    });
  }
}

// Singleton esportato per uso globale nell'applicazione frontend
export const debugLogger = new DebugLogger();
