import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
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
  ChevronDown,
  ChevronUp,
  AlertCircle,
  Copy,
  Check,
  Building,
  Layers,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { reportsApi } from '../api/reports';
import { ApiResponse } from '../types/auth';
import { TargetGroupNode } from '../types/target';
import { ReportItem, ReportTemplate, ReportFormat } from '../types/report';
import { filterReportArchive, loadReportArchive } from '../utils/reportArchive';
import { useCompany } from '../context/CompanyContext';

export const Reports: React.FC = () => {
  const {
    selectedCompanyId,
    selectedCompany,
    departments,
    companyGroupIds,
  } = useCompany();

  // Modalità di visualizzazione: Raggruppata per Azienda (default) vs Archivio Cronologico Lineare
  const [viewMode, setViewMode] = useState<'grouped' | 'flat'>('grouped');

  // Stato dati report e aziende
  const [allReports, setAllReports] = useState<ReportItem[]>([]);
  const requestId = useRef(0);
  const [groupsTree, setGroupsTree] = useState<TargetGroupNode[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Stato filtri di ricerca
  const [search, setSearch] = useState<string>('');
  const [severityFilter, setSeverityFilter] = useState<string>('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [profileFilter, setProfileFilter] = useState('');
  const [targetFilter, setTargetFilter] = useState('');
  const [groupFilter, setGroupFilter] = useState('');
  const [archivePage, setArchivePage] = useState(0);
  const scopedReports = useMemo(() => allReports.filter(item => !selectedCompanyId || (!!item.group_id && companyGroupIds.has(item.group_id))), [allReports, selectedCompanyId, companyGroupIds]);
  const reports = useMemo(() => filterReportArchive(scopedReports, {
    dateFrom, dateTo, profile: profileFilter, target: targetFilter, group: groupFilter,
    severity: severityFilter, search,
  }), [scopedReports, dateFrom, dateTo, profileFilter, targetFilter, groupFilter, severityFilter, search]);
  const filtersActive = Boolean(search.trim() || severityFilter || dateFrom || dateTo || profileFilter || targetFilter || groupFilter);
  const invalidPeriod = Boolean(dateFrom && dateTo && dateFrom > dateTo);
  useEffect(() => { setArchivePage(0); }, [search, severityFilter, dateFrom, dateTo, profileFilter, targetFilter, groupFilter, selectedCompanyId]);
  useEffect(() => { setProfileFilter(''); setTargetFilter(''); setGroupFilter(''); }, [selectedCompanyId]);
  const visiblePage = Math.min(archivePage, Math.max(0, Math.ceil(reports.length / 10) - 1));

  // Accordion stato gruppi: Set degli ID collassati (vuoto = tutti espansi)
  const [collapsedGroupIds, setCollapsedGroupIds] = useState<Set<string>>(new Set());

  // Copia IP feedback
  const [copiedIP, setCopiedIP] = useState<string | null>(null);

  // Stato operazioni di download attive
  const [downloadingId, setDownloadingId] = useState<string | null>(null);
  const [downloadSuccessMsg, setDownloadSuccessMsg] = useState<string | null>(null);
  const [downloadErrorMsg, setDownloadErrorMsg] = useState<string | null>(null);

  // Caricamento dati dal backend (report + metadati cartelle/aziende)
  const fetchData = useCallback(async (silent = false) => {
    const currentRequest = ++requestId.current;
    const isCurrent = () => requestId.current === currentRequest;
    if (!silent) setLoading(true);
    setError(null);

    const isGrouped = viewMode === 'grouped';
    const effectiveLimit = isGrouped ? 100 : 10;

    try {
      const [reportsData, groupsRes] = await Promise.all([
        loadReportArchive(params => reportsApi.list(params), effectiveLimit, isCurrent),
        apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups'),
      ]);
      if (!isCurrent()) return;
      if (!groupsRes.data.success || !Array.isArray(groupsRes.data.data)) {
        throw new Error('Impossibile caricare le cartelle dello storico.');
      }
      setAllReports(reportsData);

      if (groupsRes?.data?.data) {
        setGroupsTree(groupsRes.data.data);
      }
    } catch (err: any) {
      if (!isCurrent()) return;
      setError(err.message || 'Errore durante il recupero dei report.');
      setAllReports([]);
    } finally {
      if (isCurrent()) setLoading(false);
    }
  }, [viewMode]);

  useEffect(() => {
    fetchData();
    return () => { requestId.current++; };
  }, [fetchData]);

  // Gestione del download del Documento Consolidato Aziendale (tutti gli IP dell'azienda)
  const handleCompanyDownload = async (
    groupId: string,
    companyName: string,
    format: ReportFormat,
    template: ReportTemplate = 'executive'
  ) => {
    setDownloadingId(groupId);
    setDownloadSuccessMsg(null);
    setDownloadErrorMsg(null);

    try {
      await reportsApi.downloadCompany(groupId, companyName, format, template, {
        onSuccess: () => {
          setDownloadSuccessMsg(
            `Documento di ${companyName} scaricato con successo (${format.toUpperCase()} ${format === 'pdf' ? template : ''})`
          );
          setTimeout(() => setDownloadSuccessMsg(null), 4000);
        },
        onError: (err) => {
          setDownloadErrorMsg(err.message || `Download del report di ${companyName} fallito.`);
          setTimeout(() => setDownloadErrorMsg(null), 5000);
        },
      });
    } catch (err: any) {
      setDownloadErrorMsg(err.message || 'Download non riuscito.');
      setTimeout(() => setDownloadErrorMsg(null), 5000);
    } finally {
      setDownloadingId(null);
    }
  };

  // Copia IP
  const handleCopy = (val: string) => {
    navigator.clipboard.writeText(val);
    setCopiedIP(val);
    setTimeout(() => setCopiedIP(null), 2000);
  };

  // Reset filtri
  const handleResetFilters = () => {
    setSearch('');
    setSeverityFilter('');
  };

  // Accordion toggle per azienda
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

  // Espandi tutti / Comprimi tutti
  const handleExpandAll = () => setCollapsedGroupIds(new Set());
  const handleCollapseAll = () => {
    const allIds = new Set<string>();
    function recurse(nodes: TargetGroupNode[]) {
      for (const n of nodes) {
        allIds.add(n.id);
        if (n.children && n.children.length > 0) recurse(n.children);
      }
    }
    recurse(groupsTree);
    allIds.add('unassigned');
    setCollapsedGroupIds(allIds);
  };

  // Aggregazione e preparazione delle Aziende Clienti con i rispettivi report consolidati
  const companyDataList = useMemo(() => {
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

    // Mappa groupId -> lista di report
    const groupReportsMap = new Map<string, ReportItem[]>();
    reports.forEach((r) => {
      const gId = r.group_id || 'unassigned';
      if (!groupReportsMap.has(gId)) {
        groupReportsMap.set(gId, []);
      }
      groupReportsMap.get(gId)!.push(r);
    });

    const result: Array<{
      id: string;
      name: string;
      description?: string | null;
      color_tag: string;
      targets: Array<{ targetId: string; targetValue: string; targetType: string }>;
      reports: ReportItem[];
      totalTargets: number;
      postureScore: number;
      postureGrade: string;
      criticalCount: number;
      highCount: number;
      mediumCount: number;
      lowCount: number;
      infoCount: number;
      latestCompletedAt?: string;
    }> = [];

    // Processa tutti i gruppi/aziende registrati
    gMap.forEach((gNode, gId) => {
      const companyReports = groupReportsMap.get(gId) || [];
      const targetsMap = new Map<string, { targetId: string; targetValue: string; targetType: string }>();

      let crit = 0;
      let high = 0;
      let med = 0;
      let low = 0;
      let info = 0;
      let latestTime: number | null = null;
      let totalScore = 0;

      companyReports.forEach((r) => {
        if (!targetsMap.has(r.target_id)) {
          targetsMap.set(r.target_id, {
            targetId: r.target_id,
            targetValue: r.target_value,
            targetType: r.target_type,
          });
        }
        crit += r.summary_stats.critical_count || 0;
        high += r.summary_stats.high_count || 0;
        med += r.summary_stats.medium_count || 0;
        low += r.summary_stats.low_count || 0;
        info += r.summary_stats.info_count || 0;
        totalScore += r.posture_score;

        const t = new Date(r.completed_at || r.created_at).getTime();
        if (latestTime === null || t > latestTime) {
          latestTime = t;
        }
      });

      // Se non ci sono report ma gNode ha conteggi dal backend
      if (companyReports.length === 0 && gNode.severity_counts && !filtersActive) {
        crit = gNode.severity_counts['CRITICAL'] || 0;
        high = gNode.severity_counts['HIGH'] || 0;
        med = gNode.severity_counts['MEDIUM'] || 0;
        low = gNode.severity_counts['LOW'] || 0;
        info = gNode.severity_counts['INFO'] || 0;
      }

      // Calcola Posture Score & Grade per l'Azienda
      let postureScore = companyReports.length > 0 ? Math.round(totalScore / companyReports.length) : 100;
      if (companyReports.length === 0) {
        const deduction = crit * 25 + high * 15 + med * 5 + low * 1;
        postureScore = Math.max(0, 100 - deduction);
      }

      let postureGrade = 'A';
      if (postureScore < 25) postureGrade = 'F';
      else if (postureScore < 50) postureGrade = 'D';
      else if (postureScore < 75) postureGrade = 'C';
      else if (postureScore < 90) postureGrade = 'B';

      result.push({
        id: gId,
        name: gNode.name,
        description: gNode.description,
        color_tag: gNode.color_tag || '#06b6d4',
        targets: Array.from(targetsMap.values()),
        reports: companyReports,
        totalTargets: Math.max(gNode.direct_target_count || 0, targetsMap.size),
        postureScore,
        postureGrade,
        criticalCount: crit,
        highCount: high,
        mediumCount: med,
        lowCount: low,
        infoCount: info,
        latestCompletedAt: latestTime ? new Date(latestTime).toISOString() : undefined,
      });
    });

    // Filtro per scope attivo
    const scopedResult = selectedCompanyId
      ? result.filter((comp) => companyGroupIds.has(comp.id))
      : result.filter((comp) => !gMap.get(comp.id)?.parent_id);

    // Filtra per ricerca e severità
    return scopedResult.filter((comp) => {
      if (filtersActive) return comp.reports.length > 0;
      if (search.trim()) {
        const q = search.toLowerCase();
        const matchName = comp.name.toLowerCase().includes(q);
        const matchDesc = comp.description ? comp.description.toLowerCase().includes(q) : false;
        const matchIP = comp.targets.some((t) => t.targetValue.toLowerCase().includes(q));
        if (!matchName && !matchDesc && !matchIP) return false;
      }

      if (severityFilter) {
        if (severityFilter === 'CRITICAL' && comp.criticalCount === 0) return false;
        if (severityFilter === 'HIGH' && comp.highCount === 0 && comp.criticalCount === 0) return false;
        if (severityFilter === 'MEDIUM' && comp.mediumCount === 0 && comp.highCount === 0 && comp.criticalCount === 0) return false;
      }

      return true;
    });
  }, [groupsTree, reports, search, severityFilter, selectedCompanyId, companyGroupIds, filtersActive]);

  // Metriche sintetiche aggregate delle aziende
  const kpiStats = useMemo(() => {
    if (companyDataList.length === 0) {
      return { totalCompanies: 0, avgScore: 100, avgGrade: 'A', totalTargets: 0, highRiskCompanies: 0 };
    }

    let sumScore = 0;
    let totalT = 0;
    let highRisk = 0;

    companyDataList.forEach((c) => {
      sumScore += c.postureScore;
      totalT += c.totalTargets;
      if (c.criticalCount > 0 || c.highCount > 0) {
        highRisk++;
      }
    });

    const avgScore = Math.round(sumScore / companyDataList.length);
    let avgGrade = 'A';
    if (avgScore < 25) avgGrade = 'F';
    else if (avgScore < 50) avgGrade = 'D';
    else if (avgScore < 75) avgGrade = 'C';
    else if (avgScore < 90) avgGrade = 'B';

    return {
      totalCompanies: companyDataList.length,
      avgScore,
      avgGrade,
      totalTargets: totalT,
      highRiskCompanies: highRisk,
    };
  }, [companyDataList]);

  const getGradeBadge = (grade: string) => {
    switch (grade) {
      case 'A':
        return 'bg-emerald-500/15 text-emerald-400 border-emerald-500/40';
      case 'B':
        return 'bg-cyan-500/15 text-cyan-400 border-cyan-500/40';
      case 'C':
        return 'bg-amber-500/15 text-amber-400 border-amber-500/40';
      case 'D':
        return 'bg-orange-500/15 text-orange-400 border-orange-500/40';
      case 'F':
      default:
        return 'bg-rose-500/15 text-rose-400 border-rose-500/40';
    }
  };

  return (
    <div className="space-y-6">
      {/* Toast Feedback Notifiche */}
      {downloadSuccessMsg && (
        <div className="p-3.5 rounded-xl border border-emerald-500/40 bg-emerald-950/40 text-emerald-200 text-xs flex items-center justify-between shadow-lg animate-in fade-in">
          <div className="flex items-center space-x-2">
            <Check className="w-4 h-4 text-emerald-400 shrink-0" />
            <span className="font-medium">{downloadSuccessMsg}</span>
          </div>
          <button
            onClick={() => setDownloadSuccessMsg(null)}
            className="text-emerald-400 hover:text-white"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {downloadErrorMsg && (
        <div className="p-3.5 rounded-xl border border-rose-800/50 bg-rose-950/40 text-rose-300 text-xs flex items-center justify-between shadow-lg animate-in fade-in">
          <div className="flex items-center space-x-2">
            <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />
            <span>{downloadErrorMsg}</span>
          </div>
          <button
            onClick={() => setDownloadErrorMsg(null)}
            className="text-rose-400 hover:text-white"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Header Pagina */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Building className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-2xl font-bold tracking-tight text-white flex items-center space-x-2.5">
                <span>Dossier & Report Aziendali Consolidati</span>
                <span className="text-xs px-2.5 py-0.5 rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 font-medium">
                  Per Azienda
                </span>
              </h2>
              <p className="text-sm text-slate-400 mt-0.5">
                Generazione white-label consolidata per Azienda Cliente. Ogni documento racchiude l'audit completo di tutti gli IP e nodi del perimetro.
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-2 self-start md:self-auto">
          {/* Selettore modalità */}
          <div className="p-1 rounded-xl bg-slate-900 border border-cyber-border flex items-center space-x-1 text-xs">
            <button
              onClick={() => setViewMode('grouped')}
              className={`px-3 py-1.5 rounded-lg font-medium transition-all flex items-center space-x-1.5 ${
                viewMode === 'grouped'
                  ? 'bg-cyan-500 text-slate-950 shadow-[0_0_12px_rgba(6,182,212,0.3)]'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <Building className="w-3.5 h-3.5" />
              <span>Per Azienda</span>
            </button>
            <button
              onClick={() => setViewMode('flat')}
              className={`px-3 py-1.5 rounded-lg font-medium transition-all flex items-center space-x-1.5 ${
                viewMode === 'flat'
                  ? 'bg-cyan-500 text-slate-950 shadow-[0_0_12px_rgba(6,182,212,0.3)]'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <Clock className="w-3.5 h-3.5" />
              <span>Cronologico</span>
            </button>
          </div>

          <button
            onClick={() => fetchData()}
            disabled={loading}
            className="p-2 rounded-xl bg-slate-900 border border-cyber-border text-slate-400 hover:text-cyan-400 hover:border-cyan-500/40 transition-colors disabled:opacity-50"
            title="Ricarica report"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Banner Scope Aziendale Attivo */}
      {selectedCompany && (
        <div className="p-4 rounded-xl border border-cyan-500/30 bg-gradient-to-r from-cyan-950/40 to-slate-900/60 backdrop-blur flex flex-col sm:flex-row sm:items-center justify-between gap-4 shadow-lg">
          <div className="flex items-center space-x-3.5">
            <div
              className="p-2.5 rounded-xl border shrink-0"
              style={{
                backgroundColor: `${selectedCompany.color_tag || '#06b6d4'}20`,
                borderColor: `${selectedCompany.color_tag || '#06b6d4'}40`,
                color: selectedCompany.color_tag || '#06b6d4',
              }}
            >
              <Building className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="text-[11px] font-mono uppercase tracking-wider text-cyan-400 font-semibold">
                  Dossier Consolidato Scope Attivo
                </span>
                <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
                  {departments.length} {departments.length === 1 ? 'Reparto' : 'Reparti'}
                </span>
              </div>
              <h3 className="text-base font-bold text-white">{selectedCompany.name}</h3>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <button
              onClick={() => handleCompanyDownload(selectedCompany.id, selectedCompany.name, 'pdf', 'executive')}
              disabled={downloadingId === selectedCompany.id}
              className="px-3 py-1.5 rounded-lg border border-cyan-500/40 bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-300 text-xs font-semibold transition-colors flex items-center space-x-1.5"
            >
              <FileText className="w-3.5 h-3.5" />
              <span>Executive PDF</span>
            </button>
            <button
              onClick={() => handleCompanyDownload(selectedCompany.id, selectedCompany.name, 'pdf', 'technical')}
              disabled={downloadingId === selectedCompany.id}
              className="px-3 py-1.5 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium transition-colors flex items-center space-x-1.5"
            >
              <FileCode className="w-3.5 h-3.5" />
              <span>Tecnico PDF</span>
            </button>
            <button
              onClick={() => handleCompanyDownload(selectedCompany.id, selectedCompany.name, 'csv')}
              disabled={downloadingId === selectedCompany.id}
              className="px-3 py-1.5 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium transition-colors flex items-center space-x-1.5"
            >
              <FileSpreadsheet className="w-3.5 h-3.5" />
              <span>CSV</span>
            </button>
          </div>
        </div>
      )}

      {/* KPI Cards di Riepilogo Posture */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-4 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Aziende Monitorate</span>
            <Building className="w-4 h-4 text-cyan-400" />
          </div>
          <div className="mt-2 text-2xl font-bold font-mono text-white">
            {kpiStats.totalCompanies}
          </div>
          <div className="mt-1 text-[11px] text-slate-500">Perimetri e cartelle censite</div>
        </div>

        <div className="p-4 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Posture Score Medio</span>
            <ShieldCheck className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="mt-2 flex items-baseline space-x-2">
            <span className="text-2xl font-bold font-mono text-white">{kpiStats.avgScore}/100</span>
            <span className={`text-xs px-2 py-0.5 rounded font-mono font-bold border ${getGradeBadge(kpiStats.avgGrade)}`}>
              GRADE {kpiStats.avgGrade}
            </span>
          </div>
          <div className="mt-1 text-[11px] text-slate-500">Media ponderata su tutti i clienti</div>
        </div>

        <div className="p-4 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">IP Totali Censiti</span>
            <Layers className="w-4 h-4 text-indigo-400" />
          </div>
          <div className="mt-2 text-2xl font-bold font-mono text-indigo-300">
            {kpiStats.totalTargets}
          </div>
          <div className="mt-1 text-[11px] text-slate-500">Bersagli compresi nei report</div>
        </div>

        <div className="p-4 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400">Aziende ad Alto Rischio</span>
            <AlertTriangle className="w-4 h-4 text-rose-400" />
          </div>
          <div className="mt-2 text-2xl font-bold font-mono text-rose-400">
            {kpiStats.highRiskCompanies}
          </div>
          <div className="mt-1 text-[11px] text-slate-500">Con vulnerabilità Critiche o Alte</div>
        </div>
      </div>

      {/* Toolbar Filtri & Ricerca */}
      <div className="p-3.5 rounded-xl border border-cyber-border bg-cyber-card/60 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="relative flex-1 max-w-md">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-500">
            <Search className="w-4 h-4" />
          </div>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cerca per Azienda, descrizione o IP..."
            className="w-full pl-9 pr-8 py-2 bg-slate-950 border border-cyber-border rounded-lg text-xs text-slate-200 placeholder-slate-500 focus:outline-hidden focus:border-cyan-500"
          />
          {search && (
            <button
              onClick={() => setSearch('')}
              className="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-500 hover:text-slate-300"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>

        <div className="flex items-center space-x-2">
          <select
            aria-label="Severità massima"
            value={severityFilter}
            onChange={(e) => setSeverityFilter(e.target.value)}
            className="px-3 py-2 bg-slate-950 border border-cyber-border rounded-lg text-xs text-slate-200 focus:outline-hidden focus:border-cyan-500"
          >
            <option value="">Tutte le severità</option>
            <option value="CRITICAL">Critica</option>
            <option value="HIGH">Alta</option>
            <option value="MEDIUM">Media</option>
            <option value="LOW">Bassa</option>
            <option value="INFO">Informativa</option>
          </select>

          {viewMode === 'grouped' && (
            <div className="flex items-center space-x-1 pl-2 border-l border-cyber-border">
              <button
                onClick={handleExpandAll}
                className="px-2 py-1.5 rounded bg-slate-900 border border-cyber-border text-[11px] text-slate-400 hover:text-white"
                title="Espandi tutte le aziende"
              >
                Espandi tutti
              </button>
              <button
                onClick={handleCollapseAll}
                className="px-2 py-1.5 rounded bg-slate-900 border border-cyber-border text-[11px] text-slate-400 hover:text-white"
                title="Comprimi tutte le aziende"
              >
                Comprimi tutti
              </button>
            </div>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 text-xs">
        <label className="text-slate-300">Dal giorno (UTC)
          <input aria-label="Dal giorno (UTC)" type="date" value={dateFrom} onChange={e => setDateFrom(e.target.value)} className="block w-full mt-1 p-2 bg-slate-950 border border-cyber-border rounded" />
        </label>
        <label className="text-slate-300">Al giorno (UTC)
          <input aria-label="Al giorno (UTC)" type="date" value={dateTo} onChange={e => setDateTo(e.target.value)} className="block w-full mt-1 p-2 bg-slate-950 border border-cyber-border rounded" />
        </label>
        <label className="text-slate-300">Profilo
          <select aria-label="Profilo" value={profileFilter} onChange={e => setProfileFilter(e.target.value)} className="block w-full mt-1 p-2 bg-slate-950 border border-cyber-border rounded">
            <option value="">Tutti i profili</option>
            {Array.from(new Map(scopedReports.map(r => [r.profile_id, r.profile_name]))).map(([id, name]) => <option key={id} value={id}>{name}</option>)}
          </select>
        </label>
        <label className="text-slate-300">Target
          <select aria-label="Target" value={targetFilter} onChange={e => setTargetFilter(e.target.value)} className="block w-full mt-1 p-2 bg-slate-950 border border-cyber-border rounded">
            <option value="">Tutti i target</option>
            {Array.from(new Map(scopedReports.map(r => [r.target_id, `${r.target_value} · ${r.group_name || 'Non assegnato'}`]))).map(([id, name]) => <option key={id} value={id}>{name}</option>)}
          </select>
        </label>
        <label className="text-slate-300">Cartella
          <select aria-label="Cartella" value={groupFilter} onChange={e => setGroupFilter(e.target.value)} className="block w-full mt-1 p-2 bg-slate-950 border border-cyber-border rounded">
            <option value="">Tutte le cartelle</option>
            {Array.from(new Map(scopedReports.filter(r => r.group_id).map(r => [r.group_id!, r.group_name || r.group_id!]))).map(([id, name]) => <option key={id} value={id}>{name}</option>)}
          </select>
        </label>
      </div>
      <p className="text-xs text-slate-400">I filtri selezionano lo storico. I download aziendali includono il dossier completo dell’azienda o reparto.</p>
      {invalidPeriod && <p role="alert" className="text-sm text-rose-400">La data iniziale deve precedere o coincidere con quella finale.</p>}
      {filtersActive && <button type="button" className="text-xs text-cyan-400" onClick={() => {
        setSearch(''); setSeverityFilter(''); setDateFrom(''); setDateTo(''); setProfileFilter(''); setTargetFilter(''); setGroupFilter('');
      }}>Azzera filtri</button>}

      {/* Loading & Error States */}
      {loading && (
        <div className="p-12 text-center text-slate-400 flex flex-col items-center justify-center space-y-3">
          <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
          <span className="text-xs">Caricamento dossier e report aziendali...</span>
        </div>
      )}

      {error && !loading && (
        <div className="p-6 rounded-xl border border-rose-800/40 bg-rose-950/20 text-center space-y-3">
          <AlertCircle className="w-8 h-8 text-rose-400 mx-auto" />
          <div className="text-sm font-semibold text-rose-200">Impossibile caricare i report</div>
          <div className="text-xs text-rose-300/80">{error}</div>
          <button
            onClick={() => fetchData()}
            className="px-3 py-1.5 rounded-lg bg-rose-900/40 hover:bg-rose-900 text-rose-200 text-xs font-medium"
          >
            Riprova
          </button>
        </div>
      )}

      {/* Vista Raggruppata per Azienda Cliente (Default, Consolidato) */}
      {!loading && !error && viewMode === 'grouped' && (
        <div className="space-y-4">
          {companyDataList.length === 0 ? (
            <div className="p-12 rounded-2xl border border-cyber-border bg-cyber-card/40 text-center space-y-3">
              <Building className="w-10 h-10 text-slate-600 mx-auto" />
              <div className="text-sm font-semibold text-white">Nessuna Azienda trovata</div>
              <p className="text-xs text-slate-400 max-w-md mx-auto">
                Non sono presenti aziende o report che corrispondono ai filtri selezionati. Crea una nuova cartella/azienda in Target o reimposta i filtri.
              </p>
              <button
                onClick={handleResetFilters}
                className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs"
              >
                Reimposta filtri
              </button>
            </div>
          ) : (
            companyDataList.map((company) => {
              const isCollapsed = collapsedGroupIds.has(company.id);
              const isDownloadingThis = downloadingId === company.id;

              return (
                <div
                  key={company.id}
                  className="rounded-2xl border border-cyber-border bg-cyber-card/70 overflow-hidden shadow-xl transition-all"
                >
                  {/* Header Card Azienda Cliente */}
                  <div className="p-5 border-b border-cyber-border flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-slate-900/40">
                    <div
                      onClick={() => toggleGroupCollapse(company.id)}
                      className="flex items-start space-x-3.5 cursor-pointer flex-1"
                    >
                      <div
                        className="p-2.5 rounded-xl border mt-0.5 shrink-0"
                        style={{
                          backgroundColor: `${company.color_tag}20`,
                          borderColor: `${company.color_tag}50`,
                          color: company.color_tag,
                        }}
                      >
                        <Building className="w-5 h-5" />
                      </div>
                      <div>
                        <div className="flex items-center space-x-2.5 flex-wrap">
                          <h3 className="text-base font-bold text-white hover:text-cyan-300 transition-colors">
                            {company.name}
                          </h3>
                          <span
                            className="px-2 py-0.5 rounded text-[11px] font-medium border"
                            style={{
                              backgroundColor: `${company.color_tag}15`,
                              borderColor: `${company.color_tag}40`,
                              color: company.color_tag,
                            }}
                          >
                            {company.totalTargets} IP censiti
                          </span>
                        </div>
                        {company.description && (
                          <p className="text-xs text-slate-400 mt-1 line-clamp-1">
                            {company.description}
                          </p>
                        )}
                        <div className="flex items-center space-x-4 mt-2 text-[11px] text-slate-500">
                          {company.latestCompletedAt ? (
                            <span className="flex items-center space-x-1">
                              <Clock className="w-3 h-3 text-slate-400" />
                              <span>
                                Ultimo audit: {new Date(company.latestCompletedAt).toLocaleDateString('it-IT', {
                                  day: '2-digit',
                                  month: '2-digit',
                                  year: 'numeric',
                                  hour: '2-digit',
                                  minute: '2-digit',
                                })}
                              </span>
                            </span>
                          ) : (
                            <span className="italic text-slate-500">In attesa di primo audit di sicurezza</span>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Posture Score & Azioni di Download Consolidato per Azienda */}
                    <div className="flex flex-col sm:flex-row sm:items-center gap-4 shrink-0">
                      {/* Posture Badge Aziendale */}
                      <div className="flex items-center space-x-2.5 sm:px-3 sm:py-1.5 sm:rounded-xl sm:bg-slate-950 sm:border sm:border-cyber-border">
                        <div className="text-right">
                          <div className="text-[10px] text-slate-400 font-semibold uppercase">Security Posture</div>
                          <div className="text-sm font-bold font-mono text-white">
                            {company.postureScore}/100
                          </div>
                        </div>
                        <span
                          className={`px-2 py-0.5 rounded text-xs font-mono font-bold border ${getGradeBadge(
                            company.postureGrade
                          )}`}
                        >
                          GRADE {company.postureGrade}
                        </span>
                      </div>

                      {/* 1-Click Documento Consolidato Aziendale */}
                      <div className="flex items-center space-x-2">
                        {/* Executive PDF */}
                        <button
                          onClick={() => handleCompanyDownload(company.id, company.name, 'pdf', 'executive')}
                          disabled={isDownloadingThis}
                          className="px-3 py-1.5 rounded-lg border border-cyan-500/30 bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-300 font-medium text-xs flex items-center space-x-1.5 transition-all disabled:opacity-50"
                          title="Scarica Report PDF Executive per l'intera Azienda"
                        >
                          <FileText className="w-3.5 h-3.5 text-cyan-400" />
                          <span>Executive PDF</span>
                        </button>

                        {/* Tecnico PDF */}
                        <button
                          onClick={() => handleCompanyDownload(company.id, company.name, 'pdf', 'technical')}
                          disabled={isDownloadingThis}
                          className="px-3 py-1.5 rounded-lg border border-purple-500/30 bg-purple-500/10 hover:bg-purple-500/20 text-purple-300 font-medium text-xs flex items-center space-x-1.5 transition-all disabled:opacity-50"
                          title="Scarica Report PDF Tecnico per l'intera Azienda"
                        >
                          <FileCode className="w-3.5 h-3.5 text-purple-400" />
                          <span>Tecnico PDF</span>
                        </button>

                        {/* CSV */}
                        <button
                          onClick={() => handleCompanyDownload(company.id, company.name, 'csv', 'executive')}
                          disabled={isDownloadingThis}
                          className="px-3 py-1.5 rounded-lg border border-emerald-500/30 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-300 font-medium text-xs flex items-center space-x-1.5 transition-all disabled:opacity-50"
                          title="Scarica Evidenze CSV per l'intera Azienda"
                        >
                          <FileSpreadsheet className="w-3.5 h-3.5 text-emerald-400" />
                          <span>CSV</span>
                        </button>

                        {/* Chevron Collapse */}
                        <button
                          onClick={() => toggleGroupCollapse(company.id)}
                          className="p-1.5 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-white transition-colors ml-1"
                          title={isCollapsed ? 'Espandi dettagli perimetro' : 'Comprimi perimetro'}
                        >
                          {isCollapsed ? <ChevronDown className="w-4 h-4" /> : <ChevronUp className="w-4 h-4" />}
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Dettagli Espansi: Elenco IP inclusi nel Report Aziendale & Finding Summary */}
                  {!isCollapsed && (
                    <div className="p-5 space-y-4 animate-in fade-in bg-slate-950/40">
                      {/* Breakdown Vulnerabilità dell'Azienda */}
                      <div className="flex flex-wrap items-center justify-between gap-3 p-3 rounded-xl border border-cyber-border bg-slate-900/60 text-xs">
                        <div className="flex items-center space-x-2">
                          <AlertTriangle className="w-4 h-4 text-amber-400" />
                          <span className="font-semibold text-slate-200">
                            Vulnerabilità Rilevate nel Perimetro Aziendale:
                          </span>
                        </div>
                        <div className="flex items-center space-x-2 font-mono text-[11px]">
                          <span className="px-2 py-0.5 rounded bg-rose-500/20 text-rose-300 border border-rose-500/40">
                            {company.criticalCount} Critiche
                          </span>
                          <span className="px-2 py-0.5 rounded bg-orange-500/20 text-orange-300 border border-orange-500/40">
                            {company.highCount} Alte
                          </span>
                          <span className="px-2 py-0.5 rounded bg-amber-500/20 text-amber-300 border border-amber-500/40">
                            {company.mediumCount} Medie
                          </span>
                          <span className="px-2 py-0.5 rounded bg-blue-500/20 text-blue-300 border border-blue-500/40">
                            {company.lowCount} Basse
                          </span>
                        </div>
                      </div>

                      {/* Lista IP inclusi nel documento dell'azienda */}
                      <div>
                        <div className="text-xs font-semibold text-slate-300 mb-2.5 flex items-center justify-between">
                          <span>Bersagli e Indirizzi IP inclusi nel Documento Aziendale:</span>
                          <span className="text-[11px] text-slate-500">
                            (Il documento PDF e CSV racchiude tutti questi IP)
                          </span>
                        </div>

                        {company.targets.length === 0 ? (
                          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/40 text-center text-xs text-slate-400">
                            Nessun IP con scansioni storiche. Assegna bersagli a questa azienda nella sezione Target.
                          </div>
                        ) : (
                          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-2.5">
                            {company.targets.map((tgt) => (
                              <div
                                key={tgt.targetId}
                                className="p-2.5 rounded-xl border border-cyber-border bg-slate-900/50 flex items-center justify-between text-xs hover:border-slate-700 transition-colors"
                              >
                                <div className="flex items-center space-x-2">
                                  <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-bold bg-slate-800 text-cyan-400 border border-slate-700">
                                    {tgt.targetType}
                                  </span>
                                  <span className="font-mono font-medium text-slate-100">
                                    {tgt.targetValue}
                                  </span>
                                </div>
                                <button
                                  onClick={() => handleCopy(tgt.targetValue)}
                                  className="p-1 rounded text-slate-500 hover:text-slate-300"
                                  title="Copia IP"
                                >
                                  {copiedIP === tgt.targetValue ? (
                                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                                  ) : (
                                    <Copy className="w-3.5 h-3.5" />
                                  )}
                                </button>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              );
            })
          )}
        </div>
      )}

      {/* Vista Cronologica Lineare (Audit Runs Storiche) */}
      {!loading && !error && viewMode === 'flat' && (
        <div className="rounded-2xl border border-cyber-border bg-cyber-card/70 overflow-hidden shadow-xl">
          <div className="p-4 border-b border-cyber-border bg-slate-900/40 flex items-center justify-between">
            <div className="text-xs font-semibold text-slate-200">
              Storico Cronologico Esecuzioni di Scansione
            </div>
            <div className="text-xs text-slate-400">
              {reports.length} report registrati
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="border-b border-cyber-border bg-slate-950/80 text-slate-400 font-mono text-[11px] uppercase tracking-wider">
                  <th className="py-3 px-4">Azienda Scope</th>
                  <th className="py-3 px-4">Bersaglio / IP</th>
                  <th className="py-3 px-4">Profilo Audit</th>
                  <th className="py-3 px-4">Data Esecuzione</th>
                  <th className="py-3 px-4">Posture Score</th>
                  <th className="py-3 px-4 text-right">Download Dossier</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-cyber-border/40">
                {reports.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="py-8 text-center text-slate-500 italic">
                      Nessun report presente nello storico.
                    </td>
                  </tr>
                ) : (
                  reports.slice(visiblePage * 10, (visiblePage + 1) * 10).map((r) => (
                    <tr key={r.scan_id} className="hover:bg-slate-900/50 transition-colors">
                      <td className="py-3 px-4 font-medium text-white">
                        {r.group_name || 'Non Assegnato'}
                      </td>
                      <td className="py-3 px-4 font-mono text-cyan-400">
                        {r.target_value}
                      </td>
                      <td className="py-3 px-4 text-slate-300">
                        {r.profile_name}
                      </td>
                      <td className="py-3 px-4 text-slate-400">
                        {new Date(r.completed_at || r.created_at).toLocaleDateString('it-IT', {
                          day: '2-digit',
                          month: '2-digit',
                          year: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-0.5 rounded text-xs font-mono font-bold border ${getGradeBadge(r.posture_grade)}`}>
                          {r.posture_score}/100 ({r.posture_grade})
                        </span>
                      </td>
                      <td className="py-3 px-4 text-right">
                        {r.group_id ? (
                          <div className="flex items-center justify-end space-x-1.5">
                            <button
                              onClick={() => handleCompanyDownload(r.group_id!, r.group_name || 'Azienda', 'pdf', 'executive')}
                              className="p-1.5 rounded hover:bg-slate-800 text-cyan-400 hover:text-cyan-300"
                              title="Scarica Executive PDF Aziendale"
                            >
                              <FileText className="w-3.5 h-3.5" />
                            </button>
                            <button
                              onClick={() => handleCompanyDownload(r.group_id!, r.group_name || 'Azienda', 'pdf', 'technical')}
                              className="p-1.5 rounded hover:bg-slate-800 text-purple-400 hover:text-purple-300"
                              title="Scarica Tecnico PDF Aziendale"
                            >
                              <FileCode className="w-3.5 h-3.5" />
                            </button>
                            <button
                              onClick={() => handleCompanyDownload(r.group_id!, r.group_name || 'Azienda', 'csv', 'executive')}
                              className="p-1.5 rounded hover:bg-slate-800 text-emerald-400 hover:text-emerald-300"
                              title="Scarica CSV Aziendale"
                            >
                              <FileSpreadsheet className="w-3.5 h-3.5" />
                            </button>
                          </div>
                        ) : (
                          <span className="text-slate-500 text-[11px] italic">Non associato ad azienda</span>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
          <div className="p-3 flex items-center justify-between gap-3 text-xs text-slate-300">
            <button type="button" disabled={visiblePage === 0} onClick={() => setArchivePage(visiblePage - 1)} className="disabled:opacity-40">Pagina precedente</button>
            <span>Pagina {visiblePage + 1} di {Math.max(1, Math.ceil(reports.length / 10))}</span>
            <button type="button" disabled={(visiblePage + 1) * 10 >= reports.length} onClick={() => setArchivePage(visiblePage + 1)} className="disabled:opacity-40">Pagina successiva</button>
          </div>
        </div>
      )}
    </div>
  );
};
