import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  Crosshair,
  Plus,
  Search,
  Filter,
  Trash2,
  ShieldAlert,
  CheckCircle2,
  RefreshCw,
  AlertCircle,
  Copy,
  Check,
  Folder,
  X,
  ArrowRight,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetItem, TargetListResponseData, TargetGroupNode, TargetGroup } from '../types/target';
import { TargetModal } from '../components/TargetModal';
import { TreeExplorer, getGroupAndDescendantIds } from '../components/TreeExplorer';
import { useCompany } from '../context/CompanyContext';

export const Targets: React.FC = () => {
  const {
    selectedCompanyId,
    selectedCompany,
    setSelectedCompanyId,
    companyGroupIds,
    refreshCompanies,
  } = useCompany();

  const [targets, setTargets] = useState<TargetItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Selezione cartella / reparto Scope attiva
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
  const [selectedGroupNode, setSelectedGroupNode] = useState<TargetGroupNode | null>(null);
  const [flatGroups, setFlatGroups] = useState<Map<string, TargetGroup>>(new Map());

  // Filtri e ricerca
  const [search, setSearch] = useState<string>('');
  const [selectedType, setSelectedType] = useState<string>('ALL');
  const [selectedStatus, setSelectedStatus] = useState<string>('ALL');

  // Modali
  const [isCreateModalOpen, setIsCreateModalOpen] = useState<boolean>(false);
  const [targetToDelete, setTargetToDelete] = useState<TargetItem | null>(null);
  const [deleting, setDeleting] = useState<boolean>(false);

  // Selezione multipla per eliminazione massiva
  const [selectedTargetIds, setSelectedTargetIds] = useState<Set<string>>(new Set());
  const [isBulkDeleteModalOpen, setIsBulkDeleteModalOpen] = useState<boolean>(false);
  const [bulkDeleting, setBulkDeleting] = useState<boolean>(false);

  // Feedback copia
  const [copiedId, setCopiedId] = useState<string | null>(null);

  useEffect(() => {
    setSelectedGroupId(null);
    setSelectedGroupNode(null);
    setSelectedTargetIds(new Set());
    setTargetToDelete(null);
    setIsBulkDeleteModalOpen(false);
    setIsCreateModalOpen(false);
  }, [selectedCompanyId]);

  const fetchTargets = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<ApiResponse<TargetListResponseData>>('/targets?limit=200');
      if (response.data && response.data.success && response.data.data) {
        setTargets(response.data.data.items || []);
      } else {
        setError(response.data?.error?.message || 'Impossibile recuperare i target');
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Errore durante il caricamento dei bersagli'
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTargets();
  }, [fetchTargets]);

  // Callback quando i gruppi vengono caricati dal TreeExplorer
  const handleGroupsLoaded = useCallback((nodes: TargetGroupNode[]) => {
    const map = new Map<string, TargetGroup>();
    function recurse(list: TargetGroupNode[]) {
      for (const item of list) {
        map.set(item.id, item);
        if (item.children && item.children.length > 0) {
          recurse(item.children);
        }
      }
    }
    recurse(nodes);
    setFlatGroups(map);
  }, []);

  const handleDelete = async () => {
    if (!targetToDelete) return;
    setDeleting(true);
    try {
      const response = await apiClient.delete<ApiResponse<{ message: string }>>(`/targets/${targetToDelete.id}`);
      if (response.data && response.data.success) {
        setTargetToDelete(null);
        setSelectedTargetIds((prev) => {
          const next = new Set(prev);
          next.delete(targetToDelete.id);
          return next;
        });
        fetchTargets();
        void refreshCompanies();
      } else {
        alert(response.data?.error?.message || 'Impossibile eliminare il target');
      }
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Errore durante l\'eliminazione del target');
    } finally {
      setDeleting(false);
    }
  };

  const handleToggleSelectTarget = (id: string) => {
    setSelectedTargetIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  // Calcola gli ID ammissibili per la cartella selezionata (include tutti i sotto-rami)
  const activeSubtreeGroupIds = useMemo(() => {
    if (!selectedGroupId || !selectedGroupNode) return null;
    return new Set(getGroupAndDescendantIds(selectedGroupNode));
  }, [selectedGroupId, selectedGroupNode]);

  // Filtraggio client-side
  const filteredTargets = useMemo(() => {
    return targets.filter((t) => {
      // Filtro ricerca
      const matchesSearch =
        search.trim() === '' ||
        t.value.toLowerCase().includes(search.toLowerCase()) ||
        (t.description && t.description.toLowerCase().includes(search.toLowerCase()));

      // Filtro tipologia
      const matchesType =
        selectedType === 'ALL' || t.type.toUpperCase() === selectedType;

      // Filtro stato blacklist
      const matchesStatus =
        selectedStatus === 'ALL' ||
        (selectedStatus === 'AUTHORIZED' && !t.is_blacklisted) ||
        (selectedStatus === 'BLACKLIST' && t.is_blacklisted);

      // Filtro reparto scope (albero) o azienda attiva
      let matchesGroup = true;
      if (activeSubtreeGroupIds) {
        matchesGroup = t.group_id ? activeSubtreeGroupIds.has(t.group_id) : false;
      }
      if (selectedCompanyId) {
        matchesGroup = matchesGroup && Boolean(t.group_id && companyGroupIds.has(t.group_id));
      }

      return matchesSearch && matchesType && matchesStatus && matchesGroup;
    });
  }, [targets, search, selectedType, selectedStatus, activeSubtreeGroupIds, selectedCompanyId, companyGroupIds]);

  const isAllFilteredSelected = useMemo(() => {
    return (
      filteredTargets.length > 0 &&
      filteredTargets.every((t) => selectedTargetIds.has(t.id))
    );
  }, [filteredTargets, selectedTargetIds]);

  const handleSelectAllFiltered = () => {
    if (isAllFilteredSelected) {
      setSelectedTargetIds(new Set());
    } else {
      setSelectedTargetIds(new Set(filteredTargets.map((t) => t.id)));
    }
  };

  const handleClearSelection = () => {
    setSelectedTargetIds(new Set());
  };

  const handleBulkDelete = async () => {
    if (selectedTargetIds.size === 0) return;
    setBulkDeleting(true);
    try {
      const deletePromises = Array.from(selectedTargetIds).map((id) =>
        apiClient.delete<ApiResponse<{ message: string }>>(`/targets/${id}`)
      );
      await Promise.allSettled(deletePromises);
      setSelectedTargetIds(new Set());
      setIsBulkDeleteModalOpen(false);
      fetchTargets();
      void refreshCompanies();
    } catch (err: any) {
      alert(err.response?.data?.error?.message || 'Errore durante l\'eliminazione massiva dei target');
    } finally {
      setBulkDeleting(false);
    }
  };

  const handleCopy = (value: string, id: string) => {
    navigator.clipboard.writeText(value);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const getTypeBadge = (type: string) => {
    switch (type.toUpperCase()) {
      case 'IP':
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/30">
            IP
          </span>
        );
      case 'CIDR':
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/30">
            CIDR
          </span>
        );
      case 'RANGE':
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/30">
            RANGE
          </span>
        );
      case 'FQDN':
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
            FQDN
          </span>
        );
      default:
        return (
          <span className="px-2 py-0.5 rounded text-[11px] bg-slate-800 text-slate-300">
            {type}
          </span>
        );
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Crosshair className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-2xl font-bold tracking-tight text-white flex items-center space-x-2">
                <span>Target & Scope Management</span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-slate-800 text-cyan-400 border border-cyber-border font-mono">
                  {targets.length} totali
                </span>
              </h2>
              <p className="text-sm text-slate-400 mt-0.5">
                Organizzazione ad albero, perimetro autorizzato e analisi delle vulnerabilità per cartella
              </p>
            </div>
          </div>
        </div>

        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="self-start sm:self-auto flex items-center space-x-2 px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] hover:shadow-[0_0_20px_rgba(6,182,212,0.4)]"
        >
          <Plus className="w-4 h-4" />
          <span>Aggiungi Bersaglio</span>
        </button>
      </div>

      {/* Active Company Scope Banner */}
      {selectedCompany && (
        <div className="p-3.5 rounded-xl border border-cyan-500/40 bg-cyan-950/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs text-cyan-200">
          <div className="flex items-center space-x-2.5">
            <div
              className="w-3.5 h-3.5 rounded-full shrink-0 shadow-[0_0_8px_rgba(6,182,212,0.4)]"
              style={{ backgroundColor: selectedCompany.color_tag || '#06b6d4' }}
            />
            <span>
              Scope Aziendale Attivo: <strong className="text-white font-semibold">{selectedCompany.name}</strong> — Stai visualizzando e gestendo i reparti e i bersagli IP di questa specifica azienda.
            </span>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <button
              onClick={() => setSelectedCompanyId(null)}
              className="px-2.5 py-1 rounded-lg border border-cyan-500/40 bg-cyan-500/15 text-cyan-300 hover:bg-cyan-500/25 text-[11px] font-medium transition"
            >
              Mostra Tutte le Aziende
            </button>
            <Link
              to="/companies"
              className="px-2.5 py-1 rounded-lg border border-slate-700 bg-slate-800 text-slate-300 hover:text-white text-[11px] font-medium transition"
            >
              Gestione Aziende
            </Link>
          </div>
        </div>
      )}

      {/* Workflow Guidance Banner */}
      <div className="p-3.5 rounded-xl border border-blue-500/30 bg-blue-500/10 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs text-blue-200">
        <div className="flex items-center space-x-2.5">
          <Folder className="w-4 h-4 text-cyan-400 shrink-0" />
          <span>
            <strong>Gestione Bersagli & Aziende</strong>: Questa sezione è dedicata al censimento, organizzazione in reparti/aziende e blacklist degli host. Per avviare scansioni e audit di vulnerabilità, recati nella sezione dedicata.
          </span>
        </div>
        <Link
          to="/scans"
          className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs whitespace-nowrap transition-colors shadow-xs self-start sm:self-auto"
        >
          <span>Vai a Scansioni</span>
          <ArrowRight className="w-3.5 h-3.5" />
        </Link>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="p-4 rounded-xl border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
            <span>{error}</span>
          </div>
          <button
            onClick={fetchTargets}
            className="px-2.5 py-1 rounded bg-rose-900/50 hover:bg-rose-900 text-rose-200 font-medium"
          >
            Riprova
          </button>
        </div>
      )}

      {/* Layout a 2 Colonne: Albero a Sinistra, Tabella Bersagli a Destra */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Colonna Sinistra: Tree Explorer */}
        <div className="lg:col-span-4 xl:col-span-3">
          <TreeExplorer
            selectedGroupId={selectedGroupId}
            companyId={selectedCompanyId}
            onSelectGroup={(id, node) => {
              setSelectedGroupId(id);
              setSelectedGroupNode(node || null);
            }}
            onGroupsLoaded={handleGroupsLoaded}
            targets={targets}
          />
        </div>

        {/* Colonna Destra: Contenuto Target */}
        <div className="lg:col-span-8 xl:col-span-9 space-y-4">
          {/* Active Folder Filter Banner */}
          {selectedGroupNode && (
            <div className="p-3 rounded-xl border border-cyan-500/30 bg-cyan-500/10 flex items-center justify-between text-xs animate-in fade-in">
              <div className="flex items-center space-x-2.5">
                <div
                  className="p-1.5 rounded-lg border"
                  style={{
                    backgroundColor: `${selectedGroupNode.color_tag || '#06b6d4'}20`,
                    borderColor: `${selectedGroupNode.color_tag || '#06b6d4'}50`,
                    color: selectedGroupNode.color_tag || '#06b6d4',
                  }}
                >
                  <Folder className="w-4 h-4" />
                </div>
                <div>
                  <span className="text-slate-400 text-[11px]">Cartella Scope attiva: </span>
                  <span className="font-bold text-white ml-1">{selectedGroupNode.name}</span>
                  <span className="text-slate-400 text-[11px] ml-2">
                    ({filteredTargets.length} target in questo ramo)
                  </span>
                </div>
              </div>
              <button
                type="button"
                onClick={() => {
                  setSelectedGroupId(null);
                  setSelectedGroupNode(null);
                }}
                className="p-1 rounded-md text-slate-400 hover:text-white hover:bg-slate-800 flex items-center space-x-1 text-[11px] transition-colors"
                title="Rimuovi filtro cartella"
              >
                <span>Mostra tutti</span>
                <X className="w-3.5 h-3.5" />
              </button>
            </div>
          )}

          {/* Toolbar: Ricerca e Filtri */}
          <div className="p-4 rounded-xl border border-cyber-border bg-cyber-card/60 flex flex-col md:flex-row md:items-center justify-between gap-3">
            {/* Input Ricerca */}
            <div className="relative flex-1 max-w-md">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-500">
                <Search className="w-4 h-4" />
              </div>
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Cerca per IP, CIDR, dominio o descrizione..."
                className="w-full pl-9 pr-3 py-2 bg-slate-900/80 border border-cyber-border rounded-lg text-xs text-slate-200 placeholder-slate-500 focus:outline-hidden focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500"
              />
            </div>

            {/* Dropdown Filtri */}
            <div className="flex items-center space-x-3">
              {/* Filtro Tipo */}
              <div className="flex items-center space-x-1.5 text-xs">
                <Filter className="w-3.5 h-3.5 text-slate-400" />
                <select
                  value={selectedType}
                  onChange={(e) => setSelectedType(e.target.value)}
                  className="bg-slate-900 border border-cyber-border rounded-lg px-2.5 py-1.5 text-xs text-slate-300 focus:outline-hidden focus:border-cyan-500"
                >
                  <option value="ALL">Tutti i tipi</option>
                  <option value="IP">IP Singolo</option>
                  <option value="CIDR">Subnet CIDR</option>
                  <option value="RANGE">Range IP</option>
                  <option value="FQDN">Dominio FQDN</option>
                </select>
              </div>

              {/* Filtro Blacklist */}
              <select
                value={selectedStatus}
                onChange={(e) => setSelectedStatus(e.target.value)}
                className="bg-slate-900 border border-cyber-border rounded-lg px-2.5 py-1.5 text-xs text-slate-300 focus:outline-hidden focus:border-cyan-500"
              >
                <option value="ALL">Tutti gli stati</option>
                <option value="AUTHORIZED">Solo Autorizzati</option>
                <option value="BLACKLIST">Solo Blacklist</option>
              </select>

              {/* Refresh button */}
              <button
                onClick={fetchTargets}
                disabled={loading}
                className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
                title="Aggiorna lista"
              >
                <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
              </button>
            </div>
          </div>

          {/* Bulk Actions Floating Banner */}
          {selectedTargetIds.size > 0 && (
            <div className="p-3.5 rounded-xl border border-rose-500/30 bg-rose-950/40 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs animate-in fade-in">
              <div className="flex items-center space-x-2.5">
                <div className="p-1 rounded-md bg-rose-500/20 text-rose-400">
                  <Trash2 className="w-4 h-4" />
                </div>
                <div>
                  <span className="font-bold text-white">
                    {selectedTargetIds.size} bersaglio/i selezionato/i
                  </span>
                  <span className="text-slate-400 ml-1.5 text-[11px]">
                    (su {filteredTargets.length} visualizzati)
                  </span>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                <button
                  type="button"
                  onClick={handleClearSelection}
                  className="px-3 py-1.5 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
                >
                  Deseleziona
                </button>
                <button
                  type="button"
                  onClick={() => setIsBulkDeleteModalOpen(true)}
                  className="flex items-center space-x-1.5 px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 text-white font-bold text-xs transition-all shadow-[0_0_12px_rgba(225,29,72,0.3)]"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  <span>Elimina Selezionati ({selectedTargetIds.size})</span>
                </button>
              </div>
            </div>
          )}

          {/* Tabella Target */}
          <div className="rounded-xl border border-cyber-border bg-cyber-card/40 overflow-hidden shadow-xl">
            {loading ? (
              <div className="py-16 flex flex-col items-center justify-center text-slate-400 space-y-3">
                <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
                <span className="text-xs">Caricamento asset in corso...</span>
              </div>
            ) : filteredTargets.length === 0 ? (
              <div className="py-16 flex flex-col items-center justify-center text-center space-y-3 p-6">
                <div className="w-12 h-12 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-slate-500">
                  <Crosshair className="w-6 h-6" />
                </div>
                <div className="text-sm font-semibold text-slate-200">
                  {targets.length === 0
                    ? 'Nessun bersaglio censito'
                    : selectedGroupNode
                    ? `Nessun bersaglio trovato nella cartella "${selectedGroupNode.name}"`
                    : 'Nessun target corrisponde ai filtri'}
                </div>
                <p className="text-xs text-slate-400 max-w-sm">
                  {targets.length === 0
                    ? 'Inizia aggiungendo il primo indirizzo IP, blocco subnet o dominio autorizzato per le attività di sicurezza.'
                    : selectedGroupNode
                    ? 'Aggiungi un nuovo bersaglio direttamente in questa cartella per iniziare.'
                    : 'Prova a modificare i criteri di ricerca o azzerare i filtri applicati.'}
                </p>
                <button
                  onClick={() => setIsCreateModalOpen(true)}
                  className="mt-2 inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs transition-colors"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>
                    {selectedGroupNode
                      ? `Aggiungi Target a "${selectedGroupNode.name}"`
                      : 'Aggiungi il primo target'}
                  </span>
                </button>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="border-b border-cyber-border/80 text-[11px] font-semibold text-slate-400 uppercase tracking-wider bg-slate-900/40">
                      <th className="py-3 px-3 w-10 text-center">
                        <input
                          type="checkbox"
                          checked={isAllFilteredSelected}
                          onChange={handleSelectAllFiltered}
                          className="rounded border-cyber-border bg-slate-900 text-cyan-500 focus:ring-cyan-500 w-3.5 h-3.5 cursor-pointer accent-cyan-500"
                          title={isAllFilteredSelected ? 'Deseleziona tutti' : 'Seleziona tutti i target filtrati'}
                        />
                      </th>
                      <th className="py-3 px-4">Bersaglio / Valore</th>
                      <th className="py-3 px-4">Cartella Scope</th>
                      <th className="py-3 px-4">Tipo</th>
                      <th className="py-3 px-4">Descrizione</th>
                      <th className="py-3 px-4">Stato Autorizzativo</th>
                      <th className="py-3 px-4">Data Creazione</th>
                      <th className="py-3 px-4 text-right">Azioni</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-cyber-border/40 text-xs">
                    {filteredTargets.map((target) => {
                      const group = target.group_id ? flatGroups.get(target.group_id) : null;
                      const isSelected = selectedTargetIds.has(target.id);

                      return (
                        <tr
                          key={target.id}
                          className={`hover:bg-slate-800/25 transition-colors ${
                            isSelected ? 'bg-cyan-500/10' : ''
                          }`}
                        >
                          {/* Checkbox di Selezione */}
                          <td className="py-3.5 px-3 w-10 text-center">
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={() => handleToggleSelectTarget(target.id)}
                              className="rounded border-cyber-border bg-slate-900 text-cyan-500 focus:ring-cyan-500 w-3.5 h-3.5 cursor-pointer accent-cyan-500"
                            />
                          </td>

                          {/* Valore Target */}
                          <td className="py-3.5 px-4 font-mono font-medium text-slate-100">
                            <div className="flex items-center space-x-2">
                              <span className="text-cyan-400 font-semibold">{target.value}</span>
                              <button
                                onClick={() => handleCopy(target.value, target.id)}
                                className="text-slate-500 hover:text-slate-300 transition-colors p-1 rounded"
                                title="Copia valore"
                              >
                                {copiedId === target.id ? (
                                  <Check className="w-3 h-3 text-emerald-400" />
                                ) : (
                                  <Copy className="w-3 h-3" />
                                )}
                              </button>
                            </div>
                          </td>

                          {/* Cartella Scope */}
                          <td className="py-3.5 px-4 text-slate-300">
                            {group ? (
                              <span
                                className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-medium border"
                                style={{
                                  backgroundColor: `${group.color_tag || '#06b6d4'}15`,
                                  borderColor: `${group.color_tag || '#06b6d4'}35`,
                                  color: group.color_tag || '#06b6d4',
                                }}
                              >
                                <Folder className="w-3 h-3" />
                                <span className="truncate max-w-[120px]">{group.name}</span>
                              </span>
                            ) : (
                              <span className="text-slate-500 text-[11px] italic">
                                Non assegnato
                              </span>
                            )}
                          </td>

                          {/* Tipo */}
                          <td className="py-3.5 px-4">{getTypeBadge(target.type)}</td>

                          {/* Descrizione */}
                          <td className="py-3.5 px-4 text-slate-300 max-w-xs truncate">
                            {target.description || <span className="text-slate-500 italic">-</span>}
                          </td>

                          {/* Stato Blacklist */}
                          <td className="py-3.5 px-4">
                            {target.is_blacklisted ? (
                              <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/30">
                                <ShieldAlert className="w-3 h-3" />
                                <span>Blacklist (Escluso)</span>
                              </span>
                            ) : (
                              <span className="inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                                <CheckCircle2 className="w-3 h-3" />
                                <span>Autorizzato</span>
                              </span>
                            )}
                          </td>

                          {/* Data Creazione */}
                          <td className="py-3.5 px-4 text-slate-400">
                            {new Date(target.created_at).toLocaleDateString('it-IT', {
                              day: '2-digit',
                              month: '2-digit',
                              year: 'numeric',
                            })}
                          </td>

                          {/* Azioni */}
                          <td className="py-3.5 px-4 text-right">
                            <div className="flex items-center justify-end space-x-2">
                              {/* Delete */}
                              <button
                                onClick={() => setTargetToDelete(target)}
                                className="p-1.5 rounded-lg border border-rose-900/40 bg-rose-950/20 text-rose-400 hover:bg-rose-900/30 hover:border-rose-800 transition-colors"
                                title="Elimina Bersaglio"
                              >
                                <Trash2 className="w-3.5 h-3.5" />
                              </button>
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
        </div>
      </div>

      {/* Modale Creazione Target */}
      <TargetModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onTargetCreated={() => { void fetchTargets(); void refreshCompanies(); }}
        defaultGroupId={selectedGroupId || selectedCompanyId}
      />



      {/* Conferma Eliminazione Target Dialog */}
      {targetToDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-xs animate-in fade-in duration-150">
          <div className="max-w-md w-full p-6 rounded-2xl border border-rose-800/40 bg-[#111726] shadow-2xl space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400">
                <Trash2 className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-white">Conferma Eliminazione Target</h3>
                <p className="text-xs text-slate-400">Questa operazione non può essere annullata</p>
              </div>
            </div>

            <p className="text-xs text-slate-300">
              Sei sicuro di voler rimuovere il bersaglio{' '}
              <strong className="font-mono text-cyan-400">{targetToDelete.value}</strong> dallo scope dell'organizzazione?
            </p>

            <div className="flex items-center justify-end space-x-3 pt-2">
              <button
                type="button"
                onClick={() => setTargetToDelete(null)}
                disabled={deleting}
                className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
              >
                Annulla
              </button>
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleting}
                className="px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold transition-colors disabled:opacity-50"
              >
                {deleting ? 'Eliminazione...' : 'Elimina Definitivamente'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Conferma Eliminazione Multipla Target Dialog */}
      {isBulkDeleteModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-xs animate-in fade-in duration-150">
          <div className="max-w-md w-full p-6 rounded-2xl border border-rose-800/40 bg-[#111726] shadow-2xl space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400">
                <Trash2 className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-white">Conferma Eliminazione Multipla</h3>
                <p className="text-xs text-slate-400">Questa operazione non può essere annullata</p>
              </div>
            </div>

            <p className="text-xs text-slate-300 leading-relaxed">
              Sei sicuro di voler rimuovere <strong className="font-mono text-rose-400 font-bold">{selectedTargetIds.size}</strong> bersagli
              selezionati dallo scope dell'organizzazione?
            </p>

            <div className="flex items-center justify-end space-x-3 pt-2">
              <button
                type="button"
                onClick={() => setIsBulkDeleteModalOpen(false)}
                disabled={bulkDeleting}
                className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
              >
                Annulla
              </button>
              <button
                type="button"
                onClick={handleBulkDelete}
                disabled={bulkDeleting}
                className="px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold transition-colors disabled:opacity-50"
              >
                {bulkDeleting ? 'Eliminazione in corso...' : `Elimina Definitivamente (${selectedTargetIds.size})`}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
