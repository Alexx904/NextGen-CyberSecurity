import React, { useState, useEffect } from 'react';
import { X, Play, AlertCircle, Check, Folder, Layers } from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';

interface ScanProfileItem {
  id: string;
  name: string;
  nmap_args: string;
  description: string;
  timeout_seconds: number;
}

interface GroupScanModalProps {
  isOpen: boolean;
  onClose: () => void;
  groupId: string;
  groupName: string;
  targetCount: number;
  onScanStarted: () => void;
}

export const GroupScanModal: React.FC<GroupScanModalProps> = ({
  isOpen,
  onClose,
  groupId,
  groupName,
  targetCount,
  onScanStarted,
}) => {
  const [profiles, setProfiles] = useState<ScanProfileItem[]>([]);
  const [selectedProfileId, setSelectedProfileId] = useState<string>('');
  const [loadingProfiles, setLoadingProfiles] = useState<boolean>(false);
  const [submitting, setSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setError(null);
      setLoadingProfiles(true);

      apiClient
        .get<ApiResponse<{ items: ScanProfileItem[]; total: number }>>('/scan-profiles')
        .then((res) => {
          if (res.data?.data?.items) {
            const list = res.data.data.items;
            setProfiles(list);
            const fast = list.find((p) => p.name.includes('Fast')) || list[0];
            if (fast) setSelectedProfileId(fast.id);
          }
        })
        .catch((err) => {
          setError(
            err.response?.data?.error?.message ||
              err.message ||
              'Impossibile recuperare i profili di scansione'
          );
        })
        .finally(() => {
          setLoadingProfiles(false);
        });
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!selectedProfileId) {
      setError('Seleziona un profilo di scansione Nmap');
      return;
    }

    if (targetCount === 0) {
      setError('Questa cartella e i suoi sotto-rami non contengono target scansionabili');
      return;
    }

    setSubmitting(true);
    try {
      const response = await apiClient.post<
        ApiResponse<{
          group_id: string;
          profile_id: string;
          scans_count: number;
          created_scans: any[];
        }>
      >(`/target-groups/${groupId}/scan`, {
        profile_id: selectedProfileId,
      });

      if (response.data && response.data.success) {
        onScanStarted();
        onClose();
      } else {
        setError(
          response.data?.error?.message ||
            'Errore durante l\'avvio della scansione per l\'intera cartella'
        );
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
          err.message ||
          'Impossibile avviare la scansione di gruppo'
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="max-w-lg w-full bg-[#111726] border border-cyber-border rounded-2xl shadow-2xl overflow-hidden relative">
        {/* Header */}
        <div className="p-5 border-b border-cyber-border flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Layers className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">Scansione Ricorsiva Cartella</h3>
              <p className="text-xs text-slate-400">Avvia una scansione per tutti i target nel ramo</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Info Box Gruppo */}
        <div className="px-6 pt-5 pb-2">
          <div className="p-3.5 rounded-xl border border-cyber-border bg-slate-900/80 flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
                <Folder className="w-4 h-4" />
              </div>
              <div>
                <span className="text-xs font-semibold text-white block">{groupName}</span>
                <span className="text-[11px] text-slate-400">
                  Include la cartella e tutte le relative sottocartelle
                </span>
              </div>
            </div>
            <div className="text-right">
              <span className="text-xs font-bold text-cyan-400 block font-mono">
                {targetCount} target
              </span>
              <span className="text-[10px] text-slate-500">rilevati</span>
            </div>
          </div>
        </div>

        {/* Body Form */}
        <form onSubmit={handleSubmit} className="p-6 pt-3 space-y-4">
          {error && (
            <div className="p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-start space-x-2">
              <AlertCircle className="w-4 h-4 shrink-0 text-rose-400 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {loadingProfiles ? (
            <div className="py-8 text-center text-xs text-slate-400">
              Caricamento profili Nmap disponibili...
            </div>
          ) : (
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-2">
                Seleziona Profilo di Scansione Nmap <span className="text-red-400">*</span>
              </label>
              <div className="space-y-2 max-h-56 overflow-y-auto pr-1">
                {profiles.map((p) => {
                  const isSelected = selectedProfileId === p.id;
                  return (
                    <div
                      key={p.id}
                      onClick={() => setSelectedProfileId(p.id)}
                      className={`p-3 rounded-xl border cursor-pointer transition-all ${
                        isSelected
                          ? 'border-cyan-500 bg-cyan-500/10 shadow-[0_0_15px_rgba(6,182,212,0.15)]'
                          : 'border-cyber-border bg-slate-900/60 hover:border-slate-700'
                      }`}
                    >
                      <div className="flex items-center justify-between text-xs">
                        <div className="flex items-center space-x-2">
                          <span
                            className={`w-3.5 h-3.5 rounded-full border flex items-center justify-center ${
                              isSelected ? 'border-cyan-400 bg-cyan-400' : 'border-slate-600'
                            }`}
                          >
                            {isSelected && <Check className="w-2.5 h-2.5 text-slate-950" />}
                          </span>
                          <span className="font-semibold text-white">{p.name}</span>
                        </div>
                        <span className="text-[10px] text-slate-400 font-mono">
                          timeout: {Math.round(p.timeout_seconds / 60)}m
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-400 mt-1 pl-5.5">{p.description}</p>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {targetCount === 0 && (
            <div className="p-3 rounded-lg border border-amber-800/40 bg-amber-950/20 text-amber-300 text-xs flex items-center space-x-2">
              <AlertCircle className="w-4 h-4 shrink-0 text-amber-400" />
              <span>Attenzione: non ci sono target in questa cartella o nelle sue sottocartelle.</span>
            </div>
          )}

          {/* Footer Actions */}
          <div className="pt-2 flex items-center justify-end space-x-3 border-t border-cyber-border/60">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
            >
              Annulla
            </button>
            <button
              type="submit"
              disabled={submitting || targetCount === 0 || !selectedProfileId}
              className="px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 text-xs font-semibold transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-1.5"
            >
              <Play className="w-3.5 h-3.5" />
              <span>{submitting ? 'Avvio in corso...' : `Avvia Scansione (${targetCount} Target)`}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
