import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  BarChart3,
  RefreshCw,
  AlertCircle,
  Terminal,
  ExternalLink,
  Folder,
  Crosshair,
  Layers,
  Check,
  Copy,
  Building,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { ScanItem, ScanListResponseData } from '../types/scan';
import { FindingItem, FindingListResponseData } from '../types/finding';
import { TargetItem, TargetListResponseData, TargetGroupNode } from '../types/target';
import { FindingsTable } from '../components/FindingsTable';

export const Findings: React.FC = () => {
  const [scans, setScans] = useState<ScanItem[]>([]);
  const [targetsMap, setTargetsMap] = useState<Map<string, TargetItem>>(new Map());
  const [groupsTree, setGroupsTree] = useState<TargetGroupNode[]>([]);

  // Filtri gerarchici: Cartella / Azienda Scope -> Target IP -> Scansione
  const [selectedCompanyId, setSelectedCompanyId] = useState<string>('');
  const [selectedTargetId, setSelectedTargetId] = useState<string>('');
  const [selectedScanId, setSelectedScanId] = useState<string>('');

  const [findings, setFindings] = useState<FindingItem[]>([]);
  const [loadingData, setLoadingData] = useState<boolean>(true);
  const [loadingFindings, setLoadingFindings] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Copia IP
  const [copiedIP, setCopiedIP] = useState<string | null>(null);

  // Caricamento metadati perimetri, bersagli e scansioni
  const fetchData = useCallback(async () => {
    setLoadingData(true);
    setError(null);
    try {
      const [scansRes, targetsRes, groupsRes] = await Promise.all([
        apiClient.get<ApiResponse<ScanListResponseData>>('/scans?limit=100'),
        apiClient.get<ApiResponse<TargetListResponseData>>('/targets?limit=200'),
        apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups'),
      ]);

      if (scansRes.data?.data?.items) {
        setScans(scansRes.data.data.items);
      }
      if (targetsRes.data?.data?.items) {
        const tMap = new Map<string, TargetItem>();
        targetsRes.data.data.items.forEach((t) => tMap.set(t.id, t));
        setTargetsMap(tMap);
      }
      if (groupsRes.data?.data) {
        setGroupsTree(groupsRes.data.data);
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile recuperare lo storico scansioni e le cartelle scope'
      );
    } finally {
      setLoadingData(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Costruzione gerarchica dei gruppi / aziende
  const companyFolders = useMemo(() => {
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

    const folderMap = new Map<
      string,
      {
        id: string;
        name: string;
        color_tag: string;
        targets: Map<string, { target: TargetItem | null; scans: ScanItem[] }>;
      }
    >();

    // Inizializza cartelle note
    gMap.forEach((gNode, gId) => {
      folderMap.set(gId, {
        id: gId,
        name: gNode.name,
        color_tag: gNode.color_tag || '#06b6d4',
        targets: new Map(),
      });
    });

    // Inizializza root per target non assegnati
    folderMap.set('unassigned', {
      id: 'unassigned',
      name: 'Bersagli Non Assegnati (Root)',
      color_tag: '#64748b',
      targets: new Map(),
    });

    // Popola i target
    targetsMap.forEach((target, targetId) => {
      const gId = target.group_id || 'unassigned';
      if (!folderMap.has(gId)) {
        folderMap.set(gId, {
          id: gId,
          name: 'Cartella Scope',
          color_tag: '#06b6d4',
          targets: new Map(),
        });
      }
      folderMap.get(gId)!.targets.set(targetId, { target, scans: [] });
    });

    // Associa le scansioni storiche ai rispettivi target e cartelle
    scans.forEach((scan) => {
      const target = targetsMap.get(scan.target_id);
      const gId = target?.group_id || 'unassigned';
      if (!folderMap.has(gId)) {
        folderMap.set(gId, {
          id: gId,
          name: 'Cartella Scope',
          color_tag: '#06b6d4',
          targets: new Map(),
        });
      }
      const entry = folderMap.get(gId)!;
      if (!entry.targets.has(scan.target_id)) {
        entry.targets.set(scan.target_id, { target: target || null, scans: [] });
      }
      entry.targets.get(scan.target_id)!.scans.push(scan);
    });

    const result = Array.from(folderMap.values())
      .map((f) => {
        const targetsList = Array.from(f.targets.entries()).map(([tId, val]) => ({
          id: tId,
          value: val.target?.value || `${tId.slice(0, 8)}...`,
          type: val.target?.type || 'IP',
          description: val.target?.description,
          scans: val.scans,
        }));
        // Ordina bersagli per valore IP
        targetsList.sort((a, b) => a.value.localeCompare(b.value));

        const totalScans = targetsList.reduce((acc, t) => acc + t.scans.length, 0);
        return {
          id: f.id,
          name: f.name,
          color_tag: f.color_tag,
          targets: targetsList,
          totalScans,
        };
      })
      .filter((f) => f.targets.length > 0 || f.totalScans > 0);

    result.sort((a, b) => {
      if (a.id === 'unassigned') return 1;
      if (b.id === 'unassigned') return -1;
      return a.name.localeCompare(b.name);
    });

    return result;
  }, [groupsTree, targetsMap, scans]);

  // Seleziona automaticamente la prima cartella/azienda con scansioni se non ancora selezionata
  useEffect(() => {
    if (!selectedCompanyId && companyFolders.length > 0) {
      const firstWithScans = companyFolders.find((c) => c.totalScans > 0);
      setSelectedCompanyId(firstWithScans ? firstWithScans.id : companyFolders[0].id);
    }
  }, [companyFolders, selectedCompanyId]);

  // Cartella / Azienda attiva
  const activeCompany = useMemo(() => {
    return companyFolders.find((c) => c.id === selectedCompanyId) || companyFolders[0] || null;
  }, [companyFolders, selectedCompanyId]);

  // Target attivo all'interno dell'azienda
  const activeTarget = useMemo(() => {
    if (!activeCompany || !selectedTargetId) return null;
    return activeCompany.targets.find((t) => t.id === selectedTargetId) || null;
  }, [activeCompany, selectedTargetId]);

  // Scansioni idonee da cui recuperare i finding per la vista corrente
  const targetScanIds = useMemo(() => {
    if (selectedScanId) {
      return [selectedScanId];
    }
    if (activeTarget) {
      // Se è selezionato un IP specifico, prendi la sua scansione completata più recente
      const completed = activeTarget.scans.filter((s) => s.status === 'COMPLETED');
      if (completed.length > 0) {
        return [completed[0].id];
      }
      return activeTarget.scans.length > 0 ? [activeTarget.scans[0].id] : [];
    }
    if (activeCompany) {
      // Se sono selezionati tutti gli IP dell'azienda, raccogli la scansione più recente per ciascun target
      const ids: string[] = [];
      activeCompany.targets.forEach((t) => {
        const completed = t.scans.filter((s) => s.status === 'COMPLETED');
        if (completed.length > 0) {
          ids.push(completed[0].id);
        } else if (t.scans.length > 0) {
          ids.push(t.scans[0].id);
        }
      });
      return ids;
    }
    return [];
  }, [selectedScanId, activeTarget, activeCompany]);

  // Caricamento finding per gli scan identificati
  const fetchFindingsForScans = useCallback(async (scanIds: string[]) => {
    if (scanIds.length === 0) {
      setFindings([]);
      return;
    }
    setLoadingFindings(true);
    try {
      const results = await Promise.all(
        scanIds.slice(0, 5).map((sId) =>
          apiClient.get<ApiResponse<FindingListResponseData>>(`/scans/${sId}/findings?limit=100`)
        )
      );

      const combined: FindingItem[] = [];
      const seen = new Set<string>();
      for (const res of results) {
        if (res.data?.data?.items) {
          for (const item of res.data.data.items) {
            if (!seen.has(item.id)) {
              seen.add(item.id);
              combined.push(item);
            }
          }
        }
      }
      setFindings(combined);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Impossibile recuperare i finding di vulnerabilità');
    } finally {
      setLoadingFindings(false);
    }
  }, []);

  const targetScanIdsKey = useMemo(() => targetScanIds.join(','), [targetScanIds]);

  useEffect(() => {
    if (targetScanIds.length > 0) {
      fetchFindingsForScans(targetScanIds);
    } else {
      setFindings([]);
    }
  }, [targetScanIdsKey, fetchFindingsForScans]);

  const handleCopy = (val: string) => {
    navigator.clipboard.writeText(val);
    setCopiedIP(val);
    setTimeout(() => setCopiedIP(null), 2000);
  };

  const selectedScanObj = scans.find((s) => s.id === (selectedScanId || (targetScanIds.length === 1 ? targetScanIds[0] : '')));

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center space-x-3">
          <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400 shadow-[0_0_15px_rgba(6,182,212,0.15)]">
            <BarChart3 className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center space-x-2">
              <span>Vulnerabilità & Correlazione CVE</span>
              <span className="text-xs px-2.5 py-0.5 rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 font-mono">
                Multi-Company Scope
              </span>
            </h1>
            <p className="text-sm text-slate-400 mt-0.5">
              Catalogo vulnerabilità scoperte organizzato per Cartella/Azienda cliente ed esplorazione dei singoli IP
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-3">
          {selectedScanObj && (
            <Link
              to={`/scans/${selectedScanObj.id}`}
              className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-cyan-500/30 bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-400 text-xs font-semibold transition-colors"
            >
              <span>Dettaglio Scansione</span>
              <ExternalLink className="w-3.5 h-3.5" />
            </Link>
          )}

          <button
            onClick={() => {
              fetchData();
              if (targetScanIds.length > 0) fetchFindingsForScans(targetScanIds);
            }}
            disabled={loadingData || loadingFindings}
            className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
            title="Aggiorna dati"
          >
            <RefreshCw className={`w-4 h-4 ${loadingData || loadingFindings ? 'animate-spin text-cyan-400' : ''}`} />
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 rounded-xl border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center space-x-2">
          <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
          <span>{error}</span>
        </div>
      )}

      {/* Livello 1: Selettore Cartella Scope / Azienda Cliente */}
      <div className="rounded-xl border border-cyber-border bg-[#111726]/80 p-4 space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2 text-xs font-bold uppercase tracking-wider text-slate-400">
            <Building className="w-4 h-4 text-cyan-400" />
            <span>Seleziona Azienda o Cartella Scope</span>
          </div>
          <span className="text-[11px] text-slate-400 font-mono">
            {companyFolders.length} {companyFolders.length === 1 ? 'cartella attiva' : 'cartelle attive'}
          </span>
        </div>

        {/* Lista Orizzontale Card Aziende */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5">
          {companyFolders.map((comp) => {
            const isSelected = activeCompany?.id === comp.id;
            return (
              <button
                key={comp.id}
                type="button"
                onClick={() => {
                  setSelectedCompanyId(comp.id);
                  setSelectedTargetId('');
                  setSelectedScanId('');
                }}
                className={`p-3 rounded-xl border text-left transition-all flex items-start space-x-3 ${
                  isSelected
                    ? 'bg-cyan-500/15 border-cyan-500/50 shadow-[0_0_15px_rgba(6,182,212,0.15)] ring-1 ring-cyan-500/30'
                    : 'bg-slate-900/60 hover:bg-slate-800/60 border-cyber-border/70 text-slate-300'
                }`}
              >
                <div
                  className="p-2 rounded-lg border shrink-0 mt-0.5"
                  style={{
                    backgroundColor: `${comp.color_tag}15`,
                    borderColor: `${comp.color_tag}40`,
                    color: comp.color_tag,
                  }}
                >
                  <Folder className="w-4 h-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className={`text-xs font-bold truncate ${isSelected ? 'text-white' : 'text-slate-200'}`}>
                    {comp.name}
                  </div>
                  <div className="flex items-center space-x-2 text-[10px] text-slate-400 mt-1 font-mono">
                    <span>{comp.targets.length} Target/IP</span>
                    <span>·</span>
                    <span className={comp.totalScans > 0 ? 'text-cyan-400' : 'text-slate-500'}>
                      {comp.totalScans} audit
                    </span>
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Livello 2: Selettore Specifico Target / Indirizzo IP all'interno dell'Azienda */}
      {activeCompany && activeCompany.targets.length > 0 && (
        <div className="rounded-xl border border-cyber-border bg-[#111726]/60 p-4 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2 text-xs font-bold uppercase tracking-wider text-slate-400">
              <Crosshair className="w-4 h-4 text-cyan-400" />
              <span>Ispezione per Singolo IP o Visione Aggregata</span>
            </div>
            <span className="text-[11px] text-slate-400">
              Ambito: <strong className="text-white font-semibold">{activeCompany.name}</strong>
            </span>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {/* Opzione: Tutti gli IP dell'azienda */}
            <button
              type="button"
              onClick={() => {
                setSelectedTargetId('');
                setSelectedScanId('');
              }}
              className={`px-3 py-1.5 rounded-lg border text-xs font-semibold flex items-center space-x-2 transition-all ${
                selectedTargetId === ''
                  ? 'bg-cyan-500/20 text-cyan-300 border-cyan-500/50 shadow-sm'
                  : 'bg-slate-900 border-cyber-border text-slate-400 hover:text-white hover:bg-slate-800'
              }`}
            >
              <Layers className="w-3.5 h-3.5" />
              <span>Tutti gli IP di questa Cartella ({activeCompany.targets.length})</span>
            </button>

            {/* Pill per ciascun IP / Target censito */}
            {activeCompany.targets.map((tgt) => {
              const isSelected = selectedTargetId === tgt.id;
              return (
                <button
                  key={tgt.id}
                  type="button"
                  onClick={() => {
                    setSelectedTargetId(tgt.id);
                    setSelectedScanId('');
                  }}
                  className={`px-3 py-1.5 rounded-lg border text-xs transition-all flex items-center space-x-2 font-mono ${
                    isSelected
                      ? 'bg-cyan-500/20 text-cyan-300 border-cyan-500/50 font-bold shadow-sm'
                      : 'bg-slate-900 border-cyber-border text-slate-300 hover:text-white hover:bg-slate-800'
                  }`}
                >
                  <span className="text-cyan-400 font-semibold">{tgt.value}</span>
                  <span className="text-[9px] px-1 py-0.2 rounded bg-slate-950 border border-slate-700 text-slate-400 uppercase font-sans">
                    {tgt.type}
                  </span>
                  {tgt.scans.length > 0 && (
                    <span className="text-[10px] px-1.5 py-0.2 rounded bg-cyan-950/60 text-cyan-400 border border-cyan-800/40">
                      {tgt.scans.length}
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          {/* Dettagli del Target Selezionato e Selettore di Scansione Storica */}
          {activeTarget && (
            <div className="pt-2 border-t border-cyber-border/60 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
              <div className="flex items-center space-x-2.5">
                <span className="text-slate-400">Target IP selezionato:</span>
                <span className="font-mono font-bold text-white text-cyan-300">{activeTarget.value}</span>
                <button
                  type="button"
                  onClick={() => handleCopy(activeTarget.value)}
                  className="text-slate-500 hover:text-slate-300 p-0.5 rounded"
                  title="Copia valore IP"
                >
                  {copiedIP === activeTarget.value ? (
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                  ) : (
                    <Copy className="w-3.5 h-3.5" />
                  )}
                </button>
                {activeTarget.description && (
                  <span className="text-slate-400 italic hidden md:inline">— {activeTarget.description}</span>
                )}
              </div>

              {/* Selettore scansione se ce ne sono più di una */}
              {activeTarget.scans.length > 1 && (
                <div className="flex items-center space-x-2">
                  <span className="text-slate-400 text-[11px]">Esecuzione Scansione:</span>
                  <select
                    value={selectedScanId}
                    onChange={(e) => setSelectedScanId(e.target.value)}
                    className="bg-slate-900 border border-cyber-border rounded px-2 py-1 text-xs text-cyan-400 font-mono focus:outline-none"
                  >
                    <option value="">Ultima Scansione (Consigliata)</option>
                    {activeTarget.scans.map((s) => (
                      <option key={s.id} value={s.id}>
                        [{s.status}] {s.id.slice(0, 8)}... - {new Date(s.created_at).toLocaleDateString('it-IT')}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Tabella Findings */}
      {loadingData ? (
        <div className="py-16 rounded-xl border border-cyber-border bg-cyber-card/40 flex flex-col items-center justify-center text-slate-400 space-y-3">
          <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
          <span className="text-xs">Caricamento inventario vulnerabilità e perimetri...</span>
        </div>
      ) : scans.length === 0 ? (
        <div className="py-16 flex flex-col items-center justify-center text-center space-y-3 p-6 rounded-xl border border-cyber-border bg-cyber-card/30">
          <div className="w-12 h-12 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-slate-500">
            <Terminal className="w-6 h-6" />
          </div>
          <div className="text-sm font-semibold text-slate-200">Nessuna scansione eseguita</div>
          <p className="text-xs text-slate-400 max-w-sm">
            Esegui prima una scansione Nmap per rilevare servizi e correlare vulnerabilità CVE per questo cliente.
          </p>
        </div>
      ) : (
        <FindingsTable findings={findings} loading={loadingFindings} />
      )}
    </div>
  );
};
