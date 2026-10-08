import React, { useState, useEffect } from 'react';
import { X, Play, Calendar, Terminal, AlertCircle, Check } from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetItem, TargetListResponseData } from '../types/target';
import { ScanItem } from '../types/scan';
import { ScheduleItem } from '../types/schedule';

interface ScanProfileItem {
  id: string;
  name: string;
  nmap_args: string;
  description: string;
  timeout_seconds: number;
}

interface NewScanModalProps {
  isOpen: boolean;
  onClose: () => void;
  onScanLaunched: () => void;
  onScheduleCreated?: () => void;
}

const CRON_PRESETS = [
  { label: 'Ogni 6 ore', value: '0 */6 * * *', desc: '4 volte al giorno' },
  { label: 'Ogni notte alle 02:00', value: '0 2 * * *', desc: 'Scansione notturna quotidiana' },
  { label: 'Ogni Lunedì alle 03:00', value: '0 3 * * 1', desc: 'Scansione settimanale di audit' },
  { label: 'Primo giorno del mese (04:00)', value: '0 4 1 * *', desc: 'Verifica mensile di conformità' },
  { label: 'Personalizzato...', value: 'CUSTOM', desc: 'Definisci una stringa cron a 5 campi' },
];

export const NewScanModal: React.FC<NewScanModalProps> = ({
  isOpen,
  onClose,
  onScanLaunched,
  onScheduleCreated,
}) => {
  const [activeTab, setActiveTab] = useState<'ON_DEMAND' | 'SCHEDULED'>('ON_DEMAND');

  const [targets, setTargets] = useState<TargetItem[]>([]);
  const [profiles, setProfiles] = useState<ScanProfileItem[]>([]);
  const [loadingData, setLoadingData] = useState<boolean>(false);

  const [selectedTargetId, setSelectedTargetId] = useState<string>('');
  const [selectedProfileId, setSelectedProfileId] = useState<string>('');

  // Campi Scheduler
  const [selectedCronPreset, setSelectedCronPreset] = useState<string>('0 2 * * *');
  const [customCron, setCustomCron] = useState<string>('0 2 * * *');

  const [submitting, setSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setError(null);
      setLoadingData(true);

      Promise.all([
        apiClient.get<ApiResponse<TargetListResponseData>>('/targets?limit=100'),
        apiClient.get<ApiResponse<{ items: ScanProfileItem[]; total: number }>>('/scan-profiles'),
      ])
        .then(([targetsRes, profilesRes]) => {
          if (targetsRes.data?.data?.items) {
            const targetList = targetsRes.data.data.items;
            setTargets(targetList);
            const firstValid = targetList.find((t) => !t.is_blacklisted);
            if (firstValid) setSelectedTargetId(firstValid.id);
          }

          if (profilesRes.data?.data?.items) {
            const profileList = profilesRes.data.data.items;
            setProfiles(profileList);
            const fast = profileList.find((p) => p.name.includes('Fast')) || profileList[0];
            if (fast) setSelectedProfileId(fast.id);
          }
        })
        .catch((err) => {
          setError(err.response?.data?.error?.message || 'Impossibile recuperare target o profili');
        })
        .finally(() => {
          setLoadingData(false);
        });
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!selectedTargetId || !selectedProfileId) {
      setError('Seleziona sia un target valido che un profilo di scansione');
      return;
    }

    const selectedTarget = targets.find((t) => t.id === selectedTargetId);
    if (selectedTarget?.is_blacklisted) {
      setError('Impossibile scansionare un bersaglio contrassegnato come Blacklist');
      return;
    }

    setSubmitting(true);
    try {
      if (activeTab === 'ON_DEMAND') {
        const response = await apiClient.post<ApiResponse<ScanItem>>('/scans', {
          target_id: selectedTargetId,
          profile_id: selectedProfileId,
        });

        if (response.data && response.data.success) {
          onScanLaunched();
          onClose();
        } else {
          setError(response.data?.error?.message || 'Errore durante l\'avvio della scansione');
        }
      } else {
        // Scheduled
        const cronExpr = selectedCronPreset === 'CUSTOM' ? customCron.trim() : selectedCronPreset;
        if (!cronExpr) {
          setError('Specifica un\'espressione cron valida');
          setSubmitting(false);
          return;
        }

        const response = await apiClient.post<ApiResponse<ScheduleItem>>('/schedules', {
          target_id: selectedTargetId,
          profile_id: selectedProfileId,
          cron_expression: cronExpr,
          is_active: true,
        });

        if (response.data && response.data.success) {
          if (onScheduleCreated) onScheduleCreated();
          onClose();
        } else {
          setError(response.data?.error?.message || 'Errore durante la creazione della pianificazione');
        }
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile completare la richiesta'
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="max-w-xl w-full bg-[#111726] border border-cyber-border rounded-2xl shadow-2xl overflow-hidden relative">
        {/* Header */}
        <div className="p-5 border-b border-cyber-border flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Terminal className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">Nuova Scansione di Sicurezza</h3>
              <p className="text-xs text-slate-400">Configura un job Nmap on-demand o pianificato</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tab Selector */}
        <div className="flex border-b border-cyber-border bg-slate-900/60 px-6 pt-3">
          <button
            type="button"
            onClick={() => setActiveTab('ON_DEMAND')}
            className={`pb-3 px-3 text-xs font-semibold flex items-center space-x-2 border-b-2 transition-colors ${
              activeTab === 'ON_DEMAND'
                ? 'border-cyan-500 text-cyan-400'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            <Play className="w-3.5 h-3.5" />
            <span>Lancio Immediato (On-Demand)</span>
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('SCHEDULED')}
            className={`pb-3 px-3 text-xs font-semibold flex items-center space-x-2 border-b-2 transition-colors ${
              activeTab === 'SCHEDULED'
                ? 'border-cyan-500 text-cyan-400'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            <Calendar className="w-3.5 h-3.5" />
            <span>Pianificazione Ricorrente (Cron)</span>
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {error && (
            <div className="p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-start space-x-2">
              <AlertCircle className="w-4 h-4 shrink-0 text-rose-400 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {loadingData ? (
            <div className="py-8 text-center text-xs text-slate-400">
              Caricamento target e profili disponibili...
            </div>
          ) : (
            <>
              {/* Selezione Target */}
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1.5">
                  Bersaglio / Asset da Scansionare <span className="text-red-400">*</span>
                </label>
                {targets.length === 0 ? (
                  <div className="p-3 rounded-lg border border-cyber-border bg-slate-900 text-xs text-amber-400">
                    Nessun target configurato. Aggiungi prima un target nella sezione Target & Asset.
                  </div>
                ) : (
                  <select
                    value={selectedTargetId}
                    onChange={(e) => setSelectedTargetId(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-900 border border-cyber-border rounded-lg text-xs text-slate-200 focus:outline-none focus:border-cyan-500 font-mono"
                    required
                  >
                    {targets.map((t) => (
                      <option
                        key={t.id}
                        value={t.id}
                        disabled={t.is_blacklisted}
                        className={t.is_blacklisted ? 'text-slate-500' : 'text-slate-100'}
                      >
                        {t.value} ({t.type}) {t.description ? `- ${t.description}` : ''}{' '}
                        {t.is_blacklisted ? '[BLACKLIST - ESCLUSO]' : ''}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              {/* Selezione Profilo */}
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-2">
                  Profilo di Scansione Nmap <span className="text-red-400">*</span>
                </label>
                <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
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

              {/* Opzioni per Scheda Scheduler */}
              {activeTab === 'SCHEDULED' && (
                <div className="space-y-3 pt-2 border-t border-cyber-border/60">
                  <label className="block text-xs font-medium text-slate-300">
                    Frequenza di Esecuzione Ricorrente
                  </label>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                    {CRON_PRESETS.map((preset) => {
                      const isSelected = selectedCronPreset === preset.value;
                      return (
                        <div
                          key={preset.value}
                          onClick={() => setSelectedCronPreset(preset.value)}
                          className={`p-2.5 rounded-lg border cursor-pointer text-xs transition-colors ${
                            isSelected
                              ? 'border-cyan-500 bg-cyan-500/10 text-white'
                              : 'border-cyber-border bg-slate-900/60 text-slate-300 hover:border-slate-700'
                          }`}
                        >
                          <div className="font-medium">{preset.label}</div>
                          <div className="text-[10px] text-slate-500 mt-0.5">{preset.desc}</div>
                        </div>
                      );
                    })}
                  </div>

                  {selectedCronPreset === 'CUSTOM' && (
                    <div className="mt-2">
                      <label className="block text-[11px] text-slate-400 mb-1">
                        Espressione Cron personalizzata (5 campi: min ora giorno mese dow)
                      </label>
                      <input
                        type="text"
                        value={customCron}
                        onChange={(e) => setCustomCron(e.target.value)}
                        placeholder="es: */30 * * * *"
                        className="w-full px-3 py-2 bg-slate-900 border border-cyber-border rounded-lg text-xs font-mono text-cyan-400 focus:outline-none focus:border-cyan-500"
                        required
                      />
                    </div>
                  )}
                </div>
              )}
            </>
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
              disabled={submitting || targets.length === 0}
              className="px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 text-xs font-semibold transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-1.5"
            >
              {activeTab === 'ON_DEMAND' ? (
                <>
                  <Play className="w-3.5 h-3.5" />
                  <span>{submitting ? 'Avvio scansione...' : 'Avvia Scansione Ora'}</span>
                </>
              ) : (
                <>
                  <Calendar className="w-3.5 h-3.5" />
                  <span>{submitting ? 'Salvataggio...' : 'Crea Pianificazione'}</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
