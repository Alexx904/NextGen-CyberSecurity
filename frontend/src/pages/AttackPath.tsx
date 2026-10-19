import React, { useState, useEffect, useLayoutEffect, useCallback, useMemo, useRef } from 'react';
import { Link } from 'react-router-dom';
import {
  Network,
  ShieldAlert,
  Target,
  RefreshCw,
  SlidersHorizontal,
  Flame,
  AlertTriangle,
  ArrowRight,
  ExternalLink,
  ChevronRight,
  Copy,
  Check,
  Compass,
} from 'lucide-react';
import { attackPathApi } from '../api/attackPath';
import {
  AttackGraph,
  AttackNode,
  AttackPath as IAttackPath,
} from '../types/attackPath';
import { AttackPathGraph } from '../components/graph/AttackPathGraph';
import { useCompany } from '../context/CompanyContext';
import { scopeAttackGraph } from '../components/graph/attackPathModel';
import { AttackPathGuide } from '../components/graph/AttackPathGuide';
import { useAuth } from '../context/AuthContext';
import { ChokepointRemediationModal, supportsChokepointMitigation } from '../components/ChokepointRemediationModal';
import { FirewallManager } from '../components/FirewallManager';
import type { SSHServerItem } from '../types/ssh';

export const AttackPath: React.FC = () => {
  const { user } = useAuth();
  const canRemediate = Boolean(user && ['SuperAdmin', 'TenantAdmin', 'SecOps'].includes(user.role));
  const [remediationNode, setRemediationNode] = useState<AttackNode | null>(null);
  const [firewallServer, setFirewallServer] = useState<SSHServerItem | null>(null);
  const {
    selectedCompanyId,
    selectedCompany,
    departments,
    companyTargetValues,
    allTargets = [],
  } = useCompany();
  const [departmentId, setDepartmentId] = useState('');
  const activeDepartmentId = departments.some(d => d.id === departmentId) ? departmentId : '';
  const scopeValues = useMemo(() => activeDepartmentId
    ? new Set(allTargets.filter(t => t.group_id === activeDepartmentId).map(t => t.value))
    : companyTargetValues, [activeDepartmentId, allTargets, companyTargetValues]);

  const [rawGraph, setRawGraph] = useState<AttackGraph | null>(null);
  const requestVersion = useRef(0);
  const graph = useMemo(() => rawGraph && selectedCompanyId
    ? scopeAttackGraph(rawGraph, scopeValues) : rawGraph,
  [rawGraph, selectedCompanyId, scopeValues]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Stati di interattività e selezione
  const [selectedPathId, setSelectedPathId] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<AttackNode | null>(null);
  const [activeTab, setActiveTab] = useState<'paths' | 'chokepoints' | 'node'>('paths');
  const [copiedChokepoint, setCopiedChokepoint] = useState<string | null>(null);

  const displayedPaths = graph?.paths || [];

  useLayoutEffect(() => {
    setSelectedPathId(graph?.paths[0]?.id || null);
    setSelectedNode(null);
    setActiveTab('paths');
    setRemediationNode(null);
  }, [graph, canRemediate]);

  const fetchGraph = useCallback(async () => {
    const version = ++requestVersion.current;
    setLoading(true);
    setError(null);
    try {
      const data = await attackPathApi.getGraph();
      if (version !== requestVersion.current) return;
      setRawGraph(data);
    } catch (err) {
      if (version !== requestVersion.current) return;
      setRawGraph(null);
      setError(
        err instanceof Error ? err.message : 'Impossibile recuperare il grafo di attacco'
      );
    } finally {
      if (version === requestVersion.current) setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchGraph();
    return () => { requestVersion.current++; };
  }, [fetchGraph]);

  const handleSelectNode = (node: AttackNode | null) => {
    setSelectedNode(node);
    if (node) {
      setSelectedPathId(null);
      setActiveTab('node');
    } else {
      setActiveTab('paths');
    }
  };

  const handleSelectPath = (path: IAttackPath | null) => {
    setSelectedPathId(path ? path.id : null);
    if (path) {
      setActiveTab('paths');
    }
  };

  const handleCopyRemediation = async (text: string, id: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedChokepoint(id);
    } catch {
      setError('Copia non disponibile: seleziona e copia il testo della raccomandazione.');
    }
  };

  const mitigationAction = (node?: AttackNode) => canRemediate && node ? <button type="button"
    disabled={!supportsChokepointMitigation(node)}
    title={supportsChokepointMitigation(node) ? 'Valuta un blocco porta sul server SSH del servizio' : 'Solo servizi IPv4 TCP/UDP; per altri nodi serve una verifica manuale'}
    onClick={() => setRemediationNode(node)} aria-label={`Neutralizza Chokepoint ${node.label}`}
    className="mt-3 rounded-lg border border-amber-600 px-3 py-2 text-xs font-semibold text-amber-300 disabled:opacity-40">Neutralizza Chokepoint</button> : null;

  return (
    <div className="space-y-6">
      {/* Header Pagina con Titolo, Descrizione e Filtri Rapidi */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-2 border-b border-slate-800">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
              <Network className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white tracking-tight">
                Attack Path Mapping & Lateral Movement
              </h1>
              <p className="text-sm text-slate-400 mt-0.5">
                Mappa topologica delle vulnerabilità, correlazione dei vettori di penetrazione e chokepoint di sicurezza.
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          {/* Filtro Min CVSS */}
          <div className="flex items-center gap-2 bg-slate-900 border border-slate-800 rounded-lg px-3 py-1.5 text-xs text-slate-300">
            <SlidersHorizontal className="w-3.5 h-3.5 text-slate-400" />
            <span>Perimetro:</span>
            <select
              aria-label="Reparto Attack Path"
              value={activeDepartmentId}
              onChange={(e) => setDepartmentId(e.target.value)}
              className="bg-transparent text-cyan-400 font-mono font-medium focus:outline-hidden cursor-pointer"
            >
              <option value="" className="bg-slate-900">Intera azienda</option>
              {departments.map(d => <option key={d.id} value={d.id} className="bg-slate-900">{d.name}</option>)}
            </select>
          </div>

          {/* Pulsante Ricarica Grafo */}
          <button
            onClick={fetchGraph}
            disabled={loading}
            className="flex items-center gap-2 px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium rounded-lg border border-slate-700 transition disabled:opacity-50"
            title="Ricarica Grafo Topologico"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            <span>Aggiorna</span>
          </button>
        </div>
      </div>

      <AttackPathGuide />
      <div className="rounded-xl border border-violet-500/30 bg-violet-950/20 p-4 text-sm text-slate-300 space-y-2">
        <p className="font-semibold text-violet-200">Modello di analisi: evidenze e ipotesi, nessun exploit eseguito</p>
        <p>Le frecce tratteggiate richiedono verifica. Una CVE non dimostra RCE o lateral movement; l’indice 0–100 serve a ordinare le analisi e non è una probabilità.</p>
        {graph && <details className="text-xs text-slate-400"><summary className="cursor-pointer">Copertura e limiti dei dati</summary>
          <ul className="list-disc pl-5 space-y-1 mt-2">{(graph.analysis?.warnings || ['Verifica età delle scansioni, accessibilità di rete e prerequisiti delle CVE prima di intervenire.']).map(w => <li key={w}>{w}</li>)}</ul>
          <p className="mt-2">Generato: {new Date(graph.generated_at).toLocaleString('it-IT')}. Lo scope aziendale filtra la vista; l’autorizzazione API è per tenant.</p>
        </details>}
        {(graph?.analysis?.paths_truncated || graph?.analysis?.inventory_truncated) && <p role="status" className="text-amber-300">Copertura parziale: limite di inventario o percorsi raggiunto prima del filtro aziendale. KPI e chokepoint descrivono solo il campione visualizzato.</p>}
      </div>

      {/* Banner Scope Aziendale Attivo */}
      {selectedCompany && (
        <div className="p-3.5 rounded-xl border border-cyan-500/30 bg-cyan-950/20 backdrop-blur-xs flex items-center justify-between text-xs">
          <div className="flex items-center space-x-2.5">
            <span className="w-2 h-2 rounded-full bg-cyan-400 animate-pulse" />
            <span className="text-slate-300">
              Scope attivo:{' '}
              <strong className="text-white font-semibold">{selectedCompany.name}</strong>
              {departments.length > 0 && (
                <span className="text-slate-400 ml-1.5">
                  ({departments.length} reparti associati)
                </span>
              )}
            </span>
          </div>
          <span className="text-[11px] text-cyan-400 font-mono">
            {companyTargetValues.size} IP censiti nel perimetro aziendale
          </span>
        </div>
      )}

      {/* KPI Banner di Sintesi Topologica */}
      {graph && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-slate-400">Nodi & Connessioni</p>
              <p className="text-2xl font-bold text-slate-100 mt-1 font-mono">
                {graph.summary.total_nodes}
                <span className="text-xs text-slate-500 font-normal ml-2">
                  ({graph.summary.total_edges} archi)
                </span>
              </p>
            </div>
            <div className="p-3 bg-slate-800/80 rounded-xl text-cyan-400 border border-slate-700/50">
              <Network className="w-5 h-5" />
            </div>
          </div>

          <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-slate-400">Percorsi da verificare</p>
              <div className="flex items-center gap-2 mt-1">
                <p className="text-2xl font-bold text-slate-100 font-mono">
                  {graph.summary.total_paths}
                </p>
                {graph.summary.critical_paths > 0 && (
                  <span className="px-2 py-0.5 text-[11px] font-bold rounded-full bg-rose-500/20 text-rose-300 border border-rose-500/30">
                    {graph.summary.critical_paths} Critici
                  </span>
                )}
              </div>
            </div>
            <div className="p-3 bg-slate-800/80 rounded-xl text-rose-400 border border-slate-700/50">
              <Flame className="w-5 h-5" />
            </div>
          </div>

          <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center justify-between">
            <div>
              <p className="text-xs font-medium text-slate-400">Indice euristico Max</p>
              <p
                className={`text-2xl font-bold mt-1 font-mono ${
                  graph.summary.highest_risk_score >= 80
                    ? 'text-rose-400'
                    : graph.summary.highest_risk_score >= 50
                    ? 'text-amber-400'
                    : 'text-cyan-400'
                }`}
              >
                {graph.summary.highest_risk_score.toFixed(1)}
                <span className="text-xs text-slate-500 font-normal ml-1">/ 100</span>
              </p>
            </div>
            <div className="p-3 bg-slate-800/80 rounded-xl text-amber-400 border border-slate-700/50">
              <ShieldAlert className="w-5 h-5" />
            </div>
          </div>

          <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center justify-between">
            <div className="min-w-0 pr-2">
              <p className="text-xs font-medium text-slate-400">Priorità nel modello</p>
              <p
                className="text-sm font-bold text-amber-300 mt-1 truncate"
                title={graph.summary.top_chokepoint || 'Nessuno'}
              >
                {graph.summary.top_chokepoint || 'Nessuno identificato'}
              </p>
            </div>
            <div className="p-3 bg-slate-800/80 rounded-xl text-amber-400 border border-slate-700/50 shrink-0">
              <Target className="w-5 h-5" />
            </div>
          </div>
        </div>
      )}

      {/* Stato di Errore */}
      {error && (
        <div className="p-4 bg-rose-500/10 border border-rose-500/30 rounded-xl text-sm text-rose-300 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-5 h-5 text-rose-400 shrink-0" />
            <span>{error}</span>
          </div>
          <button
            onClick={fetchGraph}
            className="text-xs text-rose-200 underline hover:text-white"
          >
            Riprova
          </button>
        </div>
      )}

      {/* Stato di Caricamento */}
      {loading && !graph && (
        <div className="h-[500px] flex items-center justify-center bg-slate-900/40 border border-slate-800 rounded-xl">
          <div className="flex flex-col items-center gap-3 text-cyan-400">
            <div className="w-8 h-8 rounded-full border-2 border-cyan-400 border-t-transparent animate-spin" />
            <span className="text-xs font-mono">Calcolo topologia e percorsi di attacco...</span>
          </div>
        </div>
      )}

      {/* Stato Vuoto / Nessuna Scansione Completata */}
      {!loading && graph && graph.summary.total_nodes <= 1 && (
        <div className="p-12 text-center bg-slate-900/40 border border-slate-800 rounded-2xl flex flex-col items-center max-w-2xl mx-auto my-8">
          <div className="w-16 h-16 rounded-2xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 flex items-center justify-center mb-4">
            <Compass className="w-8 h-8" />
          </div>
          <h2 className="text-lg font-bold text-slate-100">
            Nessun percorso di attacco rilevato
          </h2>
          <p className="text-sm text-slate-400 mt-2 leading-relaxed">
            Il grafo delle minacce correla automaticamente i target censiti con le porte aperte e le vulnerabilità scoperte dalle scansioni Nmap. Esegui un audit con il profilo{' '}
            <strong className="text-slate-200">Audit completo</strong> o{' '}
            <strong className="text-slate-200">Vulnerabilità</strong> per arricchire le ipotesi del modello. Nessun percorso non significa assenza di rischio.
          </p>
          <Link
            to="/scans"
            className="mt-6 inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-white font-medium text-sm transition shadow-lg shadow-cyan-900/30"
          >
            <span>Vai a Scansioni Nmap</span>
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      )}

      {/* Griglia Principale: Graph Explorer + Pannello Laterale Interattivo */}
      {graph && graph.summary.total_nodes > 1 && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          {/* Canvas Grafo Interattivo (8 colonne su 12) */}
          <div className="lg:col-span-8 flex flex-col gap-2">
            <div className="flex items-center justify-between text-xs text-slate-400 px-1">
              <span className="font-mono">
                Visualizzazione Topologica Diretta ({graph.nodes.length} nodi, {graph.edges.length} archi)
              </span>
              <span className="text-[11px] text-slate-500">
                Trascina per spostare la vista · Clicca su un nodo per ispezionarlo
              </span>
            </div>

            <AttackPathGraph
              graph={graph}
              selectedPathId={selectedPathId}
              selectedNodeId={selectedNode?.id || null}
              onSelectNode={handleSelectNode}
              onSelectPath={handleSelectPath}
            />
          </div>

          {/* Pannello Laterale di Controllo (4 colonne su 12) */}
          <div className="lg:col-span-4 bg-slate-900 border border-slate-800 rounded-xl flex flex-col overflow-hidden shadow-xl">
            {/* Tab Header */}
            <div className="flex items-center border-b border-slate-800 bg-slate-950/50 p-1">
              <button
                onClick={() => setActiveTab('paths')}
                className={`flex-1 py-2 text-xs font-semibold rounded-lg transition flex items-center justify-center gap-1.5 ${
                  activeTab === 'paths'
                    ? 'bg-slate-800 text-white shadow-xs'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                <span>Percorsi ({displayedPaths.length})</span>
              </button>
              <button
                onClick={() => setActiveTab('chokepoints')}
                className={`flex-1 py-2 text-xs font-semibold rounded-lg transition flex items-center justify-center gap-1.5 ${
                  activeTab === 'chokepoints'
                    ? 'bg-slate-800 text-amber-300 shadow-xs'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                <span>Chokepoints ({graph.chokepoints?.length || 0})</span>
              </button>
              {selectedNode && (
                <button
                  onClick={() => setActiveTab('node')}
                  className={`flex-1 py-2 text-xs font-semibold rounded-lg transition flex items-center justify-center gap-1.5 ${
                    activeTab === 'node'
                      ? 'bg-slate-800 text-cyan-300 shadow-xs'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <span>Dettaglio Nodo</span>
                </button>
              )}
            </div>

            {/* Contenuto Tab 1: Lista Percorsi di Attacco */}
            {activeTab === 'paths' && (
              <div className="p-4 space-y-3 max-h-[580px] overflow-y-auto">
                <div className="text-xs text-slate-400 mb-1">
                  Seleziona una catena di attacco per visualizzarla isolata sul grafo:
                </div>

                {displayedPaths.length === 0 && (
                  <p className="text-xs text-slate-500 py-4 text-center">
                    Nessun percorso identificato per l’azienda o reparto selezionato.
                  </p>
                )}

                {displayedPaths.map((path) => {
                  const isSelected = selectedPathId === path.id;
                  const isCrit = path.severity === 'CRITICAL';
                  const isHigh = path.severity === 'HIGH';

                  return (
                    <div
                      key={path.id}
                      role="button"
                      tabIndex={0}
                      aria-label={`Seleziona percorso ${path.name}`}
                      aria-pressed={isSelected}
                      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); setSelectedPathId(isSelected ? null : path.id); } }}
                      onClick={() => setSelectedPathId(isSelected ? null : path.id)}
                      className={`p-3.5 rounded-xl border transition cursor-pointer ${
                        isSelected
                          ? 'bg-slate-800/90 border-rose-500/80 ring-1 ring-rose-500/50 shadow-lg'
                          : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <span
                              className={`text-[10px] font-bold px-2 py-0.5 rounded uppercase tracking-wider font-mono border ${
                                isCrit
                                  ? 'bg-rose-500/20 text-rose-300 border-rose-500/30'
                                  : isHigh
                                  ? 'bg-amber-500/20 text-amber-300 border-amber-500/30'
                                  : 'bg-cyan-500/20 text-cyan-300 border-cyan-500/30'
                              }`}
                            >
                              {path.severity}
                            </span>
                            <span className="text-xs font-mono text-slate-400">
                              {path.hop_count} passaggi
                            </span>
                          </div>

                          <h3 className="text-xs font-semibold text-slate-100 mt-1.5 leading-snug">
                            {path.name}
                          </h3>
                        </div>

                        <div className="text-right shrink-0">
                          <span
                            className={`text-sm font-bold font-mono ${
                              isCrit ? 'text-rose-400' : isHigh ? 'text-amber-400' : 'text-cyan-400'
                            }`}
                          >
                            {path.cumulative_risk.toFixed(1)}
                          </span>
                          <span className="text-[10px] text-slate-500 block">Indice</span>
                        </div>
                      </div>

                      <div className="mt-2.5 pt-2 border-t border-slate-800/80 flex items-center justify-between text-[11px] text-slate-400">
                        <span className="truncate">
                          Target: <strong className="text-slate-200">{path.target_asset}</strong>
                        </span>
                        <ChevronRight
                          className={`w-3.5 h-3.5 transition-transform ${
                            isSelected ? 'rotate-90 text-rose-400' : 'text-slate-600'
                          }`}
                        />
                      </div>

                      {/* Sequenza step espansa */}
                      {isSelected && (
                        <div className="mt-3 pt-2.5 border-t border-slate-700/60 space-y-1.5">
                          {path.description && <p className="text-xs text-slate-400 leading-relaxed mb-3">{path.description}</p>}
                          <p className="text-[11px] font-medium text-slate-300 mb-1">
                            Sequenza di analisi:
                          </p>
                          {path.node_ids.map((nodeId, idx) => {
                            const node = graph.nodes.find((n) => n.id === nodeId);
                            return (
                              <div
                                key={idx}
                                className="flex items-center gap-2 text-[11px] font-mono text-slate-300 pl-1"
                              >
                                <span className="text-slate-500">{idx + 1}.</span>
                                <span className="truncate text-slate-200">
                                  {node?.label || nodeId}
                                </span>
                              </div>
                            );
                          })}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}

            {/* Contenuto Tab 2: Chokepoints di Remediation */}
            {activeTab === 'chokepoints' && (
              <div className="p-4 space-y-3 max-h-[580px] overflow-y-auto">
                <div className="text-xs text-slate-400 mb-1">
                  Nodi comuni ai percorsi visualizzati. Valuta la raccomandazione e l’impatto operativo prima di intervenire.
                </div>

                {(!graph.chokepoints || graph.chokepoints.length === 0) && (
                  <p className="text-xs text-slate-500 py-4 text-center">
                    Nessun chokepoint critico rilevato per questo perimetro.
                  </p>
                )}

                {graph.chokepoints?.map((cp, idx) => {
                  const node = graph.nodes.find((n) => n.id === cp.node_id);
                  const isNodeActive = selectedNode?.id === cp.node_id;

                  return (
                    <div
                      key={idx}
                      className={`p-3.5 rounded-xl border transition ${
                        isNodeActive
                          ? 'bg-slate-800/90 border-amber-500 ring-1 ring-amber-500/40 shadow-lg'
                          : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                          <span className="text-[10px] font-mono font-bold px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-300 border border-amber-500/30">
                            Priorità #{idx + 1}
                          </span>
                          <h3 className="text-xs font-semibold text-slate-100 mt-1.5 truncate">
                            {cp.label}
                          </h3>
                        </div>

                        <button
                          onClick={() => handleSelectNode(node || null)}
                          className="text-[11px] text-cyan-400 hover:text-cyan-200 font-medium flex items-center gap-1 shrink-0"
                          title="Focalizza sul Grafo"
                        >
                          <span>Focalizza</span>
                          <ExternalLink className="w-3 h-3" />
                        </button>
                      </div>

                      <div className="grid grid-cols-2 gap-2 mt-2 py-2 border-y border-slate-800/80 text-[11px] font-mono">
                        <div>
                          <span className="text-slate-500 block text-[10px]">Nel modello</span>
                          <strong className="text-amber-300">{cp.mitigated_paths} percorsi</strong>
                        </div>
                        <div>
                          <span className="text-slate-500 block text-[10px]">Indice aggregato</span>
                          <strong className="text-rose-400">
                            {cp.mitigated_risk.toFixed(1)} pts
                          </strong>
                        </div>
                      </div>

                      <div className="mt-2.5">
                        <p className="text-[10px] text-slate-400 font-medium mb-1">
                          Raccomandazione da verificare:
                        </p>
                        <div className="p-2 bg-slate-900 rounded-lg border border-slate-800 text-[11px] text-slate-300 leading-relaxed font-sans relative group">
                          {cp.remediation}
                          <button
                            onClick={() => handleCopyRemediation(cp.remediation, cp.node_id)}
                            className="absolute top-1.5 right-1.5 p-1 bg-slate-800 hover:bg-slate-700 rounded text-slate-400 hover:text-white transition"
                            title="Copia Testo Azione"
                          >
                            {copiedChokepoint === cp.node_id ? (
                              <Check className="w-3 h-3 text-emerald-400" />
                            ) : (
                              <Copy className="w-3 h-3" />
                            )}
                          </button>
                        </div>
                        {mitigationAction(node)}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Contenuto Tab 3: Dettaglio Nodo Selezionato */}
            {activeTab === 'node' && selectedNode && (
              <div className="p-4 space-y-4 max-h-[580px] overflow-y-auto">
                <div>
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
                      {selectedNode.type}
                    </span>
                    {selectedNode.severity && (
                      <span
                        className={`text-[10px] font-bold font-mono px-2 py-0.5 rounded uppercase ${
                          selectedNode.severity === 'CRITICAL'
                            ? 'bg-rose-500/20 text-rose-300 border border-rose-500/30'
                            : selectedNode.severity === 'HIGH'
                            ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                            : 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/30'
                        }`}
                      >
                        {selectedNode.severity}
                      </span>
                    )}
                  </div>
                  <h3 className="text-sm font-bold text-white mt-1.5">{selectedNode.label}</h3>
                </div>

                <div className="space-y-2 border-t border-slate-800 pt-3 text-xs">
                  {selectedNode.ip && (
                    <div className="flex items-center justify-between">
                      <span className="text-slate-400">Indirizzo IP:</span>
                      <span className="font-mono text-slate-200">{selectedNode.ip}</span>
                    </div>
                  )}
                  {selectedNode.port && (
                    <div className="flex items-center justify-between">
                      <span className="text-slate-400">Porta & Protocollo:</span>
                      <span className="font-mono text-cyan-400">
                        {selectedNode.port}/{selectedNode.protocol || 'tcp'}
                      </span>
                    </div>
                  )}
                  {selectedNode.cvss && (
                    <div className="flex items-center justify-between">
                      <span className="text-slate-400">CVSS del finding:</span>
                      <span className="font-mono font-bold text-rose-400">
                        {selectedNode.cvss.toFixed(1)}
                      </span>
                    </div>
                  )}
                </div>

                {selectedNode.metadata && Object.keys(selectedNode.metadata).length > 0 && (
                  <div className="border-t border-slate-800 pt-3 text-xs space-y-2">
                    <h4 className="text-slate-200 font-semibold">Provenienza del finding</h4>
                    {Object.entries(selectedNode.metadata).map(([key, value]) => <div key={key} className="break-words">
                      <span className="text-slate-400">{key}: </span><span className="font-mono text-slate-200">{value === null || value === '' ? 'Non disponibile' : typeof value === 'object' ? JSON.stringify(value) : String(value)}</span>
                    </div>)}
                  </div>
                )}

                {/* Box se il nodo è anche un Chokepoint */}
                {graph.chokepoints?.some((cp) => cp.node_id === selectedNode.id) && (
                  <div className="p-3 bg-amber-500/10 border border-amber-500/30 rounded-xl space-y-1.5">
                    <div className="flex items-center gap-1.5 text-amber-300 text-xs font-bold">
                      <Target className="w-4 h-4 text-amber-400" />
                      <span>Chokepoint di Massima Priorità</span>
                    </div>
                    <p className="text-[11px] text-slate-300 leading-snug">
                      {
                        graph.chokepoints.find((cp) => cp.node_id === selectedNode.id)
                          ?.remediation
                      }
                    </p>
                    {mitigationAction(selectedNode)}
                  </div>
                )}

                <div className="pt-2 border-t border-slate-800 flex justify-between">
                  <button
                    onClick={() => handleSelectNode(null)}
                    className="text-xs text-slate-400 hover:text-white"
                  >
                    Deseleziona nodo
                  </button>
                  <button
                    onClick={() => setActiveTab('paths')}
                    className="text-xs text-cyan-400 hover:text-cyan-200 font-medium"
                  >
                    Torna ai percorsi
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
      {remediationNode && <ChokepointRemediationModal key={remediationNode.id} node={remediationNode} onClose={() => setRemediationNode(null)} onCheckFirewall={server => { setRemediationNode(null); setFirewallServer(server); }} />}
      {firewallServer && <FirewallManager server={firewallServer} onClose={() => { setFirewallServer(null); void fetchGraph(); }} />}
    </div>
  );
};
