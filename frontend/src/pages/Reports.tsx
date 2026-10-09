import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  FileText,
  FileCode,
  FileSpreadsheet,
  RefreshCw,
  Search,
  X,
  ShieldCheck,
  AlertTriangle,
  Clock,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  Folder,
  FolderOpen,
  Activity,
  AlertCircle,
  List,
  Copy,
  Check,
  Building,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { reportsApi } from '../api/reports';
import { ApiResponse } from '../types/auth';
import { TargetGroupNode } from '../types/target';
import { ReportItem, ReportFilterParams, ReportTemplate, ReportFormat } from '../types/report';
import { SeverityLevel, ScanStatus } from '../types/scan';

export const Reports: React.FC = () => {
  // Modalità di visualizzazione: Raggruppata per Cartella/Azienda Scope vs Archivio Cronologico Lineare
  const [viewMode, setViewMode] = useState<'grouped' | 'flat'>('grouped');

  // Stato dati report
  const [reports, setReports] = useState<ReportItem[]>([]);
  const [groupsTree, setGroupsTree] = useState<TargetGroupNode[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Stato filtri di ricerca
  const [search, setSearch] = useState<string>('');
  const [severityFilter, setSeverityFilter] = useState<string>('');
  const [targetTypeFilter, setTargetTypeFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [dateFrom, setDateFrom] = useState<string>('');
  const [dateTo, setDateTo] = useState<string>('');

  // Paginazione per vista lineare
  const [page, setPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(10);

  // Accordion stato gruppi: Set degli ID collassati (vuoto = tutti espansi)
  const [collapsedGroupIds, setCollapsedGroupIds] = useState<Set<string>>(new Set());

  // Copia IP
  const [copiedIP, setCopiedIP] = useState<string | null>(null);

  // Stato operazioni di download attive
  const [downloadingId, setDownloadingId] = useState<string | null>(null);
  const [downloadingFormat, setDownloadingFormat] = useState<string | null>(null);
  const [downloadSuccessMsg, setDownloadSuccessMsg] = useState<string | null>(null);
  const [downloadErrorMsg, setDownloadErrorMsg] = useState<string | null>(null);

  // Template globale di preferenza
  const [preferredTemplate, setPreferredTemplate] = useState<ReportTemplate>('executive');

  // Caricamento dati dal backend (report + metadati cartelle/aziende)
  const fetchData = useCallback(async (silent = false) => {
    if (!silent) setLoading(true);
    setError(null);

    const isGrouped = viewMode === 'grouped';
    const effectiveLimit = isGrouped ? 100 : pageSize;
    const effectiveOffset = isGrouped ? 0 : (page - 1) * pageSize;

    const params: ReportFilterParams = {
      limit: effectiveLimit,
      offset: effectiveOffset,
    };

    if (search.trim()) params.search = search.trim();
    if (severityFilter) params.severity = severityFilter;
    if (targetTypeFilter) params.target_type = targetTypeFilter;
    if (statusFilter) params.status = statusFilter;
    if (dateFrom) params.date_from = dateFrom;
    if (dateTo) params.date_to = dateTo;

    try {
      const [reportsData, groupsRes] = await Promise.all([
        reportsApi.list(params),
        apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups').catch(() => null),
      ]);

      setReports(reportsData.items || []);
      setTotal(reportsData.total || 0);

      if (groupsRes?.data?.data) {
        setGroupsTree(groupsRes.data.data);
      }
    } catch (err: any) {
      setError(err.message || 'Errore durante il recupero dei report.');
      setReports([]);
      setTotal(0);
    } finally {
      if (!silent) setLoading(false);
    }
  }, [viewMode, page, pageSize, search, severityFilter, targetTypeFilter, statusFilter, dateFrom, dateTo]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Gestione del download con feedback di caricamento
  const handleDownload = async (scanId: string, format: ReportFormat, template: ReportTemplate = 'executive') => {
    setDownloadingId(scanId);
    setDownloadingFormat(`${format}_${template}`);
    setDownloadSuccessMsg(null);
    setDownloadErrorMsg(null);

    try {
      await reportsApi.download(scanId, format, template, {
        onSuccess: () => {
          setDownloadSuccessMsg(
            `Download completato con successo (${format.toUpperCase()} ${format === 'pdf' ? template : ''})`
          );
          setTimeout(() => setDownloadSuccessMsg(null), 4000);
        },
        onError: (err) => {
          setDownloadErrorMsg(err.message || 'Download del report fallito.');
          setTimeout(() => setDownloadErrorMsg(null), 5000);
        },
      });
    } catch (err: any) {
      setDownloadErrorMsg(err.message || 'Download non riuscito.');
      setTimeout(() => setDownloadErrorMsg(null), 5000);
    } finally {
      setDownloadingId(null);
      setDownloadingFormat(null);
    }
  };

  // Copia IP
  const handleCopy = (val: string) => {
    navigator.clipboard.writeText(val);
    setCopiedIP(val);
    setTimeout(() => setCopiedIP(null), 2000);
  };

  // Reset dei filtri
  const handleResetFilters = () => {
    setSearch('');
    setSeverityFilter('');
    setTargetTypeFilter('');
    setStatusFilter('');
    setDateFrom('');
    setDateTo('');
    setPage(1);
  };

  const activeFiltersCount = useMemo(() => {
    let count = 0;
    if (search.trim()) count++;
    if (severityFilter) count++;
    if (targetTypeFilter) count++;
    if (statusFilter) count++;
    if (dateFrom) count++;
    if (dateTo) count++;
    return count;
  }, [search, severityFilter, targetTypeFilter, statusFilter, dateFrom, dateTo]);

  // Metriche sintetiche aggregate della vista corrente
  const kpiStats = useMemo(() => {
    if (!reports || reports.length === 0) {
      return { avgScore: 100, avgGrade: 'A', highRiskCount: 0, totalUpHosts: 0 };
    }
    let totalScore = 0;
    let highRisk = 0;
    let upHosts = 0;

    reports.forEach((r) => {
      totalScore += r.posture_score;
      if (r.max_severity === 'CRITICAL' || r.max_severity === 'HIGH') {
        highRisk++;
      }
      upHosts += r.summary_stats.up_hosts || 0;
    });

    const avgScore = Math.round(totalScore / reports.length);
    let avgGrade = 'A';
    if (avgScore < 25) avgGrade = 'F';
    else if (avgScore < 50) avgGrade = 'D';
    else if (avgScore < 75) avgGrade = 'C';
    else if (avgScore < 90) avgGrade = 'B';

    return { avgScore, avgGrade, highRiskCount: highRisk, totalUpHosts: upHosts };
  }, [reports]);

  // Raggruppamento per Cartella / Azienda Cliente Scope
  const groupedCompanyReports = useMemo(() => {
    const gMap = new Map<string, TargetGroupNode>();
    function recurse(nodes: TargetGroupNode[]) {
      for (const n of nodes) {
        gMap.set(n.id, n);
        if (n.children && n.children.length > 0) {
          recurse(n.children);
        }
      }
    }
    recurse(groupsTree);

    // Mappa cartella -> target -> reports
    const folderMap = new Map<
      string,
      {
        id: string;
        name: string;
        color_tag: string;
        targets: Map<
          string,
          {
            targetId: string;
            targetValue: string;
            targetType: 'IP' | 'CIDR' | 'RANGE' | 'FQDN';
            reports: ReportItem[];
          }
        >;
      }
    >();

    // Pre-popola cartelle note da Target Groups
    gMap.forEach((gNode, gId) => {
      folderMap.set(gId, {
        id: gId,
        name: gNode.name,
        color_tag: gNode.color_tag || '#06b6d4',
        targets: new Map(),
      });
    });

    // Inizializza Root / Bersagli Non Assegnati
    folderMap.set('unassigned', {
      id: 'unassigned',
      name: 'Bersagli Non Assegnati (Root)',
      color_tag: '#64748b',
      targets: new Map(),
    });

    // Popola i report
    reports.forEach((report) => {
      const gId = report.group_id || 'unassigned';
      if (!folderMap.has(gId)) {
        folderMap.set(gId, {
          id: gId,
          name: report.group_name || 'Cartella Scope',
          color_tag: '#06b6d4',
          targets: new Map(),
        });
      }

      const folder = folderMap.get(gId)!;
      if (!folder.targets.has(report.target_id)) {
        folder.targets.set(report.target_id, {
          targetId: report.target_id,
          targetValue: report.target_value,
          targetType: report.target_type,
          reports: [],
        });
      }

      folder.targets.get(report.target_id)!.reports.push(report);
    });

    const result = Array.from(folderMap.values())
      .map((f) => {
        const targetList = Array.from(f.targets.values()).map((t) => {
          // Ordina i report del target dal più recente al più vecchio
          const sortedReports = [...t.reports].sort((a, b) => {
            const timeA = new Date(a.completed_at || a.created_at).getTime();
            const timeB = new Date(b.completed_at || b.created_at).getTime();
            return timeB - timeA;
          });
          return {
            ...t,
            reports: sortedReports,
            latestReport: sortedReports[0] || null,
          };
        });

        // Ordina target per valore IP
        targetList.sort((a, b) => a.targetValue.localeCompare(b.targetValue));

        const allReportsInFolder = targetList.flatMap((t) => t.reports);
        const totalAudits = allReportsInFolder.length;

        // Posture Score medio per l'azienda
        let avgScore = 100;
        let avgGrade = 'A';
        let maxSev: SeverityLevel = 'INFO';
        const sevPriority: Record<SeverityLevel, number> = {
          CRITICAL: 5,
          HIGH: 4,
          MEDIUM: 3,
          LOW: 2,
          INFO: 1,
        };

        if (totalAudits > 0) {
          const sumScore = allReportsInFolder.reduce((acc, r) => acc + r.posture_score, 0);
          avgScore = Math.round(sumScore / totalAudits);
          if (avgScore < 25) avgGrade = 'F';
          else if (avgScore < 50) avgGrade = 'D';
          else if (avgScore < 75) avgGrade = 'C';
          else if (avgScore < 90) avgGrade = 'B';

          let topSevScore = 0;
          for (const r of allReportsInFolder) {
            const curP = sevPriority[r.max_severity] || 1;
            if (curP > topSevScore) {
              topSevScore = curP;
              maxSev = r.max_severity;
            }
          }
        }

        return {
          id: f.id,
          name: f.name,
          color_tag: f.color_tag,
          targetList,
          totalAudits,
          avgScore,
          avgGrade,
          maxSeverity: maxSev,
        };
      })
      .filter((f) => f.targetList.length > 0 || f.totalAudits > 0);

    result.sort((a, b) => {
      if (a.id === 'unassigned') return 1;
      if (b.id === 'unassigned') return -1;
      return a.name.localeCompare(b.name);
    });

    return result;
  }, [groupsTree, reports]);

  // Gestione espansione accordion
  const toggleGroupCollapse = (groupId: string) => {
    setCollapsedGroupIds((prev) => {
      const next = new Set(prev);
      if (next.has(groupId)) {
        next.delete(groupId);
      } else {
        next.add(groupId);
      }
      return next;
    });
  };

  const expandAllGroups = () => {
    setCollapsedGroupIds(new Set());
  };

  const collapseAllGroups = () => {
    setCollapsedGroupIds(new Set(groupedCompanyReports.map((g) => g.id)));
  };

  // Formattatori
  const formatDate = (dateStr?: string): string => {
    if (!dateStr) return 'N/D';
    const d = new Date(dateStr);
    return isNaN(d.getTime()) ? 'N/D' : d.toLocaleString('it-IT', { dateStyle: 'short', timeStyle: 'short' });
  };

  const formatDuration = (seconds?: number): string => {
    if (seconds === undefined || seconds === null) return 'N/D';
    if (seconds < 60) return `${Math.round(seconds)}s`;
    const mins = Math.floor(seconds / 60);
    const secs = Math.round(seconds % 60);
    return `${mins}m ${secs}s`;
  };

  const getPostureColor = (score: number) => {
    if (score >= 90) return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30';
    if (score >= 75) return 'text-cyan-400 bg-cyan-500/10 border-cyan-500/30';
    if (score >= 50) return 'text-amber-400 bg-amber-500/10 border-amber-500/30';
    if (score >= 25) return 'text-orange-400 bg-orange-500/10 border-orange-500/30';
    return 'text-rose-400 bg-rose-500/10 border-rose-500/30';
  };

  const getSeverityBadge = (sev: SeverityLevel) => {
    switch (sev) {
      case 'CRITICAL':
        return 'bg-red-500/20 text-red-400 border-red-500/30';
      case 'HIGH':
        return 'bg-orange-500/20 text-orange-400 border-orange-500/30';
      case 'MEDIUM':
        return 'bg-amber-500/20 text-amber-400 border-amber-500/30';
      case 'LOW':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
      case 'INFO':
      default:
        return 'bg-slate-500/20 text-slate-400 border-slate-500/30';
    }
  };

  const getStatusBadge = (status: ScanStatus) => {
    switch (status) {
      case 'COMPLETED':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
      case 'RUNNING':
        return 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20 animate-pulse';
      case 'FAILED':
        return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
      case 'CANCELLED':
        return 'bg-slate-500/10 text-slate-400 border-slate-500/20';
      default:
        return 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20';
    }
  };

  const totalPages = Math.ceil(total / pageSize) || 1;

  return (
    <div className="space-y-6">
      {/* Intestazione Pagina */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center space-x-3">
          <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400 shadow-[0_0_15px_rgba(6,182,212,0.15)]">
            <FileText className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              <span>Hub Centralizzato Reportistica</span>
              <span className="text-xs px-2.5 py-0.5 rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 font-mono">
                White-Label
              </span>
            </h1>
            <p className="text-sm text-slate-400 mt-0.5">
              Archivio storico degli audit, download PDF (Executive & Tecnico) ed esportazioni CSV per azienda e singolo IP.
            </p>
          </div>
        </div>

        {/* Azioni Globali */}
        <div className="flex flex-wrap items-center gap-2.5">
          {/* Switcher Modalità: Vista per Azienda vs Archivio Lineare */}
          <div className="flex items-center bg-[#111726] border border-cyber-border rounded-lg p-1 text-xs">
            <button
              onClick={() => setViewMode('grouped')}
              className={`flex items-center space-x-1.5 px-2.5 py-1 rounded-md transition-colors ${
                viewMode === 'grouped'
                  ? 'bg-cyan-500/20 text-cyan-300 font-semibold border border-cyan-500/40 shadow-sm'
                  : 'text-slate-400 hover:text-white'
              }`}
              title="Organizza i report per Cartella / Azienda Cliente ed ispeziona ogni IP"
            >
              <Building className="w-3.5 h-3.5" />
              <span>Per Azienda</span>
            </button>
            <button
              onClick={() => setViewMode('flat')}
              className={`flex items-center space-x-1.5 px-2.5 py-1 rounded-md transition-colors ${
                viewMode === 'flat'
                  ? 'bg-cyan-500/20 text-cyan-300 font-semibold border border-cyan-500/40 shadow-sm'
                  : 'text-slate-400 hover:text-white'
              }`}
              title="Visualizzazione tabellare cronologica classica"
            >
              <List className="w-3.5 h-3.5" />
              <span>Cronologico</span>
            </button>
          </div>

          {/* Selettore Template Predefinito */}
          <div className="flex items-center bg-[#111726] border border-cyber-border rounded-lg p-1 text-xs">
            <span className="text-slate-400 px-2 font-medium hidden md:inline">PDF:</span>
            <button
              onClick={() => setPreferredTemplate('executive')}
              className={`px-2.5 py-1 rounded-md transition-colors ${
                preferredTemplate === 'executive'
                  ? 'bg-cyan-500/20 text-cyan-300 font-semibold border border-cyan-500/40'
                  : 'text-slate-400 hover:text-white'
              }`}
              title="Template manageriale sintetico con Posture Index, KRIs e priorità"
            >
              Executive
            </button>
            <button
              onClick={() => setPreferredTemplate('technical')}
              className={`px-2.5 py-1 rounded-md transition-colors ${
                preferredTemplate === 'technical'
                  ? 'bg-purple-500/20 text-purple-300 font-semibold border border-purple-500/40'
                  : 'text-slate-400 hover:text-white'
              }`}
              title="Template tecnico esteso con dettagli porte, censimento e raw banner"
            >
              Tecnico
            </button>
          </div>

          <button
            onClick={() => fetchData(false)}
            disabled={loading}
            className="flex items-center space-x-2 px-3 py-1.5 bg-[#111726] hover:bg-slate-800 border border-cyber-border rounded-lg text-slate-300 hover:text-white transition-all text-xs font-medium"
            title="Aggiorna elenco report"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
            <span className="hidden sm:inline">Aggiorna</span>
          </button>
        </div>
      </div>

      {/* Banner di notifica per download riuscito o fallito */}
      {downloadSuccessMsg && (
        <div className="p-3.5 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-sm flex items-center justify-between animate-in fade-in">
          <div className="flex items-center space-x-2">
            <ShieldCheck className="w-4 h-4 text-emerald-400 shrink-0" />
            <span>{downloadSuccessMsg}</span>
          </div>
          <button onClick={() => setDownloadSuccessMsg(null)} className="text-emerald-400 hover:text-white">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}
      {downloadErrorMsg && (
        <div className="p-3.5 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-300 text-sm flex items-center justify-between animate-in fade-in">
          <div className="flex items-center space-x-2">
            <AlertTriangle className="w-4 h-4 text-rose-400 shrink-0" />
            <span>{downloadErrorMsg}</span>
          </div>
          <button onClick={() => setDownloadErrorMsg(null)} className="text-rose-400 hover:text-white">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* KPI Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Totale Report */}
        <div className="p-4 rounded-xl border border-cyber-border bg-[#111726]/60 backdrop-blur-md flex items-center space-x-4">
          <div className="p-3 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
            <FileText className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xs font-medium text-slate-400">Report Disponibili</div>
            <div className="text-2xl font-bold text-white mt-0.5">{total}</div>
          </div>
        </div>

        {/* Postura Media */}
        <div className="p-4 rounded-xl border border-cyber-border bg-[#111726]/60 backdrop-blur-md flex items-center space-x-4">
          <div className={`p-3 rounded-lg border ${getPostureColor(kpiStats.avgScore)}`}>
            <ShieldCheck className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xs font-medium text-slate-400">Postura Media (Vista)</div>
            <div className="flex items-center space-x-2 mt-0.5">
              <span className="text-2xl font-bold text-white">{kpiStats.avgScore}</span>
              <span className="text-xs text-slate-400">/100</span>
              <span className={`text-xs px-2 py-0.5 rounded-md font-bold border ${getPostureColor(kpiStats.avgScore)}`}>
                GRADE {kpiStats.avgGrade}
              </span>
            </div>
          </div>
        </div>

        {/* Report ad Alto Rischio */}
        <div className="p-4 rounded-xl border border-cyber-border bg-[#111726]/60 backdrop-blur-md flex items-center space-x-4">
          <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400">
            <AlertTriangle className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xs font-medium text-slate-400">Audit ad Alto Rischio</div>
            <div className="text-2xl font-bold text-white mt-0.5">{kpiStats.highRiskCount}</div>
          </div>
        </div>

        {/* Host Monitorati */}
        <div className="p-4 rounded-xl border border-cyber-border bg-[#111726]/60 backdrop-blur-md flex items-center space-x-4">
          <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
            <Activity className="w-5 h-5" />
          </div>
          <div>
            <div className="text-xs font-medium text-slate-400">Host Operativi Censiti</div>
            <div className="text-2xl font-bold text-white mt-0.5">{kpiStats.totalUpHosts}</div>
          </div>
        </div>
      </div>

      {/* Toolbar Filtri Istantanei */}
      <div className="p-4 rounded-xl border border-cyber-border bg-[#111726]/70 backdrop-blur-md space-y-3">
        <div className="flex flex-col md:flex-row gap-3 items-stretch md:items-center">
          {/* Ricerca Testuale */}
          <div className="relative flex-1">
            <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              placeholder="Cerca per azienda, target IP, subnet CIDR o profilo..."
              className="w-full pl-9 pr-3 py-2 bg-[#0a0d14] border border-cyber-border rounded-lg text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500/50"
            />
            {search && (
              <button
                onClick={() => {
                  setSearch('');
                  setPage(1);
                }}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>

          {/* Filtro Severità Massima */}
          <div className="w-full sm:w-44">
            <select
              value={severityFilter}
              onChange={(e) => {
                setSeverityFilter(e.target.value);
                setPage(1);
              }}
              className="w-full px-3 py-2 bg-[#0a0d14] border border-cyber-border rounded-lg text-xs text-slate-300 focus:outline-none focus:border-cyan-500/50"
            >
              <option value="">Severità (Tutte)</option>
              <option value="CRITICAL">Solo Critiche</option>
              <option value="HIGH">Solo Alte</option>
              <option value="MEDIUM">Medie</option>
              <option value="LOW">Basse</option>
              <option value="INFO">Informative</option>
            </select>
          </div>

          {/* Filtro Tipologia Target */}
          <div className="w-full sm:w-36">
            <select
              value={targetTypeFilter}
              onChange={(e) => {
                setTargetTypeFilter(e.target.value);
                setPage(1);
              }}
              className="w-full px-3 py-2 bg-[#0a0d14] border border-cyber-border rounded-lg text-xs text-slate-300 focus:outline-none focus:border-cyan-500/50"
            >
              <option value="">Tipo Target (Tutti)</option>
              <option value="IP">IP Singolo</option>
              <option value="CIDR">Subnet CIDR</option>
              <option value="RANGE">Range IP</option>
              <option value="FQDN">Dominio FQDN</option>
            </select>
          </div>

          {/* Filtro Stato Scansione */}
          <div className="w-full sm:w-36">
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setPage(1);
              }}
              className="w-full px-3 py-2 bg-[#0a0d14] border border-cyber-border rounded-lg text-xs text-slate-300 focus:outline-none focus:border-cyan-500/50"
            >
              <option value="">Stato (Tutti)</option>
              <option value="COMPLETED">Completata</option>
              <option value="RUNNING">In Corso</option>
              <option value="FAILED">Fallita</option>
              <option value="CANCELLED">Annullata</option>
            </select>
          </div>
        </div>

        {/* Seconda riga: Date, controlli espansione ed eventuale Reset */}
        <div className="flex flex-wrap items-center justify-between gap-3 pt-2 border-t border-slate-800/80">
          <div className="flex flex-wrap items-center gap-3 text-xs text-slate-400">
            <div className="flex items-center space-x-1.5">
              <span>Dal:</span>
              <input
                type="date"
                value={dateFrom}
                onChange={(e) => {
                  setDateFrom(e.target.value);
                  setPage(1);
                }}
                className="px-2 py-1 bg-[#0a0d14] border border-cyber-border rounded text-slate-300 focus:outline-none focus:border-cyan-500/50"
              />
            </div>
            <div className="flex items-center space-x-1.5">
              <span>Al:</span>
              <input
                type="date"
                value={dateTo}
                onChange={(e) => {
                  setDateTo(e.target.value);
                  setPage(1);
                }}
                className="px-2 py-1 bg-[#0a0d14] border border-cyber-border rounded text-slate-300 focus:outline-none focus:border-cyan-500/50"
              />
            </div>
          </div>

          <div className="flex items-center space-x-3 text-xs">
            {viewMode === 'grouped' && groupedCompanyReports.length > 0 && (
              <div className="flex items-center space-x-2 text-slate-400">
                <button
                  type="button"
                  onClick={expandAllGroups}
                  className="hover:text-cyan-400 transition-colors"
                >
                  Espandi tutti
                </button>
                <span>·</span>
                <button
                  type="button"
                  onClick={collapseAllGroups}
                  className="hover:text-cyan-400 transition-colors"
                >
                  Comprimi tutti
                </button>
              </div>
            )}

            {activeFiltersCount > 0 && (
              <button
                onClick={handleResetFilters}
                className="flex items-center space-x-1 text-xs text-cyan-400 hover:text-cyan-300 underline font-medium"
              >
                <X className="w-3.5 h-3.5" />
                <span>Azzera {activeFiltersCount} filtri</span>
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Contenuto Principale */}
      {loading ? (
        <div className="p-16 text-center text-slate-400 flex flex-col items-center justify-center space-y-3 rounded-xl border border-cyber-border bg-[#111726]/60">
          <RefreshCw className="w-8 h-8 animate-spin text-cyan-400" />
          <p className="text-sm">Caricamento dell'archivio report di sicurezza...</p>
        </div>
      ) : error ? (
        <div className="p-12 text-center text-rose-400 flex flex-col items-center justify-center space-y-3 rounded-xl border border-cyber-border bg-[#111726]/60">
          <AlertCircle className="w-8 h-8 text-rose-500" />
          <p className="text-sm font-medium">{error}</p>
          <button
            onClick={() => fetchData(false)}
            className="px-4 py-1.5 bg-rose-500/20 hover:bg-rose-500/30 text-rose-300 border border-rose-500/40 rounded-lg text-xs"
          >
            Riprova
          </button>
        </div>
      ) : reports.length === 0 ? (
        <div className="p-16 text-center text-slate-400 flex flex-col items-center justify-center space-y-3 rounded-xl border border-cyber-border bg-[#111726]/60">
          <FileText className="w-10 h-10 text-slate-600" />
          <p className="text-base font-semibold text-white">Nessun report trovato</p>
          <p className="text-xs text-slate-500 max-w-sm">
            Non sono presenti report corrispondenti ai filtri impostati o non è ancora stata completata alcuna scansione.
          </p>
          {activeFiltersCount > 0 && (
            <button
              onClick={handleResetFilters}
              className="mt-2 px-3 py-1.5 bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-300 border border-cyan-500/30 rounded-lg text-xs transition-colors"
            >
              Azzera filtri
            </button>
          )}
        </div>
      ) : viewMode === 'grouped' ? (
        /* VISTA RAGGRUPPATA PER CARTELLA / AZIENDA CLIENTE */
        <div className="space-y-4">
          {groupedCompanyReports.map((company) => {
            const isExpanded = !collapsedGroupIds.has(company.id);

            return (
              <div
                key={company.id}
                className="rounded-xl border border-cyber-border/80 bg-[#111726]/75 shadow-lg overflow-hidden transition-all duration-200"
              >
                {/* Header Accordion Gruppo / Azienda */}
                <div
                  onClick={() => toggleGroupCollapse(company.id)}
                  className="px-5 py-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 cursor-pointer hover:bg-slate-800/40 transition-colors select-none"
                >
                  <div className="flex items-center space-x-3.5">
                    <div
                      className="p-2.5 rounded-lg border flex items-center justify-center"
                      style={{
                        backgroundColor: `${company.color_tag}15`,
                        borderColor: `${company.color_tag}40`,
                        color: company.color_tag,
                      }}
                    >
                      {isExpanded ? <FolderOpen className="w-5 h-5" /> : <Folder className="w-5 h-5" />}
                    </div>

                    <div>
                      <div className="flex items-center space-x-2.5">
                        <span className="font-bold text-white text-base tracking-wide">{company.name}</span>
                        <span className="text-[10px] px-2 py-0.5 rounded-full font-mono bg-slate-900 border border-slate-700 text-slate-300">
                          {company.targetList.length} {company.targetList.length === 1 ? 'target' : 'target'}
                        </span>
                      </div>
                      <div className="flex items-center space-x-2 text-xs text-slate-400 mt-0.5">
                        <span className="text-cyan-400 font-semibold">{company.totalAudits} report totali</span>
                        <span>·</span>
                        <span>Scope cliente isolato</span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center space-x-3 sm:self-center">
                    {/* Posture Score medio azienda */}
                    <div className="flex items-center space-x-2">
                      <span className="text-xs text-slate-400 hidden md:inline">Posture Score:</span>
                      <div className={`px-2.5 py-1 rounded-lg border font-mono text-xs font-bold ${getPostureColor(company.avgScore)}`}>
                        <span>{company.avgScore}/100</span>
                        <span className="ml-1 px-1 rounded bg-slate-950/40 text-[10px]">GRADE {company.avgGrade}</span>
                      </div>
                    </div>

                    {/* Max Severity Badge */}
                    <span className={`text-[10px] px-2 py-0.5 rounded font-bold border ${getSeverityBadge(company.maxSeverity)}`}>
                      MAX: {company.maxSeverity}
                    </span>

                    {/* Freccia espansione */}
                    <button
                      type="button"
                      className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors ml-1"
                      aria-label={isExpanded ? 'Comprimi gruppo' : 'Espandi gruppo'}
                    >
                      {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                    </button>
                  </div>
                </div>

                {/* Contenuto Accordion: Tabella Target IP e Report White-Label dedicati */}
                {isExpanded && (
                  <div className="border-t border-cyber-border/60 bg-[#0c101a]/90">
                    <div className="overflow-x-auto">
                      <table className="w-full text-left border-collapse text-xs">
                        <thead>
                          <tr className="border-b border-cyber-border/60 bg-[#0d121f]/90 text-slate-400 uppercase tracking-wider text-[11px] font-semibold">
                            <th className="py-2.5 px-4">Target IP / Valore</th>
                            <th className="py-2.5 px-4">Profilo Scansione</th>
                            <th className="py-2.5 px-4 text-center">Security Posture</th>
                            <th className="py-2.5 px-4">Vulnerabilità (Findings)</th>
                            <th className="py-2.5 px-4 text-center">Host UP</th>
                            <th className="py-2.5 px-4">Data Ultimo Audit</th>
                            <th className="py-2.5 px-4 text-center">Stato</th>
                            <th className="py-2.5 px-4 text-right">Download White-Label</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-cyber-border/30">
                          {company.targetList.map((target) => {
                            const rep = target.latestReport;
                            if (!rep) return null;
                            const isCurrentItemDownloading = downloadingId === rep.scan_id;

                            return (
                              <tr
                                key={target.targetId}
                                className="hover:bg-slate-800/25 transition-colors group"
                              >
                                {/* Target IP e pulsante copia */}
                                <td className="py-3 px-4">
                                  <div className="flex items-center space-x-2">
                                    <span className="font-mono font-bold text-white group-hover:text-cyan-300 transition-colors">
                                      {target.targetValue}
                                    </span>
                                    <span className="text-[10px] px-1.5 py-0.2 rounded bg-slate-800 text-slate-300 border border-slate-700 uppercase">
                                      {target.targetType}
                                    </span>
                                    <button
                                      type="button"
                                      onClick={() => handleCopy(target.targetValue)}
                                      className="text-slate-500 hover:text-slate-300 p-0.5 rounded transition-colors"
                                      title="Copia valore IP"
                                    >
                                      {copiedIP === target.targetValue ? (
                                        <Check className="w-3.5 h-3.5 text-emerald-400" />
                                      ) : (
                                        <Copy className="w-3.5 h-3.5" />
                                      )}
                                    </button>
                                  </div>
                                  {target.reports.length > 1 && (
                                    <div className="text-[10px] text-slate-400 font-mono mt-0.5">
                                      {target.reports.length} report archiviati per questo IP
                                    </div>
                                  )}
                                </td>

                                {/* Profilo & Trigger */}
                                <td className="py-3 px-4">
                                  <div className="flex flex-col">
                                    <span className="text-slate-200 font-medium">{rep.profile_name}</span>
                                    <span className="text-[10px] text-slate-500 uppercase mt-0.5">
                                      {rep.trigger_type === 'SCHEDULED' ? 'Pianificata' : 'Manuale'}
                                    </span>
                                  </div>
                                </td>

                                {/* Posture Score */}
                                <td className="py-3 px-4 text-center">
                                  <div className="inline-flex items-center space-x-2 px-2 py-0.5 rounded-lg border font-mono font-semibold">
                                    <span className={getPostureColor(rep.posture_score).split(' ')[0]}>
                                      {rep.posture_score}
                                    </span>
                                    <span className="text-slate-500">/</span>
                                    <span className={`text-[10px] px-1 rounded ${getPostureColor(rep.posture_score)}`}>
                                      {rep.posture_grade}
                                    </span>
                                  </div>
                                </td>

                                {/* Findings */}
                                <td className="py-3 px-4">
                                  <div className="flex flex-col space-y-1">
                                    <div className="flex items-center space-x-1.5">
                                      <span className={`text-[10px] px-2 py-0.5 rounded font-bold border ${getSeverityBadge(rep.max_severity)}`}>
                                        MAX: {rep.max_severity}
                                      </span>
                                    </div>
                                    <div className="flex items-center space-x-2 text-[10px] text-slate-400 font-mono">
                                      {rep.summary_stats.critical_count > 0 && (
                                        <span className="text-red-400 font-bold">{rep.summary_stats.critical_count}C</span>
                                      )}
                                      {rep.summary_stats.high_count > 0 && (
                                        <span className="text-orange-400 font-bold">{rep.summary_stats.high_count}H</span>
                                      )}
                                      {rep.summary_stats.medium_count > 0 && (
                                        <span className="text-amber-400">{rep.summary_stats.medium_count}M</span>
                                      )}
                                      {rep.summary_stats.low_count > 0 && (
                                        <span className="text-blue-400">{rep.summary_stats.low_count}L</span>
                                      )}
                                      {rep.summary_stats.critical_count === 0 &&
                                        rep.summary_stats.high_count === 0 &&
                                        rep.summary_stats.medium_count === 0 &&
                                        rep.summary_stats.low_count === 0 && (
                                          <span className="text-emerald-400">Pulito (0 vulns)</span>
                                        )}
                                    </div>
                                  </div>
                                </td>

                                {/* Host Monitorati */}
                                <td className="py-3 px-4 text-center">
                                  <div className="font-mono text-slate-300">
                                    {rep.summary_stats.up_hosts || 1}
                                    <span className="text-slate-500 text-[10px]">
                                      {rep.summary_stats.total_hosts > 1 ? ` / ${rep.summary_stats.total_hosts}` : ' host'} UP
                                    </span>
                                  </div>
                                </td>

                                {/* Data & Durata */}
                                <td className="py-3 px-4">
                                  <div className="flex flex-col text-slate-300">
                                    <span>{formatDate(rep.completed_at || rep.created_at)}</span>
                                    <span className="text-[10px] text-slate-500 flex items-center gap-1 mt-0.5">
                                      <Clock className="w-3 h-3" />
                                      {formatDuration(rep.summary_stats.duration_seconds)}
                                    </span>
                                  </div>
                                </td>

                                {/* Stato */}
                                <td className="py-3 px-4 text-center">
                                  <span className={`text-[10px] px-2 py-0.5 rounded-full font-medium border ${getStatusBadge(rep.status)}`}>
                                    {rep.status}
                                  </span>
                                </td>

                                {/* Azioni di Esportazione 1-Click White-Label */}
                                <td className="py-3 px-4 text-right">
                                  <div className="flex items-center justify-end space-x-1.5">
                                    {/* Scarica PDF Executive */}
                                    <button
                                      onClick={() => handleDownload(rep.scan_id, 'pdf', 'executive')}
                                      disabled={isCurrentItemDownloading || rep.status !== 'COMPLETED'}
                                      className={`p-1.5 rounded-lg border transition-all text-cyan-400 bg-cyan-500/10 border-cyan-500/30 hover:bg-cyan-500/20 ${
                                        isCurrentItemDownloading && downloadingFormat === 'pdf_executive' ? 'animate-pulse' : ''
                                      } disabled:opacity-40 disabled:cursor-not-allowed`}
                                      title="Scarica Report PDF Executive per questo IP (Sintesi per Leadership)"
                                    >
                                      <FileText className="w-3.5 h-3.5" />
                                    </button>

                                    {/* Scarica PDF Tecnico */}
                                    <button
                                      onClick={() => handleDownload(rep.scan_id, 'pdf', 'technical')}
                                      disabled={isCurrentItemDownloading || rep.status !== 'COMPLETED'}
                                      className={`p-1.5 rounded-lg border transition-all text-purple-400 bg-purple-500/10 border-purple-500/30 hover:bg-purple-500/20 ${
                                        isCurrentItemDownloading && downloadingFormat === 'pdf_technical' ? 'animate-pulse' : ''
                                      } disabled:opacity-40 disabled:cursor-not-allowed`}
                                      title="Scarica Report PDF Tecnico Deep-Dive per questo IP (Porte & Banner)"
                                    >
                                      <FileCode className="w-3.5 h-3.5" />
                                    </button>

                                    {/* Scarica CSV */}
                                    <button
                                      onClick={() => handleDownload(rep.scan_id, 'csv')}
                                      disabled={isCurrentItemDownloading || rep.status !== 'COMPLETED'}
                                      className={`p-1.5 rounded-lg border transition-all text-emerald-400 bg-emerald-500/10 border-emerald-500/30 hover:bg-emerald-500/20 ${
                                        isCurrentItemDownloading && downloadingFormat?.startsWith('csv') ? 'animate-pulse' : ''
                                      } disabled:opacity-40 disabled:cursor-not-allowed`}
                                      title="Esporta Finding di questo IP in CSV Tabellare (RFC 4180)"
                                    >
                                      <FileSpreadsheet className="w-3.5 h-3.5" />
                                    </button>

                                    {/* Dettaglio Scansione */}
                                    <Link
                                      to={`/scans/${rep.scan_id}`}
                                      className="p-1.5 rounded-lg border border-slate-700 bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 transition-colors"
                                      title="Apri Dettaglio Scansione Operativa"
                                    >
                                      <ExternalLink className="w-3.5 h-3.5" />
                                    </Link>
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
            );
          })}
        </div>
      ) : (
        /* VISTA ARCHIVIO CRONOLOGICO LINEARE (FLAT TABLE) */
        <div className="rounded-xl border border-cyber-border bg-[#111726]/60 backdrop-blur-md overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="border-b border-cyber-border bg-[#0d121f]/80 text-slate-400 font-semibold uppercase tracking-wider">
                  <th className="py-3 px-4">Target / Scope</th>
                  <th className="py-3 px-4">Profilo Scansione</th>
                  <th className="py-3 px-4 text-center">Security Posture</th>
                  <th className="py-3 px-4">Vulnerabilità (Findings)</th>
                  <th className="py-3 px-4 text-center">Host Monitorati</th>
                  <th className="py-3 px-4">Data Audit</th>
                  <th className="py-3 px-4 text-center">Stato</th>
                  <th className="py-3 px-4 text-right">Download White-Label</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-cyber-border/40">
                {reports.map((item) => {
                  const isCurrentItemDownloading = downloadingId === item.scan_id;

                  return (
                    <tr
                      key={item.scan_id}
                      className="hover:bg-slate-800/30 transition-colors group"
                    >
                      {/* Target & Cartella */}
                      <td className="py-3.5 px-4">
                        <div className="flex flex-col">
                          <div className="flex items-center space-x-2">
                            <span className="font-mono font-medium text-white group-hover:text-cyan-300 transition-colors">
                              {item.target_value}
                            </span>
                            <span className="text-[10px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
                              {item.target_type}
                            </span>
                          </div>
                          {item.group_name && (
                            <div className="flex items-center space-x-1 text-[11px] text-slate-400 mt-1">
                              <Folder className="w-3 h-3 text-cyan-400" />
                              <span>{item.group_name}</span>
                            </div>
                          )}
                        </div>
                      </td>

                      {/* Profilo & Trigger */}
                      <td className="py-3.5 px-4">
                        <div className="flex flex-col">
                          <span className="text-slate-200 font-medium">{item.profile_name}</span>
                          <span className="text-[10px] text-slate-500 uppercase mt-0.5">
                            {item.trigger_type === 'SCHEDULED' ? 'Pianificata (Cron)' : 'Manuale On-Demand'}
                          </span>
                        </div>
                      </td>

                      {/* Posture Score Badge */}
                      <td className="py-3.5 px-4 text-center">
                        <div className="inline-flex items-center space-x-2 px-2.5 py-1 rounded-lg border font-mono font-semibold">
                          <span className={getPostureColor(item.posture_score).split(' ')[0]}>
                            {item.posture_score}
                          </span>
                          <span className="text-slate-500">/</span>
                          <span className={`text-[10px] px-1 rounded ${getPostureColor(item.posture_score)}`}>
                            {item.posture_grade}
                          </span>
                        </div>
                      </td>

                      {/* Vulnerabilità e Max Severity */}
                      <td className="py-3.5 px-4">
                        <div className="flex flex-col space-y-1">
                          <div className="flex items-center space-x-1.5">
                            <span className={`text-[10px] px-2 py-0.5 rounded font-bold border ${getSeverityBadge(item.max_severity)}`}>
                              MAX: {item.max_severity}
                            </span>
                          </div>
                          {/* Conteggi sintetici */}
                          <div className="flex items-center space-x-2 text-[10px] text-slate-400 font-mono">
                            {item.summary_stats.critical_count > 0 && (
                              <span className="text-red-400 font-bold">{item.summary_stats.critical_count}C</span>
                            )}
                            {item.summary_stats.high_count > 0 && (
                              <span className="text-orange-400 font-bold">{item.summary_stats.high_count}H</span>
                            )}
                            {item.summary_stats.medium_count > 0 && (
                              <span className="text-amber-400">{item.summary_stats.medium_count}M</span>
                            )}
                            {item.summary_stats.low_count > 0 && (
                              <span className="text-blue-400">{item.summary_stats.low_count}L</span>
                            )}
                            {item.summary_stats.critical_count === 0 &&
                              item.summary_stats.high_count === 0 &&
                              item.summary_stats.medium_count === 0 &&
                              item.summary_stats.low_count === 0 && (
                                <span className="text-emerald-400">Pulito (0 vulns)</span>
                              )}
                          </div>
                        </div>
                      </td>

                      {/* Host Monitorati */}
                      <td className="py-3.5 px-4 text-center">
                        <div className="font-mono text-slate-300">
                          {item.summary_stats.up_hosts || 1}
                          <span className="text-slate-500 text-[10px]">
                            {item.summary_stats.total_hosts > 1 ? ` / ${item.summary_stats.total_hosts}` : ' host'} UP
                          </span>
                        </div>
                      </td>

                      {/* Data & Durata */}
                      <td className="py-3.5 px-4">
                        <div className="flex flex-col text-slate-300">
                          <span>{formatDate(item.completed_at || item.created_at)}</span>
                          <span className="text-[10px] text-slate-500 flex items-center gap-1 mt-0.5">
                            <Clock className="w-3 h-3" />
                            {formatDuration(item.summary_stats.duration_seconds)}
                          </span>
                        </div>
                      </td>

                      {/* Stato Scansione */}
                      <td className="py-3.5 px-4 text-center">
                        <span className={`text-[10px] px-2 py-0.5 rounded-full font-medium border ${getStatusBadge(item.status)}`}>
                          {item.status}
                        </span>
                      </td>

                      {/* Azioni di Esportazione 1-Click */}
                      <td className="py-3.5 px-4 text-right">
                        <div className="flex items-center justify-end space-x-1.5">
                          {/* Scarica PDF Executive */}
                          <button
                            onClick={() => handleDownload(item.scan_id, 'pdf', 'executive')}
                            disabled={isCurrentItemDownloading || item.status !== 'COMPLETED'}
                            className={`p-1.5 rounded-lg border transition-all text-cyan-400 bg-cyan-500/10 border-cyan-500/30 hover:bg-cyan-500/20 ${
                              isCurrentItemDownloading && downloadingFormat === 'pdf_executive' ? 'animate-pulse' : ''
                            } disabled:opacity-40 disabled:cursor-not-allowed`}
                            title="Scarica Report PDF Executive (Sintesi per Leadership)"
                          >
                            <FileText className="w-3.5 h-3.5" />
                          </button>

                          {/* Scarica PDF Tecnico */}
                          <button
                            onClick={() => handleDownload(item.scan_id, 'pdf', 'technical')}
                            disabled={isCurrentItemDownloading || item.status !== 'COMPLETED'}
                            className={`p-1.5 rounded-lg border transition-all text-purple-400 bg-purple-500/10 border-purple-500/30 hover:bg-purple-500/20 ${
                              isCurrentItemDownloading && downloadingFormat === 'pdf_technical' ? 'animate-pulse' : ''
                            } disabled:opacity-40 disabled:cursor-not-allowed`}
                            title="Scarica Report PDF Tecnico Deep-Dive (Dettaglio Porte & Banner)"
                          >
                            <FileCode className="w-3.5 h-3.5" />
                          </button>

                          {/* Scarica CSV */}
                          <button
                            onClick={() => handleDownload(item.scan_id, 'csv')}
                            disabled={isCurrentItemDownloading || item.status !== 'COMPLETED'}
                            className={`p-1.5 rounded-lg border transition-all text-emerald-400 bg-emerald-500/10 border-emerald-500/30 hover:bg-emerald-500/20 ${
                              isCurrentItemDownloading && downloadingFormat?.startsWith('csv') ? 'animate-pulse' : ''
                            } disabled:opacity-40 disabled:cursor-not-allowed`}
                            title="Esporta Finding in CSV Tabellare (RFC 4180)"
                          >
                            <FileSpreadsheet className="w-3.5 h-3.5" />
                          </button>

                          {/* Dettaglio Scansione */}
                          <Link
                            to={`/scans/${item.scan_id}`}
                            className="p-1.5 rounded-lg border border-slate-700 bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 transition-colors"
                            title="Apri Dettaglio Scansione Operativa"
                          >
                            <ExternalLink className="w-3.5 h-3.5" />
                          </Link>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Paginazione */}
          {total > 0 && (
            <div className="p-3.5 border-t border-cyber-border bg-[#0d121f]/80 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-400">
              <div className="flex items-center space-x-2">
                <span>Righe per pagina:</span>
                <select
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                  }}
                  className="bg-[#0a0d14] border border-cyber-border rounded px-2 py-0.5 text-slate-300 focus:outline-none"
                >
                  <option value={10}>10</option>
                  <option value={20}>20</option>
                  <option value={50}>50</option>
                </select>
                <span>
                  Visualizzati {(page - 1) * pageSize + 1} - {Math.min(page * pageSize, total)} di {total} report
                </span>
              </div>

              <div className="flex items-center space-x-1.5">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="p-1 rounded border border-cyber-border bg-[#0a0d14] text-slate-300 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-slate-800"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <span className="px-2 font-mono">
                  Pagina {page} di {totalPages}
                </span>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                  className="p-1 rounded border border-cyber-border bg-[#0a0d14] text-slate-300 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-slate-800"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
