import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  Folder,
  FolderOpen,
  ChevronRight,
  ChevronDown,
  Plus,
  Play,
  Edit2,
  Trash2,
  RefreshCw,
  Layers,
  AlertCircle,
  Globe,
  FolderPlus,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetGroup, TargetGroupNode } from '../types/target';
import { GroupModal } from './GroupModal';
import { GroupScanModal } from './GroupScanModal';

interface TreeExplorerProps {
  selectedGroupId: string | null;
  onSelectGroup: (groupId: string | null, groupNode?: TargetGroupNode | null) => void;
  onGroupsLoaded?: (groups: TargetGroupNode[]) => void;
  className?: string;
}

// Funzione helper esportata per ricavare l'ID del gruppo e di tutti i suoi discendenti
export function getGroupAndDescendantIds(group: TargetGroupNode): string[] {
  const ids: string[] = [group.id];
  if (group.children && group.children.length > 0) {
    for (const child of group.children) {
      ids.push(...getGroupAndDescendantIds(child));
    }
  }
  return ids;
}

// Appiattisce l'albero per il selettore parent in GroupModal
function flattenTree(nodes: TargetGroupNode[]): TargetGroup[] {
  const result: TargetGroup[] = [];
  function recurse(list: TargetGroupNode[]) {
    for (const n of list) {
      result.push(n);
      if (n.children && n.children.length > 0) {
        recurse(n.children);
      }
    }
  }
  recurse(nodes);
  return result;
}

export const TreeExplorer: React.FC<TreeExplorerProps> = ({
  selectedGroupId,
  onSelectGroup,
  onGroupsLoaded,
  className = '',
}) => {
  const [treeData, setTreeData] = useState<TargetGroupNode[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Espansione nodi
  const [expandedNodeIds, setExpandedNodeIds] = useState<Set<string>>(new Set());

  // Stato Modali
  const [isGroupModalOpen, setIsGroupModalOpen] = useState<boolean>(false);
  const [groupToEdit, setGroupToEdit] = useState<TargetGroup | null>(null);
  const [defaultParentId, setDefaultParentId] = useState<string | null>(null);

  // Modale Scansione Ricorsiva
  const [scanModalGroup, setScanModalGroup] = useState<TargetGroupNode | null>(null);

  // Modale Eliminazione
  const [groupToDelete, setGroupToDelete] = useState<TargetGroupNode | null>(null);
  const [deleting, setDeleting] = useState<boolean>(false);

  const fetchTree = useCallback(async (silent = false) => {
    if (!silent) setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups');
      if (response.data && response.data.success && response.data.data) {
        const nodes = response.data.data;
        setTreeData(nodes);
        if (onGroupsLoaded) {
          onGroupsLoaded(nodes);
        }

        // Auto-espandi i primi nodi se non ancora espansi
        setExpandedNodeIds((prev) => {
          if (prev.size === 0) {
            const initial = new Set<string>();
            nodes.forEach((n) => initial.add(n.id));
            return initial;
          }
          return prev;
        });
      } else {
        setError(response.data?.error?.message || 'Impossibile recuperare i gruppi');
      }
    } catch (err: any) {
      if (!silent) {
        setError(
          err.response?.data?.error?.message ||
            err.message ||
            'Errore durante il caricamento della gerarchia gruppi'
        );
      }
    } finally {
      if (!silent) setLoading(false);
    }
  }, [onGroupsLoaded]);

  useEffect(() => {
    fetchTree();
  }, [fetchTree]);

  const toggleExpand = (nodeId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setExpandedNodeIds((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  };

  const handleOpenCreateRoot = () => {
    setGroupToEdit(null);
    setDefaultParentId(null);
    setIsGroupModalOpen(true);
  };

  const handleOpenCreateSub = (parent: TargetGroupNode, e: React.MouseEvent) => {
    e.stopPropagation();
    setGroupToEdit(null);
    setDefaultParentId(parent.id);
    setIsGroupModalOpen(true);
    // Espandi anche il nodo padre
    setExpandedNodeIds((prev) => new Set(prev).add(parent.id));
  };

  const handleOpenEdit = (group: TargetGroupNode, e: React.MouseEvent) => {
    e.stopPropagation();
    setGroupToEdit(group);
    setDefaultParentId(group.parent_id || null);
    setIsGroupModalOpen(true);
  };

  const handleOpenScan = (group: TargetGroupNode, e: React.MouseEvent) => {
    e.stopPropagation();
    setScanModalGroup(group);
  };

  const handleOpenDelete = (group: TargetGroupNode, e: React.MouseEvent) => {
    e.stopPropagation();
    setGroupToDelete(group);
  };

  const handleDeleteConfirm = async () => {
    if (!groupToDelete) return;
    setDeleting(true);
    try {
      const response = await apiClient.delete<ApiResponse<{ message: string }>>(
        `/target-groups/${groupToDelete.id}`
      );
      if (response.data && response.data.success) {
        if (selectedGroupId === groupToDelete.id) {
          onSelectGroup(null, null);
        }
        setGroupToDelete(null);
        fetchTree(true);
      } else {
        alert(response.data?.error?.message || 'Impossibile eliminare la cartella');
      }
    } catch (err: any) {
      alert(
        err.response?.data?.error?.message ||
          err.message ||
          'Errore durante l\'eliminazione della cartella'
      );
    } finally {
      setDeleting(false);
    }
  };

  const allFlatGroups = useMemo(() => flattenTree(treeData), [treeData]);

  // Calcolo totale complessivo target
  const totalScopeTargets = useMemo(() => {
    return treeData.reduce((acc, n) => acc + (n.target_count || 0), 0);
  }, [treeData]);

  const renderRiskBadge = (risk: string) => {
    switch (risk?.toUpperCase()) {
      case 'CRITICAL':
        return (
          <span className="px-1.5 py-0.2 rounded text-[10px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/30">
            CRIT
          </span>
        );
      case 'HIGH':
        return (
          <span className="px-1.5 py-0.2 rounded text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/30">
            HIGH
          </span>
        );
      case 'MEDIUM':
        return (
          <span className="px-1.5 py-0.2 rounded text-[10px] font-semibold bg-yellow-500/10 text-yellow-400 border border-yellow-500/30">
            MED
          </span>
        );
      case 'LOW':
        return (
          <span className="px-1.5 py-0.2 rounded text-[10px] font-medium bg-blue-500/10 text-blue-400 border border-blue-500/30">
            LOW
          </span>
        );
      case 'INFO':
        return (
          <span className="px-1.5 py-0.2 rounded text-[10px] font-medium bg-cyan-500/10 text-cyan-400 border border-cyan-500/30">
            INFO
          </span>
        );
      default:
        return null;
    }
  };

  const renderNode = (node: TargetGroupNode, depth = 0) => {
    const hasChildren = node.children && node.children.length > 0;
    const isExpanded = expandedNodeIds.has(node.id);
    const isSelected = selectedGroupId === node.id;
    const nodeColor = node.color_tag || '#06b6d4';

    return (
      <div key={node.id} className="select-none">
        <div
          onClick={() => onSelectGroup(node.id, node)}
          className={`group flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs cursor-pointer transition-all ${
            isSelected
              ? 'bg-cyan-500/15 border border-cyan-500/30 text-white font-medium shadow-[0_0_12px_rgba(6,182,212,0.15)]'
              : 'hover:bg-slate-800/60 text-slate-300 border border-transparent'
          }`}
          style={{ paddingLeft: `${Math.max(10, depth * 18 + 10)}px` }}
        >
          {/* Sinistra: Chevron + Icona Cartella + Nome */}
          <div className="flex items-center space-x-1.5 min-w-0 mr-2">
            {hasChildren ? (
              <button
                type="button"
                onClick={(e) => toggleExpand(node.id, e)}
                className="p-0.5 rounded hover:bg-slate-700/60 text-slate-400 hover:text-white transition-colors -ml-1"
                title={isExpanded ? 'Comprimi' : 'Espandi'}
              >
                {isExpanded ? (
                  <ChevronDown className="w-3.5 h-3.5" />
                ) : (
                  <ChevronRight className="w-3.5 h-3.5" />
                )}
              </button>
            ) : (
              <span className="w-3.5 inline-block -ml-1" />
            )}

            <span
              className="shrink-0 transition-transform group-hover:scale-105"
              style={{ color: nodeColor }}
            >
              {isExpanded && hasChildren ? (
                <FolderOpen className="w-4 h-4" />
              ) : (
                <Folder className="w-4 h-4" />
              )}
            </span>

            <span
              className={`truncate text-xs ${isSelected ? 'text-cyan-300 font-semibold' : 'text-slate-200'}`}
              title={node.name}
            >
              {node.name}
            </span>
          </div>

          {/* Destra: Badges & Hover Actions */}
          <div className="flex items-center space-x-1 shrink-0">
            {/* Risk Badge */}
            {renderRiskBadge(node.aggregate_risk)}

            {/* Target count badge */}
            <span
              className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-900 border border-cyber-border text-slate-400"
              title={`${node.target_count} target totali (${node.direct_target_count} diretti)`}
            >
              {node.target_count}
            </span>

            {/* Quick Actions (Visibili all'hover o se selezionato) */}
            <div className="hidden group-hover:flex items-center space-x-0.5 bg-slate-900/90 rounded px-1 border border-cyber-border/80 ml-1">
              {/* Nuova Sottocartella */}
              <button
                type="button"
                onClick={(e) => handleOpenCreateSub(node, e)}
                className="p-1 rounded text-slate-400 hover:text-cyan-400 hover:bg-slate-800 transition-colors"
                title="Nuova sottocartella"
              >
                <Plus className="w-3 h-3" />
              </button>

              {/* Scansiona cartella */}
              <button
                type="button"
                onClick={(e) => handleOpenScan(node, e)}
                className="p-1 rounded text-slate-400 hover:text-emerald-400 hover:bg-slate-800 transition-colors"
                title="Avvia scansione ricorsiva cartella"
              >
                <Play className="w-3 h-3" />
              </button>

              {/* Modifica */}
              <button
                type="button"
                onClick={(e) => handleOpenEdit(node, e)}
                className="p-1 rounded text-slate-400 hover:text-amber-400 hover:bg-slate-800 transition-colors"
                title="Modifica cartella"
              >
                <Edit2 className="w-3 h-3" />
              </button>

              {/* Elimina */}
              <button
                type="button"
                onClick={(e) => handleOpenDelete(node, e)}
                className="p-1 rounded text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition-colors"
                title="Elimina cartella"
              >
                <Trash2 className="w-3 h-3" />
              </button>
            </div>
          </div>
        </div>

        {/* Nodi Figli */}
        {hasChildren && isExpanded && (
          <div className="relative border-l border-slate-800/80 ml-4 my-0.5">
            {node.children.map((child) => renderNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div
      className={`rounded-xl border border-cyber-border bg-[#111726]/80 flex flex-col overflow-hidden shadow-lg ${className}`}
    >
      {/* Header Widget */}
      <div className="p-3.5 border-b border-cyber-border/80 bg-slate-900/60 flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Layers className="w-4 h-4 text-cyan-400" />
          <h3 className="text-xs font-bold text-white tracking-wide uppercase">
            Albero Scope & Cartelle
          </h3>
        </div>

        <div className="flex items-center space-x-1">
          <button
            type="button"
            onClick={handleOpenCreateRoot}
            className="p-1.5 rounded-lg border border-cyan-500/30 bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 text-[11px] font-semibold flex items-center space-x-1 transition-colors"
            title="Aggiungi Cartella Principale"
          >
            <FolderPlus className="w-3.5 h-3.5" />
            <span className="hidden sm:inline">Nuova</span>
          </button>

          <button
            type="button"
            onClick={() => fetchTree(false)}
            disabled={loading}
            className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
            title="Aggiorna Albero"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
          </button>
        </div>
      </div>

      {/* Body Widget */}
      <div className="p-2 space-y-1 flex-1 overflow-y-auto max-h-[620px]">
        {error && (
          <div className="p-3 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center justify-between mb-2">
            <div className="flex items-center space-x-2">
              <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
              <span>{error}</span>
            </div>
            <button
              onClick={() => fetchTree(false)}
              className="text-[11px] px-2 py-0.5 rounded bg-rose-900/60 hover:bg-rose-900 text-rose-200"
            >
              Riprova
            </button>
          </div>
        )}

        {/* Voce Radice: "Tutti i Bersagli" */}
        <div
          onClick={() => onSelectGroup(null, null)}
          className={`flex items-center justify-between px-3 py-2 rounded-lg text-xs cursor-pointer transition-all ${
            selectedGroupId === null
              ? 'bg-cyan-500/15 border border-cyan-500/30 text-cyan-300 font-semibold shadow-[0_0_12px_rgba(6,182,212,0.15)]'
              : 'hover:bg-slate-800/60 text-slate-300 border border-transparent'
          }`}
        >
          <div className="flex items-center space-x-2">
            <Globe className="w-4 h-4 text-cyan-400" />
            <span className="font-medium">Tutti i Target</span>
          </div>
          <span className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-900 border border-cyber-border text-slate-400">
            {totalScopeTargets}
          </span>
        </div>

        {/* Caricamento */}
        {loading && treeData.length === 0 ? (
          <div className="py-8 text-center text-slate-400 text-xs space-y-2">
            <RefreshCw className="w-5 h-5 animate-spin text-cyan-400 mx-auto" />
            <p>Caricamento struttura gerarchica...</p>
          </div>
        ) : treeData.length === 0 ? (
          <div className="py-8 text-center p-4">
            <div className="w-10 h-10 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-slate-500 mx-auto mb-2">
              <Folder className="w-5 h-5" />
            </div>
            <p className="text-xs text-slate-300 font-medium">Nessuna cartella creata</p>
            <p className="text-[11px] text-slate-500 mt-1">
              Organizza la tua infrastruttura in cartelle logiche per facilitare le scansioni.
            </p>
            <button
              onClick={handleOpenCreateRoot}
              className="mt-3 px-3 py-1.5 rounded-lg bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-400 border border-cyan-500/30 text-xs font-semibold inline-flex items-center space-x-1"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Crea Prima Cartella</span>
            </button>
          </div>
        ) : (
          <div className="space-y-0.5 pt-1">
            {treeData.map((node) => renderNode(node, 0))}
          </div>
        )}
      </div>

      {/* Modale Creazione / Modifica Gruppo */}
      <GroupModal
        isOpen={isGroupModalOpen}
        onClose={() => {
          setIsGroupModalOpen(false);
          setGroupToEdit(null);
          setDefaultParentId(null);
        }}
        onGroupSaved={() => {
          fetchTree(true);
        }}
        groupToEdit={groupToEdit}
        defaultParentId={defaultParentId}
        existingGroups={allFlatGroups}
      />

      {/* Modale Scansione Ricorsiva Cartella */}
      {scanModalGroup && (
        <GroupScanModal
          isOpen={!!scanModalGroup}
          onClose={() => setScanModalGroup(null)}
          groupId={scanModalGroup.id}
          groupName={scanModalGroup.name}
          targetCount={scanModalGroup.target_count}
          onScanStarted={() => {
            fetchTree(true);
          }}
        />
      )}

      {/* Dialog Conferma Eliminazione Cartella */}
      {groupToDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150">
          <div className="max-w-md w-full p-6 rounded-2xl border border-rose-800/40 bg-[#111726] shadow-2xl space-y-4">
            <div className="flex items-center space-x-3">
              <div className="p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400">
                <Trash2 className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-white">Elimina Cartella Scope</h3>
                <p className="text-xs text-slate-400">Conferma rimozione del gruppo</p>
              </div>
            </div>

            <p className="text-xs text-slate-300">
              Sei sicuro di voler eliminare la cartella{' '}
              <strong className="text-cyan-400 font-semibold">"{groupToDelete.name}"</strong>?
            </p>

            <div className="p-3 rounded-lg border border-amber-800/40 bg-amber-950/20 text-amber-300 text-[11px] leading-relaxed">
              I target e le eventuali sottocartelle non verranno eliminati: i target associati saranno
              spostati alla cartella radice (non assegnati), e i sotto-rami saranno riagganciati
              automaticamente.
            </div>

            <div className="flex items-center justify-end space-x-3 pt-2">
              <button
                type="button"
                onClick={() => setGroupToDelete(null)}
                disabled={deleting}
                className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
              >
                Annulla
              </button>
              <button
                type="button"
                onClick={handleDeleteConfirm}
                disabled={deleting}
                className="px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold transition-colors disabled:opacity-50"
              >
                {deleting ? 'Eliminazione...' : 'Elimina Cartella'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

