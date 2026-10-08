import React, { useState, useEffect, useCallback } from 'react';
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
  Check
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { ScanItem } from '../types/scan';
import { FindingItem, FindingListResponseData } from '../types/finding';
import { ScanProgressBadge } from '../components/ScanProgressBadge';
import { SeverityChart } from '../components/SeverityChart';
import { FindingsTable } from '../components/FindingsTable';
import { downloadReportFile } from '../utils/download';

export const ScanDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  const [scan, setScan] = useState<ScanItem | null>(null);
  const [findings, setFindings] = useState<FindingItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [copiedId, setCopiedId] = useState<boolean>(false);
  const [downloadingPdf, setDownloadingPdf] = useState<boolean>(false);
  const [downloadingCsv, setDownloadingCsv] = useState<boolean>(false);
  const [downloadError, setDownloadError] = useState<string | null>(null);

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
                  <ScanProgressBadge status={scan.status} />
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
              <span>Host Responsivi</span>
              <Server className="w-4 h-4 text-cyan-400" />
            </div>
            <div className="text-xl font-bold text-white font-mono">
              {stats.up_hosts} <span className="text-xs text-slate-500 font-normal">/ {stats.total_hosts}</span>
            </div>
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

