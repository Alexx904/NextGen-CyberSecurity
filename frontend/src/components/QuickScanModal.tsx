import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { X, Play, Clock, Terminal, AlertCircle } from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetItem } from '../types/target';
import { ScanItem } from '../types/scan';

interface ScanProfileItem {
  id: string;
  name: string;
  nmap_args: string;
  description: string;
  timeout_seconds: number;
}

interface QuickScanModalProps {
  isOpen: boolean;
  target: TargetItem | null;
  onClose: () => void;
  onScanLaunched?: () => void;
}

export const QuickScanModal: React.FC<QuickScanModalProps> = ({
  isOpen,
  target,
  onClose,
  onScanLaunched,
}) => {
  const navigate = useNavigate();
  const [profiles, setProfiles] = useState<ScanProfileItem[]>([]);
  const [selectedProfileId, setSelectedProfileId] = useState<string>('');
  const [loadingProfiles, setLoadingProfiles] = useState<boolean>(false);
  const [launching, setLaunching] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && target) {
      setError(null);
      setLoadingProfiles(true);
      apiClient
        .get<ApiResponse<{ items: ScanProfileItem[]; total: number }>>('/scan-profiles')
        .then((res) => {
          if (res.data?.data?.items) {
            const list = res.data.data.items;
            setProfiles(list);
            // Pre-seleziona Fast Discovery se presente, altrimenti il primo
            const fast = list.find((p) => p.name.includes('Fast')) || list[0];
            if (fast) setSelectedProfileId(fast.id);
          }
        })
        .catch((err) => {
          setError(err.response?.data?.error?.message || 'Impossibile caricare i profili di scansione');
        })
        .finally(() => {
          setLoadingProfiles(false);
        });
    }
  }, [isOpen, target]);

  if (!isOpen || !target) return null;

  const handleLaunch = async () => {
    if (!selectedProfileId) {
      setError('Seleziona un profilo di scansione');
      return;
    }

    setLaunching(true);
    setError(null);
    try {
      const response = await apiClient.post<ApiResponse<ScanItem>>('/scans', {
        target_id: target.id,
        profile_id: selectedProfileId,
      });

      if (response.data && response.data.success) {
        if (onScanLaunched) onScanLaunched();
        onClose();
        navigate('/scans');
      } else {
        setError(response.data?.error?.message || 'Errore durante il lancio della scansione');
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile avviare la scansione'
      );
    } finally {
      setLaunching(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="max-w-lg w-full bg-[#111726] border border-cyber-border rounded-2xl shadow-2xl overflow-hidden relative">
        {/* Header */}
        <div className="p-5 border-b border-cyber-border flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Play className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">Scansione Rapida Nmap</h3>
              <p className="text-xs text-slate-400">
                Bersaglio: <span className="font-mono text-cyan-400 font-semibold">{target.value}</span> ({target.type})
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-5">
          {error && (
            <div className="p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-start space-x-2">
              <AlertCircle className="w-4 h-4 shrink-0 text-rose-400 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-2">
              Seleziona Profilo di Scansione
            </label>

            {loadingProfiles ? (
              <div className="p-6 text-center text-xs text-slate-400">
                Caricamento profili in corso...
              </div>
            ) : profiles.length === 0 ? (
              <div className="p-4 rounded-lg bg-slate-900 border border-cyber-border text-xs text-slate-400">
                Nessun profilo di scansione disponibile nel sistema.
              </div>
            ) : (
              <div className="space-y-2.5 max-h-64 overflow-y-auto pr-1">
                {profiles.map((profile) => {
                  const isSelected = selectedProfileId === profile.id;
                  return (
                    <div
                      key={profile.id}
                      onClick={() => setSelectedProfileId(profile.id)}
                      className={`p-3.5 rounded-xl border cursor-pointer transition-all ${
                        isSelected
                          ? 'border-cyan-500 bg-cyan-500/10 shadow-[0_0_15px_rgba(6,182,212,0.15)]'
                          : 'border-cyber-border bg-slate-900/60 hover:border-slate-700'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <span
                            className={`w-3.5 h-3.5 rounded-full border flex items-center justify-center ${
                              isSelected
                                ? 'border-cyan-400 bg-cyan-400'
                                : 'border-slate-600'
                            }`}
                          >
                            {isSelected && <span className="w-1.5 h-1.5 rounded-full bg-slate-950" />}
                          </span>
                          <span className="font-semibold text-xs text-white">
                            {profile.name}
                          </span>
                        </div>
                        <span className="text-[10px] text-slate-400 font-mono flex items-center space-x-1">
                          <Clock className="w-3 h-3" />
                          <span>{Math.round(profile.timeout_seconds / 60)} min max</span>
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-400 mt-1 pl-5.5">
                        {profile.description}
                      </p>
                      <div className="mt-2 pl-5.5 flex items-center space-x-2">
                        <span className="text-[10px] text-slate-500 flex items-center space-x-1 font-mono bg-slate-950 px-2 py-0.5 rounded border border-cyber-border">
                          <Terminal className="w-2.5 h-2.5 text-cyan-400" />
                          <span>nmap {profile.nmap_args}</span>
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Footer Actions */}
          <div className="pt-2 flex items-center justify-end space-x-3 border-t border-cyber-border/60">
            <button
              type="button"
              onClick={onClose}
              disabled={launching}
              className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
            >
              Annulla
            </button>
            <button
              type="button"
              onClick={handleLaunch}
              disabled={launching || !selectedProfileId}
              className="px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 text-xs font-semibold transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-1.5"
            >
              <Play className="w-3.5 h-3.5" />
              <span>{launching ? 'Avvio scansione...' : 'Avvia Scansione Ora'}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
