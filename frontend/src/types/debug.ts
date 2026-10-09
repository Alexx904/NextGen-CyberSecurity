/**
 * NextGen CyberSecurity - Debug & Testing Console Types
 * Definisce i contratti di tipo per il logger client-side e il drawer di diagnostica.
 */

export type DebugLogLevel = 'info' | 'warn' | 'error';
export type DebugLogSource = 'api' | 'sse' | 'ws' | 'runtime' | 'app';

export interface DebugEntry {
  id: string;
  timestamp: string; // ISO 8601 string
  level: DebugLogLevel;
  source: DebugLogSource;
  title: string;
  url?: string;
  method?: string;
  status?: number;
  durationMs?: number;
  requestData?: unknown;
  responseData?: unknown;
  error?: string;
  stack?: string;
  metadata?: Record<string, unknown>;
}

export type DebugListener = (entry: DebugEntry, allEntries: DebugEntry[]) => void;
export type Unsubscribe = () => void;

export interface DebugLoggerStats {
  total: number;
  errors: number;
  warnings: number;
  info: number;
  oldestTimestamp?: string;
  newestTimestamp?: string;
}

export interface DebugFilterOptions {
  level?: DebugLogLevel | 'all';
  source?: DebugLogSource | 'all';
  search?: string;
}
