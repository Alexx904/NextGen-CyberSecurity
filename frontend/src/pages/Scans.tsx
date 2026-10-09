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
  FolderOpen,
  ChevronDown,
  ChevronRight,
  List,
  Search,
  Crosshair,
  Layers,
  X,
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

  // Modalità di visualizzazione: Raggruppata per Cartella/Azienda vs Tabella Lineare
  const [viewMode, setViewMode] = useState<'GROUPED' | 'TABLE'>('GROUPED');
  const [collapsedGroupIds, setCollapsedGroupIds] = useState<Set<string>>(new Set());
  const [searchFilter, setSearchFilter] = useState<string>('');

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

  interface ScansByTarget {
    targetId: string;
    targetValue: string;
    targetType: string;
    targetDescription?: string;
    scans: ScanItem[];
  }

  interface ScansGroup {
    id: string;
    name: string;
    color_tag: string;
    targets: ScansByTarget[];
    totalScans: number;
    totalHostsUp: number;
    criticalCount: number;
    highCount: number;
    mediumCount: number;
    lowCount: number;
    hasRunning: boolean;
  }

  // Raggruppamento per Cartella/Azienda Scope e poi per Target/IP
  const groupedScans = useMemo(() => {
    const term = searchFilter.trim().toLowerCase();
    const filtered = displayedScans.filter((s) => {
      if (!term) return true;
      const target = targetsMap.get(s.target_id);
      const targetVal = target?.value.toLowerCase() || '';
      const desc = target?.description?.toLowerCase() || '';
      const scanId = s.id.toLowerCase();
      const groupName = target?.group_id ? groupsMap.get(target.group_id)?.name.toLowerCase() || '' : '';
      return targetVal.includes(term) || desc.includes(term) || scanId.includes(term) || groupName.includes(term);
    });

    const groupsData = new Map<string, Map<string, ScanItem[]>>();

    filtered.forEach((scan) => {
      const target = targetsMap.get(scan.target_id);
      const groupId = target?.group_id || 'unassigned';
      if (!groupsData.has(groupId)) {
        groupsData.set(groupId, new Map());
      }
      const targetMap = groupsData.get(groupId)!;
      if (!targetMap.has(scan.target_id)) {
        targetMap.set(scan.target_id, []);
      }
      targetMap.get(scan.target_id)!.push(scan);
    });

    const result: ScansGroup[] = [];

    groupsData.forEach((targetsMapForGroup, groupId) => {
      let name = 'Bersagli Non Assegnati (Root Scope)';
      let color_tag = '#64748b';
      if (groupId !== 'unassigned') {
        const gNode = groupsMap.get(groupId);
        if (gNode) {
          name = gNode.name;
          color_tag = gNode.color_tag || '#06b6d4';
        } else {
          name = 'Cartella / Ambito Sconosciuto';
        }
      }

      const targets: ScansByTarget[] = [];
      let totalScans = 0;
      let totalHostsUp = 0;
      let criticalCount = 0;
      let highCount = 0;
      let mediumCount = 0;
      let lowCount = 0;
      let hasRunning = false;

      targetsMapForGroup.forEach((scansForTarget, targetId) => {
        const targetObj = targetsMap.get(targetId);
        targets.push({
          targetId,
          targetValue: targetObj?.value || `${targetId.slice(0, 8)}...`,
          targetType: targetObj?.type || 'IP',
          targetDescription: targetObj?.description,
          scans: scansForTarget,
        });

        totalScans += scansForTarget.length;
        scansForTarget.forEach((s) => {
          totalHostsUp += s.summary_stats?.up_hosts ?? 0;
          criticalCount += s.summary_stats?.critical_count ?? 0;
          highCount += s.summary_stats?.high_count ?? 0;
          mediumCount += s.summary_stats?.medium_count ?? 0;
          lowCount += s.summary_stats?.low_count ?? 0;
          if (s.status === 'RUNNING' || s.status === 'QUEUED') {
            hasRunning = true;
          }
        });
      });

      // Ordina bersagli per indirizzo / valore
      targets.sort((a, b) => a.targetValue.localeCompare(b.targetValue));

      result.push({
        id: groupId,
        name,
        color_tag,
        targets,
        totalScans,
        totalHostsUp,
        criticalCount,
        highCount,
        mediumCount,
        lowCount,
        hasRunning,
      });
    });

    result.sort((a, b) => {
      if (a.id === 'unassigned') return 1;
      if (b.id === 'unassigned') return -1;
      return a.name.localeCompare(b.name);
    });

    return result;
  }, [displayedScans, targetsMap, groupsMap, searchFilter]);

  const toggleGroup = (groupId: string) => {
    setCollapsedGroupIds((prev) => {
      const next = new Set(prev);
      if (next.has(groupId)) next.delete(groupId);
      else next.add(groupId);
      return next;
    });
  };

  const expandAllGroups = () => {
    setCollapsedGroupIds(new Set());
  };

  const collapseAllGroups = () => {
    setCollapsedGroupIds(new Set(groupedScans.map((g) => g.id)));
  };

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

      {/* Sotto-Toolbar di Ricerca e Selezione Vista (Tab SCANS) */}
      {activeTab === 'SCANS' && (
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-3 rounded-xl border border-cyber-border bg-[#111726]/60 backdrop-blur-md text-xs">
          {/* Ricerca per IP o Cartella */}
          <div className="relative flex-1 max-w-md">
            <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={searchFilter}
              onChange={(e) => setSearchFilter(e.target.value)}
              placeholder="Cerca per IP, host, cartella scope o ID scansione..."
              className="w-full pl-9 pr-7 py-1.5 rounded-lg bg-slate-950 border border-cyber-border text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500 text-xs font-mono"
            />
            {searchFilter && (
              <button
                type="button"
                onClick={() => setSearchFilter('')}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>

          {/* Switch Modalità: Raggruppata per Cartella/Azienda vs Tabellare */}
          <div className="flex items-center space-x-2 self-end sm:self-auto">
            {viewMode === 'GROUPED' && groupedScans.length > 0 && (
              <div className="flex items-center space-x-1.5 mr-2">
                <button
                  type="button"
                  onClick={expandAllGroups}
                  className="px-2 py-1 rounded bg-slate-900 hover:bg-slate-800 text-slate-300 border border-cyber-border text-[11px] font-medium transition-colors"
                  title="Espandi tutti i gruppi"
                >
                  Espandi tutti
                </button>
                <button
                  type="button"
                  onClick={collapseAllGroups}
                  className="px-2 py-1 rounded bg-slate-900 hover:bg-slate-800 text-slate-300 border border-cyber-border text-[11px] font-medium transition-colors"
                  title="Comprimi tutti i gruppi"
                >
                  Comprimi tutti
                </button>
              </div>
            )}

            <div className="flex items-center rounded-lg border border-cyber-border bg-slate-950 p-0.5">
              <button
                type="button"
                onClick={() => setViewMode('GROUPED')}
                className={`flex items-center space-x-1.5 px-2.5 py-1 rounded-md text-[11px] font-semibold transition-colors ${
                  viewMode === 'GROUPED'
                    ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-sm'
                    : 'text-slate-400 hover:text-white'
                }`}
                title="Visualizzazione organizzata per Cartella / Azienda e singoli IP"
              >
                <Layers className="w-3.5 h-3.5" />
                <span>Vista per Cartella/Azienda</span>
              </button>
              <button
                type="button"
                onClick={() => setViewMode('TABLE')}
                className={`flex items-center space-x-1.5 px-2.5 py-1 rounded-md text-[11px] font-semibold transition-colors ${
                  viewMode === 'TABLE'
                    ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-sm'
                    : 'text-slate-400 hover:text-white'
                }`}
                title="Visualizzazione tabellare lineare"
              >
                <List className="w-3.5 h-3.5" />
                <span>Tabella Lineare</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Scheda 1: Scansioni Eseguite & Live */}
      {activeTab === 'SCANS' && (
        <div className="space-y-4">
          {scansError && (
            <div className="p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center justify-between">
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
            <div className="py-16 rounded-xl border border-cyber-border bg-cyber-card/40 flex flex-col items-center justify-center text-slate-400 space-y-3">
              <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
              <span className="text-xs">Caricamento storico scansioni e perimetri...</span>
            </div>
          ) : displayedScans.length === 0 || (viewMode === 'GROUPED' && groupedScans.length === 0) ? (
            <div className="py-16 rounded-xl border border-cyber-border bg-cyber-card/40 flex flex-col items-center justify-center text-center space-y-3 p-6">
              <div className="w-12 h-12 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-slate-500">
                <Terminal className="w-6 h-6" />
              </div>
              <div className="text-sm font-semibold text-slate-200">
                {scans.length === 0
                  ? 'Nessuna scansione presente'
                  : searchFilter
                  ? `Nessun risultato per "${searchFilter}"`
                  : 'Nessuna scansione per la cartella selezionata'}
              </div>
              <p className="text-xs text-slate-400 max-w-sm">
                {scans.length === 0
                  ? 'Non è ancora stato avviato alcun processo di analisi di rete per questa organizzazione.'
                  : searchFilter
                  ? 'Verifica i termini inseriti o prova a cercare per IP o ID parziale.'
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
                  onClick={() => {
                    setSelectedGroupFilter('');
                    setSearchFilter('');
                  }}
                  className="mt-2 inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs transition-colors"
                >
                  <span>Azzera filtri e mostra tutto</span>
                </button>
              )}
            </div>
          ) : viewMode === 'GROUPED' ? (
            /* Vista Raggruppata per Cartella/Azienda Scope */
            <div className="space-y-4">
              {groupedScans.map((group) => {
                const isExpanded = !collapsedGroupIds.has(group.id);

                return (
                  <div
                    key={group.id}
                    className="rounded-xl border border-cyber-border bg-[#111726]/80 backdrop-blur-md overflow-hidden shadow-lg transition-all"
                  >
                    {/* Header Cartella / Azienda */}
                    <div
                      onClick={() => toggleGroup(group.id)}
                      className="p-3.5 bg-slate-900/70 hover:bg-slate-900 border-b border-cyber-border/80 cursor-pointer flex flex-col sm:flex-row sm:items-center justify-between gap-3 select-none"
                    >
                      <div className="flex items-center space-x-3">
                        <button
                          type="button"
                          className="p-1 rounded text-slate-400 hover:text-white transition-colors"
                          title={isExpanded ? 'Comprimi cartella' : 'Espandi cartella'}
                        >
                          {isExpanded ? (
                            <ChevronDown className="w-4 h-4 text-cyan-400" />
                          ) : (
                            <ChevronRight className="w-4 h-4 text-slate-400" />
                          )}
                        </button>

                        <div
                          className="p-2 rounded-lg border flex items-center justify-center shrink-0"
                          style={{
                            backgroundColor: `${group.color_tag}15`,
                            borderColor: `${group.color_tag}40`,
                            color: group.color_tag,
                          }}
                        >
                          {isExpanded ? <FolderOpen className="w-4 h-4" /> : <Folder className="w-4 h-4" />}
                        </div>

                        <div>
                          <div className="flex items-center space-x-2">
                            <h3 className="text-sm font-bold text-white tracking-wide">{group.name}</h3>
                            <span className="text-[11px] px-2 py-0.5 rounded-full font-mono bg-slate-800 text-slate-300 border border-slate-700">
                              {group.targets.length} {group.targets.length === 1 ? 'Target / IP' : 'Target / IP'}
                            </span>
                            {group.hasRunning && (
                              <span className="text-[10px] px-2 py-0.5 rounded-full bg-cyan-500/15 text-cyan-400 border border-cyan-500/30 animate-pulse font-medium">
                                Job in corso
                              </span>
                            )}
                          </div>
                          <p className="text-[11px] text-slate-400 mt-0.5">
                            {group.totalScans} {group.totalScans === 1 ? 'scansione totale' : 'scansioni totali'} · {group.totalHostsUp} host UP
                          </p>
                        </div>
                      </div>

                      {/* Badges metriche e conteggi vulnerabilità */}
                      <div className="flex items-center space-x-2 shrink-0">
                        <div className="flex items-center space-x-1.5 text-[11px] font-mono">
                          {group.criticalCount > 0 && (
                            <span className="px-2 py-0.5 rounded bg-red-500/15 text-red-400 border border-red-500/30 font-bold">
                              {group.criticalCount} Crit
                            </span>
                          )}
                          {group.highCount > 0 && (
                            <span className="px-2 py-0.5 rounded bg-orange-500/15 text-orange-400 border border-orange-500/30 font-bold">
                              {group.highCount} High
                            </span>
                          )}
                          {group.mediumCount > 0 && (
                            <span className="px-2 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">
                              {group.mediumCount} Med
                            </span>
                          )}
                          {group.criticalCount === 0 && group.highCount === 0 && group.mediumCount === 0 && (
                            <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 text-[10px]">
                              0 Crit/High
                            </span>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Contenuto espanso: Targets e Scansioni */}
                    {isExpanded && (
                      <div className="p-3 sm:p-4 space-y-4 bg-slate-950/40">
                        {group.targets.map((tgt) => (
                          <div
                            key={tgt.targetId}
                            className="rounded-lg border border-cyber-border/70 bg-[#0d121f]/90 overflow-hidden shadow-sm"
                          >
                            {/* Header per singolo Target / IP */}
                            <div className="px-3.5 py-2.5 bg-slate-900/60 border-b border-cyber-border/60 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                              <div className="flex items-center space-x-2.5">
                                <Crosshair className="w-4 h-4 text-cyan-400 shrink-0" />
                                <span className="font-mono font-bold text-white text-xs text-cyan-300">
                                  {tgt.targetValue}
                                </span>
                                <button
                                  type="button"
                                  onClick={() => handleCopy(tgt.targetValue)}
                                  className="text-slate-500 hover:text-slate-300 p-0.5 rounded transition-colors"
                                  title="Copia valore target"
                                >
                                  {copiedId === tgt.targetValue ? (
                                    <Check className="w-3 h-3 text-emerald-400" />
                                  ) : (
                                    <Copy className="w-3 h-3" />
                                  )}
                                </button>
                                <span className="text-[10px] px-1.5 py-0.2 rounded bg-slate-800 text-slate-300 border border-slate-700">
                                  {tgt.targetType}
                                </span>
                                {tgt.targetDescription && (
                                  <span className="text-slate-400 text-xs truncate max-w-xs italic hidden md:inline">
                                    — {tgt.targetDescription}
                                  </span>
                                )}
                              </div>
                              <div className="text-[11px] text-slate-400 font-mono">
                                {tgt.scans.length} {tgt.scans.length === 1 ? 'audit storico' : 'audit storici'}
                              </div>
                            </div>

                            {/* Elenco scansioni per questo IP */}
                            <div className="overflow-x-auto">
                              <table className="w-full text-left border-collapse text-xs">
                                <thead>
                                  <tr className="border-b border-cyber-border/40 text-[10px] font-semibold text-slate-400 uppercase tracking-wider bg-slate-900/20">
                                    <th className="py-2.5 px-3">Scansione ID</th>
                                    <th className="py-2.5 px-3">Trigger</th>
                                    <th className="py-2.5 px-3">Stato</th>
                                    <th className="py-2.5 px-3">Host Up</th>
                                    <th className="py-2.5 px-3">Vulnerabilità</th>
                                    <th className="py-2.5 px-3">Durata</th>
                                    <th className="py-2.5 px-3">Data</th>
                                    <th className="py-2.5 px-3 text-right">Azioni</th>
                                  </tr>
                                </thead>
                                <tbody className="divide-y divide-cyber-border/30">
                                  {tgt.scans.map((scan) => {
                                    const isRunningOrQueued = scan.status === 'RUNNING' || scan.status === 'QUEUED';
                                    return (
                                      <tr key={scan.id} className="hover:bg-slate-800/20 transition-colors">
                                        <td className="py-2.5 px-3 font-mono text-cyan-400 font-medium">
                                          <div className="flex items-center space-x-1.5">
                                            <span>{scan.id.slice(0, 8)}...</span>
                                            <button
                                              onClick={() => handleCopy(scan.id)}
                                              className="text-slate-500 hover:text-slate-300 p-0.5 rounded"
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
                                        <td className="py-2.5 px-3 text-slate-300">
                                          <span className="px-1.5 py-0.5 rounded bg-slate-900 border border-cyber-border text-[10px]">
                                            {scan.trigger_type}
                                          </span>
                                        </td>
                                        <td className="py-2.5 px-3">
                                          <ScanProgressBadge status={scan.status} />
                                        </td>
                                        <td className="py-2.5 px-3 text-slate-300 font-mono">
                                          {scan.summary_stats?.up_hosts ?? 0} / {scan.summary_stats?.total_hosts ?? 0}
                                        </td>
                                        <td className="py-2.5 px-3">
                                          <div className="flex items-center space-x-1 text-[10px]">
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
                                        <td className="py-2.5 px-3 text-slate-400 font-mono text-[11px]">
                                          {scan.summary_stats?.duration_seconds
                                            ? `${scan.summary_stats.duration_seconds.toFixed(1)}s`
                                            : '-'}
                                        </td>
                                        <td className="py-2.5 px-3 text-slate-400 text-[11px]">
                                          {new Date(scan.created_at).toLocaleString('it-IT', {
                                            day: '2-digit',
                                            month: '2-digit',
                                            hour: '2-digit',
                                            minute: '2-digit',
                                          })}
                                        </td>
                                        <td className="py-2.5 px-3 text-right">
                                          <div className="flex items-center justify-end space-x-1.5">
                                            <Link
                                              to={`/scans/${scan.id}`}
                                              className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-cyan-400 transition-colors"
                                              title="Dettaglio Scansione & Vulnerabilità"
                                            >
                                              <ExternalLink className="w-3.5 h-3.5" />
                                            </Link>
                                            {scan.status === 'COMPLETED' && (
                                              <>
                                                <button
                                                  type="button"
                                                  onClick={() => handleDownloadPdf(scan.id)}
                                                  disabled={downloadingScanId === scan.id}
                                                  className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-emerald-400 transition-colors disabled:opacity-50"
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
                                                  className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-blue-400 transition-colors disabled:opacity-50"
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
                                            {isRunningOrQueued && (
                                              <button
                                                onClick={() => setScanToCancel(scan)}
                                                className="p-1.5 rounded-lg border border-rose-900/40 bg-rose-950/20 text-rose-400 hover:bg-rose-900/40 transition-colors"
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
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          ) : (
            /* Vista Tabellare Lineare Classica */
            <div className="rounded-xl border border-cyber-border bg-cyber-card/40 overflow-hidden shadow-xl">
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
                          <td className="py-3.5 px-4 text-slate-300">
                            <span className="px-2 py-0.5 rounded bg-slate-900 border border-cyber-border text-[11px]">
                              {scan.trigger_type}
                            </span>
                          </td>
                          <td className="py-3.5 px-4">
                            <ScanProgressBadge status={scan.status} />
                          </td>
                          <td className="py-3.5 px-4 text-slate-300 font-mono">
                            {scan.summary_stats?.up_hosts ?? 0} / {scan.summary_stats?.total_hosts ?? 0}
                          </td>
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
                          <td className="py-3.5 px-4 text-slate-400 font-mono">
                            {scan.summary_stats?.duration_seconds
                              ? `${scan.summary_stats.duration_seconds.toFixed(1)}s`
                              : '-'}
                          </td>
                          <td className="py-3.5 px-4 text-slate-400">
                            {new Date(scan.created_at).toLocaleString('it-IT', {
                              day: '2-digit',
                              month: '2-digit',
                              year: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit',
                            })}
                          </td>
                          <td className="py-3.5 px-4 text-right">
                            <div className="flex items-center justify-end space-x-1.5">
                              <Link
                                to={`/scans/${scan.id}`}
                                className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-cyan-400 transition-colors"
                                title="Dettaglio Scansione & Vulnerabilità"
                              >
                                <ExternalLink className="w-3.5 h-3.5" />
                              </Link>
                              {scan.status === 'COMPLETED' && (
                                <>
                                  <button
                                    type="button"
                                    onClick={() => handleDownloadPdf(scan.id)}
                                    disabled={downloadingScanId === scan.id}
                                    className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-emerald-400 transition-colors disabled:opacity-50"
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
                                    className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-blue-400 transition-colors disabled:opacity-50"
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
                              {isRunningOrQueued && (
                                <button
                                  onClick={() => setScanToCancel(scan)}
                                  className="p-1.5 rounded-lg border border-rose-900/40 bg-rose-950/20 text-rose-400 hover:bg-rose-900/40 transition-colors"
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
