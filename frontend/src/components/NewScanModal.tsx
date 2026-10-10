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
  FileCheck2,
  Sliders,
  Sparkles,
  Save,
  Trash2,
  Compass,
  HelpCircle,
  Info,
  ChevronDown,
  ChevronUp,
  Layers,
  Server,
  Cpu,
  Folder,
  Bug,
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetItem, TargetListResponseData, TargetGroupNode } from '../types/target';
import { ScanItem, ScanProfileItem, ScanProfileCategory } from '../types/scan';
import { ScheduleItem } from '../types/schedule';
import { getGroupAndDescendantIds } from './TreeExplorer';

interface NewScanModalProps {
  isOpen: boolean;
  onClose: () => void;
  onScanLaunched: () => void;
  onScheduleCreated?: () => void;
  defaultGroupId?: string;
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
    id: 'ALL_IN_ONE',
    name: 'Tutto in Uno (360°)',
    badge: 'Consigliato · Completo',
    icon: Compass,
    description: 'Audit generalista a 360°: host vivi, porte, versioni demoni e rilevamento OS',
    color: 'from-blue-600/30 to-indigo-600/20 border-blue-500/50 text-blue-300 ring-1 ring-blue-500/40',
  },
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

const CATEGORY_HELP_GUIDE: Record<
  ScanProfileCategory,
  {
    simpleExplanation: string;
    whenToUse: string;
    estimatedDuration: string;
    networkImpact: string;
    whatItDetects: string;
  }
> = {
  ALL_IN_ONE: {
    simpleExplanation:
      'Controllo a 360° globale: trova tutti gli host accesi, tutte le porte principali, le versioni precise dei servizi e il sistema operativo.',
    whenToUse:
      'Consigliato se vuoi un quadro completo e affidabile su un server o dispositivo senza tralasciare nulla.',
    estimatedDuration: '15 - 45 minuti',
    networkImpact: 'Bilanciato e sicuro (nessun disturbo ai servizi di produzione)',
    whatItDetects: 'OS (Linux/Windows), porte aperte, versioni demoni, vulnerabilità di superficie',
  },
  PEN_TEST: {
    simpleExplanation:
      'Simula un attaccante esterno: invia pacchetti discreti (SYN stealth) per mappare la rete eludendo o testando i firewall perimetrali.',
    whenToUse:
      'Per verificare se i tuoi sistemi di difesa (firewall/IDS) bloccano accessi indesiderati.',
    estimatedDuration: '30 - 60 minuti',
    networkImpact: 'Basso rumore, pacchetti half-open senza handshake completo',
    whatItDetects: 'Filtri firewall permissivi, porte fantasma, apparati esposti',
  },
  SYSTEM_FLAWS: {
    simpleExplanation:
      'Scava nei demoni di gestione di sistema (Windows RDP, Linux SSH, database MySQL/Postgres, cartelle SMB).',
    whenToUse:
      'Se vuoi identificare software obsoleto o falle critiche sfruttabili per esecuzione di codice remoto (RCE).',
    estimatedDuration: '20 - 40 minuti',
    networkImpact: 'Mirato sui servizi applicativi di sistema',
    whatItDetects: 'Versioni software vulnerabili a CVE note, protocolli di gestione aperti',
  },
  WEB_APP: {
    simpleExplanation:
      'Controlla solo le porte dedicate al web (HTTP, HTTPS, API REST, reverse proxy e microservizi).',
    whenToUse:
      'Se gestisci un sito web aziendale, un e-commerce o API e vuoi verificare i server web prima o dopo il deploy.',
    estimatedDuration: '10 - 20 minuti',
    networkImpact: 'Molto leggero, solo traffico web standard',
    whatItDetects: 'Versioni web server (Nginx/Apache), intestazioni TLS, endpoint API esposti',
  },
  DISCOVERY: {
    simpleExplanation:
      'Mappatura lampo: invia ping rapidi per dirti chi è acceso nella rete locale senza fare scansioni di porte pesanti.',
    whenToUse:
      'Per inventariare rapidamente gli apparati (PC, stampanti, server) in pochi secondi.',
    estimatedDuration: '1 - 3 minuti',
    networkImpact: 'Minimo assoluto, traffico trascurabile',
    whatItDetects: 'Elenco indirizzi IP attivi e host vivi nella LAN',
  },
  COMPLIANCE: {
    simpleExplanation:
      'Verifica di conformità e sicurezza standard: cerca vecchi protocolli insicuri che trasmettono dati in chiaro.',
    whenToUse:
      'Per audit periodici di conformità aziendale (ISO 27001, NIS2) per eliminare software deprecato.',
    estimatedDuration: '5 - 15 minuti',
    networkImpact: 'Molto leggero',
    whatItDetects: 'Protocolli obsoleti (Telnet, FTP, TFTP, condivisioni legacy)',
  },
  CUSTOM: {
    simpleExplanation:
      'I profili personalizzati creati e salvati dal tuo team con parametri specifici su misura.',
    whenToUse:
      'Per riapplicare una configurazione frequente con un solo clic.',
    estimatedDuration: 'In base al timeout configurato',
    networkImpact: 'Dipende dai parametri scelti',
    whatItDetects: 'Definito dalla configurazione personalizzata',
  },
};

interface ObjectiveOption {
  id: string;
  label: string;
  simpleDesc: string;
  ports?: string[];
  flags?: string[];
  icon: React.ComponentType<{ className?: string }>;
}

const OBJECTIVE_OPTIONS: ObjectiveOption[] = [
  {
    id: 'OBJ_WEB',
    label: 'Siti Web, Portali & API',
    simpleDesc: 'Ispeziona server HTTP/HTTPS, proxy e API (porte 80, 443, 8000, 8080, 8443, 9000)',
    ports: ['80', '443', '8000', '8080', '8443', '8888', '3000', '5000', '9000'],
    icon: Globe,
  },
  {
    id: 'OBJ_SERVICES',
    label: 'Servizi Critici & Accesso Remoto',
    simpleDesc: 'Controlla database (MySQL, Postgres, MSSQL), desktop remoto (RDP), SSH e cartelle SMB',
    ports: ['21', '22', '23', '25', '53', '110', '135', '139', '143', '445', '1433', '1521', '3306', '3389', '5432', '5900', '6379'],
    icon: Server,
  },
  {
    id: 'OBJ_VERSIONS',
    label: 'Identifica Versioni Servizi (-sV)',
    simpleDesc: 'Interroga i banner dei programmi per identificare la versione esatta e scovare vulnerabilità CVE',
    flags: ['-sV'],
    icon: ShieldAlert,
  },
  {
    id: 'OBJ_OS',
    label: 'Riconosci Sistema Operativo (-O)',
    simpleDesc: 'Identifica se il dispositivo monta Windows, Linux, macOS o router/firewall di rete',
    flags: ['-O'],
    icon: Cpu,
  },
  {
    id: 'OBJ_STEALTH',
    label: 'Modalità Silenziosa Stealth (-sS)',
    simpleDesc: 'Invia pacchetti SYN senza completare la connessione per eludere o non intasare i log dei firewall',
    flags: ['-sS'],
    icon: Crosshair,
  },
  {
    id: 'OBJ_LEGACY',
    label: 'Controllo Protocolli Legacy Insicuri',
    simpleDesc: 'Scova servizi non crittografati che trasmettono credenziali in chiaro (Telnet, FTP, TFTP)',
    ports: ['21', '23', '69', '79', '110', '143', '512', '513', '514'],
    icon: FileCheck2,
  },
  {
    id: 'OBJ_NO_PING',
    label: 'Bypass Filtro Ping (-Pn)',
    simpleDesc: 'Forza la scansione anche se l’host blocca i pacchetti ping ICMP o ha firewall attivi',
    flags: ['-Pn'],
    icon: Terminal,
  },
  {
    id: 'OBJ_VULN_SCRIPTS',
    label: 'Script Vulnerabilità Nmap (--script=vuln)',
    simpleDesc: 'Esegue script attivi Nmap per scovare CVE note (EternalBlue, SSL, RCE, default credentials)',
    flags: ['--script=vuln'],
    icon: Bug,
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
  defaultGroupId,
}) => {
  const [activeTab, setActiveTab] = useState<'ON_DEMAND' | 'SCHEDULED'>('ON_DEMAND');
  const [configMode, setConfigMode] = useState<'GUIDED' | 'MULTI_OBJECTIVE' | 'PROFESSIONAL'>('GUIDED');
  const [showHelpGuide, setShowHelpGuide] = useState<boolean>(false);

  const [targetGroups, setTargetGroups] = useState<TargetGroupNode[]>([]);
  const [targets, setTargets] = useState<TargetItem[]>([]);
  const [profiles, setProfiles] = useState<ScanProfileItem[]>([]);
  const [loadingData, setLoadingData] = useState<boolean>(false);

  const [selectedGroupId, setSelectedGroupId] = useState<string>(defaultGroupId || '');
  const [selectedProfileId, setSelectedProfileId] = useState<string>('');
  const [selectedCategory, setSelectedCategory] = useState<ScanProfileCategory>('ALL_IN_ONE');

  // Campi Modalità Multi-Obiettivo
  const [selectedObjectiveIds, setSelectedObjectiveIds] = useState<string[]>([
    'OBJ_WEB',
    'OBJ_SERVICES',
    'OBJ_VERSIONS',
    'OBJ_OS',
    'OBJ_VULN_SCRIPTS',
  ]);

  // Campi Configuratore Professionale (Avanzato)
  const [selectedPortPreset, setSelectedPortPreset] = useState<string>('FAST');
  const [customPortInput, setCustomPortInput] = useState<string>('80,443,8080-8090');
  const [timingTemplate, setTimingTemplate] = useState<string>('-T4');
  const [flagSynStealth, setFlagSynStealth] = useState<boolean>(true);
  const [flagAckProbe, setFlagAckProbe] = useState<boolean>(false);
  const [flagServiceVersion, setFlagServiceVersion] = useState<boolean>(true);
  const [flagDeepVersion, setFlagDeepVersion] = useState<boolean>(false);
  const [flagOsDetection, setFlagOsDetection] = useState<boolean>(false);
  const [flagVulnScripts, setFlagVulnScripts] = useState<boolean>(true);
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
        apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups'),
        apiClient.get<ApiResponse<TargetListResponseData>>('/targets?limit=200'),
        apiClient.get<ApiResponse<{ items: ScanProfileItem[]; total: number }>>('/scan-profiles'),
      ])
        .then(([groupsRes, targetsRes, profilesRes]) => {
          let loadedGroups: TargetGroupNode[] = [];
          if (groupsRes.data?.data) {
            loadedGroups = groupsRes.data.data;
            setTargetGroups(loadedGroups);
          }

          if (targetsRes.data?.data?.items) {
            setTargets(targetsRes.data.data.items);
          }

          if (defaultGroupId) {
            setSelectedGroupId(defaultGroupId);
          } else if (loadedGroups.length > 0) {
            setSelectedGroupId((prev) => (prev ? prev : loadedGroups[0].id));
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
          setError(err.response?.data?.error?.message || 'Impossibile recuperare cartelle o profili');
        })
        .finally(() => {
          setLoadingData(false);
        });
    }
  }, [isOpen, defaultGroupId]);

  // Filtra profili per categoria nella modalità guidata
  const filteredProfiles = useMemo(() => {
    return profiles.filter((p) => {
      if (selectedCategory === 'CUSTOM') {
        return p.is_custom === true;
      }
      return p.category === selectedCategory;
    });
  }, [profiles, selectedCategory]);

  // Helper per appiattire l'albero dei gruppi con profondità e bersagli associati
  const flattenedGroups = useMemo(() => {
    const list: {
      id: string;
      name: string;
      depth: number;
      color_tag?: string;
      targetCount: number;
      targets: TargetItem[];
    }[] = [];

    const recurse = (nodes: TargetGroupNode[], depth = 0) => {
      for (const node of nodes) {
        const descendantIds = new Set(getGroupAndDescendantIds(node));
        const matchingTargets = targets.filter((t) => t.group_id && descendantIds.has(t.group_id));
        list.push({
          id: node.id,
          name: node.name,
          depth,
          color_tag: node.color_tag || undefined,
          targetCount: matchingTargets.length,
          targets: matchingTargets,
        });
        if (node.children && node.children.length > 0) {
          recurse(node.children, depth + 1);
        }
      }
    };

    recurse(targetGroups);
    return list;
  }, [targetGroups, targets]);

  const selectedGroupInfo = useMemo(() => {
    return flattenedGroups.find((g) => g.id === selectedGroupId) || null;
  }, [flattenedGroups, selectedGroupId]);

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
    if (flagVulnScripts) parts.push('--script=vuln');
    if (flagNoPing) parts.push('-Pn');
    if (flagOnlyOpen) parts.push('--open');

    return parts.join(' ');
  }, [
    flagSynStealth,
    flagAckProbe,
    flagServiceVersion,
    flagDeepVersion,
    flagOsDetection,
    flagVulnScripts,
    timingTemplate,
    selectedPortPreset,
    customPortInput,
    flagNoPing,
    flagOnlyOpen,
  ]);

  // Costruzione dinamica del comando Nmap per la modalità Multi-Obiettivo
  const generatedMultiObjectiveArgs = useMemo(() => {
    const selectedObjs = OBJECTIVE_OPTIONS.filter((o) => selectedObjectiveIds.includes(o.id));

    const parts: string[] = [];

    // Stealth SYN
    if (selectedObjectiveIds.includes('OBJ_STEALTH')) {
      parts.push('-sS');
    }

    // Version Detection
    if (selectedObjectiveIds.includes('OBJ_VERSIONS')) {
      parts.push('-sV');
    }

    // OS Detection
    if (selectedObjectiveIds.includes('OBJ_OS')) {
      parts.push('-O');
    }

    // Timing
    parts.push('-T4');

    // Porte da unire
    const allPortsSet = new Set<number>();
    selectedObjs.forEach((o) => {
      if (o.ports) {
        o.ports.forEach((p) => {
          const num = parseInt(p, 10);
          if (!isNaN(num)) allPortsSet.add(num);
        });
      }
    });

    if (allPortsSet.size > 0) {
      const sortedPorts = Array.from(allPortsSet).sort((a, b) => a - b);
      parts.push(`-p ${sortedPorts.join(',')}`);
    }

    // Script di vulnerabilità Nmap
    if (selectedObjectiveIds.includes('OBJ_VULN_SCRIPTS')) {
      parts.push('--script=vuln');
    }

    // No ping se richiesto
    if (selectedObjectiveIds.includes('OBJ_NO_PING')) {
      parts.push('-Pn');
    }

    parts.push('--open');

    return parts.join(' ');
  }, [selectedObjectiveIds]);

  const targetIpPreview = useMemo(() => {
    if (selectedGroupInfo) {
      if (selectedGroupInfo.targets.length > 0) {
        const preview = selectedGroupInfo.targets.map((t) => t.value).slice(0, 3).join(', ');
        return `${selectedGroupInfo.name} [${selectedGroupInfo.targets.length} IP: ${preview}${selectedGroupInfo.targets.length > 3 ? '...' : ''}]`;
      }
      return `${selectedGroupInfo.name} [0 IP censiti]`;
    }
    return '<CARTELLA_AZIENDA>';
  }, [selectedGroupInfo]);

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

    if (!selectedGroupId) {
      setError('Seleziona un\'azienda o cartella valida da scansionare');
      return;
    }

    if (!selectedGroupInfo || selectedGroupInfo.targets.length === 0) {
      setError('La cartella o azienda selezionata non contiene alcun indirizzo IP. Aggiungi target prima di avviare la scansione.');
      return;
    }

    setSubmitting(true);
    try {
      let finalProfileId = selectedProfileId;

      // In modalità Multi-Obiettivo: creiamo automaticamente un profilo con la combinazione selezionata
      if (configMode === 'MULTI_OBJECTIVE') {
        if (selectedObjectiveIds.length === 0) {
          setError('Seleziona almeno un obiettivo di scansione per continuare');
          setSubmitting(false);
          return;
        }

        const selectedObjs = OBJECTIVE_OPTIONS.filter((o) => selectedObjectiveIds.includes(o.id));
        const objNames = selectedObjs.map((o) => o.label).join(', ');
        const profileName = `Multi-Audit ${new Date().toISOString().slice(11, 19)} (${selectedObjectiveIds.length} Obiettivi)`;
        const profileDesc = `Scansione multi-obiettivo combinata: ${objNames}`;

        const createRes = await apiClient.post<ApiResponse<ScanProfileItem>>('/scan-profiles', {
          name: profileName,
          nmap_args: generatedMultiObjectiveArgs,
          description: profileDesc,
          timeout_seconds: 3600,
          category: 'CUSTOM',
        });

        if (!createRes.data || !createRes.data.success || !createRes.data.data) {
          throw new Error(createRes.data?.error?.message || 'Errore nella creazione del profilo multi-obiettivo');
        }

        finalProfileId = createRes.data.data.id;
        fetchProfiles();
      } else if (configMode === 'PROFESSIONAL') {
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
        const response = await apiClient.post<ApiResponse<ScanItem[] | ScanItem>>('/scans', {
          group_id: selectedGroupId,
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

        const validTargets = selectedGroupInfo.targets.filter((t) => !t.is_blacklisted);
        if (validTargets.length === 0) {
          setError('Nessun target valido (non in blacklist) nella cartella da pianificare');
          setSubmitting(false);
          return;
        }

        await Promise.all(
          validTargets.map((t) =>
            apiClient.post<ApiResponse<ScheduleItem>>('/schedules', {
              target_id: t.id,
              profile_id: finalProfileId,
              cron_expression: cronExpr,
              is_active: true,
            })
          )
        );

        if (onScheduleCreated) onScheduleCreated();
        onClose();
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

          {/* Mode Selector: Guidata vs Multi-Obiettivo vs Professionale & Guida Semplice */}
          <div className="pb-2 flex flex-wrap items-center gap-1.5">
            <div className="flex items-center space-x-1 bg-slate-950/60 p-1 rounded-lg border border-cyber-border">
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
                <span>Guidata (1-Click)</span>
              </button>
              <button
                type="button"
                onClick={() => setConfigMode('MULTI_OBJECTIVE')}
                className={`px-2.5 py-1 rounded text-xs font-medium flex items-center space-x-1.5 transition-all ${
                  configMode === 'MULTI_OBJECTIVE'
                    ? 'bg-cyan-500 text-slate-950 font-semibold shadow-sm'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                <Layers className="w-3 h-3" />
                <span>Componi Multi-Obiettivo</span>
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
                <span>Professionale (Pro)</span>
              </button>
            </div>

            <button
              type="button"
              onClick={() => setShowHelpGuide(!showHelpGuide)}
              className={`px-2.5 py-1.5 rounded-lg text-xs font-medium flex items-center space-x-1.5 transition-all ${
                showHelpGuide
                  ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40 shadow-sm'
                  : 'bg-slate-950/60 text-slate-400 hover:text-amber-300 border border-cyber-border hover:border-amber-500/30'
              }`}
              title="Mostra spiegazione semplice su ogni categoria e obiettivo"
            >
              <HelpCircle className="w-3.5 h-3.5 text-amber-400" />
              <span className="hidden sm:inline">Guida Semplice</span>
              {showHelpGuide ? <ChevronUp className="w-3 h-3 ml-0.5" /> : <ChevronDown className="w-3 h-3 ml-0.5" />}
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

          {/* Pannello Espandibile: Guida Semplice alla Scelta */}
          {showHelpGuide && (
            <div className="p-4 rounded-xl border border-amber-500/30 bg-amber-950/20 text-xs space-y-3 animate-in fade-in">
              <div className="flex items-center justify-between pb-2 border-b border-amber-500/20">
                <div className="flex items-center space-x-2 text-amber-300 font-semibold">
                  <HelpCircle className="w-4 h-4 text-amber-400 shrink-0" />
                  <span>💡 Guida Semplice: Come Scegliere la Scansione Ideale per la Tua Rete</span>
                </div>
                <button
                  type="button"
                  onClick={() => setShowHelpGuide(false)}
                  className="text-amber-400 hover:text-amber-200 p-1"
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              </div>
              <p className="text-[11px] text-slate-300 leading-relaxed">
                Non sai quale scegliere? Consulta la guida rapida qui sotto per capire quale obiettivo fa al caso tuo:
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5 max-h-60 overflow-y-auto pr-1">
                {CATEGORIES.map((cat) => {
                  const guide = CATEGORY_HELP_GUIDE[cat.id];
                  if (!guide) return null;
                  return (
                    <div key={cat.id} className="p-2.5 rounded-lg border border-slate-800 bg-slate-900/90 space-y-1">
                      <div className="flex items-center justify-between">
                        <span className="font-semibold text-white">{cat.name}</span>
                        <span className="text-[10px] text-cyan-400 font-mono">{guide.estimatedDuration}</span>
                      </div>
                      <p className="text-[11px] text-slate-300 leading-snug">{guide.simpleExplanation}</p>
                      <div className="text-[10px] text-slate-400 pt-1 border-t border-slate-800">
                        <strong className="text-amber-300">Quando usarlo:</strong> {guide.whenToUse}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {loadingData ? (
            <div className="py-12 text-center text-xs text-slate-400">
              <div className="inline-block animate-spin rounded-full h-6 w-6 border-2 border-cyan-500 border-t-transparent mb-2"></div>
              <p>Caricamento asset e catalogo profili...</p>
            </div>
          ) : (
            <>
              {/* Selezione Azienda / Cartella Target */}
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="block text-xs font-medium text-slate-300">
                    Azienda / Cartella Target da Scansionare <span className="text-red-400">*</span>
                  </label>
                  <span className="text-[11px] text-slate-400">
                    Audit consolidato per tutti gli IP della cartella
                  </span>
                </div>

                {flattenedGroups.length === 0 ? (
                  <div className="p-3 rounded-lg border border-cyber-border bg-slate-900 text-xs text-amber-400">
                    Nessuna cartella o azienda configurata. Crea prima una cartella nella sezione Target & Scope.
                  </div>
                ) : (
                  <select
                    value={selectedGroupId}
                    onChange={(e) => setSelectedGroupId(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-900 border border-cyber-border rounded-lg text-xs text-slate-200 focus:outline-none focus:border-cyan-500 font-sans"
                    required
                  >
                    <option value="" disabled>Seleziona un'azienda o cartella...</option>
                    {flattenedGroups.map((g) => (
                      <option key={g.id} value={g.id}>
                        {'\u00A0'.repeat(g.depth * 3) + (g.depth > 0 ? '↳ 📁 ' : '🏢 ') + g.name} ({g.targetCount} IP censiti)
                      </option>
                    ))}
                  </select>
                )}

                {/* Card preview dell'azienda / cartella selezionata */}
                {selectedGroupInfo && (
                  <div className="mt-2.5 p-3 rounded-xl border border-cyan-500/30 bg-cyan-950/20 text-xs">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Folder className="w-4 h-4 text-cyan-400 shrink-0" />
                        <span className="font-bold text-white">{selectedGroupInfo.name}</span>
                        <span className="text-[11px] px-2 py-0.5 rounded-full bg-slate-900 border border-cyber-border text-cyan-300 font-mono">
                          {selectedGroupInfo.targetCount} IP inclusi
                        </span>
                      </div>
                    </div>
                    {selectedGroupInfo.targets.length > 0 ? (
                      <div className="mt-2 flex flex-wrap gap-1.5 max-h-20 overflow-y-auto pr-1">
                        {selectedGroupInfo.targets.map((t) => (
                          <span
                            key={t.id}
                            className={`px-2 py-0.5 rounded text-[11px] font-mono border ${
                              t.is_blacklisted
                                ? 'bg-rose-950/30 text-rose-400 border-rose-800/40 line-through'
                                : 'bg-slate-900 text-slate-300 border-cyber-border'
                            }`}
                          >
                            {t.value}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <p className="mt-1.5 text-[11px] text-amber-400/90">
                        ⚠️ Questa cartella non ha ancora alcun indirizzo IP assegnato. Aggiungi prima dei bersagli per avviare la scansione.
                      </p>
                    )}
                  </div>
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

                  {/* Spiegazione semplice per la Categoria Selezionata */}
                  {CATEGORY_HELP_GUIDE[selectedCategory] && (
                    <div className="p-3 rounded-xl border border-cyan-500/20 bg-slate-900/60 text-xs space-y-1.5 animate-in fade-in">
                      <div className="flex items-center justify-between text-slate-300">
                        <span className="flex items-center space-x-1.5 text-cyan-400 font-semibold">
                          <Info className="w-3.5 h-3.5" />
                          <span>Obiettivo Categoria: {CATEGORIES.find((c) => c.id === selectedCategory)?.name}</span>
                        </span>
                        <span className="text-[10px] font-mono text-cyan-300 bg-cyan-950/60 px-2 py-0.5 rounded border border-cyan-800/40">
                          {CATEGORY_HELP_GUIDE[selectedCategory].estimatedDuration}
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-300 leading-relaxed">
                        {CATEGORY_HELP_GUIDE[selectedCategory].simpleExplanation}
                      </p>
                      <div className="text-[10px] text-slate-400 pt-1 border-t border-slate-800/80 flex flex-wrap items-center justify-between gap-1">
                        <span><strong className="text-amber-400">Consigliato quando:</strong> {CATEGORY_HELP_GUIDE[selectedCategory].whenToUse}</span>
                        <span className="text-slate-500 font-mono hidden sm:inline">{CATEGORY_HELP_GUIDE[selectedCategory].networkImpact}</span>
                      </div>
                    </div>
                  )}

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

              {/* SEZIONE 2: MODALITÀ MULTI-OBIETTIVO (COMPOSIZIONE LIBERA) */}
              {configMode === 'MULTI_OBJECTIVE' && (
                <div className="space-y-4">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div>
                      <label className="text-xs font-medium text-slate-300">
                        Seleziona uno o più Obiettivi di Scansione
                      </label>
                      <p className="text-[11px] text-slate-400">
                        Scegli gli audit desiderati: il motore unisce le porte e attiva i flag Nmap sicuri.
                      </p>
                    </div>
                    <div className="flex space-x-2 text-[11px]">
                      <button
                        type="button"
                        onClick={() => setSelectedObjectiveIds(OBJECTIVE_OPTIONS.map((o) => o.id))}
                        className="text-cyan-400 hover:underline"
                      >
                        Seleziona tutti
                      </button>
                      <span className="text-slate-600">|</span>
                      <button
                        type="button"
                        onClick={() =>
                          setSelectedObjectiveIds([
                            'OBJ_WEB',
                            'OBJ_SERVICES',
                            'OBJ_VERSIONS',
                            'OBJ_OS',
                          ])
                        }
                        className="text-slate-400 hover:text-white"
                      >
                        Standard 360°
                      </button>
                      <span className="text-slate-600">|</span>
                      <button
                        type="button"
                        onClick={() => setSelectedObjectiveIds([])}
                        className="text-rose-400 hover:underline"
                      >
                        Deseleziona
                      </button>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                    {OBJECTIVE_OPTIONS.map((obj) => {
                      const Icon = obj.icon;
                      const isChecked = selectedObjectiveIds.includes(obj.id);
                      return (
                        <label
                          key={obj.id}
                          className={`flex items-start space-x-3 p-3 rounded-xl border cursor-pointer transition-all ${
                            isChecked
                              ? 'border-cyan-500 bg-cyan-500/10 shadow-[0_0_12px_rgba(6,182,212,0.15)] ring-1 ring-cyan-500/40'
                              : 'border-cyber-border bg-slate-900/60 hover:border-slate-700'
                          }`}
                        >
                          <input
                            type="checkbox"
                            checked={isChecked}
                            onChange={(e) => {
                              if (e.target.checked) {
                                setSelectedObjectiveIds((prev) => [...prev, obj.id]);
                              } else {
                                setSelectedObjectiveIds((prev) => prev.filter((id) => id !== obj.id));
                              }
                            }}
                            className="mt-1 rounded border-slate-700 text-cyan-500 focus:ring-0"
                          />
                          <div className="flex-1">
                            <div className="flex items-center space-x-1.5">
                              <Icon className={`w-3.5 h-3.5 ${isChecked ? 'text-cyan-400' : 'text-slate-400'}`} />
                              <span className="text-xs font-semibold text-white">{obj.label}</span>
                            </div>
                            <p className="text-[11px] text-slate-400 mt-0.5 leading-relaxed">
                              {obj.simpleDesc}
                            </p>
                            {obj.ports && (
                              <div className="mt-1 text-[10px] text-slate-500 font-mono">
                                Porte: {obj.ports.slice(0, 6).join(', ')}
                                {obj.ports.length > 6 ? ` (+altre ${obj.ports.length - 6})` : ''}
                              </div>
                            )}
                          </div>
                        </label>
                      );
                    })}
                  </div>

                  <div className="p-3 rounded-xl border border-slate-800 bg-slate-900/60 flex items-center justify-between text-xs text-slate-300">
                    <div className="flex items-center space-x-2">
                      <span className="w-2 h-2 rounded-full bg-cyan-400 animate-pulse"></span>
                      <span>
                        Obiettivi attivi: <strong className="text-white font-mono">{selectedObjectiveIds.length}</strong> su{' '}
                        {OBJECTIVE_OPTIONS.length}
                      </span>
                    </div>
                    <span className="text-[11px] text-cyan-300 font-mono bg-cyan-950/60 px-2 py-0.5 rounded border border-cyan-800/40">
                      Unione porte & flag automatica
                    </span>
                  </div>
                </div>
              )}

              {/* SEZIONE 3: MODALITÀ PROFESSIONALE (CONFIGURATORE AVANZATO) */}
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

                      {/* Script Vulnerabilità Nmap */}
                      <label className="flex items-center space-x-2 p-2 rounded-lg border border-cyber-border bg-slate-900/50 cursor-pointer hover:border-slate-700 sm:col-span-2">
                        <input
                          type="checkbox"
                          checked={flagVulnScripts}
                          onChange={(e) => setFlagVulnScripts(e.target.checked)}
                          className="rounded border-slate-700 text-cyan-500 focus:ring-0"
                        />
                        <div className="flex-1">
                          <span className="text-white font-medium">Script Vulnerabilità Nmap (--script=vuln)</span>
                          <p className="text-[10px] text-slate-500">Esegue script attivi di rilevamento vulnerabilità e falle critiche CVE note</p>
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
                      ? profiles.find((p) => p.id === selectedProfileId)?.nmap_args || '-sS -sV -O -T4 --open'
                      : configMode === 'MULTI_OBJECTIVE'
                      ? generatedMultiObjectiveArgs
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
              disabled={Boolean(submitting || !selectedGroupId || (selectedGroupInfo && selectedGroupInfo.targets.length === 0))}
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
