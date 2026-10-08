import React, { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Crosshair,
  Terminal,
  Server,
  ShieldAlert,
  RefreshCw,
  ArrowRight,
  Clock,
  CheckCircle2,
  AlertCircle,
  Play,
  Layers,
  Radio
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { ScanListResponseData, ScanItem } from '../types/scan';
import { TargetListResponseData } from '../types/target';
import { SeverityChart } from '../components/SeverityChart';
import { useAuth } from '../context/AuthContext';

export const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const [scans, setScans] = useState<ScanItem[]>([]);
  const [totalScans, setTotalScans] = useState<number>(0);
  const [totalTargets, setTotalTargets] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDashboardData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [scansRes, targetsRes] = await Promise.all([
        apiClient.get<ApiResponse<ScanListResponseData>>('/scans?limit=10'),
        apiClient.get<ApiResponse<TargetListResponseData>>('/targets?limit=1'),
      ]);

      if (scansRes.data?.data) {
        setScans(scansRes.data.data.items || []);
        setTotalScans(scansRes.data.data.total || 0);
      }

      if (targetsRes.data?.data) {
        setTotalTargets(targetsRes.data.data.total || 0);
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile caricare i dati della dashboard'
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDashboardData();
  }, [fetchDashboardData]);

  // Calcolo metriche aggregate
  const activeScansCount = scans.filter(
    (s) => s.status === 'RUNNING' || s.status === 'QUEUED'
  ).length;

  const upHostsCount = scans.reduce(
    (acc, s) => acc + (s.summary_stats?.up_hosts || 0),
    0
  );

  const totalOpenPorts = scans.reduce(
    (acc, s) => acc + (s.summary_stats?.total_open_ports || 0),
    0
  );

  const severityTotals = scans.reduce(
    (acc, s) => {
      const stats = s.summary_stats;
      if (stats) {
        acc.critical += stats.critical_count || 0;
        acc.high += stats.high_count || 0;
        acc.medium += stats.medium_count || 0;
        acc.low += stats.low_count || 0;
        acc.info += stats.info_count || 0;
      }
      return acc;
    },
    { critical: 0, high: 0, medium: 0, low: 0, info: 0 }
  );

  const totalVulnerabilities =
    severityTotals.critical +
    severityTotals.high +
    severityTotals.medium +
    severityTotals.low +
    severityTotals.info;

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'COMPLETED':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
            <CheckCircle2 className="w-3 h-3" />
            <span>Completata</span>
          </span>
        );
      case 'RUNNING':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 animate-pulse">
            <Radio className="w-3 h-3 animate-spin" />
            <span>In Esecuzione</span>
          </span>
        );
      case 'QUEUED':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/30">
            <Clock className="w-3 h-3" />
            <span>In Coda</span>
          </span>
        );
      case 'FAILED':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-rose-500/10 text-rose-400 border border-rose-500/30">
            <AlertCircle className="w-3 h-3" />
            <span>Fallita</span>
          </span>
        );
      case 'CANCELLED':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-500/10 text-slate-400 border border-slate-500/30">
            <span>Annullata</span>
          </span>
        );
      default:
        return (
          <span className="px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-800 text-slate-300">
            {status}
          </span>
        );
    }
  };

  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white flex items-center space-x-2">
            <span>Security Operations Dashboard</span>
          </h2>
          <p className="text-sm text-slate-400 mt-1">
            Superficie d'attacco, scansioni attive e vulnerabilità per l'organizzazione{' '}
            <strong className="text-slate-200">
              {user?.tenant_id ? `(Tenant: ${user.tenant_id.slice(0, 8)}...)` : 'Default'}
            </strong>
          </p>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={fetchDashboardData}
            disabled={loading}
            className="flex items-center space-x-2 px-3 py-1.5 rounded-lg border border-cyber-border bg-slate-900/80 hover:bg-slate-800 text-slate-300 hover:text-white transition-colors text-xs font-medium disabled:opacity-50"
            title="Aggiorna metriche"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
            <span>{loading ? 'Sincronizzazione...' : 'Aggiorna Dati'}</span>
          </button>

          <Link
            to="/scans"
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs transition-colors shadow-[0_0_15px_rgba(6,182,212,0.3)]"
          >
            <Play className="w-3.5 h-3.5" />
            <span>Nuova Scansione</span>
          </Link>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="p-4 rounded-xl border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
            <span>{error}</span>
          </div>
          <button
            onClick={fetchDashboardData}
            className="px-2.5 py-1 rounded bg-rose-900/50 hover:bg-rose-900 text-rose-200 font-medium"
          >
            Riprova
          </button>
        </div>
      )}

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* KPI: Target Censiti */}
        <Link
          to="/targets"
          className="p-5 rounded-xl border border-cyber-border bg-cyber-card/60 hover:border-cyan-500/40 transition-colors group relative overflow-hidden"
        >
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
              Target Censiti
            </span>
            <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 group-hover:scale-105 transition-transform">
              <Crosshair className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-white tracking-tight">
            {loading ? '...' : totalTargets}
          </div>
          <div className="mt-3 pt-3 border-t border-cyber-border/60 flex items-center justify-between text-xs text-slate-400">
            <span>Asset configurati</span>
            <ArrowRight className="w-3.5 h-3.5 text-cyan-400 group-hover:translate-x-1 transition-transform" />
          </div>
        </Link>

        {/* KPI: Scansioni Eseguite */}
        <Link
          to="/scans"
          className="p-5 rounded-xl border border-cyber-border bg-cyber-card/60 hover:border-cyan-500/40 transition-colors group relative overflow-hidden"
        >
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
              Scansioni Totali
            </span>
            <div className="p-2 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 group-hover:scale-105 transition-transform">
              <Terminal className="w-4 h-4" />
            </div>
          </div>
          <div className="flex items-baseline space-x-2">
            <span className="text-2xl font-bold text-white tracking-tight">
              {loading ? '...' : totalScans}
            </span>
            {activeScansCount > 0 && (
              <span className="text-xs font-medium text-cyan-400 animate-pulse">
                ({activeScansCount} attive)
              </span>
            )}
          </div>
          <div className="mt-3 pt-3 border-t border-cyber-border/60 flex items-center justify-between text-xs text-slate-400">
            <span>Storico esecuzioni</span>
            <ArrowRight className="w-3.5 h-3.5 text-cyan-400 group-hover:translate-x-1 transition-transform" />
          </div>
        </Link>

        {/* KPI: Host Attivi */}
        <div className="p-5 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
              Host Attivi
            </span>
            <div className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <Server className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-white tracking-tight">
            {loading ? '...' : upHostsCount}
          </div>
          <div className="mt-3 pt-3 border-t border-cyber-border/60 flex items-center justify-between text-xs text-slate-400">
            <span>Rilevati online</span>
            <span className="text-emerald-400 font-medium">{totalOpenPorts} porte aperte</span>
          </div>
        </div>

        {/* KPI: Vulnerabilità Rilevate */}
        <Link
          to="/findings"
          className="p-5 rounded-xl border border-cyber-border bg-cyber-card/60 hover:border-red-500/40 transition-colors group relative overflow-hidden"
        >
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
              Vulnerabilità (CVE)
            </span>
            <div className="p-2 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 group-hover:scale-105 transition-transform">
              <ShieldAlert className="w-4 h-4" />
            </div>
          </div>
          <div className="flex items-baseline space-x-2">
            <span className="text-2xl font-bold text-white tracking-tight">
              {loading ? '...' : totalVulnerabilities}
            </span>
            {severityTotals.critical > 0 && (
              <span className="text-xs font-bold text-red-400">
                ({severityTotals.critical} critiche)
              </span>
            )}
          </div>
          <div className="mt-3 pt-3 border-t border-cyber-border/60 flex items-center justify-between text-xs text-slate-400">
            <span>Catalogo riscontrato</span>
            <ArrowRight className="w-3.5 h-3.5 text-red-400 group-hover:translate-x-1 transition-transform" />
          </div>
        </Link>
      </div>

      {/* Main Grid: Charts & Surface Details */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Severity Chart Container */}
        <div className="lg:col-span-7 p-6 rounded-xl border border-cyber-border bg-cyber-card/40 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-cyber-border/60">
            <div>
              <h3 className="text-base font-semibold text-white">
                Distribuzione Severità Vulnerabilità (CVE)
              </h3>
              <p className="text-xs text-slate-400">
                Correlazione CVSS v3.1 rilevata sui banner dei servizi scansionati
              </p>
            </div>
            <Link
              to="/findings"
              className="text-xs text-cyan-400 hover:text-cyan-300 font-medium flex items-center space-x-1"
            >
              <span>Dettagli</span>
              <ArrowRight className="w-3 h-3" />
            </Link>
          </div>

          <SeverityChart
            critical={severityTotals.critical}
            high={severityTotals.high}
            medium={severityTotals.medium}
            low={severityTotals.low}
            info={severityTotals.info}
          />
        </div>

        {/* Attack Surface Summary Card */}
        <div className="lg:col-span-5 p-6 rounded-xl border border-cyber-border bg-cyber-card/40 space-y-5">
          <div className="pb-3 border-b border-cyber-border/60">
            <h3 className="text-base font-semibold text-white">
              Sintesi Esposizione Rete
            </h3>
            <p className="text-xs text-slate-400">
              Indicatori di superficie emersi dalle attività di discovery Nmap
            </p>
          </div>

          <div className="space-y-4">
            <div className="p-4 rounded-lg bg-slate-900/60 border border-cyber-border flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
                  <Layers className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs text-slate-400">Porte Aperte Rilevate</div>
                  <div className="text-lg font-bold text-white">{totalOpenPorts}</div>
                </div>
              </div>
              <span className="text-[11px] text-slate-500 font-mono">TCP / UDP</span>
            </div>

            <div className="p-4 rounded-lg bg-slate-900/60 border border-cyber-border flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
                  <Server className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs text-slate-400">Host Responsivi (Up)</div>
                  <div className="text-lg font-bold text-white">{upHostsCount}</div>
                </div>
              </div>
              <span className="text-[11px] text-emerald-400 font-medium">Online</span>
            </div>

            <div className="p-4 rounded-lg bg-slate-900/60 border border-cyber-border flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <div className="p-2 rounded-lg bg-red-500/10 text-red-400">
                  <ShieldAlert className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs text-slate-400">Severità Critica + Alta</div>
                  <div className="text-lg font-bold text-red-400">
                    {severityTotals.critical + severityTotals.high}
                  </div>
                </div>
              </div>
              <span className="text-[11px] text-red-400 font-medium">Priorità Alta</span>
            </div>
          </div>

          <div className="pt-2">
            <Link
              to="/reports"
              className="w-full flex items-center justify-center space-x-2 py-2.5 px-4 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 hover:text-white transition-colors text-xs font-semibold border border-cyber-border"
            >
              <span>Genera Report Audit White-Label (PDF)</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
          </div>
        </div>
      </div>

      {/* Recent Scans Table Section */}
      <div className="p-6 rounded-xl border border-cyber-border bg-cyber-card/40 space-y-4">
        <div className="flex items-center justify-between pb-3 border-b border-cyber-border/60">
          <div>
            <h3 className="text-base font-semibold text-white">
              Ultime Scansioni di Rete Eseguite
            </h3>
            <p className="text-xs text-slate-400">
              Cronologia recente dei job di ricognizione e vulnerabilità
            </p>
          </div>
          <Link
            to="/scans"
            className="text-xs text-cyan-400 hover:text-cyan-300 font-medium flex items-center space-x-1"
          >
            <span>Tutte le scansioni</span>
            <ArrowRight className="w-3 h-3" />
          </Link>
        </div>

        {loading ? (
          <div className="py-12 flex flex-col items-center justify-center text-slate-400 space-y-2">
            <RefreshCw className="w-5 h-5 animate-spin text-cyan-400" />
            <span className="text-xs">Caricamento storico scansioni...</span>
          </div>
        ) : scans.length === 0 ? (
          <div className="py-12 flex flex-col items-center justify-center text-center space-y-3">
            <div className="w-12 h-12 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-slate-500">
              <Terminal className="w-6 h-6" />
            </div>
            <div className="text-sm font-semibold text-slate-300">
              Nessuna scansione presente
            </div>
            <p className="text-xs text-slate-500 max-w-sm">
              Non è ancora stata avviata alcuna scansione di rete. Configura un target e avvia la prima analisi Nmap.
            </p>
            <Link
              to="/scans"
              className="mt-2 inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs transition-colors"
            >
              <Play className="w-3.5 h-3.5" />
              <span>Avvia la prima scansione</span>
            </Link>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-cyber-border/80 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                  <th className="py-3 px-4">Scansione ID</th>
                  <th className="py-3 px-4">Trigger</th>
                  <th className="py-3 px-4">Stato</th>
                  <th className="py-3 px-4">Host Up</th>
                  <th className="py-3 px-4">Vulnerabilità</th>
                  <th className="py-3 px-4">Durata</th>
                  <th className="py-3 px-4">Data Creazione</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-cyber-border/40 text-xs">
                {scans.map((scan) => (
                  <tr
                    key={scan.id}
                    className="hover:bg-slate-800/30 transition-colors"
                  >
                    <td className="py-3 px-4 font-mono text-cyan-400 font-medium">
                      {scan.id.slice(0, 8)}...
                    </td>
                    <td className="py-3 px-4 text-slate-300">
                      <span className="px-2 py-0.5 rounded bg-slate-900 border border-cyber-border text-[11px]">
                        {scan.trigger_type}
                      </span>
                    </td>
                    <td className="py-3 px-4">{getStatusBadge(scan.status)}</td>
                    <td className="py-3 px-4 text-slate-300">
                      {scan.summary_stats?.up_hosts ?? 0} / {scan.summary_stats?.total_hosts ?? 0}
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center space-x-1 text-[11px]">
                        {(scan.summary_stats?.critical_count ?? 0) > 0 && (
                          <span className="px-1.5 py-0.5 rounded bg-red-500/10 text-red-400 border border-red-500/30 font-semibold">
                            C:{scan.summary_stats.critical_count}
                          </span>
                        )}
                        {(scan.summary_stats?.high_count ?? 0) > 0 && (
                          <span className="px-1.5 py-0.5 rounded bg-orange-500/10 text-orange-400 border border-orange-500/30 font-semibold">
                            H:{scan.summary_stats.high_count}
                          </span>
                        )}
                        {(scan.summary_stats?.medium_count ?? 0) > 0 && (
                          <span className="px-1.5 py-0.5 rounded bg-yellow-500/10 text-yellow-400 border border-yellow-500/30">
                            M:{scan.summary_stats.medium_count}
                          </span>
                        )}
                        {(scan.summary_stats?.low_count ?? 0) > 0 && (
                          <span className="px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/30">
                            L:{scan.summary_stats.low_count}
                          </span>
                        )}
                        {(scan.summary_stats?.critical_count ?? 0) === 0 &&
                          (scan.summary_stats?.high_count ?? 0) === 0 &&
                          (scan.summary_stats?.medium_count ?? 0) === 0 &&
                          (scan.summary_stats?.low_count ?? 0) === 0 && (
                            <span className="text-slate-500">-</span>
                          )}
                      </div>
                    </td>
                    <td className="py-3 px-4 text-slate-400 font-mono">
                      {scan.summary_stats?.duration_seconds
                        ? `${scan.summary_stats.duration_seconds.toFixed(1)}s`
                        : '-'}
                    </td>
                    <td className="py-3 px-4 text-slate-400">
                      {new Date(scan.created_at).toLocaleString('it-IT', {
                        day: '2-digit',
                        month: '2-digit',
                        year: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
