import React, { useState } from 'react';
import {
  X,
  Search,
  Trash2,
  Download,
  Sparkles,
  Check,
  ChevronDown,
  ChevronRight,
  AlertCircle,
  AlertTriangle,
  Info,
  Copy,
  Clock,
  Radio,
  Server,
  Activity,
  Code2,
} from 'lucide-react';
import { useDebugConsole } from '../../context/DebugContext';
import { DebugEntry, DebugLogLevel, DebugLogSource } from '../../types/debug';
import {
  generateAiMarkdownReport,
  exportLogsToJson,
  copyTextToClipboard,
  triggerFileDownload,
} from '../../utils/markdownExporter';

export const DebugConsoleDrawer: React.FC = () => {
  const {
    isOpen,
    closeDrawer,
    filteredEntries,
    entries,
    stats,
    filterOptions,
    setFilterOptions,
    clearLogs,
  } = useDebugConsole();

  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [copiedMd, setCopiedMd] = useState(false);
  const [copiedJson, setCopiedJson] = useState(false);
  const [copiedSingleId, setCopiedSingleId] = useState<string | null>(null);

  if (!isOpen) {
    return null;
  }

  const toggleExpand = (id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const handleCopyAiMarkdown = async () => {
    const md = generateAiMarkdownReport(entries, stats);
    const success = await copyTextToClipboard(md);
    if (success) {
      setCopiedMd(true);
      setTimeout(() => setCopiedMd(false), 2000);
    }
  };

  const handleExportJson = () => {
    const jsonStr = exportLogsToJson(entries);
    triggerFileDownload(
      jsonStr,
      `nextgen-debug-logs-${new Date().toISOString().replace(/[:.]/g, '-')}.json`,
      'application/json'
    );
    setCopiedJson(true);
    setTimeout(() => setCopiedJson(false), 2000);
  };

  const handleCopySingle = async (e: React.MouseEvent, entry: DebugEntry) => {
    e.stopPropagation();
    const jsonStr = JSON.stringify(entry, null, 2);
    const success = await copyTextToClipboard(jsonStr);
    if (success) {
      setCopiedSingleId(entry.id);
      setTimeout(() => setCopiedSingleId(null), 1500);
    }
  };

  const levelBadges: Array<{ key: DebugLogLevel | 'all'; label: string; count?: number; color: string }> = [
    { key: 'all', label: 'Tutti', count: stats.total, color: 'text-slate-300' },
    { key: 'error', label: 'Error', count: stats.errors, color: 'text-red-400' },
    { key: 'warn', label: 'Warn', count: stats.warnings, color: 'text-yellow-400' },
    { key: 'info', label: 'Info', count: stats.info, color: 'text-cyan-400' },
  ];

  const sourceBadges: Array<{ key: DebugLogSource | 'all'; label: string }> = [
    { key: 'all', label: 'Tutte le sorgenti' },
    { key: 'api', label: 'API REST' },
    { key: 'sse', label: 'Live SSE' },
    { key: 'ws', label: 'WebSocket' },
    { key: 'runtime', label: 'Runtime JS' },
    { key: 'app', label: 'App Logs' },
  ];

  const getSourceIcon = (source: DebugLogSource) => {
    switch (source) {
      case 'api':
        return <Server className="w-3.5 h-3.5 text-blue-400" />;
      case 'sse':
        return <Radio className="w-3.5 h-3.5 text-emerald-400" />;
      case 'ws':
        return <Activity className="w-3.5 h-3.5 text-purple-400" />;
      case 'runtime':
        return <AlertCircle className="w-3.5 h-3.5 text-rose-400" />;
      case 'app':
        return <Code2 className="w-3.5 h-3.5 text-amber-400" />;
      default:
        return null;
    }
  };

  const getLevelIcon = (level: DebugLogLevel) => {
    switch (level) {
      case 'error':
        return <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />;
      case 'warn':
        return <AlertTriangle className="w-4 h-4 text-yellow-400 shrink-0" />;
      case 'info':
        return <Info className="w-4 h-4 text-cyan-400 shrink-0" />;
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 backdrop-blur-xs transition-opacity"
        onClick={closeDrawer}
        aria-hidden="true"
      />

      {/* Slide-over panel */}
      <div className="fixed inset-y-0 right-0 max-w-full flex pl-6 sm:pl-10">
        <div className="w-screen max-w-2xl lg:max-w-3xl bg-[#0b0f19] border-l border-cyan-500/30 text-slate-100 flex flex-col shadow-2xl">
          {/* Header */}
          <div className="p-4 sm:p-5 border-b border-cyber-border bg-[#101524]/90 flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
                <Activity className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-base sm:text-lg font-bold text-white flex items-center space-x-2">
                  <span>Live Debug & Diagnostics Console</span>
                  <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700">
                    Ctrl+Shift+D
                  </span>
                </h2>
                <p className="text-xs text-slate-400">
                  Buffer circolare real-time (Max 500) • Dati sensibili mascherati [REDACTED]
                </p>
              </div>
            </div>

            <button
              onClick={closeDrawer}
              className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/80 transition-colors"
              aria-label="Chiudi console"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Action Toolbar */}
          <div className="p-3 sm:p-4 border-b border-cyber-border bg-slate-900/60 space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-2">
              {/* Pulsante Copia per AI (Markdown) */}
              <button
                onClick={handleCopyAiMarkdown}
                data-testid="copy-ai-btn"
                className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-cyan-500/20 hover:bg-cyan-500/30 border border-cyan-500/40 text-cyan-300 text-xs font-semibold shadow-xs transition-all"
                title="Copia report Markdown strutturato pronto per la chat AI"
              >
                {copiedMd ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Sparkles className="w-3.5 h-3.5" />}
                <span>{copiedMd ? 'Copiato per AI! ✓' : 'Copia per AI (Markdown)'}</span>
              </button>

              <div className="flex items-center space-x-2">
                {/* Esporta JSON */}
                <button
                  onClick={handleExportJson}
                  data-testid="export-json-btn"
                  className="flex items-center space-x-1 px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 border border-cyber-border text-slate-300 text-xs font-medium transition-colors"
                  title="Scarica log in JSON"
                >
                  {copiedJson ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Download className="w-3.5 h-3.5" />}
                  <span>JSON</span>
                </button>

                {/* Svuota Buffer */}
                <button
                  onClick={clearLogs}
                  data-testid="clear-logs-btn"
                  className="flex items-center space-x-1 px-2.5 py-1.5 rounded-lg bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 text-xs font-medium transition-colors"
                  title="Svuota buffer log"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  <span>Pulisci ({stats.total})</span>
                </button>
              </div>
            </div>

            {/* Ricerca e Filtri */}
            <div className="flex flex-col sm:flex-row gap-2">
              {/* Input Cerca */}
              <div className="relative flex-1">
                <Search className="w-4 h-4 absolute left-3 top-2.5 text-slate-500" />
                <input
                  type="text"
                  placeholder="Cerca per titolo, URL, o errore..."
                  value={filterOptions.search || ''}
                  onChange={(e) => setFilterOptions({ search: e.target.value })}
                  className="w-full pl-9 pr-3 py-1.5 bg-[#0e1320] border border-cyber-border rounded-lg text-xs text-white placeholder-slate-500 focus:outline-hidden focus:border-cyan-500/60"
                />
              </div>

              {/* Selettore Sorgente */}
              <select
                value={filterOptions.source || 'all'}
                onChange={(e) => setFilterOptions({ source: e.target.value as DebugLogSource | 'all' })}
                className="bg-[#0e1320] border border-cyber-border rounded-lg px-2.5 py-1.5 text-xs text-slate-300 focus:outline-hidden focus:border-cyan-500/60"
              >
                {sourceBadges.map((s) => (
                  <option key={s.key} value={s.key}>
                    {s.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Badge Filtri di Livello */}
            <div className="flex items-center space-x-1.5 pt-1">
              {levelBadges.map((lvl) => {
                const isActive = (filterOptions.level || 'all') === lvl.key;
                return (
                  <button
                    key={lvl.key}
                    data-testid={`filter-level-${lvl.key}`}
                    onClick={() => setFilterOptions({ level: lvl.key })}
                    className={`px-2.5 py-1 rounded-md text-xs font-medium transition-all ${
                      isActive
                        ? 'bg-slate-700 border border-cyan-500/50 text-white shadow-xs'
                        : 'bg-slate-900/80 border border-transparent text-slate-400 hover:text-slate-200'
                    }`}
                  >
                    <span className={lvl.color}>{lvl.label}</span>
                    {lvl.count !== undefined && (
                      <span className="ml-1.5 text-[10px] text-slate-400 font-mono">
                        ({lvl.count})
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          </div>

          {/* Log List View */}
          <div className="flex-1 overflow-y-auto p-3 sm:p-4 space-y-2">
            {filteredEntries.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-48 text-center text-slate-500 text-xs">
                <Info className="w-8 h-8 text-slate-600 mb-2" />
                <p>Nessun log corrisponde ai criteri di filtro correnti.</p>
                <p className="text-[11px] mt-1 text-slate-600">
                  Totale log in memoria: {stats.total}
                </p>
              </div>
            ) : (
              filteredEntries.map((entry) => {
                const isExpanded = expandedIds.has(entry.id);
                const timeStr = entry.timestamp.split('T')[1]?.replace('Z', '') || entry.timestamp;

                return (
                  <div
                    key={entry.id}
                    data-testid={`log-entry-${entry.id}`}
                    onClick={() => toggleExpand(entry.id)}
                    className={`p-3 rounded-lg border transition-all cursor-pointer ${
                      entry.level === 'error'
                        ? 'bg-red-950/20 border-red-500/30 hover:border-red-500/50'
                        : entry.level === 'warn'
                        ? 'bg-yellow-950/20 border-yellow-500/30 hover:border-yellow-500/50'
                        : 'bg-slate-900/50 border-cyber-border hover:border-cyan-500/30'
                    }`}
                  >
                    {/* Header Riga Log */}
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex items-start space-x-2.5 flex-1 min-w-0">
                        <div className="mt-0.5">{getLevelIcon(entry.level)}</div>

                        <div className="flex-1 min-w-0">
                          <div className="flex flex-wrap items-center gap-1.5 text-xs">
                            {/* Badge Sorgente */}
                            <span className="inline-flex items-center space-x-1 px-1.5 py-0.5 rounded bg-slate-800 text-[10px] text-slate-300 border border-slate-700">
                              {getSourceIcon(entry.source)}
                              <span>{entry.source.toUpperCase()}</span>
                            </span>

                            {/* Badge Metodo HTTP o Status */}
                            {entry.method && (
                              <span className="font-mono text-[10px] px-1 py-0.5 rounded bg-cyan-950/60 text-cyan-300 border border-cyan-800/40">
                                {entry.method}
                              </span>
                            )}

                            {entry.status !== undefined && (
                              <span
                                className={`font-mono text-[10px] px-1.5 py-0.5 rounded font-bold ${
                                  entry.status >= 500
                                    ? 'bg-red-900/80 text-red-200'
                                    : entry.status >= 400
                                    ? 'bg-amber-900/80 text-amber-200'
                                    : 'bg-emerald-900/80 text-emerald-200'
                                }`}
                              >
                                {entry.status}
                              </span>
                            )}

                            {/* Latenza */}
                            {entry.durationMs !== undefined && (
                              <span className="inline-flex items-center text-[10px] text-slate-400">
                                <Clock className="w-2.5 h-2.5 mr-0.5" />
                                {entry.durationMs}ms
                              </span>
                            )}

                            {/* Orario */}
                            <span className="text-[10px] text-slate-500 font-mono ml-auto">
                              {timeStr}
                            </span>
                          </div>

                          {/* Titolo Principale */}
                          <div className="text-xs font-medium text-slate-200 mt-1 truncate">
                            {entry.title}
                          </div>

                          {/* Messaggio di Errore Rapido */}
                          {entry.error && (
                            <div className="text-[11px] text-red-400 mt-0.5 font-mono truncate">
                              {entry.error}
                            </div>
                          )}
                        </div>
                      </div>

                      {/* Icone Espandi / Copia */}
                      <div className="flex items-center space-x-1 shrink-0 ml-2">
                        <button
                          onClick={(e) => handleCopySingle(e, entry)}
                          className="p-1 rounded text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-colors"
                          title="Copia record JSON"
                        >
                          {copiedSingleId === entry.id ? (
                            <Check className="w-3.5 h-3.5 text-emerald-400" />
                          ) : (
                            <Copy className="w-3.5 h-3.5" />
                          )}
                        </button>
                        <div className="text-slate-500">
                          {isExpanded ? (
                            <ChevronDown className="w-4 h-4" />
                          ) : (
                            <ChevronRight className="w-4 h-4" />
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Dettaglio Espanso */}
                    {isExpanded && (
                      <div className="mt-3 pt-3 border-t border-slate-800/80 space-y-2 text-xs">
                        {entry.url && (
                          <div>
                            <span className="text-[11px] text-slate-400 font-semibold block">URL Endpoint:</span>
                            <span className="font-mono text-[11px] text-cyan-300 break-all select-all">
                              {entry.url}
                            </span>
                          </div>
                        )}

                        {entry.requestData !== undefined && (
                          <div>
                            <span className="text-[11px] text-slate-400 font-semibold block mb-0.5">
                              Request Data (Sanificato):
                            </span>
                            <pre className="p-2 rounded bg-black/60 border border-slate-800 font-mono text-[10px] text-slate-300 overflow-x-auto max-h-40">
                              {JSON.stringify(entry.requestData, null, 2)}
                            </pre>
                          </div>
                        )}

                        {entry.responseData !== undefined && (
                          <div>
                            <span className="text-[11px] text-slate-400 font-semibold block mb-0.5">
                              Response Data (Sanificato):
                            </span>
                            <pre className="p-2 rounded bg-black/60 border border-slate-800 font-mono text-[10px] text-slate-300 overflow-x-auto max-h-40">
                              {JSON.stringify(entry.responseData, null, 2)}
                            </pre>
                          </div>
                        )}

                        {entry.stack && (
                          <div>
                            <span className="text-[11px] text-red-400 font-semibold block mb-0.5">
                              Stack Trace:
                            </span>
                            <pre className="p-2 rounded bg-black/60 border border-red-950 font-mono text-[10px] text-red-300 overflow-x-auto max-h-40">
                              {entry.stack}
                            </pre>
                          </div>
                        )}

                        {entry.metadata && Object.keys(entry.metadata).length > 0 && (
                          <div>
                            <span className="text-[11px] text-slate-400 font-semibold block mb-0.5">
                              Metadati:
                            </span>
                            <pre className="p-2 rounded bg-black/60 border border-slate-800 font-mono text-[10px] text-slate-300 overflow-x-auto max-h-40">
                              {JSON.stringify(entry.metadata, null, 2)}
                            </pre>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
