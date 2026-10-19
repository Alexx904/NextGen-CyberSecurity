import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  ArrowLeft,
  Terminal,
  FileText,
  FileSpreadsheet,
  Server,
  Layers,
  ShieldAlert,
  Clock,
  RefreshCw,
  AlertCircle,
  Copy,
  Check,
  Zap,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { scanApi } from '../api/scan';
import { useSSE } from '../hooks/useSSE';
import { ApiResponse } from '../types/auth';
import {
  ScanItem,
  ScanPhasePayload,
  ScanProgressPayload,
  ScanFindingPayload,
  ScanStatusPayload,
  ScanErrorPayload,
} from '../types/scan';
import { FindingItem, FindingListResponseData } from '../types/finding';
import { ScanProgressBadge } from '../components/ScanProgressBadge';
import { ProgressEstimate, ScanTotalProgress } from '../components/scans/ScanTotalProgress';
import { SeverityChart } from '../components/SeverityChart';
import { FindingsTable } from '../components/FindingsTable';
import { downloadReportFile } from '../utils/download';

const SCAN_PHASES = [
  { key: 'HOST_DISCOVERY', label: 'Host Discovery', desc: 'Verifica host alive e ping sweep' },
  { key: 'PORT_SCAN', label: 'Port Scan', desc: 'Analisi porte aperte TCP/UDP' },
  { key: 'SERVICE_DETECTION', label: 'Service Detection', desc: 'Rilevamento servizi e versioni (sV)' },
  { key: 'NSE', label: 'Script NSE', desc: 'Esecuzione controlli script' },
  { key: 'CVE_CORRELATION', label: 'CVE Correlation', desc: 'Correlazione CPE/CVE vulnerabilità' },
];

export const ScanDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  const [scan, setScan] = useState<ScanItem | null>(null);
  const [findings, setFindings] = useState<FindingItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Stato monitoraggio streaming live SSE
  const [livePhase, setLivePhase] = useState<{
    phase: string;
    percentage: number;
    message: string;
  } | null>(null);
  const [liveError, setLiveError] = useState<string | null>(null);
  const [progressEstimate, setProgressEstimate] = useState<ProgressEstimate | null>(null);
  useEffect(() => { setProgressEstimate(null); setLivePhase(null); }, [id]);
  const [, setNewlyDiscoveredFindingIds] = useState<Set<string>>(new Set());

  const [copiedId, setCopiedId] = useState<boolean>(false);
  const [downloadingPdf, setDownloadingPdf] = useState<boolean>(false);
  const [downloadingCsv, setDownloadingCsv] = useState<boolean>(false);
  const [downloadError, setDownloadError] = useState<string | null>(null);

  const isLive = scan?.status === 'RUNNING' || scan?.status === 'QUEUED';

  // URL dello stream Server-Sent Events per scansioni attive
  const streamUrl = useMemo(() => {
    return id && isLive ? scanApi.getStreamUrl(id) : null;
  }, [id, isLive]);

  // Hook Server-Sent Events per il monitoraggio reattivo del job
  useSSE({
    url: streamUrl,
    enabled: Boolean(isLive && streamUrl),
    onEvent: (eventType, data) => {
      if (eventType === 'phase') {
        const payload = data as ScanPhasePayload;
        if (payload) {
          setProgressEstimate((prev) => ({ percentage: Math.max(prev?.percentage || 0, Math.min(99, payload.percentage)), receivedAt: Date.now() }));
          setLivePhase({
            phase: payload.phase,
            percentage: payload.percentage,
            message: payload.message,
          });
        }
      } else if (eventType === 'progress') {
        const payload = data as ScanProgressPayload;
        if (payload) {
          if (!Number.isFinite(payload.percentage)) return;
          setProgressEstimate((prev) => ({
            percentage: Math.max(prev?.percentage || 0, Math.max(0, Math.min(99, payload.percentage))),
            remainingSeconds: payload.remaining_seconds, phasePercentage: payload.phase_percentage, receivedAt: Date.now(),
          }));
          setLivePhase((prev) => ({
            phase: payload.phase || prev?.phase || 'PORT_SCAN',
            percentage: payload.percentage,
            message: payload.message || prev?.message || '',
          }));
        }
      } else if (eventType === 'finding_discovered') {
        const payload = data as ScanFindingPayload;
        if (payload?.finding) {
          setFindings((prev) => {
            if (prev.some((f) => f.id === payload.finding.id)) {
              return prev;
            }
            return [payload.finding, ...prev];
          });
          setNewlyDiscoveredFindingIds((prev) => new Set(prev).add(payload.finding.id));

          // Incremento dinamico delle metriche di riepilogo
          setScan((prev) => {
            if (!prev) return prev;
            const updatedStats = { ...prev.summary_stats };
            updatedStats.total_open_ports += 1;
            const sev = payload.finding.severity;
            if (sev === 'CRITICAL') updatedStats.critical_count += 1;
            else if (sev === 'HIGH') updatedStats.high_count += 1;
            else if (sev === 'MEDIUM') updatedStats.medium_count += 1;
            else if (sev === 'LOW') updatedStats.low_count += 1;
            else if (sev === 'INFO') updatedStats.info_count += 1;

            return {
              ...prev,
              summary_stats: updatedStats,
            };
          });
        }
      } else if (eventType === 'status_change') {
        const payload = data as ScanStatusPayload;
        if (payload) {
          setScan((prev) => {
            if (!prev) return prev;
            return {
              ...prev,
              status: payload.status,
              summary_stats: payload.summary_stats || prev.summary_stats,
            };
          });
          if (
            payload.status === 'COMPLETED' ||
            payload.status === 'FAILED' ||
            payload.status === 'CANCELLED'
          ) {
            setLivePhase(null);
          }
        }
      } else if (eventType === 'error') {
        const payload = data as ScanErrorPayload;
        if (payload?.message) {
          setLiveError(payload.message);
        }
      }
    },
  });

  const fetchScanDetails = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const [scanRes, findingsRes] = await Promise.all([
        apiClient.get<ApiResponse<ScanItem>>(`/scans/${id}`),
        apiClient.get<ApiResponse<FindingListResponseData>>(`/scans/${id}/findings?limit=100`),
      ]);

      if (scanRes.data?.data) {
        setScan(scanRes.data.data);
        const snapshot = scanRes.data.data.progress;
        if (snapshot && Number.isFinite(snapshot.percentage)) {
          setProgressEstimate({ percentage: snapshot.percentage, remainingSeconds: snapshot.remaining_seconds,
            phasePercentage: snapshot.phase_percentage, receivedAt: Date.parse(snapshot.observed_at || '') || Date.now() - 31000 });
        }
      }
      if (findingsRes.data?.data) {
        setFindings(findingsRes.data.data.items || []);
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile caricare il dettaglio della scansione'
      );
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchScanDetails();
  }, [fetchScanDetails]);

  const handleCopyId = () => {
    if (id) {
      navigator.clipboard.writeText(id);
      setCopiedId(true);
      setTimeout(() => setCopiedId(false), 2000);
    }
  };

  const handleDownloadPdf = async () => {
    if (!scan) return;
    setDownloadError(null);
    try {
      await downloadReportFile({
        url: `/scans/${scan.id}/reports/pdf`,
        defaultFilename: `report_scan_${scan.id.slice(0, 8)}.pdf`,
        onStart: () => setDownloadingPdf(true),
        onSuccess: () => setDownloadingPdf(false),
        onError: (err) => {
          setDownloadingPdf(false);
          setDownloadError(err.message);
        },
      });
    } catch {
      // Errore già gestito in onError
    }
  };

  const handleDownloadCsv = async () => {
    if (!scan) return;
    setDownloadError(null);
    try {
      await downloadReportFile({
        url: `/scans/${scan.id}/reports/csv`,
        defaultFilename: `findings_scan_${scan.id.slice(0, 8)}.csv`,
        onStart: () => setDownloadingCsv(true),
        onSuccess: () => setDownloadingCsv(false),
        onError: (err) => {
          setDownloadingCsv(false);
          setDownloadError(err.message);
        },
      });
    } catch {
      // Errore già gestito in onError
    }
  };

  if (loading && !scan) {
    return (
      <div className="py-24 flex flex-col items-center justify-center text-slate-400 space-y-3">
        <RefreshCw className="w-8 h-8 animate-spin text-cyan-400" />
        <span className="text-sm">Caricamento audit e vulnerabilità in corso...</span>
      </div>
    );
  }

  if (error || !scan) {
    return (
      <div className="space-y-4">
        <Link
          to="/scans"
          className="inline-flex items-center space-x-1.5 text-xs text-slate-400 hover:text-white transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Torna all'elenco scansioni</span>
        </Link>
        <div className="p-6 rounded-2xl border border-rose-800/50 bg-rose-950/20 text-rose-300 text-sm space-y-3">
          <div className="flex items-center space-x-2 font-bold text-rose-400">
            <AlertCircle className="w-5 h-5" />
            <span>Errore nel caricamento della scansione</span>
          </div>
          <p>{error || 'Scansione non trovata o ID non valido.'}</p>
        </div>
      </div>
    );
  }

  const stats = scan.summary_stats || {
    total_hosts: 0,
    up_hosts: 0,
    down_hosts: 0,
    total_open_ports: 0,
    critical_count: 0,
    high_count: 0,
    medium_count: 0,
    low_count: 0,
    info_count: 0,
    duration_seconds: 0,
  };

  return (
    <div className="space-y-8">
      {/* Breadcrumb & Navigation */}
      <div className="flex items-center justify-between">
        <Link
          to="/scans"
          className="inline-flex items-center space-x-1.5 text-xs font-medium text-slate-400 hover:text-cyan-400 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Torna all'elenco Scansioni</span>
        </Link>

        <button
          onClick={fetchScanDetails}
          className="flex items-center space-x-1.5 px-3 py-1 rounded-lg border border-cyber-border bg-slate-900/60 hover:bg-slate-800 text-slate-400 hover:text-white text-xs font-medium transition-colors"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
          <span>Aggiorna Dati</span>
        </button>
      </div>

      {/* Main Header Card */}
      <div className="p-6 rounded-2xl border border-cyber-border bg-cyber-card/60 shadow-xl space-y-6">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div className="space-y-2">
            <div className="flex items-center space-x-3">
              <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
                <Terminal className="w-6 h-6" />
              </div>
              <div>
                <div className="flex items-center space-x-2.5">
                  <h2 className="text-xl font-bold text-white font-mono">
                    Scansione {scan.id.slice(0, 8)}...
                  </h2>
                  <button
                    onClick={handleCopyId}
                    className="p-1 rounded text-slate-500 hover:text-slate-300 transition-colors"
                    title="Copia ID completo"
                  >
                    {copiedId ? (
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                    ) : (
                      <Copy className="w-3.5 h-3.5" />
                    )}
                  </button>
                  <ScanProgressBadge
                    status={scan.status}
                    phase={livePhase?.phase}
                    percentage={livePhase?.percentage}
                  />
                </div>
                <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-400 mt-1">
                  <span>
                    Origine: <strong className="text-slate-200">{scan.trigger_type}</strong>
                  </span>
                  <span>•</span>
                  <span>
                    Creata il:{' '}
                    <strong className="text-slate-200">
                      {new Date(scan.created_at).toLocaleString('it-IT')}
                    </strong>
                  </span>
                  {stats.duration_seconds > 0 && (
                    <>
                      <span>•</span>
                      <span className="flex items-center space-x-1">
                        <Clock className="w-3 h-3 text-cyan-400" />
                        <span>Durata: {stats.duration_seconds.toFixed(1)}s</span>
                      </span>
                    </>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Action Buttons: Reports */}
          {scan.status === 'COMPLETED' && (
            <div className="flex flex-col sm:flex-row sm:items-center gap-2.5">
              <div className="flex items-center gap-2.5">
                <button
                  type="button"
                  onClick={handleDownloadPdf}
                  disabled={downloadingPdf || downloadingCsv}
                  className="flex items-center space-x-2 px-3.5 py-2 rounded-lg bg-emerald-500/15 hover:bg-emerald-500/25 text-emerald-300 border border-emerald-500/30 font-semibold text-xs transition-colors shadow-[0_0_15px_rgba(16,185,129,0.15)] disabled:opacity-50 disabled:cursor-not-allowed"
                  title="Scarica Report PDF completo"
                >
                  {downloadingPdf ? (
                    <RefreshCw className="w-4 h-4 animate-spin text-emerald-400" />
                  ) : (
                    <FileText className="w-4 h-4 text-emerald-400" />
                  )}
                  <span>{downloadingPdf ? 'Generazione PDF...' : 'Scarica Report PDF (White-Label)'}</span>
                </button>

                <button
                  type="button"
                  onClick={handleDownloadCsv}
                  disabled={downloadingPdf || downloadingCsv}
                  className="flex items-center space-x-2 px-3.5 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-cyber-border font-semibold text-xs transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  title="Esporta findings in formato CSV"
                >
                  {downloadingCsv ? (
                    <RefreshCw className="w-4 h-4 animate-spin text-blue-400" />
                  ) : (
                    <FileSpreadsheet className="w-4 h-4 text-blue-400" />
                  )}
                  <span>{downloadingCsv ? 'Esportazione...' : 'Esporta CSV'}</span>
                </button>
              </div>

              {downloadError && (
                <div className="text-xs text-rose-400 flex items-center space-x-1">
                  <AlertCircle className="w-3.5 h-3.5" />
                  <span>{downloadError}</span>
                </div>
              )}
            </div>
          )}
        </div>

        {/* KPI Strip */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-4 border-t border-cyber-border/60">
          <div className="p-3.5 rounded-xl bg-slate-900/60 border border-cyber-border">
            <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
              <span>Host Operativi (UP)</span>
              <Server className="w-4 h-4 text-cyan-400" />
            </div>
            <div className="flex items-baseline space-x-1.5">
              <span className="text-xl font-bold text-white font-mono">
                {stats.up_hosts}
              </span>
              <span className="text-xs text-slate-500 font-normal">
                / {stats.total_hosts || stats.up_hosts}
              </span>
            </div>
            {stats.down_hosts !== undefined && stats.down_hosts > 0 ? (
              <div className="text-[10px] text-amber-400/90 mt-1 flex items-center space-x-1">
                <span>{stats.down_hosts} offline esclusi</span>
              </div>
            ) : (
              <div className="text-[10px] text-emerald-400/80 mt-1 flex items-center space-x-1">
                <span>100% responsivi</span>
              </div>
            )}
          </div>

          <div className="p-3.5 rounded-xl bg-slate-900/60 border border-cyber-border">
            <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
              <span>Porte Aperte</span>
              <Layers className="w-4 h-4 text-blue-400" />
            </div>
            <div className="text-xl font-bold text-white font-mono">
              {stats.total_open_ports}
            </div>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-900/60 border border-cyber-border">
            <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
              <span>Vulnerabilità Totali</span>
              <ShieldAlert className="w-4 h-4 text-amber-400" />
            </div>
            <div className="text-xl font-bold text-white font-mono">
              {stats.critical_count + stats.high_count + stats.medium_count + stats.low_count + stats.info_count}
            </div>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-900/60 border border-cyber-border">
            <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
              <span>Critiche + Alte</span>
              <ShieldAlert className="w-4 h-4 text-red-400" />
            </div>
            <div className="text-xl font-bold text-red-400 font-mono">
              {stats.critical_count + stats.high_count}
            </div>
          </div>
        </div>
      </div>

      {/* Notifica di errore live SSE */}
      {liveError && (
        <div className="p-4 rounded-xl border border-rose-900/50 bg-rose-950/20 flex items-center space-x-3 text-rose-300">
          <AlertCircle className="w-5 h-5 shrink-0 text-rose-400" />
          <div className="text-sm font-medium">{liveError}</div>
        </div>
      )}

      {/* Live Scan Execution Card & Stepper */}
      {scan && <ScanTotalProgress status={scan.status} estimate={progressEstimate} />}
      {isLive && (
        <div className="p-5 rounded-2xl border border-cyan-500/40 bg-cyan-950/20 shadow-xl space-y-4 animate-in fade-in duration-300">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="flex items-center space-x-3 text-cyan-300">
              <Zap className="w-5 h-5 text-cyan-400 animate-pulse shrink-0" />
              <div>
                <div className="flex items-center space-x-2">
                  <span className="font-bold text-white text-sm">
                    Scansione in tempo reale attiva
                  </span>
                  <span className="px-2 py-0.5 rounded-full bg-cyan-500/20 text-cyan-300 border border-cyan-500/30 text-[10px] font-mono font-semibold uppercase animate-pulse">
                    LIVE STREAM
                  </span>
                </div>
                <p className="text-cyan-400/80 text-xs mt-0.5">
                  {livePhase?.message || 'Esecuzione job di scansione e streaming findings live...'}
                </p>
              </div>
            </div>

            <div className="flex items-center space-x-3 text-xs font-mono self-end sm:self-auto">
              <span className="text-slate-400">Progresso stimato:</span>
              <span className="text-cyan-400 font-bold text-base">
                {progressEstimate ? `${progressEstimate.percentage.toFixed(0)}%` : 'In calcolo'}
              </span>
            </div>
          </div>

          {/* Stepper a 4 fasi */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 pt-1">
            {SCAN_PHASES.map((p, idx) => {
              const currentPhaseIdx = SCAN_PHASES.findIndex((sp) => sp.key === livePhase?.phase);
              const isPast = currentPhaseIdx > idx || scan?.status === 'COMPLETED';
              const isCurrent = currentPhaseIdx === idx && scan?.status === 'RUNNING';

              return (
                <div
                  key={p.key}
                  className={`p-2.5 rounded-xl border text-xs transition-all ${
                    isCurrent
                      ? 'bg-cyan-500/10 border-cyan-400/60 shadow-[0_0_15px_rgba(6,182,212,0.2)]'
                      : isPast
                      ? 'bg-emerald-950/20 border-emerald-500/40 text-emerald-300'
                      : 'bg-slate-900/40 border-cyber-border/60 text-slate-500'
                  }`}
                >
                  <div className="flex items-center space-x-2 mb-1">
                    <span
                      className={`w-2 h-2 rounded-full ${
                        isCurrent
                          ? 'bg-cyan-400 animate-ping'
                          : isPast
                          ? 'bg-emerald-400'
                          : 'bg-slate-600'
                      }`}
                    />
                    <span
                      className={`font-semibold ${
                        isCurrent ? 'text-white' : isPast ? 'text-emerald-300' : 'text-slate-400'
                      }`}
                    >
                      {p.label}
                    </span>
                  </div>
                  <p className="text-[10px] text-slate-400 leading-tight truncate">{p.desc}</p>
                </div>
              );
            })}
          </div>

        </div>
      )}

      {/* Grafico Severità */}
      <div className="p-6 rounded-2xl border border-cyber-border bg-cyber-card/40 space-y-4">
        <div className="pb-3 border-b border-cyber-border/60">
          <h3 className="text-base font-semibold text-white">
            Distribuzione Severità CVE Rilevate
          </h3>
          <p className="text-xs text-slate-400">
            Analisi del rischio secondo il framework di correlazione CVSS v3.1 per questa scansione
          </p>
        </div>

        <SeverityChart
          critical={stats.critical_count}
          high={stats.high_count}
          medium={stats.medium_count}
          low={stats.low_count}
          info={stats.info_count}
        />
      </div>

      {/* Tabella Dettagli Findings */}
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-bold text-white">
            Vulnerabilità e Servizi Rilevati ({findings.length})
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Elenco analitico dei servizi scoperti e delle CVE correlate tramite banner Nmap
          </p>
        </div>

        <FindingsTable findings={findings} loading={loading} />
      </div>
    </div>
  );
};

