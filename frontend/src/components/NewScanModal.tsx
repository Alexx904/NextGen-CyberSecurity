import React, { useState, useEffect, useMemo } from 'react';
import {
  X,
  Play,
  Calendar,
  Terminal,
  AlertCircle,
  Check,
  ShieldAlert,
  Crosshair,
  Globe,
  Radio,
  FileCheck2,
  Sliders,
  Sparkles,
  Save,
  Trash2,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetItem, TargetListResponseData } from '../types/target';
import { ScanItem, ScanProfileItem, ScanProfileCategory } from '../types/scan';
import { ScheduleItem } from '../types/schedule';

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

const CATEGORIES: {
  id: ScanProfileCategory;
  name: string;
  badge: string;
  icon: React.ComponentType<{ className?: string }>;
  description: string;
  color: string;
}[] = [
  {
    id: 'PEN_TEST',
    name: 'Penetration Testing',
    badge: 'Offensive & Probing',
    icon: Crosshair,
    description: 'Stealth SYN probing, bypass firewall ed enumeration aggressiva',
    color: 'from-rose-500/20 to-red-600/10 border-rose-500/40 text-rose-400',
  },
  {
    id: 'SYSTEM_FLAWS',
    name: 'Falle di Sistema',
    badge: 'Vulnerability Audit',
    icon: ShieldAlert,
    description: 'Audit OS, fingerprinting demoni e rilevamento superfici RCE critiche',
    color: 'from-amber-500/20 to-orange-600/10 border-amber-500/40 text-amber-400',
  },
  {
    id: 'WEB_APP',
    name: 'Web App & API',
    badge: 'Web Surface',
    icon: Globe,
    description: 'Scansione porte HTTP/HTTPS/API e fingerprinting stack applicativi',
    color: 'from-emerald-500/20 to-teal-600/10 border-emerald-500/40 text-emerald-400',
  },
  {
    id: 'DISCOVERY',
    name: 'Network Discovery',
    badge: 'Host Sweep & Fast',
    icon: Radio,
    description: 'Sweep ICMP/TCP ping veloce e mappatura host vivi senza rumore',
    color: 'from-cyan-500/20 to-blue-600/10 border-cyan-500/40 text-cyan-400',
  },
  {
    id: 'COMPLIANCE',
    name: 'Compliance & Hardening',
    badge: 'Baseline Security',
    icon: FileCheck2,
    description: 'Verifica protocolli legacy insicuri (Telnet, FTP, SMB) e standard baseline',
    color: 'from-purple-500/20 to-indigo-600/10 border-purple-500/40 text-purple-400',
  },
  {
    id: 'CUSTOM',
    name: 'Profili Custom',
    badge: 'Tenant Presets',
    icon: Sliders,
    description: 'Profili personalizzati configurati e salvati dal tuo team',
    color: 'from-slate-500/20 to-slate-700/10 border-slate-500/40 text-slate-300',
  },
];

const PORT_PRESETS = [
  { id: 'FAST', label: 'Top 100 Porte (-F)', value: '-F' },
  { id: 'STANDARD', label: 'Top 1000 Standard', value: '' },
  {
    id: 'CRITICAL',
    label: 'Servizi Critici & RCE',
    value: '-p 21,22,23,25,53,80,110,135,139,143,443,445,1433,1521,3306,3389,5432,5900,6379,8080,8443',
  },
  {
    id: 'WEB',
    label: 'Web & API Ports',
    value: '-p 80,443,8000,8080,8443,8888,3000,5000,9000',
  },
  { id: 'ALL', label: 'Tutte le 65535 Porte (-p-)', value: '-p-' },
  { id: 'CUSTOM', label: 'Porte Personalizzate...', value: 'CUSTOM' },
];

export const NewScanModal: React.FC<NewScanModalProps> = ({
  isOpen,
  onClose,
  onScanLaunched,
  onScheduleCreated,
}) => {
  const [activeTab, setActiveTab] = useState<'ON_DEMAND' | 'SCHEDULED'>('ON_DEMAND');
  const [configMode, setConfigMode] = useState<'GUIDED' | 'PROFESSIONAL'>('GUIDED');

  const [targets, setTargets] = useState<TargetItem[]>([]);
  const [profiles, setProfiles] = useState<ScanProfileItem[]>([]);
  const [loadingData, setLoadingData] = useState<boolean>(false);

  const [selectedTargetId, setSelectedTargetId] = useState<string>('');
  const [selectedProfileId, setSelectedProfileId] = useState<string>('');
  const [selectedCategory, setSelectedCategory] = useState<ScanProfileCategory>('PEN_TEST');

  // Campi Configuratore Professionale (Avanzato)
  const [selectedPortPreset, setSelectedPortPreset] = useState<string>('FAST');
  const [customPortInput, setCustomPortInput] = useState<string>('80,443,8080-8090');
  const [timingTemplate, setTimingTemplate] = useState<string>('-T4');
  const [flagSynStealth, setFlagSynStealth] = useState<boolean>(true);
  const [flagAckProbe, setFlagAckProbe] = useState<boolean>(false);
  const [flagServiceVersion, setFlagServiceVersion] = useState<boolean>(true);
  const [flagDeepVersion, setFlagDeepVersion] = useState<boolean>(false);
  const [flagOsDetection, setFlagOsDetection] = useState<boolean>(false);
  const [flagOnlyOpen, setFlagOnlyOpen] = useState<boolean>(true);
  const [flagNoPing, setFlagNoPing] = useState<boolean>(false);

  // Salvataggio profilo custom
  const [saveAsCustomProfile, setSaveAsCustomProfile] = useState<boolean>(false);
  const [customProfileName, setCustomProfileName] = useState<string>('');
  const [customProfileDesc, setCustomProfileDesc] = useState<string>('');
  const [customProfileCategory, setCustomProfileCategory] = useState<ScanProfileCategory>('PEN_TEST');
  const [customTimeoutMinutes, setCustomTimeoutMinutes] = useState<number>(30);

  // Campi Scheduler
  const [selectedCronPreset, setSelectedCronPreset] = useState<string>('0 2 * * *');
  const [customCron, setCustomCron] = useState<string>('0 2 * * *');

  const [submitting, setSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Caricamento dati
  const fetchProfiles = async () => {
    try {
      const res = await apiClient.get<ApiResponse<{ items: ScanProfileItem[]; total: number }>>('/scan-profiles');
      if (res.data?.data?.items) {
        setProfiles(res.data.data.items);
      }
    } catch (err: any) {
      console.error('Errore recupero profili:', err);
    }
  };

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

            // Seleziona un profilo iniziale della categoria di default PEN_TEST se presente
            const defaultProf =
              profileList.find((p) => p.category === 'PEN_TEST') || profileList[0];
            if (defaultProf) setSelectedProfileId(defaultProf.id);
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

  // Filtra profili per categoria nella modalità guidata
  const filteredProfiles = useMemo(() => {
    return profiles.filter((p) => {
      if (selectedCategory === 'CUSTOM') {
        return p.is_custom === true;
      }
      return p.category === selectedCategory;
    });
  }, [profiles, selectedCategory]);

  // Seleziona il primo profilo disponibile quando cambia la categoria
  useEffect(() => {
    if (configMode === 'GUIDED' && filteredProfiles.length > 0) {
      const alreadyIncluded = filteredProfiles.some((p) => p.id === selectedProfileId);
      if (!alreadyIncluded) {
        setSelectedProfileId(filteredProfiles[0].id);
      }
    }
  }, [selectedCategory, filteredProfiles, configMode, selectedProfileId]);

  // Costruzione dinamica del comando Nmap per la modalità Professionale
  const generatedNmapArgs = useMemo(() => {
    const parts: string[] = [];

    // Scan type
    if (flagSynStealth) {
      parts.push('-sS');
    } else if (flagAckProbe) {
      parts.push('-sA');
    }

    // Version Detection
    if (flagServiceVersion) {
      parts.push('-sV');
      if (flagDeepVersion) {
        parts.push('--version-intensity 7');
      }
    }

    // OS Detection
    if (flagOsDetection) {
      parts.push('-O');
    }

    // Timing
    if (timingTemplate) {
      parts.push(timingTemplate);
    }

    // Ports
    if (selectedPortPreset === 'CUSTOM') {
      const cleaned = customPortInput.trim().replace(/\s+/g, '');
      if (cleaned) {
        parts.push(`-p ${cleaned}`);
      }
    } else {
      const preset = PORT_PRESETS.find((p) => p.id === selectedPortPreset);
      if (preset && preset.value) {
        parts.push(preset.value);
      }
    }

    // Flags opzionali
    if (flagNoPing) parts.push('-Pn');
    if (flagOnlyOpen) parts.push('--open');

    return parts.join(' ');
  }, [
    flagSynStealth,
    flagAckProbe,
    flagServiceVersion,
    flagDeepVersion,
    flagOsDetection,
    timingTemplate,
    selectedPortPreset,
    customPortInput,
    flagNoPing,
    flagOnlyOpen,
  ]);

  const targetIpPreview = useMemo(() => {
    const t = targets.find((item) => item.id === selectedTargetId);
    return t ? t.value : '<TARGET_IP>';
  }, [targets, selectedTargetId]);

  if (!isOpen) return null;

  // Cancellazione profilo custom
  const handleDeleteCustomProfile = async (profileId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!window.confirm('Sei sicuro di voler eliminare questo profilo personalizzato?')) return;

    try {
      await apiClient.delete(`/scan-profiles/${profileId}`);
      await fetchProfiles();
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Errore durante la cancellazione del profilo custom');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!selectedTargetId) {
      setError('Seleziona un bersaglio valido da scansionare');
      return;
    }

    const selectedTarget = targets.find((t) => t.id === selectedTargetId);
    if (selectedTarget?.is_blacklisted) {
      setError('Impossibile scansionare un bersaglio contrassegnato come Blacklist');
      return;
    }

    setSubmitting(true);
    try {
      let finalProfileId = selectedProfileId;

      // In modalità Professionale: creiamo prima un profilo custom se richiesto o un profilo al volo
      if (configMode === 'PROFESSIONAL') {
        const profileName = saveAsCustomProfile && customProfileName.trim()
          ? customProfileName.trim()
          : `Custom PenTest ${new Date().toISOString().slice(11, 19)}`;

        const profileDesc = customProfileDesc.trim() || 'Configurazione specialistica avanzata Nmap';
        const timeoutSeconds = Math.max(60, customTimeoutMinutes * 60);

        const createRes = await apiClient.post<ApiResponse<ScanProfileItem>>('/scan-profiles', {
          name: profileName,
          nmap_args: generatedNmapArgs,
          description: profileDesc,
          timeout_seconds: timeoutSeconds,
          category: customProfileCategory,
        });

        if (!createRes.data || !createRes.data.success || !createRes.data.data) {
          throw new Error(createRes.data?.error?.message || 'Errore nella creazione del profilo custom');
        }

        finalProfileId = createRes.data.data.id;
        // Ricarica la lista dei profili per future selezioni
        fetchProfiles();
      }

      if (!finalProfileId) {
        setError('Nessun profilo di scansione selezionato o creato');
        setSubmitting(false);
        return;
      }

      // Esecuzione lancio On-Demand o Creazione Scheduler
      if (activeTab === 'ON_DEMAND') {
        const response = await apiClient.post<ApiResponse<ScanItem>>('/scans', {
          target_id: selectedTargetId,
          profile_id: finalProfileId,
        });

        if (response.data && response.data.success) {
          onScanLaunched();
          onClose();
        } else {
          setError(response.data?.error?.message || "Errore durante l'avvio della scansione");
        }
      } else {
        // Scheduled
        const cronExpr = selectedCronPreset === 'CUSTOM' ? customCron.trim() : selectedCronPreset;
        if (!cronExpr) {
          setError("Specifica un'espressione cron valida");
          setSubmitting(false);
          return;
        }

        const response = await apiClient.post<ApiResponse<ScheduleItem>>('/schedules', {
          target_id: selectedTargetId,
          profile_id: finalProfileId,
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
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200">
      <div className="max-w-3xl w-full max-h-[92vh] flex flex-col bg-[#111726] border border-cyber-border rounded-2xl shadow-2xl overflow-hidden relative">
        {/* Header */}
        <div className="p-4 sm:p-5 border-b border-cyber-border flex items-center justify-between shrink-0 bg-slate-900/40">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400 shadow-[0_0_12px_rgba(6,182,212,0.2)]">
              <Terminal className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h3 className="text-base font-bold text-white">Nuova Scansione di Vulnerabilità</h3>
                <span className="text-[10px] uppercase font-mono tracking-wider px-2 py-0.5 rounded-full bg-cyan-500/20 text-cyan-300 border border-cyan-500/40">
                  v2.0 Pro
                </span>
              </div>
              <p className="text-xs text-slate-400">
                Lancio on-demand e pianificazione con profili specialistici o configurazione avanzata
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tab Selector: On-Demand vs Scheduled & Modalità Semplice vs Professionale */}
        <div className="flex flex-wrap items-center justify-between border-b border-cyber-border bg-slate-900/70 px-4 sm:px-6 pt-2 shrink-0 gap-2">
          {/* On-Demand / Scheduled */}
          <div className="flex space-x-1 sm:space-x-2">
            <button
              type="button"
              onClick={() => setActiveTab('ON_DEMAND')}
              className={`pb-2.5 px-3 text-xs font-semibold flex items-center space-x-1.5 border-b-2 transition-colors ${
                activeTab === 'ON_DEMAND'
                  ? 'border-cyan-500 text-cyan-400'
                  : 'border-transparent text-slate-400 hover:text-slate-200'
              }`}
            >
              <Play className="w-3.5 h-3.5" />
              <span>Lancio On-Demand</span>
            </button>

            <button
              type="button"
              onClick={() => setActiveTab('SCHEDULED')}
              className={`pb-2.5 px-3 text-xs font-semibold flex items-center space-x-1.5 border-b-2 transition-colors ${
                activeTab === 'SCHEDULED'
                  ? 'border-cyan-500 text-cyan-400'
                  : 'border-transparent text-slate-400 hover:text-slate-200'
              }`}
            >
              <Calendar className="w-3.5 h-3.5" />
              <span>Schedulazione Cron</span>
            </button>
          </div>

          {/* Toggle Semplice / Professionale */}
          <div className="pb-2 flex items-center space-x-1 bg-slate-950/60 p-1 rounded-lg border border-cyber-border">
            <button
              type="button"
              onClick={() => setConfigMode('GUIDED')}
              className={`px-2.5 py-1 rounded text-xs font-medium flex items-center space-x-1.5 transition-all ${
                configMode === 'GUIDED'
                  ? 'bg-cyan-500 text-slate-950 font-semibold shadow-sm'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Sparkles className="w-3 h-3" />
              <span>Modalità Guidata</span>
            </button>
            <button
              type="button"
              onClick={() => setConfigMode('PROFESSIONAL')}
              className={`px-2.5 py-1 rounded text-xs font-medium flex items-center space-x-1.5 transition-all ${
                configMode === 'PROFESSIONAL'
                  ? 'bg-cyan-500 text-slate-950 font-semibold shadow-sm'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Sliders className="w-3 h-3" />
              <span>Modalità Professionale</span>
            </button>
          </div>
        </div>

        {/* Scrollable Form Body */}
        <form onSubmit={handleSubmit} className="p-4 sm:p-6 space-y-5 overflow-y-auto flex-1">
          {error && (
            <div className="p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-start space-x-2 animate-in fade-in">
              <AlertCircle className="w-4 h-4 shrink-0 text-rose-400 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {loadingData ? (
            <div className="py-12 text-center text-xs text-slate-400">
              <div className="inline-block animate-spin rounded-full h-6 w-6 border-2 border-cyan-500 border-t-transparent mb-2"></div>
              <p>Caricamento asset e catalogo profili...</p>
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

              {/* SEZIONE 1: MODALITÀ GUIDATA (PER CATEGORIE) */}
              {configMode === 'GUIDED' && (
                <div className="space-y-4">
                  {/* Selettore Categorie (Pills con icone) */}
                  <div>
                    <label className="block text-xs font-medium text-slate-300 mb-2">
                      1. Seleziona Obiettivo & Categoria di Scansione
                    </label>
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                      {CATEGORIES.map((cat) => {
                        const Icon = cat.icon;
                        const isSelected = selectedCategory === cat.id;
                        const count = profiles.filter((p) =>
                          cat.id === 'CUSTOM' ? p.is_custom : p.category === cat.id
                        ).length;

                        return (
                          <div
                            key={cat.id}
                            onClick={() => setSelectedCategory(cat.id)}
                            className={`p-2.5 rounded-xl border cursor-pointer transition-all ${
                              isSelected
                                ? `bg-gradient-to-br ${cat.color} shadow-[0_0_15px_rgba(6,182,212,0.15)] ring-1 ring-cyan-500/60`
                                : 'border-cyber-border bg-slate-900/60 hover:border-slate-700 text-slate-400'
                            }`}
                          >
                            <div className="flex items-center justify-between mb-1">
                              <Icon
                                className={`w-4 h-4 ${
                                  isSelected ? 'text-cyan-400' : 'text-slate-400'
                                }`}
                              />
                              <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-slate-800 text-slate-300 font-mono">
                                {count}
                              </span>
                            </div>
                            <div className="font-semibold text-xs text-white leading-tight">
                              {cat.name}
                            </div>
                            <div className="text-[10px] text-slate-400 mt-0.5 line-clamp-1">
                              {cat.badge}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Lista Profili per la Categoria Selezionata */}
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <label className="text-xs font-medium text-slate-300">
                        2. Scegli il Profilo Specialistico ({filteredProfiles.length} disponibili)
                      </label>
                    </div>

                    {filteredProfiles.length === 0 ? (
                      <div className="p-4 rounded-xl border border-cyber-border bg-slate-900/40 text-center text-xs text-slate-400">
                        Nessun profilo trovato per questa categoria. Passa alla Modalità Professionale per creare un profilo custom.
                      </div>
                    ) : (
                      <div className="space-y-2 max-h-56 overflow-y-auto pr-1">
                        {filteredProfiles.map((p) => {
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
                              <div className="flex items-center justify-between text-xs mb-1">
                                <div className="flex items-center space-x-2">
                                  <span
                                    className={`w-3.5 h-3.5 rounded-full border flex items-center justify-center ${
                                      isSelected
                                        ? 'border-cyan-400 bg-cyan-400'
                                        : 'border-slate-600'
                                    }`}
                                  >
                                    {isSelected && (
                                      <Check className="w-2.5 h-2.5 text-slate-950" />
                                    )}
                                  </span>
                                  <span className="font-semibold text-white">{p.name}</span>
                                  {p.is_custom && (
                                    <span className="text-[9px] px-1.5 py-0.2 rounded bg-purple-500/20 text-purple-300 border border-purple-500/40 font-mono">
                                      CUSTOM
                                    </span>
                                  )}
                                </div>
                                <div className="flex items-center space-x-2">
                                  <span className="text-[10px] text-slate-400 font-mono">
                                    timeout: {Math.round(p.timeout_seconds / 60)}m
                                  </span>
                                  {p.is_custom && (
                                    <button
                                      type="button"
                                      onClick={(e) => handleDeleteCustomProfile(p.id, e)}
                                      className="p-1 text-slate-500 hover:text-rose-400 hover:bg-slate-800 rounded transition-colors"
                                      title="Elimina profilo custom"
                                    >
                                      <Trash2 className="w-3 h-3" />
                                    </button>
                                  )}
                                </div>
                              </div>
                              <p className="text-[11px] text-slate-400 pl-5.5 leading-relaxed">
                                {p.description}
                              </p>
                              <div className="mt-2 pl-5.5 flex items-center space-x-2">
                                <code className="text-[10px] font-mono bg-slate-950/80 px-2 py-0.5 rounded text-cyan-300 border border-cyan-900/40">
                                  {p.nmap_args}
                                </code>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* SEZIONE 2: MODALITÀ PROFESSIONALE (CONFIGURATORE AVANZATO) */}
              {configMode === 'PROFESSIONAL' && (
                <div className="space-y-4">
                  {/* Selettore Scope Porte */}
                  <div>
                    <label className="block text-xs font-medium text-slate-300 mb-1.5">
                      Scope Porte Bersaglio
                    </label>
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 mb-2">
                      {PORT_PRESETS.map((preset) => {
                        const isSelected = selectedPortPreset === preset.id;
                        return (
                          <button
                            type="button"
                            key={preset.id}
                            onClick={() => setSelectedPortPreset(preset.id)}
                            className={`p-2 rounded-lg border text-left text-xs transition-colors ${
                              isSelected
                                ? 'border-cyan-500 bg-cyan-500/10 text-white font-medium'
                                : 'border-cyber-border bg-slate-900/60 text-slate-400 hover:border-slate-700'
                            }`}
                          >
                            <div className="truncate">{preset.label}</div>
                          </button>
                        );
                      })}
                    </div>

                    {selectedPortPreset === 'CUSTOM' && (
                      <div className="mt-2">
                        <input
                          type="text"
                          value={customPortInput}
                          onChange={(e) => setCustomPortInput(e.target.value)}
                          placeholder="es: 22,80,443,8000-8090,3389"
                          className="w-full px-3 py-2 bg-slate-900 border border-cyber-border rounded-lg text-xs font-mono text-cyan-400 focus:outline-none focus:border-cyan-500"
                        />
                        <span className="text-[10px] text-slate-500 mt-1 block">
                          Formati accettati: numeri singoli separati da virgola o range (es. 1000-2000).
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Timing Template (T1 - T5) */}
                  <div>
                    <div className="flex items-center justify-between mb-1.5">
                      <label className="text-xs font-medium text-slate-300">
                        Timing & Aggressiveness Template
                      </label>
                      <span className="text-[11px] font-mono text-cyan-400">
                        {timingTemplate === '-T1' && 'T1: Sneaky (Max Evasion)'}
                        {timingTemplate === '-T2' && 'T2: Polite (Low Bandwidth)'}
                        {timingTemplate === '-T3' && 'T3: Normal (Balanced)'}
                        {timingTemplate === '-T4' && 'T4: Aggressive (Recommended)'}
                        {timingTemplate === '-T5' && 'T5: Insane (Lab/LAN Only)'}
                      </span>
                    </div>
                    <div className="grid grid-cols-5 gap-1.5">
                      {[
                        { flag: '-T1', label: 'T1', sub: 'Sneaky' },
                        { flag: '-T2', label: 'T2', sub: 'Polite' },
                        { flag: '-T3', label: 'T3', sub: 'Normal' },
                        { flag: '-T4', label: 'T4', sub: 'Aggressive' },
                        { flag: '-T5', label: 'T5', sub: 'Insane' },
                      ].map((t) => {
                        const isSelected = timingTemplate === t.flag;
                        return (
                          <button
                            type="button"
                            key={t.flag}
                            onClick={() => setTimingTemplate(t.flag)}
                            className={`py-1.5 rounded-lg border text-center transition-all ${
                              isSelected
                                ? 'border-cyan-500 bg-cyan-500/20 text-white font-bold shadow-[0_0_10px_rgba(6,182,212,0.2)]'
                                : 'border-cyber-border bg-slate-900/60 text-slate-400 hover:border-slate-700'
                            }`}
                          >
                            <div className="text-xs">{t.label}</div>
                            <div className="text-[9px] text-slate-500">{t.sub}</div>
                          </button>
                        );
                      })}
                    </div>
                  </div>

                  {/* Penetration Testing & Scanner Flags (Checkboxes) */}
                  <div>
                    <label className="block text-xs font-medium text-slate-300 mb-2">
                      Flag di Scansione & Penetration Testing
                    </label>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                      {/* SYN Stealth */}
                      <label className="flex items-center space-x-2 p-2 rounded-lg border border-cyber-border bg-slate-900/50 cursor-pointer hover:border-slate-700">
                        <input
                          type="checkbox"
                          checked={flagSynStealth}
                          onChange={(e) => {
                            setFlagSynStealth(e.target.checked);
                            if (e.target.checked) setFlagAckProbe(false);
                          }}
                          className="rounded border-slate-700 text-cyan-500 focus:ring-0"
                        />
                        <div className="flex-1">
                          <span className="text-white font-medium">SYN Stealth (-sS)</span>
                          <p className="text-[10px] text-slate-500">Half-open TCP scan senza handshake</p>
                        </div>
                      </label>

                      {/* ACK Probe */}
                      <label className="flex items-center space-x-2 p-2 rounded-lg border border-cyber-border bg-slate-900/50 cursor-pointer hover:border-slate-700">
                        <input
                          type="checkbox"
                          checked={flagAckProbe}
                          onChange={(e) => {
                            setFlagAckProbe(e.target.checked);
                            if (e.target.checked) setFlagSynStealth(false);
                          }}
                          className="rounded border-slate-700 text-cyan-500 focus:ring-0"
                        />
                        <div className="flex-1">
                          <span className="text-white font-medium">ACK Firewall Probe (-sA)</span>
                          <p className="text-[10px] text-slate-500">Mappa regole firewall stateful/filtrate</p>
                        </div>
                      </label>

                      {/* Service Version */}
                      <label className="flex items-center space-x-2 p-2 rounded-lg border border-cyber-border bg-slate-900/50 cursor-pointer hover:border-slate-700">
                        <input
                          type="checkbox"
                          checked={flagServiceVersion}
                          onChange={(e) => setFlagServiceVersion(e.target.checked)}
                          className="rounded border-slate-700 text-cyan-500 focus:ring-0"
                        />
                        <div className="flex-1">
                          <span className="text-white font-medium">Service Versioning (-sV)</span>
                          <p className="text-[10px] text-slate-500">Banner grabbing e versione demoni</p>
                        </div>
                      </label>

                      {/* OS Detection */}
                      <label className="flex items-center space-x-2 p-2 rounded-lg border border-cyber-border bg-slate-900/50 cursor-pointer hover:border-slate-700">
                        <input
                          type="checkbox"
                          checked={flagOsDetection}
                          onChange={(e) => setFlagOsDetection(e.target.checked)}
                          className="rounded border-slate-700 text-cyan-500 focus:ring-0"
                        />
                        <div className="flex-1">
                          <span className="text-white font-medium">OS Fingerprinting (-O)</span>
                          <p className="text-[10px] text-slate-500">Rileva Sistema Operativo remoto</p>
                        </div>
                      </label>

                      {/* Version Intensity */}
                      <label className="flex items-center space-x-2 p-2 rounded-lg border border-cyber-border bg-slate-900/50 cursor-pointer hover:border-slate-700">
                        <input
                          type="checkbox"
                          checked={flagDeepVersion}
                          disabled={!flagServiceVersion}
                          onChange={(e) => setFlagDeepVersion(e.target.checked)}
                          className="rounded border-slate-700 text-cyan-500 focus:ring-0 disabled:opacity-40"
                        />
                        <div className="flex-1">
                          <span className="text-white font-medium">Deep Intensity (Level 7)</span>
                          <p className="text-[10px] text-slate-500">Sonde aggressive per CVE detection</p>
                        </div>
                      </label>

                      {/* No Ping */}
                      <label className="flex items-center space-x-2 p-2 rounded-lg border border-cyber-border bg-slate-900/50 cursor-pointer hover:border-slate-700">
                        <input
                          type="checkbox"
                          checked={flagNoPing}
                          onChange={(e) => setFlagNoPing(e.target.checked)}
                          className="rounded border-slate-700 text-cyan-500 focus:ring-0"
                        />
                        <div className="flex-1">
                          <span className="text-white font-medium">No Ping / Treat Up (-Pn)</span>
                          <p className="text-[10px] text-slate-500">Salta ICMP se l'host blocca ping</p>
                        </div>
                      </label>

                      {/* Only Open */}
                      <label className="flex items-center space-x-2 p-2 rounded-lg border border-cyber-border bg-slate-900/50 cursor-pointer hover:border-slate-700 sm:col-span-2">
                        <input
                          type="checkbox"
                          checked={flagOnlyOpen}
                          onChange={(e) => setFlagOnlyOpen(e.target.checked)}
                          className="rounded border-slate-700 text-cyan-500 focus:ring-0"
                        />
                        <div className="flex-1">
                          <span className="text-white font-medium">Mostra solo porte aperte (--open)</span>
                          <p className="text-[10px] text-slate-500">Ignora porte chiuse/filtrate per ridurre rumore</p>
                        </div>
                      </label>
                    </div>
                  </div>

                  {/* Opzione Salva come Profilo Custom */}
                  <div className="p-3 rounded-xl border border-cyber-border bg-slate-900/40 space-y-3">
                    <label className="flex items-center space-x-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={saveAsCustomProfile}
                        onChange={(e) => setSaveAsCustomProfile(e.target.checked)}
                        className="rounded border-slate-700 text-cyan-500 focus:ring-0"
                      />
                      <span className="text-xs font-semibold text-white flex items-center space-x-1.5">
                        <Save className="w-3.5 h-3.5 text-cyan-400" />
                        <span>Salva questa configurazione come nuovo Profilo Custom</span>
                      </span>
                    </label>

                    {saveAsCustomProfile && (
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5 pt-2 border-t border-cyber-border/60">
                        <div>
                          <label className="block text-[11px] text-slate-400 mb-1">
                            Nome Profilo <span className="text-red-400">*</span>
                          </label>
                          <input
                            type="text"
                            value={customProfileName}
                            onChange={(e) => setCustomProfileName(e.target.value)}
                            placeholder="es. Internal Web & DB Audit"
                            className="w-full px-2.5 py-1.5 bg-slate-950 border border-cyber-border rounded-lg text-xs text-white focus:outline-none focus:border-cyan-500"
                            required={saveAsCustomProfile}
                          />
                        </div>

                        <div>
                          <label className="block text-[11px] text-slate-400 mb-1">
                            Categoria di Riferimento
                          </label>
                          <select
                            value={customProfileCategory}
                            onChange={(e) => setCustomProfileCategory(e.target.value as ScanProfileCategory)}
                            className="w-full px-2.5 py-1.5 bg-slate-950 border border-cyber-border rounded-lg text-xs text-slate-200 focus:outline-none focus:border-cyan-500"
                          >
                            <option value="PEN_TEST">Penetration Testing</option>
                            <option value="SYSTEM_FLAWS">Falle di Sistema</option>
                            <option value="WEB_APP">Web App & API</option>
                            <option value="DISCOVERY">Network Discovery</option>
                            <option value="COMPLIANCE">Compliance</option>
                          </select>
                        </div>

                        <div className="sm:col-span-2">
                          <label className="block text-[11px] text-slate-400 mb-1">
                            Descrizione
                          </label>
                          <input
                            type="text"
                            value={customProfileDesc}
                            onChange={(e) => setCustomProfileDesc(e.target.value)}
                            placeholder="es. Audit porte applicative e probing versioni per testing interno"
                            className="w-full px-2.5 py-1.5 bg-slate-950 border border-cyber-border rounded-lg text-xs text-slate-200 focus:outline-none focus:border-cyan-500"
                          />
                        </div>

                        <div>
                          <label className="block text-[11px] text-slate-400 mb-1">
                            Timeout Massimo (minuti)
                          </label>
                          <input
                            type="number"
                            min="1"
                            max="360"
                            value={customTimeoutMinutes}
                            onChange={(e) => setCustomTimeoutMinutes(Number(e.target.value))}
                            className="w-full px-2.5 py-1.5 bg-slate-950 border border-cyber-border rounded-lg text-xs text-white focus:outline-none focus:border-cyan-500"
                          />
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* LIVE COMMAND PREVIEW */}
              <div className="p-3 rounded-xl border border-cyan-900/40 bg-slate-950 font-mono text-xs">
                <div className="flex items-center justify-between text-[11px] text-slate-400 mb-1">
                  <span className="flex items-center space-x-1.5 text-cyan-400">
                    <Terminal className="w-3.5 h-3.5" />
                    <span className="font-semibold">Live Nmap Command Preview</span>
                  </span>
                  <span className="text-[10px] text-emerald-400 bg-emerald-950/60 px-2 py-0.5 rounded border border-emerald-800/40">
                    Parametri Convalidati & Sicuri
                  </span>
                </div>
                <div className="text-slate-300 break-all select-all pt-1">
                  <span className="text-cyan-400 font-bold">$ nmap </span>
                  <span className="text-amber-300">
                    {configMode === 'GUIDED'
                      ? profiles.find((p) => p.id === selectedProfileId)?.nmap_args || '-T4 -F'
                      : generatedNmapArgs}
                  </span>{' '}
                  <span className="text-emerald-300 font-semibold">{targetIpPreview}</span>
                </div>
              </div>

              {/* OPZIONI SCHEDA SCHEDULER (CRON) */}
              {activeTab === 'SCHEDULED' && (
                <div className="space-y-3 pt-3 border-t border-cyber-border/60">
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
          <div className="pt-3 flex items-center justify-end space-x-3 border-t border-cyber-border/60 shrink-0">
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
              className="px-5 py-2.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 text-xs font-semibold transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-1.5"
            >
              {activeTab === 'ON_DEMAND' ? (
                <>
                  <Play className="w-3.5 h-3.5" />
                  <span>{submitting ? 'Avvio in corso...' : 'Avvia Scansione Ora'}</span>
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
