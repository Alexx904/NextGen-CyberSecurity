import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  Terminal,
  Plus,
  Play,
  Calendar,
  RefreshCw,
  AlertCircle,
  FileText,
  FileSpreadsheet,
  StopCircle,
  Trash2,
  ExternalLink,
  Power,
  Copy,
  Check,
  Folder,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { ScanItem, ScanListResponseData } from '../types/scan';
import { ScheduleItem, ScheduleListResponseData } from '../types/schedule';
import { TargetItem, TargetListResponseData, TargetGroupNode } from '../types/target';
import { ScanProgressBadge } from '../components/ScanProgressBadge';
import { NewScanModal } from '../components/NewScanModal';
import { downloadReportFile } from '../utils/download';

export const Scans: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'SCANS' | 'SCHEDULES'>('SCANS');

  // Stato Scansioni
  const [scans, setScans] = useState<ScanItem[]>([]);
  const [totalScans, setTotalScans] = useState<number>(0);
  const [loadingScans, setLoadingScans] = useState<boolean>(true);
  const [scansError, setScansError] = useState<string | null>(null);

  // Filtro cartella scope per scansioni
  const [selectedGroupFilter, setSelectedGroupFilter] = useState<string>('');
  const [targetsMap, setTargetsMap] = useState<Map<string, TargetItem>>(new Map());
  const [groupsTree, setGroupsTree] = useState<TargetGroupNode[]>([]);

  // Stato download report
  const [downloadingScanId, setDownloadingScanId] = useState<string | null>(null);
  const [downloadType, setDownloadType] = useState<'pdf' | 'csv' | null>(null);

  // Stato Schedulazioni
  const [schedules, setSchedules] = useState<ScheduleItem[]>([]);
  const [loadingSchedules, setLoadingSchedules] = useState<boolean>(false);
  const [schedulesError, setSchedulesError] = useState<string | null>(null);

  // Modali & Azioni
  const [isNewScanModalOpen, setIsNewScanModalOpen] = useState<boolean>(false);
  const [scanToCancel, setScanToCancel] = useState<ScanItem | null>(null);
  const [cancelling, setCancelling] = useState<boolean>(false);

  const [scheduleToDelete, setScheduleToDelete] = useState<ScheduleItem | null>(null);
  const [deletingSchedule, setDeletingSchedule] = useState<boolean>(false);

  // Auto-refresh control
  const [autoRefreshEnabled, setAutoRefreshEnabled] = useState<boolean>(true);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // Caricamento elenco scansioni
  const fetchScans = useCallback(async (silent = false) => {
    if (!silent) setLoadingScans(true);
    setScansError(null);
    try {
      const response = await apiClient.get<ApiResponse<ScanListResponseData>>('/scans?limit=50');
      if (response.data?.data) {
        setScans(response.data.data.items || []);
        setTotalScans(response.data.data.total || 0);
      }
    } catch (err: any) {
      if (!silent) {
        setScansError(
          err.response?.data?.error?.message ||
          err.message ||
          'Impossibile recuperare lo storico delle scansioni'
        );
      }
    } finally {
      if (!silent) setLoadingScans(false);
    }
  }, []);

  // Caricamento elenco schedulazioni
  const fetchSchedules = useCallback(async () => {
    setLoadingSchedules(true);
    setSchedulesError(null);
    try {
      const response = await apiClient.get<ApiResponse<ScheduleListResponseData>>('/schedules');
      if (response.data?.data) {
        setSchedules(response.data.data.items || []);
      }
    } catch (err: any) {
      setSchedulesError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile recuperare le pianificazioni attive'
      );
    } finally {
      setLoadingSchedules(false);
    }
  }, []);

  // Caricamento metadati target e gruppi scope per arricchimento vista
  const fetchMetadata = useCallback(async () => {
    try {
      const [targetsRes, groupsRes] = await Promise.all([
        apiClient.get<ApiResponse<TargetListResponseData>>('/targets?limit=200'),
        apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups'),
      ]);
      if (targetsRes.data?.data?.items) {
        const map = new Map<string, TargetItem>();
        targetsRes.data.data.items.forEach((t) => map.set(t.id, t));
        setTargetsMap(map);
      }
      if (groupsRes.data?.data) {
        setGroupsTree(groupsRes.data.data);
      }
    } catch {
      // Ignora errori non bloccanti
    }
  }, []);

  useEffect(() => {
    fetchScans();
    fetchMetadata();
  }, [fetchScans, fetchMetadata]);

  useEffect(() => {
    if (activeTab === 'SCHEDULES') {
      fetchSchedules();
    }
  }, [activeTab, fetchSchedules]);

  // Lista appiattita dei gruppi per dropdown
  const flatGroups = useMemo(() => {
    const list: { id: string; name: string }[] = [];
    function recurse(nodes: TargetGroupNode[], prefix = '') {
      for (const n of nodes) {
        list.push({ id: n.id, name: prefix ? `${prefix} / ${n.name}` : n.name });
        if (n.children && n.children.length > 0) {
          recurse(n.children, prefix ? `${prefix} / ${n.name}` : n.name);
        }
      }
    }
    recurse(groupsTree);
    return list;
  }, [groupsTree]);

  // Mappa dei gruppi per ID
  const groupsMap = useMemo(() => {
    const map = new Map<string, TargetGroupNode>();
    function recurse(nodes: TargetGroupNode[]) {
      for (const n of nodes) {
        map.set(n.id, n);
        if (n.children && n.children.length > 0) {
          recurse(n.children);
        }
      }
    }
    recurse(groupsTree);
    return map;
  }, [groupsTree]);

  // Target ID validi per la cartella selezionata (inclusi i sotto-rami)
  const targetIdsInSelectedGroup = useMemo(() => {
    if (!selectedGroupFilter) return null;
    const selectedNode = groupsMap.get(selectedGroupFilter);
    if (!selectedNode) return null;

    const groupIds = new Set<string>();
    function collectIds(node: TargetGroupNode) {
      groupIds.add(node.id);
      if (node.children) {
        node.children.forEach(collectIds);
      }
    }
    collectIds(selectedNode);

    const validTargetIds = new Set<string>();
    targetsMap.forEach((target, id) => {
      if (target.group_id && groupIds.has(target.group_id)) {
        validTargetIds.add(id);
      }
    });
    return validTargetIds;
  }, [selectedGroupFilter, groupsMap, targetsMap]);

  // Scansioni filtrate in base alla cartella target selezionata
  const displayedScans = useMemo(() => {
    if (!targetIdsInSelectedGroup) return scans;
    return scans.filter((s) => targetIdsInSelectedGroup.has(s.target_id));
  }, [scans, targetIdsInSelectedGroup]);

  // Monitoraggio intelligente per polling automatico se vi sono scansioni attive
  const hasActiveScans = useMemo(() => {
    return scans.some((s) => s.status === 'RUNNING' || s.status === 'QUEUED');
  }, [scans]);

  useEffect(() => {
    if (!autoRefreshEnabled || !hasActiveScans) return;

    const interval = setInterval(() => {
      fetchScans(true);
    }, 3000);

    return () => clearInterval(interval);
  }, [autoRefreshEnabled, hasActiveScans, fetchScans]);

  // Cancellazione immediata scansione (SIGTERM)
  const handleCancelScan = async () => {
    if (!scanToCancel) return;
    setCancelling(true);
    try {
      const response = await apiClient.post<ApiResponse<{ message: string; scan_id: string }>>(
        `/scans/${scanToCancel.id}/cancel`
      );
      if (response.data && response.data.success) {
        setScanToCancel(null);
        fetchScans(true);
      } else {
        alert(response.data?.error?.message || 'Impossibile annullare la scansione');
      }
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Errore durante la cancellazione della scansione');
    } finally {
      setCancelling(false);
    }
  };

  // Toggle Schedulazione
  const handleToggleSchedule = async (scheduleId: string) => {
    try {
      const response = await apiClient.put<ApiResponse<ScheduleItem>>(`/schedules/${scheduleId}/toggle`);
      if (response.data && response.data.success) {
        fetchSchedules();
      }
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Errore durante l\'aggiornamento della pianificazione');
    }
  };

  // Eliminazione Schedulazione
  const handleDeleteSchedule = async () => {
    if (!scheduleToDelete) return;
    setDeletingSchedule(true);
    try {
      const response = await apiClient.delete<ApiResponse<{ message: string }>>(`/schedules/${scheduleToDelete.id}`);
      if (response.data && response.data.success) {
        setScheduleToDelete(null);
        fetchSchedules();
      }
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Errore durante l\'eliminazione della pianificazione');
    } finally {
      setDeletingSchedule(false);
    }
  };

  const handleCopy = (id: string) => {
    navigator.clipboard.writeText(id);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleDownloadPdf = async (scanId: string) => {
    try {
      await downloadReportFile({
        url: `/scans/${scanId}/reports/pdf`,
        defaultFilename: `report_scan_${scanId.slice(0, 8)}.pdf`,
        onStart: () => {
          setDownloadingScanId(scanId);
          setDownloadType('pdf');
        },
        onSuccess: () => {
          setDownloadingScanId(null);
          setDownloadType(null);
        },
        onError: (err) => {
          setDownloadingScanId(null);
          setDownloadType(null);
          alert(`Errore durante il download del PDF: ${err.message}`);
        },
      });
    } catch {
      // Gestito in onError
    }
  };

  const handleDownloadCsv = async (scanId: string) => {
    try {
      await downloadReportFile({
        url: `/scans/${scanId}/reports/csv`,
        defaultFilename: `findings_scan_${scanId.slice(0, 8)}.csv`,
        onStart: () => {
          setDownloadingScanId(scanId);
          setDownloadType('csv');
        },
        onSuccess: () => {
          setDownloadingScanId(null);
          setDownloadType(null);
        },
        onError: (err) => {
          setDownloadingScanId(null);
          setDownloadType(null);
          alert(`Errore durante l'esportazione del CSV: ${err.message}`);
        },
      });
    } catch {
      // Gestito in onError
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Terminal className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-2xl font-bold tracking-tight text-white flex items-center space-x-2">
                <span>Network Scans & Engine</span>
                {hasActiveScans && (
                  <span className="text-xs px-2.5 py-0.5 rounded-full bg-cyan-500/15 text-cyan-400 border border-cyan-500/30 animate-pulse font-medium">
                    Job in esecuzione
                  </span>
                )}
              </h2>
              <p className="text-sm text-slate-400 mt-0.5">
                Esecuzione sicura di Nmap, monitoraggio live con polling e pianificazioni cron
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={() => setIsNewScanModalOpen(true)}
            className="flex items-center space-x-2 px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] hover:shadow-[0_0_20px_rgba(6,182,212,0.4)]"
          >
            <Plus className="w-4 h-4" />
            <span>Nuova Scansione</span>
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center justify-between border-b border-cyber-border bg-slate-900/40 rounded-xl px-4 py-2.5">
        <div className="flex space-x-2">
          <button
            onClick={() => setActiveTab('SCANS')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold flex items-center space-x-2 transition-colors ${
              activeTab === 'SCANS'
                ? 'bg-cyan-500/15 text-cyan-400 border border-cyan-500/30'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <Play className="w-3.5 h-3.5" />
            <span>Scansioni Eseguite & Live ({totalScans})</span>
          </button>

          <button
            onClick={() => setActiveTab('SCHEDULES')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold flex items-center space-x-2 transition-colors ${
              activeTab === 'SCHEDULES'
                ? 'bg-cyan-500/15 text-cyan-400 border border-cyan-500/30'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
            }`}
          >
            <Calendar className="w-3.5 h-3.5" />
            <span>Pianificazioni Ricorrenti ({schedules.length})</span>
          </button>
        </div>

        {/* Toolbar destra */}
        <div className="flex items-center space-x-3 text-xs">
          {activeTab === 'SCANS' && flatGroups.length > 0 && (
            <div className="flex items-center space-x-1.5 text-xs">
              <Folder className="w-3.5 h-3.5 text-slate-400" />
              <select
                value={selectedGroupFilter}
                onChange={(e) => setSelectedGroupFilter(e.target.value)}
                className="bg-slate-900 border border-cyber-border rounded-lg px-2.5 py-1 text-xs text-slate-300 focus:outline-none focus:border-cyan-500"
              >
                <option value="">Tutte le cartelle scope</option>
                {flatGroups.map((g) => (
                  <option key={g.id} value={g.id}>
                    📂 {g.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {activeTab === 'SCANS' && (
            <label className="hidden md:flex items-center space-x-2 text-slate-400 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={autoRefreshEnabled}
                onChange={(e) => setAutoRefreshEnabled(e.target.checked)}
                className="rounded border-slate-700 bg-slate-900 text-cyan-500 focus:ring-0"
              />
              <span className="text-[11px]">Auto-Refresh (3s)</span>
            </label>
          )}

          <button
            onClick={() => (activeTab === 'SCANS' ? fetchScans() : fetchSchedules())}
            className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
            title="Aggiorna ora"
          >
            <RefreshCw
              className={`w-3.5 h-3.5 ${
                loadingScans || loadingSchedules ? 'animate-spin text-cyan-400' : ''
              }`}
            />
          </button>
        </div>
      </div>

      {/* Scheda 1: Scansioni Eseguite & Live */}
      {activeTab === 'SCANS' && (
        <div className="rounded-xl border border-cyber-border bg-cyber-card/40 overflow-hidden shadow-xl">
          {scansError && (
            <div className="m-4 p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
                <span>{scansError}</span>
              </div>
              <button
                onClick={() => fetchScans()}
                className="px-2 py-0.5 rounded bg-rose-900/50 hover:bg-rose-900 text-rose-200 text-xs"
              >
                Ricarica
              </button>
            </div>
          )}

          {loadingScans ? (
            <div className="py-16 flex flex-col items-center justify-center text-slate-400 space-y-3">
              <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
              <span className="text-xs">Caricamento elenco scansioni...</span>
            </div>
          ) : displayedScans.length === 0 ? (
            <div className="py-16 flex flex-col items-center justify-center text-center space-y-3 p-6">
              <div className="w-12 h-12 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-slate-500">
                <Terminal className="w-6 h-6" />
              </div>
              <div className="text-sm font-semibold text-slate-200">
                {scans.length === 0 ? 'Nessuna scansione presente' : 'Nessuna scansione per la cartella selezionata'}
              </div>
              <p className="text-xs text-slate-400 max-w-sm">
                {scans.length === 0
                  ? 'Non è ancora stato avviato alcun processo di analisi di rete per questa organizzazione.'
                  : 'Nessun job di scansione registrato per i target appartenenti a questa cartella scope.'}
              </p>
              {scans.length === 0 ? (
                <button
                  onClick={() => setIsNewScanModalOpen(true)}
                  className="mt-2 inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs transition-colors"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Avvia prima scansione</span>
                </button>
              ) : (
                <button
                  onClick={() => setSelectedGroupFilter('')}
                  className="mt-2 inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs transition-colors"
                >
                  <span>Mostra tutte le scansioni</span>
                </button>
              )}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-cyber-border/80 text-[11px] font-semibold text-slate-400 uppercase tracking-wider bg-slate-900/40">
                    <th className="py-3 px-4">Scansione ID</th>
                    <th className="py-3 px-4">Target / Cartella</th>
                    <th className="py-3 px-4">Trigger</th>
                    <th className="py-3 px-4">Stato Esecuzione</th>
                    <th className="py-3 px-4">Host Up</th>
                    <th className="py-3 px-4">Vulnerabilità Rilevate</th>
                    <th className="py-3 px-4">Durata</th>
                    <th className="py-3 px-4">Data Creazione</th>
                    <th className="py-3 px-4 text-right">Azioni</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-cyber-border/40 text-xs">
                  {displayedScans.map((scan) => {
                    const isRunningOrQueued = scan.status === 'RUNNING' || scan.status === 'QUEUED';
                    const target = targetsMap.get(scan.target_id);
                    const group = target?.group_id ? groupsMap.get(target.group_id) : null;
                    return (
                      <tr key={scan.id} className="hover:bg-slate-800/25 transition-colors">
                        {/* ID Scansione */}
                        <td className="py-3.5 px-4 font-mono font-medium text-slate-100">
                          <div className="flex items-center space-x-2">
                            <span className="text-cyan-400">{scan.id.slice(0, 8)}...</span>
                            <button
                              onClick={() => handleCopy(scan.id)}
                              className="text-slate-500 hover:text-slate-300 p-1 rounded"
                              title="Copia ID scansione"
                            >
                              {copiedId === scan.id ? (
                                <Check className="w-3 h-3 text-emerald-400" />
                              ) : (
                                <Copy className="w-3 h-3" />
                              )}
                            </button>
                          </div>
                        </td>

                        {/* Target & Cartella */}
                        <td className="py-3.5 px-4">
                          <div>
                            <span className="font-mono text-slate-200 block text-xs font-semibold">
                              {target ? target.value : `${scan.target_id.slice(0, 8)}...`}
                            </span>
                            {group ? (
                              <span
                                className="inline-flex items-center space-x-1 px-1.5 py-0.2 rounded text-[10px] font-medium border mt-0.5"
                                style={{
                                  backgroundColor: `${group.color_tag || '#06b6d4'}15`,
                                  borderColor: `${group.color_tag || '#06b6d4'}35`,
                                  color: group.color_tag || '#06b6d4',
                                }}
                              >
                                <Folder className="w-2.5 h-2.5" />
                                <span className="truncate max-w-[120px]">{group.name}</span>
                              </span>
                            ) : (
                              <span className="text-[10px] text-slate-500 italic">
                                Non assegnato
                              </span>
                            )}
                          </div>
                        </td>

                        {/* Trigger */}
                        <td className="py-3.5 px-4 text-slate-300">
                          <span className="px-2 py-0.5 rounded bg-slate-900 border border-cyber-border text-[11px]">
                            {scan.trigger_type}
                          </span>
                        </td>

                        {/* Stato */}
                        <td className="py-3.5 px-4">
                          <ScanProgressBadge status={scan.status} />
                        </td>

                        {/* Host Up */}
                        <td className="py-3.5 px-4 text-slate-300 font-mono">
                          {scan.summary_stats?.up_hosts ?? 0} / {scan.summary_stats?.total_hosts ?? 0}
                        </td>

                        {/* Chip Vulnerabilità */}
                        <td className="py-3.5 px-4">
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

                        {/* Durata */}
                        <td className="py-3.5 px-4 text-slate-400 font-mono">
                          {scan.summary_stats?.duration_seconds
                            ? `${scan.summary_stats.duration_seconds.toFixed(1)}s`
                            : '-'}
                        </td>

                        {/* Data */}
                        <td className="py-3.5 px-4 text-slate-400">
                          {new Date(scan.created_at).toLocaleString('it-IT', {
                            day: '2-digit',
                            month: '2-digit',
                            year: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </td>

                        {/* Azioni */}
                        <td className="py-3.5 px-4 text-right">
                          <div className="flex items-center justify-end space-x-1.5">
                            {/* Visualizza Dettaglio e Findings */}
                            <Link
                              to={`/scans/${scan.id}`}
                              className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-cyan-400 transition-colors"
                              title="Dettaglio Scansione & Vulnerabilità"
                            >
                              <ExternalLink className="w-3.5 h-3.5" />
                            </Link>

                            {/* Download PDF White-Label */}
                            {scan.status === 'COMPLETED' && (
                              <>
                                <button
                                  type="button"
                                  onClick={() => handleDownloadPdf(scan.id)}
                                  disabled={downloadingScanId === scan.id}
                                  className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-emerald-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                  title="Scarica Report PDF White-Label"
                                >
                                  {downloadingScanId === scan.id && downloadType === 'pdf' ? (
                                    <RefreshCw className="w-3.5 h-3.5 animate-spin text-emerald-400" />
                                  ) : (
                                    <FileText className="w-3.5 h-3.5" />
                                  )}
                                </button>

                                <button
                                  type="button"
                                  onClick={() => handleDownloadCsv(scan.id)}
                                  disabled={downloadingScanId === scan.id}
                                  className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-blue-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                  title="Scarica CSV Findings"
                                >
                                  {downloadingScanId === scan.id && downloadType === 'csv' ? (
                                    <RefreshCw className="w-3.5 h-3.5 animate-spin text-blue-400" />
                                  ) : (
                                    <FileSpreadsheet className="w-3.5 h-3.5" />
                                  )}
                                </button>
                              </>
                            )}

                            {/* Annulla Scansione (SIGTERM) */}
                            {isRunningOrQueued && (
                              <button
                                onClick={() => setScanToCancel(scan)}
                                className="p-1.5 rounded-lg border border-rose-900/40 bg-rose-950/20 text-rose-400 hover:bg-rose-900/40 hover:border-rose-800 transition-colors"
                                title="Interrompi Scansione"
                              >
                                <StopCircle className="w-3.5 h-3.5" />
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Scheda 2: Pianificazioni Ricorrenti (Cron) */}
      {activeTab === 'SCHEDULES' && (
        <div className="rounded-xl border border-cyber-border bg-cyber-card/40 overflow-hidden shadow-xl">
          {schedulesError && (
            <div className="m-4 p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
                <span>{schedulesError}</span>
              </div>
              <button
                onClick={fetchSchedules}
                className="px-2 py-0.5 rounded bg-rose-900/50 hover:bg-rose-900 text-rose-200 text-xs"
              >
                Ricarica
              </button>
            </div>
          )}

          {loadingSchedules ? (
            <div className="py-16 flex flex-col items-center justify-center text-slate-400 space-y-3">
              <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
              <span className="text-xs">Caricamento pianificazioni in corso...</span>
            </div>
          ) : schedules.length === 0 ? (
            <div className="py-16 flex flex-col items-center justify-center text-center space-y-3 p-6">
              <div className="w-12 h-12 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-slate-500">
                <Calendar className="w-6 h-6" />
              </div>
              <div className="text-sm font-semibold text-slate-200">Nessuna schedulazione configurata</div>
              <p className="text-xs text-slate-400 max-w-sm">
                Puoi impostare scansioni periodiche automatiche (es. ogni notte o settimanali) per garantire monitoraggio continuo.
              </p>
              <button
                onClick={() => setIsNewScanModalOpen(true)}
                className="mt-2 inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs transition-colors"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Crea prima pianificazione</span>
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-cyber-border/80 text-[11px] font-semibold text-slate-400 uppercase tracking-wider bg-slate-900/40">
                    <th className="py-3 px-4">Schedulazione ID</th>
                    <th className="py-3 px-4">Espressione Cron</th>
                    <th className="py-3 px-4">Stato</th>
                    <th className="py-3 px-4">Ultima Esecuzione</th>
                    <th className="py-3 px-4">Prossima Esecuzione</th>
                    <th className="py-3 px-4">Creata il</th>
                    <th className="py-3 px-4 text-right">Azioni</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-cyber-border/40 text-xs">
                  {schedules.map((schedule) => (
                    <tr key={schedule.id} className="hover:bg-slate-800/25 transition-colors">
                      <td className="py-3.5 px-4 font-mono font-medium text-cyan-400">
                        {schedule.id.slice(0, 8)}...
                      </td>
                      <td className="py-3.5 px-4 font-mono font-medium text-slate-200">
                        <span className="px-2 py-0.5 rounded bg-slate-900 border border-cyber-border text-cyan-400">
                          {schedule.cron_expression}
                        </span>
                      </td>
                      <td className="py-3.5 px-4">
                        <button
                          onClick={() => handleToggleSchedule(schedule.id)}
                          className={`inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold border transition-colors ${
                            schedule.is_active
                              ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 hover:bg-emerald-500/20'
                              : 'bg-slate-800 text-slate-400 border-slate-700 hover:bg-slate-700'
                          }`}
                          title="Clicca per attivare/disattivare"
                        >
                          <Power className="w-3 h-3" />
                          <span>{schedule.is_active ? 'Attivo' : 'Sospeso'}</span>
                        </button>
                      </td>
                      <td className="py-3.5 px-4 text-slate-400">
                        {schedule.last_run_at
                          ? new Date(schedule.last_run_at).toLocaleString('it-IT')
                          : <span className="text-slate-600">Mai eseguita</span>}
                      </td>
                      <td className="py-3.5 px-4 text-slate-300 font-mono text-[11px]">
                        {schedule.next_run_at
                          ? new Date(schedule.next_run_at).toLocaleString('it-IT')
                          : <span className="text-slate-600">-</span>}
                      </td>
                      <td className="py-3.5 px-4 text-slate-400">
                        {new Date(schedule.created_at).toLocaleDateString('it-IT')}
                      </td>
                      <td className="py-3.5 px-4 text-right">
                        <button
                          onClick={() => setScheduleToDelete(schedule)}
                          className="p-1.5 rounded-lg border border-rose-900/40 bg-rose-950/20 text-rose-400 hover:bg-rose-900/40 hover:border-rose-800 transition-colors"
                          title="Elimina Schedulazione"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Modale Nuova Scansione */}
      <NewScanModal
        isOpen={isNewScanModalOpen}
        onClose={() => setIsNewScanModalOpen(false)}
        onScanLaunched={() => {
          fetchScans();
          setActiveTab('SCANS');
        }}
        onScheduleCreated={() => {
          fetchSchedules();
          setActiveTab('SCHEDULES');
        }}
      />

      {/* Dialog Conferma Annullamento Scansione */}
      {scanToCancel && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150">
          <div className="max-w-md w-full p-6 rounded-2xl border border-rose-800/40 bg-[#111726] shadow-2xl space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400">
                <StopCircle className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-white">Interrompi Scansione</h3>
                <p className="text-xs text-slate-400">Invia segnale di cancellazione SIGTERM a Nmap</p>
              </div>
            </div>

            <p className="text-xs text-slate-300">
              Sei sicuro di voler terminare forzatamente la scansione{' '}
              <strong className="font-mono text-cyan-400">{scanToCancel.id.slice(0, 8)}...</strong>? Il processo verrà interrotto immediatamente e le risorse liberate.
            </p>

            <div className="flex items-center justify-end space-x-3 pt-2">
              <button
                type="button"
                onClick={() => setScanToCancel(null)}
                disabled={cancelling}
                className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
              >
                Annulla
              </button>
              <button
                type="button"
                onClick={handleCancelScan}
                disabled={cancelling}
                className="px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold transition-colors disabled:opacity-50"
              >
                {cancelling ? 'Interruzione...' : 'Interrompi Subito'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Dialog Conferma Eliminazione Schedulazione */}
      {scheduleToDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150">
          <div className="max-w-md w-full p-6 rounded-2xl border border-rose-800/40 bg-[#111726] shadow-2xl space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400">
                <Trash2 className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-white">Elimina Pianificazione</h3>
                <p className="text-xs text-slate-400">Rimuove l'esecuzione automatica del job cron</p>
              </div>
            </div>

            <p className="text-xs text-slate-300">
              Sei sicuro di voler rimuovere la schedulazione{' '}
              <strong className="font-mono text-cyan-400">{scheduleToDelete.cron_expression}</strong>? Nessuna nuova scansione automatica verrà più generata.
            </p>

            <div className="flex items-center justify-end space-x-3 pt-2">
              <button
                type="button"
                onClick={() => setScheduleToDelete(null)}
                disabled={deletingSchedule}
                className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
              >
                Annulla
              </button>
              <button
                type="button"
                onClick={handleDeleteSchedule}
                disabled={deletingSchedule}
                className="px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold transition-colors disabled:opacity-50"
              >
                {deletingSchedule ? 'Eliminazione...' : 'Elimina Schedulazione'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
