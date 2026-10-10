import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Radar,
  Play,
  RefreshCw,
  Search,
  CheckCircle2,
  AlertCircle,
  Copy,
  Check,
  FolderPlus,
  ArrowRight,
  Server,
  Zap,
  Clock,
  X,
} from 'lucide-react';
import { discoveryApi } from '../api/discovery';
import { apiClient } from '../api/client';
import { useSSE } from '../hooks/useSSE';
import { ApiResponse } from '../types/auth';
import {
  DiscoveryResult,
  ImportTargetsResponseData,
  DeviceDiscoveredPayload,
  DiscoveryProgressPayload,
  DiscoveryCompletePayload,
  DiscoveryErrorPayload,
} from '../types/discovery';
import { TargetGroupNode } from '../types/target';

interface FlattenedGroup {
  id: string;
  name: string;
  color_tag?: string | null;
  depth: number;
}

export const NetworkDiscovery: React.FC = () => {
  const navigate = useNavigate();

  // Stato input e parametri scansione
  const [subnet, setSubnet] = useState<string>('192.168.1.0/24');
  const [includePorts, setIncludePorts] = useState<boolean>(true);
  const [isScanning, setIsScanning] = useState<boolean>(false);
  const [loadingInitial, setLoadingInitial] = useState<boolean>(true);

  // Stato sessione streaming SSE e avanzamento real-time
  const [streamSessionId, setStreamSessionId] = useState<string | null>(null);
  const [scanProgress, setScanProgress] = useState<{
    percentage: number;
    scannedHosts: number;
    totalHosts: number;
    currentIP: string;
  } | null>(null);
  const [newlyDiscoveredIPs, setNewlyDiscoveredIPs] = useState<Set<string>>(new Set());

  // Risultati ricognizione
  const [discoveryResult, setDiscoveryResult] = useState<DiscoveryResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  // URL per lo stream SSE
  const streamUrl = useMemo(
    () => (streamSessionId ? discoveryApi.getStreamUrl(streamSessionId) : null),
    [streamSessionId]
  );

  // Hook Server-Sent Events per aggiornamento live
  useSSE({
    url: streamUrl,
    enabled: isScanning && Boolean(streamSessionId),
    onEvent: (eventType, data) => {
      if (eventType === 'device_discovered') {
        const payload = data as DeviceDiscoveredPayload;
        if (payload?.device) {
          setDiscoveryResult((prev) => {
            const currentDevices = prev?.devices || [];
            if (currentDevices.some((d) => d.ip === payload.device.ip)) {
              return prev;
            }
            const updated = [payload.device, ...currentDevices];
            return {
              id: prev?.id || payload.session_id,
              tenant_id: prev?.tenant_id || '',
              subnet: prev?.subnet || subnet,
              devices: updated,
              total_discovered: updated.length,
              duration_seconds: prev?.duration_seconds || 0,
              started_at: prev?.started_at || new Date().toISOString(),
              completed_at: prev?.completed_at || '',
            };
          });
          setNewlyDiscoveredIPs((prev) => new Set(prev).add(payload.device.ip));
        }
      } else if (eventType === 'progress') {
        const payload = data as DiscoveryProgressPayload;
        if (payload) {
          setScanProgress({
            percentage: payload.percentage,
            scannedHosts: payload.scanned_hosts,
            totalHosts: payload.total_hosts,
            currentIP: payload.current_ip,
          });
        }
      } else if (eventType === 'complete') {
        const payload = data as DiscoveryCompletePayload;
        if (payload) {
          setDiscoveryResult((prev) => {
            if (!prev) return prev;
            return {
              ...prev,
              total_discovered: payload.total_discovered,
              duration_seconds: payload.duration_seconds,
              completed_at: new Date().toISOString(),
            };
          });
        }
      } else if (eventType === 'error') {
        const payload = data as DiscoveryErrorPayload;
        if (payload?.message) {
          setError(payload.message);
        }
      }
    },
  });

  // Filtro ricerca client-side
  const [searchTerm, setSearchTerm] = useState<string>('');

  // Selezione dispositivi
  const [selectedIPs, setSelectedIPs] = useState<Set<string>>(new Set());
  const [copiedIP, setCopiedIP] = useState<string | null>(null);

  // Gruppi ad albero per importazione
  const [groupOptions, setGroupOptions] = useState<FlattenedGroup[]>([]);
  const [targetGroupId, setTargetGroupId] = useState<string>('');

  // Creazione nuova cartella durante l'importazione
  const [importMode, setImportMode] = useState<'select' | 'create'>('select');
  const [newFolderName, setNewFolderName] = useState<string>('');
  const [newFolderDesc, setNewFolderDesc] = useState<string>('');
  const [newFolderColor, setNewFolderColor] = useState<string>('#06b6d4');

  // Stato modale importazione & feedback
  const [isImportModalOpen, setIsImportModalOpen] = useState<boolean>(false);
  const [isImporting, setIsImporting] = useState<boolean>(false);
  const [importResult, setImportResult] = useState<ImportTargetsResponseData | null>(null);
  const [importError, setImportError] = useState<string | null>(null);

  // Preset rapidi per le subnet più diffuse
  const presets = ['192.168.1.0/24', '192.168.0.0/24', '10.0.0.0/24', '172.16.0.0/24'];

  // Caricamento gruppi cartelle per l'importazione
  const fetchGroups = useCallback(async () => {
    try {
      const res = await apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups');
      if (res.data?.success && res.data.data) {
        const flatten = (nodes: TargetGroupNode[], depth = 0): FlattenedGroup[] => {
          let acc: FlattenedGroup[] = [];
          for (const node of nodes) {
            acc.push({
              id: node.id,
              name: node.name,
              color_tag: node.color_tag,
              depth,
            });
            if (node.children && node.children.length > 0) {
              acc = acc.concat(flatten(node.children, depth + 1));
            }
          }
          return acc;
        };
        setGroupOptions(flatten(res.data.data));
      }
    } catch {
      // Ignora o lascia lista gruppi vuota
    }
  }, []);

  // Caricamento ultimi risultati memorizzati al mount
  const loadLastResults = useCallback(async () => {
    setLoadingInitial(true);
    setError(null);
    try {
      const res = await discoveryApi.getLastResults();
      if (res) {
        setDiscoveryResult(res);
        if (res.subnet) {
          setSubnet(res.subnet);
        }
      }
    } catch (err: any) {
      // Se 404 è normale al primo avvio
    } finally {
      setLoadingInitial(false);
    }
  }, []);

  useEffect(() => {
    loadLastResults();
    fetchGroups();
  }, [loadLastResults, fetchGroups]);

  // Avvio nuova scansione rapida LAN
  const handleStartScan = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const clean = subnet.trim();
    if (!clean) {
      setError('Inserisci una subnet o range IP valido');
      return;
    }

    const sessionId = crypto.randomUUID();
    setStreamSessionId(sessionId);
    setIsScanning(true);
    setError(null);
    setSelectedIPs(new Set());
    setNewlyDiscoveredIPs(new Set());
    setScanProgress({
      percentage: 0,
      scannedHosts: 0,
      totalHosts: 0,
      currentIP: '',
    });

    // Inizializza un contenitore per accogliere i dispositivi in tempo reale
    setDiscoveryResult({
      id: sessionId,
      tenant_id: '',
      subnet: clean,
      devices: [],
      total_discovered: 0,
      duration_seconds: 0,
      started_at: new Date().toISOString(),
      completed_at: '',
    });

    try {
      const res = await discoveryApi.scan({
        subnet: clean,
        include_common_ports: includePorts,
        session_id: sessionId,
      });
      setDiscoveryResult(res);
    } catch (err: any) {
      setError(err.message || 'Errore durante la scansione LAN');
    } finally {
      setIsScanning(false);
      setScanProgress(null);
    }
  };

  // Dispositivi filtrati
  const filteredDevices = useMemo(() => {
    if (!discoveryResult || !discoveryResult.devices) return [];
    if (!searchTerm.trim()) return discoveryResult.devices;

    const term = searchTerm.toLowerCase();
    return discoveryResult.devices.filter(
      (d) =>
        d.ip.toLowerCase().includes(term) ||
        (d.hostname && d.hostname.toLowerCase().includes(term)) ||
        (d.vendor && d.vendor.toLowerCase().includes(term)) ||
        (d.mac && d.mac.toLowerCase().includes(term))
    );
  }, [discoveryResult, searchTerm]);

  // Gestione selezione checkbox
  const handleToggleSelect = (ip: string) => {
    setSelectedIPs((prev) => {
      const next = new Set(prev);
      if (next.has(ip)) {
        next.delete(ip);
      } else {
        next.add(ip);
      }
      return next;
    });
  };

  const handleSelectAll = () => {
    if (selectedIPs.size === filteredDevices.length && filteredDevices.length > 0) {
      setSelectedIPs(new Set());
    } else {
      setSelectedIPs(new Set(filteredDevices.map((d) => d.ip)));
    }
  };

  // Copia IP negli appunti
  const handleCopyIP = (ip: string) => {
    navigator.clipboard.writeText(ip);
    setCopiedIP(ip);
    setTimeout(() => setCopiedIP(null), 2000);
  };

  // Importazione massiva nei target
  const handleOpenImportModal = () => {
    setImportError(null);
    setImportResult(null);
    setImportMode('select');
    setNewFolderName('');
    setNewFolderDesc('');
    setNewFolderColor('#06b6d4');
    setIsImportModalOpen(true);
  };

  const handleExecuteImport = async () => {
    if (!discoveryResult || selectedIPs.size === 0) return;

    setIsImporting(true);
    setImportError(null);

    let finalGroupId: string | null = targetGroupId ? targetGroupId : null;

    if (importMode === 'create') {
      const trimmedName = newFolderName.trim();
      if (!trimmedName) {
        setImportError('Inserisci il nome della nuova cartella / azienda');
        setIsImporting(false);
        return;
      }

      try {
        const createRes = await apiClient.post<ApiResponse<TargetGroupNode>>('/target-groups', {
          name: trimmedName,
          description: newFolderDesc.trim() || undefined,
          color_tag: newFolderColor,
        });

        if (createRes.data?.success && createRes.data.data) {
          finalGroupId = createRes.data.data.id;
          fetchGroups();
        } else {
          setImportError(createRes.data?.error?.message || 'Impossibile creare la nuova cartella');
          setIsImporting(false);
          return;
        }
      } catch (err: any) {
        setImportError(
          err.response?.data?.error?.message ||
          err.message ||
          'Errore durante la creazione della nuova cartella'
        );
        setIsImporting(false);
        return;
      }
    }

    const devicesToImport = discoveryResult.devices
      .filter((d) => selectedIPs.has(d.ip))
      .map((d) => ({
        ip: d.ip,
        hostname: d.hostname || undefined,
        vendor: d.vendor || undefined,
      }));

    try {
      const res = await discoveryApi.importTargets({
        group_id: finalGroupId,
        devices: devicesToImport,
      });
      setImportResult(res);
    } catch (err: any) {
      setImportError(err.message || 'Errore durante l\'importazione dei target');
    } finally {
      setIsImporting(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Intestazione e Statistiche rapide */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-cyber-border pb-5">
        <div>
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center text-cyan-400 shadow-[0_0_20px_rgba(6,182,212,0.2)]">
              <Radar className="w-5 h-5 animate-pulse" />
            </div>
            <div>
              <h1 className="text-xl lg:text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                Network Discovery
                <span className="text-xs font-mono font-medium px-2 py-0.5 rounded-full bg-cyan-950 text-cyan-400 border border-cyan-800">
                  Advanced IP Scanner
                </span>
              </h1>
              <p className="text-xs lg:text-sm text-slate-400">
                Ricognizione LAN ad altissima velocità, identificazione hardware OUI e importazione bersagli
              </p>
            </div>
          </div>
        </div>

        {/* Badge metriche riepilogative */}
        {discoveryResult && (
          <div className="flex flex-wrap items-center gap-3 text-xs font-mono">
            <div className="px-3 py-1.5 rounded-lg bg-slate-900/90 border border-cyber-border flex items-center space-x-2">
              <span className="text-slate-500">Subnet:</span>
              <span className="text-cyan-400 font-semibold">{discoveryResult.subnet}</span>
            </div>
            <div className="px-3 py-1.5 rounded-lg bg-slate-900/90 border border-cyber-border flex items-center space-x-2">
              <span className="text-slate-500">Dispositivi UP:</span>
              <span className="text-emerald-400 font-semibold">{discoveryResult.total_discovered}</span>
            </div>
            <div className="px-3 py-1.5 rounded-lg bg-slate-900/90 border border-cyber-border flex items-center space-x-2">
              <Clock className="w-3.5 h-3.5 text-slate-500" />
              <span className="text-slate-300">{discoveryResult.duration_seconds.toFixed(2)}s</span>
            </div>
          </div>
        )}
      </div>

      {/* Barra Controlli Scansione */}
      <div className="rounded-xl border border-cyber-border bg-[#0d121f] p-4 lg:p-5 shadow-lg space-y-4">
        <form onSubmit={handleStartScan} className="flex flex-col lg:flex-row gap-4 items-stretch lg:items-center">
          <div className="flex-1 relative">
            <label htmlFor="subnet-input" className="block text-xs font-medium text-slate-400 mb-1">
              Subnet o Range IP Bersaglio (CIDR / Range)
            </label>
            <div className="relative">
              <input
                id="subnet-input"
                type="text"
                value={subnet}
                onChange={(e) => setSubnet(e.target.value)}
                placeholder="es. 192.168.1.0/24 oppure 10.0.0.1-50"
                disabled={isScanning}
                className="w-full px-3.5 py-2.5 rounded-lg bg-slate-950/80 border border-cyber-border text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500 transition-colors font-mono"
              />
            </div>
          </div>

          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 pt-0 lg:pt-5">
            <label className="flex items-center space-x-2.5 px-3 py-2 rounded-lg bg-slate-900/80 border border-cyber-border text-xs text-slate-300 cursor-pointer hover:bg-slate-850 select-none">
              <input
                type="checkbox"
                checked={includePorts}
                onChange={(e) => setIncludePorts(e.target.checked)}
                disabled={isScanning}
                className="rounded border-slate-700 bg-slate-950 text-cyan-500 focus:ring-cyan-500/20 w-4 h-4"
              />
              <span className="font-medium">Top Porte TCP (-F)</span>
            </label>

            <button
              type="submit"
              disabled={isScanning}
              className={`flex items-center justify-center space-x-2 px-5 py-2.5 rounded-lg font-semibold text-sm transition-all shadow-md ${
                isScanning
                  ? 'bg-cyan-900/40 text-cyan-300 border border-cyan-700/50 cursor-not-allowed'
                  : 'bg-cyan-500 hover:bg-cyan-400 text-slate-950 shadow-[0_0_15px_rgba(6,182,212,0.3)] hover:shadow-[0_0_20px_rgba(6,182,212,0.5)]'
              }`}
            >
              {isScanning ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin text-cyan-400" />
                  <span>Ricognizione in corso...</span>
                </>
              ) : (
                <>
                  <Play className="w-4 h-4 fill-current" />
                  <span>Avvia Discovery</span>
                </>
              )}
            </button>

            <button
              type="button"
              onClick={loadLastResults}
              disabled={isScanning}
              className="p-2.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors"
              title="Ricarica ultimi risultati memorizzati"
            >
              <RefreshCw className="w-4 h-4" />
            </button>
          </div>
        </form>

        {/* Preset veloci Subnet */}
        <div className="flex items-center space-x-2 text-xs text-slate-500 pt-1">
          <span>Preset rapidi:</span>
          <div className="flex flex-wrap gap-1.5">
            {presets.map((p) => (
              <button
                key={p}
                type="button"
                onClick={() => setSubnet(p)}
                disabled={isScanning}
                className={`px-2.5 py-1 rounded bg-slate-950 border text-[11px] font-mono transition-colors ${
                  subnet === p
                    ? 'border-cyan-500/50 text-cyan-400 bg-cyan-950/30'
                    : 'border-cyber-border text-slate-400 hover:text-slate-200 hover:bg-slate-900'
                }`}
              >
                {p}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Banner di scansione attiva con barra di avanzamento real-time */}
      {isScanning && (
        <div className="p-4 rounded-xl border border-cyan-500/40 bg-cyan-950/20 shadow-md space-y-3 animate-in fade-in duration-300">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div className="flex items-center space-x-3 text-cyan-300">
              <Zap className="w-5 h-5 text-cyan-400 animate-pulse flex-shrink-0" />
              <div>
                <span className="font-semibold text-white text-xs sm:text-sm">
                  Ricognizione LAN Live in corso su {subnet}
                </span>
                {scanProgress?.currentIP ? (
                  <p className="text-cyan-400/90 text-xs font-mono mt-0.5">
                    Probe attivo: <span className="text-white font-bold">{scanProgress.currentIP}</span>{' '}
                    <span className="text-slate-400">
                      ({scanProgress.scannedHosts} di {scanProgress.totalHosts} host analizzati)
                    </span>
                  </p>
                ) : (
                  <p className="text-cyan-400/80 text-xs mt-0.5">
                    Identificazione host attivi, risoluzione hostname e lookup vendor OUI IEEE...
                  </p>
                )}
              </div>
            </div>
            <div className="flex items-center space-x-2 text-xs font-mono self-end sm:self-auto">
              <span className="text-slate-400">Progresso:</span>
              <span className="text-cyan-400 font-bold text-sm">
                {(scanProgress?.percentage ?? 0).toFixed(1)}%
              </span>
            </div>
          </div>

          {/* Barra progresso animata */}
          <div className="w-full h-2 rounded-full bg-slate-900 border border-cyan-900/50 overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-cyan-500 to-emerald-400 transition-all duration-300 ease-out shadow-[0_0_10px_rgba(6,182,212,0.5)]"
              style={{ width: `${Math.min(100, Math.max(2, scanProgress?.percentage ?? 0))}%` }}
            />
          </div>
        </div>
      )}

      {/* Notifica di errore */}
      {error && (
        <div className="p-4 rounded-xl border border-rose-900/50 bg-rose-950/20 flex items-center space-x-3 text-rose-300">
          <AlertCircle className="w-5 h-5 flex-shrink-0 text-rose-400" />
          <div className="text-sm font-medium">{error}</div>
        </div>
      )}

      {/* Azioni di selezione e ricerca */}
      {discoveryResult && discoveryResult.devices.length > 0 && (
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-[#0d121f] p-3 rounded-xl border border-cyber-border">
          <div className="relative flex-1 max-w-md">
            <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Cerca per IP, Hostname, Produttore o MAC..."
              className="w-full pl-9 pr-3 py-1.5 rounded-lg bg-slate-950 border border-cyber-border text-xs text-slate-200 placeholder-slate-600 focus:outline-none focus:border-cyan-500"
            />
          </div>

          <div className="flex items-center space-x-3 text-xs">
            <span className="text-slate-400">
              Mostrati <strong className="text-white">{filteredDevices.length}</strong> di{' '}
              <strong className="text-white">{discoveryResult.devices.length}</strong> host
            </span>

            {selectedIPs.size > 0 && (
              <button
                type="button"
                onClick={handleOpenImportModal}
                className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-semibold shadow-[0_0_15px_rgba(16,185,129,0.25)] transition-all"
              >
                <FolderPlus className="w-3.5 h-3.5" />
                <span>Importa Selezionati ({selectedIPs.size})</span>
              </button>
            )}
          </div>
        </div>
      )}

      {/* Tabella Dispositivi Rilevati */}
      <div className="rounded-xl border border-cyber-border bg-[#0d121f] overflow-hidden shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-[#111726] border-b border-cyber-border text-slate-400 font-semibold uppercase tracking-wider text-[11px]">
              <tr>
                <th className="py-3 px-4 w-10 text-center">
                  <input
                    type="checkbox"
                    checked={
                      filteredDevices.length > 0 && selectedIPs.size === filteredDevices.length
                    }
                    onChange={handleSelectAll}
                    disabled={filteredDevices.length === 0}
                    className="rounded border-slate-700 bg-slate-950 text-cyan-500 focus:ring-cyan-500/20 w-4 h-4 cursor-pointer"
                    title="Seleziona tutti i dispositivi visualizzati"
                  />
                </th>
                <th className="py-3 px-4">Stato</th>
                <th className="py-3 px-4">Indirizzo IP</th>
                <th className="py-3 px-4">Produttore (Vendor OUI)</th>
                <th className="py-3 px-4">Hostname</th>
                <th className="py-3 px-4">MAC Address</th>
                <th className="py-3 px-4">Latenza (RTT)</th>
                <th className="py-3 px-4">Porte Aperte</th>
                <th className="py-3 px-4 text-right">Azioni Rapide</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-cyber-border/40 font-mono">
              {loadingInitial ? (
                <tr>
                  <td colSpan={9} className="py-12 text-center text-slate-500">
                    <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2 text-cyan-500" />
                    Caricamento stato di ricognizione...
                  </td>
                </tr>
              ) : filteredDevices.length === 0 ? (
                <tr>
                  <td colSpan={9} className="py-12 text-center text-slate-500">
                    {discoveryResult ? (
                      <div>
                        <Server className="w-8 h-8 text-slate-600 mx-auto mb-2" />
                        <p className="text-slate-300 font-medium">Nessun dispositivo attivo corrispondente ai criteri</p>
                        <p className="text-slate-500 text-[11px] mt-1 font-sans">
                          {searchTerm
                            ? 'Prova a modificare o cancellare il filtro di ricerca'
                            : 'Nessun host responsivo (alive) rilevato nella subnet specificata'}
                        </p>
                      </div>
                    ) : (
                      <div>
                        <Radar className="w-8 h-8 text-slate-600 mx-auto mb-2" />
                        <p className="text-slate-300 font-medium">Nessuna scansione discovery eseguita</p>
                        <p className="text-slate-500 text-[11px] mt-1 font-sans">
                          Specifica la subnet nel box superiore e premi "Avvia Discovery" per censire la LAN locale.
                        </p>
                      </div>
                    )}
                  </td>
                </tr>
              ) : (
                filteredDevices.map((dev) => {
                  const isSelected = selectedIPs.has(dev.ip);
                  const isNew = newlyDiscoveredIPs.has(dev.ip);
                  return (
                    <tr
                      key={dev.ip}
                      onClick={() => handleToggleSelect(dev.ip)}
                      className={`cursor-pointer transition-colors ${
                        isSelected
                          ? 'bg-cyan-500/10 hover:bg-cyan-500/15'
                          : isNew
                          ? 'bg-emerald-950/20 hover:bg-emerald-950/30'
                          : 'hover:bg-slate-900/60'
                      }`}
                    >
                      {/* Checkbox */}
                      <td className="py-3 px-4 text-center" onClick={(e) => e.stopPropagation()}>
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleToggleSelect(dev.ip)}
                          className="rounded border-slate-700 bg-slate-950 text-cyan-500 focus:ring-cyan-500/20 w-4 h-4 cursor-pointer"
                        />
                      </td>

                      {/* Stato UP */}
                      <td className="py-3 px-4">
                        <div className="flex items-center space-x-1.5">
                          <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded-full bg-emerald-950 text-emerald-400 border border-emerald-800 text-[10px] font-semibold">
                            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                            <span>UP</span>
                          </span>
                          {isNew && (
                            <span className="px-1.5 py-0.2 rounded bg-cyan-950 text-cyan-400 border border-cyan-800 text-[9px] font-sans font-bold uppercase animate-pulse">
                              LIVE
                            </span>
                          )}
                        </div>
                      </td>

                      {/* IP con pulsante copia */}
                      <td className="py-3 px-4 font-bold text-white flex items-center space-x-2">
                        <span>{dev.ip}</span>
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCopyIP(dev.ip);
                          }}
                          className="text-slate-500 hover:text-cyan-400 transition-colors p-1"
                          title="Copia IP"
                        >
                          {copiedIP === dev.ip ? (
                            <Check className="w-3.5 h-3.5 text-emerald-400" />
                          ) : (
                            <Copy className="w-3.5 h-3.5" />
                          )}
                        </button>
                      </td>

                      {/* Vendor OUI */}
                      <td className="py-3 px-4">
                        {dev.vendor && dev.vendor !== 'Unknown' ? (
                          <span className="inline-flex items-center px-2 py-0.5 rounded bg-slate-900 border border-cyan-800/40 text-cyan-300 text-[11px] font-sans font-medium">
                            {dev.vendor}
                          </span>
                        ) : (
                          <span className="text-slate-600 font-sans italic text-[11px]">Sconosciuto</span>
                        )}
                      </td>

                      {/* Hostname */}
                      <td className="py-3 px-4 text-slate-300 font-sans">
                        {dev.hostname ? (
                          <span className="font-medium text-slate-200">{dev.hostname}</span>
                        ) : (
                          <span className="text-slate-600 italic text-[11px]">-</span>
                        )}
                      </td>

                      {/* MAC */}
                      <td className="py-3 px-4 text-slate-400 text-[11px]">
                        {dev.mac || <span className="text-slate-600 italic">-</span>}
                      </td>

                      {/* Latenza RTT */}
                      <td className="py-3 px-4">
                        <span
                          className={`font-semibold ${
                            dev.latency_ms < 5
                              ? 'text-emerald-400'
                              : dev.latency_ms < 20
                              ? 'text-yellow-400'
                              : 'text-amber-500'
                          }`}
                        >
                          {dev.latency_ms.toFixed(2)} ms
                        </span>
                      </td>

                      {/* Porte Aperte */}
                      <td className="py-3 px-4">
                        {dev.open_ports && dev.open_ports.length > 0 ? (
                          <div className="flex flex-wrap gap-1">
                            {dev.open_ports.map((p) => (
                              <span
                                key={p}
                                className="px-1.5 py-0.2 rounded bg-slate-900 border border-slate-700 text-[10px] text-cyan-400 font-mono"
                              >
                                {p}
                              </span>
                            ))}
                          </div>
                        ) : (
                          <span className="text-slate-600 text-[11px] italic">-</span>
                        )}
                      </td>

                      {/* Azione rapida singola */}
                      <td className="py-3 px-4 text-right" onClick={(e) => e.stopPropagation()}>
                        <button
                          type="button"
                          onClick={() => {
                            setSelectedIPs(new Set([dev.ip]));
                            handleOpenImportModal();
                          }}
                          className="px-2.5 py-1 rounded-md bg-slate-900 hover:bg-slate-800 border border-cyber-border text-cyan-400 hover:text-cyan-300 text-[11px] font-sans font-medium transition-colors"
                        >
                          Importa
                        </button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Barra mobile/sticky se ci sono elementi selezionati */}
      {selectedIPs.size > 0 && (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-30 bg-[#111726]/95 border border-cyan-500/50 backdrop-blur-md rounded-2xl px-6 py-3.5 shadow-2xl flex items-center space-x-4 max-w-xl w-[90%] justify-between animate-in fade-in slide-in-from-bottom-4 duration-200">
          <div className="flex items-center space-x-2">
            <span className="w-2.5 h-2.5 rounded-full bg-cyan-400 animate-ping" />
            <span className="text-xs sm:text-sm font-semibold text-white">
              {selectedIPs.size} dispositivo/i selezionato/i
            </span>
          </div>

          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={() => setSelectedIPs(new Set())}
              className="px-3 py-1.5 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-850 text-slate-300 text-xs transition-colors"
            >
              Annulla
            </button>
            <button
              type="button"
              onClick={handleOpenImportModal}
              className="flex items-center space-x-1.5 px-4 py-1.5 rounded-lg bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold text-xs shadow-lg transition-all"
            >
              <FolderPlus className="w-3.5 h-3.5" />
              <span>Importa nei Target</span>
            </button>
          </div>
        </div>
      )}

      {/* Modal di Importazione Massiva */}
      {isImportModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#0d121f] border border-cyber-border rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-5 animate-in zoom-in-95 duration-150">
            <div className="flex items-center justify-between border-b border-cyber-border pb-3">
              <div className="flex items-center space-x-2 text-white font-bold text-base">
                <FolderPlus className="w-5 h-5 text-cyan-400" />
                <span>Importazione Bersagli LAN</span>
              </div>
              <button
                type="button"
                onClick={() => setIsImportModalOpen(false)}
                className="text-slate-400 hover:text-white p-1 rounded-lg"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Contenuto modale: prima dell'importazione */}
            {!importResult ? (
              <div className="space-y-4">
                <p className="text-xs text-slate-300 leading-relaxed">
                  Stai per importare <strong className="text-cyan-400">{selectedIPs.size}</strong> dispositivo/i come
                  Target ufficiali monitorati per questa organizzazione.
                </p>

                {/* Scelta Modalità: Cartella Esistente vs Crea Nuova Cartella */}
                <div className="space-y-2">
                  <label className="block text-xs font-medium text-slate-400">
                    Destinazione Bersagli
                  </label>
                  <div className="grid grid-cols-2 gap-2 p-1 bg-slate-950 rounded-lg border border-cyber-border">
                    <button
                      type="button"
                      onClick={() => setImportMode('select')}
                      className={`py-1.5 px-3 rounded-md text-xs font-semibold transition-all ${
                        importMode === 'select'
                          ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-sm'
                          : 'text-slate-400 hover:text-white'
                      }`}
                    >
                      Cartella Esistente
                    </button>
                    <button
                      type="button"
                      onClick={() => setImportMode('create')}
                      className={`py-1.5 px-3 rounded-md text-xs font-semibold transition-all ${
                        importMode === 'create'
                          ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-sm'
                          : 'text-slate-400 hover:text-white'
                      }`}
                    >
                      + Nuova Cartella (Azienda)
                    </button>
                  </div>
                </div>

                {importMode === 'select' ? (
                  /* Selezione cartella Scope Esistente */
                  <div className="space-y-1.5">
                    <label htmlFor="target-group-select" className="block text-xs font-medium text-slate-400">
                      Cartella Scope (Gerarchia ad Albero)
                    </label>
                    <div className="relative">
                      <select
                        id="target-group-select"
                        value={targetGroupId}
                        onChange={(e) => setTargetGroupId(e.target.value)}
                        disabled={isImporting}
                        className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-xs text-slate-200 focus:outline-none focus:border-cyan-500"
                      >
                        <option value="">📁 Cartella Principale (Root / Senza Cartella)</option>
                        {groupOptions.map((g) => (
                          <option key={g.id} value={g.id}>
                            {'\u00A0'.repeat(g.depth * 3) + (g.depth > 0 ? '↳ ' : '📁 ') + g.name}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                ) : (
                  /* Creazione Istantanea Nuova Cartella */
                  <div className="space-y-3 p-3.5 rounded-xl border border-cyber-border/80 bg-slate-900/50">
                    <div className="space-y-1">
                      <label className="block text-[11px] font-medium text-slate-300">
                        Nome Cartella / Azienda *
                      </label>
                      <input
                        type="text"
                        value={newFolderName}
                        onChange={(e) => setNewFolderName(e.target.value)}
                        placeholder="es. ACME Corporation, Sede Centrale, Server DMZ..."
                        className="w-full px-3 py-1.5 rounded-lg bg-slate-950 border border-cyber-border text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                      />
                    </div>

                    <div className="space-y-1">
                      <label className="block text-[11px] font-medium text-slate-300">
                        Descrizione (opzionale)
                      </label>
                      <input
                        type="text"
                        value={newFolderDesc}
                        onChange={(e) => setNewFolderDesc(e.target.value)}
                        placeholder="es. Rete uffici e postazioni lavoro..."
                        className="w-full px-3 py-1.5 rounded-lg bg-slate-950 border border-cyber-border text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
                      />
                    </div>

                    <div className="space-y-1.5">
                      <label className="block text-[11px] font-medium text-slate-300">
                        Etichetta Colore
                      </label>
                      <div className="flex items-center space-x-2">
                        {['#06b6d4', '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'].map((color) => (
                          <button
                            key={color}
                            type="button"
                            onClick={() => setNewFolderColor(color)}
                            className={`w-5 h-5 rounded-full transition-transform border ${
                              newFolderColor === color
                                ? 'scale-125 border-white ring-2 ring-cyan-500/50'
                                : 'border-transparent opacity-70 hover:opacity-100 hover:scale-110'
                            }`}
                            style={{ backgroundColor: color }}
                          />
                        ))}
                      </div>
                    </div>
                  </div>
                )}

                <p className="text-[11px] text-slate-500">
                  I dispositivi manterranno il Vendor OUI e l'Hostname censiti nella descrizione.
                </p>

                {importError && (
                  <div className="p-3 rounded-lg border border-rose-900/50 bg-rose-950/20 text-rose-300 text-xs flex items-center space-x-2">
                    <AlertCircle className="w-4 h-4 flex-shrink-0" />
                    <span>{importError}</span>
                  </div>
                )}

                <div className="flex items-center justify-end space-x-3 pt-3 border-t border-cyber-border">
                  <button
                    type="button"
                    onClick={() => setIsImportModalOpen(false)}
                    disabled={isImporting}
                    className="px-4 py-2 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
                  >
                    Annulla
                  </button>
                  <button
                    type="button"
                    onClick={handleExecuteImport}
                    disabled={isImporting}
                    className="flex items-center space-x-1.5 px-4 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold text-xs transition-all shadow-md"
                  >
                    {isImporting ? (
                      <>
                        <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                        <span>Importazione in corso...</span>
                      </>
                    ) : (
                      <>
                        <CheckCircle2 className="w-3.5 h-3.5" />
                        <span>Conferma Importazione</span>
                      </>
                    )}
                  </button>
                </div>
              </div>
            ) : (
              /* Risultato dell'importazione */
              <div className="space-y-4">
                <div className="p-4 rounded-xl bg-slate-900/80 border border-cyber-border space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-slate-400">Target Importati con Successo:</span>
                    <span className="text-sm font-bold text-emerald-400 font-mono">
                      {importResult.total_imported}
                    </span>
                  </div>
                  {importResult.total_skipped > 0 && (
                    <div className="flex items-center justify-between border-t border-cyber-border/40 pt-2">
                      <span className="text-xs text-slate-400">Target Scartati (Duplicati / Blacklist):</span>
                      <span className="text-sm font-bold text-amber-400 font-mono">
                        {importResult.total_skipped}
                      </span>
                    </div>
                  )}
                </div>

                {importResult.errors && importResult.errors.length > 0 && (
                  <div className="max-h-36 overflow-y-auto space-y-1.5 pr-1">
                    {importResult.errors.map((err, idx) => (
                      <div
                        key={idx}
                        className="text-[11px] p-2 rounded bg-slate-950 border border-cyber-border flex items-center justify-between text-slate-400 font-mono"
                      >
                        <span className="text-slate-200">{err.ip}</span>
                        <span className="text-amber-400 font-sans">{err.reason}</span>
                      </div>
                    ))}
                  </div>
                )}

                <div className="flex items-center justify-between pt-3 border-t border-cyber-border">
                  <button
                    type="button"
                    onClick={() => {
                      setIsImportModalOpen(false);
                      setSelectedIPs(new Set());
                    }}
                    className="px-4 py-2 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium"
                  >
                    Chiudi
                  </button>
                  <button
                    type="button"
                    onClick={() => navigate('/targets')}
                    className="flex items-center space-x-1.5 px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-bold text-xs"
                  >
                    <span>Vai a Target & Asset</span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
