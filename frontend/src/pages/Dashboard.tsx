import React, { useEffect, useState, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  Building,
  Crosshair,
  Radio,
  Terminal,
  ShieldAlert,
  RefreshCw,
  ArrowRight,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  Layers,
  FileText,
  Copy,
  Check,
  Search,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetGroupNode, TargetItem, TargetListResponseData } from '../types/target';
import { ReportListResponseData, ReportItem } from '../types/report';
import { ScanListResponseData, ScanItem } from '../types/scan';

export const Dashboard: React.FC = () => {

  const [groups, setGroups] = useState<TargetGroupNode[]>([]);
  const [targets, setTargets] = useState<TargetItem[]>([]);
  const [reports, setReports] = useState<ReportItem[]>([]);
  const [recentScans, setRecentScans] = useState<ScanItem[]>([]);

  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [copiedIP, setCopiedIP] = useState<string | null>(null);
  const [companySearch, setCompanySearch] = useState<string>('');

  const fetchDashboardData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [groupsRes, targetsRes, reportsRes, scansRes] = await Promise.all([
        apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups').catch(() => null),
        apiClient.get<ApiResponse<TargetListResponseData>>('/targets?limit=200').catch(() => null),
        apiClient.get<ApiResponse<ReportListResponseData>>('/reports?limit=100').catch(() => null),
        apiClient.get<ApiResponse<ScanListResponseData>>('/scans?limit=50').catch(() => null),
      ]);

      if (groupsRes?.data?.data) {
        setGroups(groupsRes.data.data);
      }
      if (targetsRes?.data?.data) {
        setTargets(targetsRes.data.data.items || []);
      }
      if (reportsRes?.data?.data) {
        setReports(reportsRes.data.data.items || []);
      }
      if (scansRes?.data?.data) {
        setRecentScans(scansRes.data.data.items || []);
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

  const handleCopy = (ip: string) => {
    navigator.clipboard.writeText(ip);
    setCopiedIP(ip);
    setTimeout(() => setCopiedIP(null), 2000);
  };

  // Mappatura delle singole aziende con metriche aggregate dedicate
  const companyCards = useMemo(() => {
    const flatGroupsMap = new Map<string, TargetGroupNode>();
    function recurse(nodes: TargetGroupNode[]) {
      for (const n of nodes) {
        flatGroupsMap.set(n.id, n);
        if (n.children && n.children.length > 0) {
          recurse(n.children);
        }
      }
    }
    recurse(groups);

    // Mappa target per group
    const targetsByGroup = new Map<string, TargetItem[]>();
    const unassignedTargets: TargetItem[] = [];

    targets.forEach((t) => {
      if (t.group_id) {
        if (!targetsByGroup.has(t.group_id)) {
          targetsByGroup.set(t.group_id, []);
        }
        targetsByGroup.get(t.group_id)!.push(t);
      } else {
        unassignedTargets.push(t);
      }
    });

    // Mappa report per group
    const reportsByGroup = new Map<string, ReportItem[]>();
    reports.forEach((r) => {
      if (r.group_id) {
        if (!reportsByGroup.has(r.group_id)) {
          reportsByGroup.set(r.group_id, []);
        }
        reportsByGroup.get(r.group_id)!.push(r);
      }
    });

    // Mappa scansioni per target
    const scansByTarget = new Map<string, ScanItem[]>();
    recentScans.forEach((s) => {
      if (!scansByTarget.has(s.target_id)) {
        scansByTarget.set(s.target_id, []);
      }
      scansByTarget.get(s.target_id)!.push(s);
    });

    const result: Array<{
      id: string;
      name: string;
      description?: string | null;
      color_tag: string;
      targets: TargetItem[];
      criticalCount: number;
      highCount: number;
      mediumCount: number;
      lowCount: number;
      postureScore: number;
      postureGrade: string;
      latestScan?: ScanItem;
    }> = [];

    flatGroupsMap.forEach((gNode, gId) => {
      const gTargets = targetsByGroup.get(gId) || [];
      const gReports = reportsByGroup.get(gId) || [];

      let crit = 0;
      let high = 0;
      let med = 0;
      let low = 0;

      gReports.forEach((r) => {
        crit += r.summary_stats.critical_count || 0;
        high += r.summary_stats.high_count || 0;
        med += r.summary_stats.medium_count || 0;
        low += r.summary_stats.low_count || 0;
      });

      if (gReports.length === 0 && gNode.severity_counts) {
        crit = gNode.severity_counts['CRITICAL'] || 0;
        high = gNode.severity_counts['HIGH'] || 0;
        med = gNode.severity_counts['MEDIUM'] || 0;
        low = gNode.severity_counts['LOW'] || 0;
      }

      // Calcolo Posture Score e Grade
      const deduction = crit * 25 + high * 15 + med * 5 + low * 1;
      let postureScore = Math.max(0, 100 - deduction);
      if (gReports.length > 0) {
        const avgR = Math.round(gReports.reduce((acc, r) => acc + r.posture_score, 0) / gReports.length);
        postureScore = avgR;
      }

      let postureGrade = 'A';
      if (postureScore < 25) postureGrade = 'F';
      else if (postureScore < 50) postureGrade = 'D';
      else if (postureScore < 75) postureGrade = 'C';
      else if (postureScore < 90) postureGrade = 'B';

      // Ultima scansione per uno qualsiasi dei target dell'azienda
      let latestScan: ScanItem | undefined;
      let latestTime = 0;
      gTargets.forEach((t) => {
        const tScans = scansByTarget.get(t.id) || [];
        tScans.forEach((s) => {
          const tVal = new Date(s.completed_at || s.created_at).getTime();
          if (tVal > latestTime) {
            latestTime = tVal;
            latestScan = s;
          }
        });
      });

      result.push({
        id: gId,
        name: gNode.name,
        description: gNode.description,
        color_tag: gNode.color_tag || '#06b6d4',
        targets: gTargets,
        criticalCount: crit,
        highCount: high,
        mediumCount: med,
        lowCount: low,
        postureScore,
        postureGrade,
        latestScan,
      });
    });

    return {
      companies: result.filter((c) =>
        companySearch.trim()
          ? c.name.toLowerCase().includes(companySearch.toLowerCase()) ||
            (c.description && c.description.toLowerCase().includes(companySearch.toLowerCase())) ||
            c.targets.some((t) => t.value.toLowerCase().includes(companySearch.toLowerCase()))
          : true
      ),
      unassignedTargets,
    };
  }, [groups, targets, reports, recentScans, companySearch]);

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

  const getStatusBadge = (status?: string) => {
    if (!status) {
      return (
        <span className="px-2 py-0.5 rounded text-[11px] font-medium bg-slate-800 text-slate-400 border border-slate-700">
          In attesa di primo audit
        </span>
      );
    }
    switch (status) {
      case 'COMPLETED':
        return (
          <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
            <CheckCircle2 className="w-3 h-3" />
            <span>Audit Completato</span>
          </span>
        );
      case 'RUNNING':
        return (
          <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded text-[11px] font-medium bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 animate-pulse">
            <Radio className="w-3 h-3 animate-spin" />
            <span>Audit In Corso</span>
          </span>
        );
      case 'FAILED':
        return (
          <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded text-[11px] font-medium bg-rose-500/10 text-rose-400 border border-rose-500/30">
            <AlertCircle className="w-3 h-3" />
            <span>Fallito</span>
          </span>
        );
      default:
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-medium bg-slate-800 text-slate-300">
            {status}
          </span>
        );
    }
  };

  return (
    <div className="space-y-8">
      {/* Header Panoramica */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Building className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-2xl font-bold tracking-tight text-white flex items-center space-x-2.5">
                <span>Panoramica Sicurezza per Azienda</span>
                <span className="text-xs px-2.5 py-0.5 rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 font-medium">
                  MSSP Multi-Company Posture
                </span>
              </h2>
              <p className="text-sm text-slate-400 mt-0.5">
                Cruscotto operativo per il monitoraggio della postura di sicurezza e conformità delle singole aziende clienti gestite.
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-2 self-start sm:self-auto">
          <button
            onClick={fetchDashboardData}
            disabled={loading}
            className="flex items-center space-x-2 px-3 py-1.5 rounded-xl border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white transition-colors text-xs font-medium disabled:opacity-50"
            title="Aggiorna dati"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
            <span>{loading ? 'Caricamento...' : 'Aggiorna'}</span>
          </button>
        </div>
      </div>

      {/* Guida al Flusso Operativo Lineare (Nessuna confusione o ridondanza) */}
      <div className="p-4 rounded-2xl border border-cyber-border bg-slate-950/80 shadow-lg space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold text-slate-200 uppercase tracking-wider flex items-center space-x-1.5">
            <Layers className="w-4 h-4 text-cyan-400" />
            <span>Flusso Operativo Lineare della Piattaforma</span>
          </span>
          <span className="text-[11px] text-slate-500">
            5 passaggi lineari: Discovery → Target → Scansioni → Vulnerabilità → Report
          </span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-5 gap-3 text-xs">
          {/* Step 1 */}
          <Link
            to="/discovery"
            className="p-3 rounded-xl border border-cyber-border bg-slate-900/60 hover:border-cyan-500/50 hover:bg-cyan-500/5 transition-all group"
          >
            <div className="flex items-center justify-between text-slate-400 group-hover:text-cyan-400 font-mono text-[11px]">
              <span>PASSO 1</span>
              <Radio className="w-3.5 h-3.5" />
            </div>
            <div className="font-bold text-white mt-1 group-hover:text-cyan-300">Network Discovery</div>
            <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
              Scansiona la subnet LAN e rileva gli IP attivi nella rete locale.
            </p>
          </Link>

          {/* Step 2 */}
          <Link
            to="/targets"
            className="p-3 rounded-xl border border-cyber-border bg-slate-900/60 hover:border-cyan-500/50 hover:bg-cyan-500/5 transition-all group"
          >
            <div className="flex items-center justify-between text-slate-400 group-hover:text-cyan-400 font-mono text-[11px]">
              <span>PASSO 2</span>
              <Crosshair className="w-3.5 h-3.5" />
            </div>
            <div className="font-bold text-white mt-1 group-hover:text-cyan-300">Target & Aziende</div>
            <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
              Assegna gli IP alle rispettive Aziende e gestisci le blacklist.
            </p>
          </Link>

          {/* Step 3 */}
          <Link
            to="/scans"
            className="p-3 rounded-xl border border-cyber-border bg-slate-900/60 hover:border-cyan-500/50 hover:bg-cyan-500/5 transition-all group"
          >
            <div className="flex items-center justify-between text-slate-400 group-hover:text-cyan-400 font-mono text-[11px]">
              <span>PASSO 3</span>
              <Terminal className="w-3.5 h-3.5" />
            </div>
            <div className="font-bold text-white mt-1 group-hover:text-cyan-300">Scansioni Nmap</div>
            <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
              Esegui audit di sicurezza (All-in-One 360°, Web, Falle, Pen Test).
            </p>
          </Link>

          {/* Step 4 */}
          <Link
            to="/findings"
            className="p-3 rounded-xl border border-cyber-border bg-slate-900/60 hover:border-cyan-500/50 hover:bg-cyan-500/5 transition-all group"
          >
            <div className="flex items-center justify-between text-slate-400 group-hover:text-cyan-400 font-mono text-[11px]">
              <span>PASSO 4</span>
              <ShieldAlert className="w-3.5 h-3.5" />
            </div>
            <div className="font-bold text-white mt-1 group-hover:text-cyan-300">Vulnerabilità CVE</div>
            <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
              Analizza le falle scoperte per azienda e verifica la remediation.
            </p>
          </Link>

          {/* Step 5 */}
          <Link
            to="/reports"
            className="p-3 rounded-xl border border-cyber-border bg-slate-900/60 hover:border-cyan-500/50 hover:bg-cyan-500/5 transition-all group"
          >
            <div className="flex items-center justify-between text-slate-400 group-hover:text-cyan-400 font-mono text-[11px]">
              <span>PASSO 5</span>
              <FileText className="w-3.5 h-3.5" />
            </div>
            <div className="font-bold text-white mt-1 group-hover:text-cyan-300">Report Aziendali</div>
            <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
              Scarica il dossier consolidato (Executive PDF, Tecnico PDF, CSV).
            </p>
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

      {/* Ricerca e Filtro Aziende */}
      <div className="flex items-center justify-between gap-4">
        <div className="relative flex-1 max-w-sm">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-500">
            <Search className="w-4 h-4" />
          </div>
          <input
            type="text"
            value={companySearch}
            onChange={(e) => setCompanySearch(e.target.value)}
            placeholder="Filtra aziende o IP perimetrali..."
            className="w-full pl-9 pr-3 py-2 bg-slate-950 border border-cyber-border rounded-lg text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
          />
        </div>

        <div className="text-xs text-slate-400">
          Mostrando <strong className="text-white">{companyCards.companies.length}</strong> aziende clienti
        </div>
      </div>

      {/* Griglia Card delle Singole Aziende */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        {companyCards.companies.length === 0 ? (
          <div className="col-span-full p-12 rounded-2xl border border-cyber-border bg-cyber-card/40 text-center space-y-3">
            <Building className="w-10 h-10 text-slate-600 mx-auto" />
            <div className="text-sm font-semibold text-white">Nessuna Azienda registrata</div>
            <p className="text-xs text-slate-400 max-w-md mx-auto">
              Non è ancora presente alcuna azienda censita nel sistema. Accedi a Target per creare la prima cartella/azienda e assegnarvi i bersagli di rete.
            </p>
            <Link
              to="/targets"
              className="inline-flex items-center space-x-1.5 px-4 py-2 rounded-lg bg-cyan-500 text-slate-950 font-semibold text-xs"
            >
              <span>Configura Aziende in Target</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
          </div>
        ) : (
          companyCards.companies.map((company) => (
            <div
              key={company.id}
              className="rounded-2xl border border-cyber-border bg-cyber-card/80 p-5 flex flex-col justify-between shadow-xl hover:border-slate-700 transition-all space-y-4"
            >
              {/* Header Azienda */}
              <div className="space-y-2">
                <div className="flex items-start justify-between gap-3">
                  <div className="flex items-center space-x-2.5">
                    <div
                      className="p-2 rounded-xl border shrink-0"
                      style={{
                        backgroundColor: `${company.color_tag}20`,
                        borderColor: `${company.color_tag}50`,
                        color: company.color_tag,
                      }}
                    >
                      <Building className="w-5 h-5" />
                    </div>
                    <div>
                      <h3 className="text-base font-bold text-white line-clamp-1">{company.name}</h3>
                      <div className="text-[11px] text-slate-400">
                        {company.targets.length} IP / Target gestiti
                      </div>
                    </div>
                  </div>

                  {/* Posture Badge */}
                  <span
                    className={`px-2 py-0.5 rounded text-xs font-mono font-bold border ${getGradeBadge(
                      company.postureGrade
                    )}`}
                  >
                    GRADE {company.postureGrade}
                  </span>
                </div>

                {company.description && (
                  <p className="text-xs text-slate-400 line-clamp-2">{company.description}</p>
                )}
              </div>

              {/* Security Posture Index & Progress Bar */}
              <div className="p-3 rounded-xl bg-slate-900/70 border border-cyber-border/80 space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-slate-400 font-medium">Posture Score:</span>
                  <span className="font-mono font-bold text-white">{company.postureScore}/100</span>
                </div>
                <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all"
                    style={{
                      width: `${company.postureScore}%`,
                      backgroundColor:
                        company.postureScore >= 80
                          ? '#10b981'
                          : company.postureScore >= 50
                          ? '#f59e0b'
                          : '#f43f5e',
                    }}
                  />
                </div>
              </div>

              {/* Vulnerabilità Rilevate per l'Azienda */}
              <div className="space-y-1.5">
                <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                  Vulnerabilità Perimetro:
                </div>
                <div className="grid grid-cols-4 gap-1.5 text-center font-mono text-xs">
                  <div className="p-1.5 rounded-lg bg-rose-500/10 border border-rose-500/25 text-rose-400">
                    <div className="font-bold">{company.criticalCount}</div>
                    <div className="text-[9px] text-rose-300">Critiche</div>
                  </div>
                  <div className="p-1.5 rounded-lg bg-orange-500/10 border border-orange-500/25 text-orange-400">
                    <div className="font-bold">{company.highCount}</div>
                    <div className="text-[9px] text-orange-300">Alte</div>
                  </div>
                  <div className="p-1.5 rounded-lg bg-amber-500/10 border border-amber-500/25 text-amber-400">
                    <div className="font-bold">{company.mediumCount}</div>
                    <div className="text-[9px] text-amber-300">Medie</div>
                  </div>
                  <div className="p-1.5 rounded-lg bg-blue-500/10 border border-blue-500/25 text-blue-400">
                    <div className="font-bold">{company.lowCount}</div>
                    <div className="text-[9px] text-blue-300">Basse</div>
                  </div>
                </div>
              </div>

              {/* IP inclusi nel perimetro */}
              <div className="space-y-1.5">
                <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                  Indirizzi IP Censiti ({company.targets.length}):
                </div>
                {company.targets.length === 0 ? (
                  <div className="p-2 rounded-lg bg-slate-900/40 border border-slate-800 text-[11px] text-slate-500 italic">
                    Nessun IP associato. Aggiungine uno in Target.
                  </div>
                ) : (
                  <div className="flex flex-wrap gap-1.5 max-h-20 overflow-y-auto">
                    {company.targets.map((tgt) => (
                      <span
                        key={tgt.id}
                        className="inline-flex items-center space-x-1 px-2 py-0.5 rounded text-[11px] font-mono bg-slate-900 border border-slate-700 text-slate-300"
                      >
                        <span>{tgt.value}</span>
                        <button
                          onClick={() => handleCopy(tgt.value)}
                          className="hover:text-cyan-400"
                          title="Copia IP"
                        >
                          {copiedIP === tgt.value ? (
                            <Check className="w-2.5 h-2.5 text-emerald-400" />
                          ) : (
                            <Copy className="w-2.5 h-2.5" />
                          )}
                        </button>
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {/* Stato Ultimo Audit */}
              <div className="pt-2 border-t border-cyber-border flex items-center justify-between text-xs">
                <span className="text-slate-400">Ultimo Audit:</span>
                <div>{getStatusBadge(company.latestScan?.status)}</div>
              </div>

              {/* Azioni Rapide per l'Azienda */}
              <div className="grid grid-cols-3 gap-2 pt-1 text-center">
                <Link
                  to="/scans"
                  className="py-1.5 px-2 rounded-lg bg-cyan-500/10 hover:bg-cyan-500/20 border border-cyan-500/30 text-cyan-300 font-semibold text-xs transition-colors flex items-center justify-center space-x-1"
                >
                  <Terminal className="w-3 h-3" />
                  <span>Audit</span>
                </Link>

                <Link
                  to="/findings"
                  className="py-1.5 px-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium transition-colors flex items-center justify-center space-x-1"
                >
                  <ShieldAlert className="w-3 h-3 text-red-400" />
                  <span>Falle</span>
                </Link>

                <Link
                  to="/reports"
                  className="py-1.5 px-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium transition-colors flex items-center justify-center space-x-1"
                >
                  <FileText className="w-3 h-3 text-cyan-400" />
                  <span>Dossier</span>
                </Link>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Bersagli Non Assegnati ad alcuna Azienda (Callout Opzionale) */}
      {companyCards.unassignedTargets.length > 0 && (
        <div className="p-4 rounded-2xl border border-amber-500/30 bg-amber-500/10 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs text-amber-200">
          <div className="flex items-center space-x-2.5">
            <AlertTriangle className="w-5 h-5 text-amber-400 shrink-0" />
            <div>
              <span className="font-bold">Attenzione: {companyCards.unassignedTargets.length} bersagli non sono ancora assegnati ad alcuna Azienda.</span>
              <p className="text-[11px] text-amber-300/80 mt-0.5">
                Organizza gli host nelle rispettive cartelle perimetrali in Target per consentire la corretta aggregazione dei report aziendali.
              </p>
            </div>
          </div>
          <Link
            to="/targets"
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-amber-400 hover:bg-amber-300 text-slate-950 font-semibold text-xs whitespace-nowrap transition-colors shadow-sm self-start sm:self-auto"
          >
            <span>Assegna in Target</span>
            <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>
      )}
    </div>
  );
};
