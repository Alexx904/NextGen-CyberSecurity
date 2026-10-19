import React, { useEffect, useMemo, useRef, useState } from 'react';
import { X, Play, Terminal } from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetGroupNode, TargetItem, TargetListResponseData } from '../types/target';
import { ScanItem, ScanProfileItem } from '../types/scan';
import { getAllDescendantGroupIds, useCompany } from '../context/CompanyContext';
import { SCAN_PRESETS, canonicalArgs, commandError, isIntrusive, normalizeNmapInput, profileDurationLabel } from './scans/scanConfiguration';
import { scanAuthorization } from './scans/scanAuthorization';

interface NewScanModalProps {
  isOpen: boolean;
  onClose: () => void;
  onScanLaunched: (scans?: ScanItem[]) => void;
  onScheduleCreated?: () => void;
  defaultGroupId?: string;
  targetValues?: string[];
}

const CRON_PRESETS = [
  { label: 'Ogni 6 ore', value: '0 */6 * * *' },
  { label: 'Ogni notte alle 02:00', value: '0 2 * * *' },
  { label: 'Ogni lunedì alle 03:00', value: '0 3 * * 1' },
  { label: 'Primo giorno del mese alle 04:00', value: '0 4 1 * *' },
  { label: 'Personalizzato', value: 'CUSTOM' },
];

function errorMessage(error: unknown): string {
  const response = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
  return response.response?.data?.error?.message || response.message || 'Impossibile completare la richiesta.';
}

const fieldClass = 'w-full rounded-lg border border-cyber-border bg-slate-950 p-2.5 text-sm text-slate-100 focus:outline-hidden focus:ring-2 focus:ring-cyan-500';

export const NewScanModal: React.FC<NewScanModalProps> = ({ isOpen, onClose, onScanLaunched, onScheduleCreated, defaultGroupId, targetValues }) => {
  const { selectedCompanyId } = useCompany();
  const dialogRef = useRef<HTMLElement>(null);
  const [mode, setMode] = useState<'PRESETS' | 'COMMAND'>('PRESETS');
  const [groups, setGroups] = useState<TargetGroupNode[]>([]);
  const [targets, setTargets] = useState<TargetItem[]>([]);
  const [profiles, setProfiles] = useState<ScanProfileItem[]>([]);
  const [groupId, setGroupId] = useState('');
  const [presetId, setPresetId] = useState(SCAN_PRESETS[0].id);
  const [command, setCommand] = useState(SCAN_PRESETS[0].args);
  const [profileName, setProfileName] = useState('');
  const [timeout, setTimeoutMinutes] = useState(60);
  const [scheduled, setScheduled] = useState(false);
  const [cronPreset, setCronPreset] = useState('0 2 * * *');
  const [customCron, setCustomCron] = useState('0 2 * * *');
  const [confirmed, setConfirmed] = useState(false);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadFailed, setLoadFailed] = useState(false);
  const [completedScheduleIds, setCompletedScheduleIds] = useState<Set<string>>(new Set());
  const [completedScanIds, setCompletedScanIds] = useState<Set<string>>(new Set());
  const [launchUncertain, setLaunchUncertain] = useState(false);
  const selectionKey = targetValues === undefined ? null : JSON.stringify([...new Set(targetValues)].sort());

  useEffect(() => {
    if (!isOpen) return;
    const previous = document.activeElement as HTMLElement | null;
    dialogRef.current?.focus();
    return () => { previous?.focus(); };
  }, [isOpen]);

  function handleDialogKey(event: React.KeyboardEvent<HTMLElement>) {
    if (event.key === 'Escape' && !submitting) { onClose(); return; }
    if (event.key !== 'Tab') return;
    const focusable = Array.from(dialogRef.current?.querySelectorAll<HTMLElement>('button, select, input, textarea, summary') || [])
      .filter((element) => !element.matches(':disabled') && !element.closest('details:not([open]) > :not(summary)'));
    const first = focusable[0]; const last = focusable[focusable.length - 1];
    if (!first) { event.preventDefault(); return; }
    if (event.shiftKey && (document.activeElement === first || document.activeElement === dialogRef.current)) { event.preventDefault(); last.focus(); }
    else if (!event.shiftKey && (document.activeElement === last || document.activeElement === dialogRef.current)) { event.preventDefault(); first.focus(); }
  }

  useEffect(() => {
    if (!isOpen) return;
    let cancelled = false;
    setLoading(true); setLoadFailed(false); setError(null); setConfirmed(false); setCompletedScheduleIds(new Set());
    async function load() {
      try {
        const [groupRes, profileRes] = await Promise.all([
          apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups'),
          apiClient.get<ApiResponse<{ items: ScanProfileItem[] }>>('/scan-profiles'),
        ]);
        if (!groupRes.data.success || !groupRes.data.data || !profileRes.data.success || !profileRes.data.data) throw new Error('Impossibile caricare gruppi e profili.');
        const loadedTargets: TargetItem[] = [];
        let offset = 0;
        while (true) {
          const res = await apiClient.get<ApiResponse<TargetListResponseData>>(`/targets?limit=200&offset=${offset}`);
          if (!res.data.success || !res.data.data) throw new Error('Impossibile caricare i target.');
          const page = res.data.data;
          loadedTargets.push(...page.items);
          offset += page.items.length;
          if (offset >= page.total) break;
          if (!page.items.length) throw new Error('Elenco target incompleto. Riapri la finestra per riprovare.');
        }
        if (cancelled) return;
        setGroups(groupRes.data.data); setProfiles(profileRes.data.data.items); setTargets(loadedTargets);
        setGroupId(defaultGroupId || selectedCompanyId || groupRes.data.data[0]?.id || '');
      } catch (err) {
        if (!cancelled) { setError(errorMessage(err)); setLoadFailed(true); }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void load();
    return () => { cancelled = true; };
  }, [isOpen, defaultGroupId, selectedCompanyId]);

  const availableGroups = useMemo(() => {
    const all: { node: TargetGroupNode; depth: number }[] = [];
    const visit = (nodes: TargetGroupNode[], depth: number) => nodes.forEach((node) => {
      all.push({ node, depth }); visit(node.children || [], depth + 1);
    });
    visit(groups, 0);
    const company = all.find(({ node }) => node.id === selectedCompanyId)?.node;
    if (selectedCompanyId && !company) return [];
    const scope = company ? new Set(getAllDescendantGroupIds(company)) : null;
    return all.filter(({ node }) => !scope || scope.has(node.id));
  }, [groups, selectedCompanyId]);
  const selectedGroup = availableGroups.find(({ node }) => node.id === groupId)?.node;
  const scopedTargets = useMemo(() => {
    if (!selectedGroup) return [];
    const ids = new Set(getAllDescendantGroupIds(selectedGroup));
    const selectedValues: Set<string> | null = selectionKey === null ? null : new Set(JSON.parse(selectionKey));
    const candidates = targets.filter((target) => target.group_id && ids.has(target.group_id) && (!selectedValues || selectedValues.has(target.value)));
    if (!selectedValues) return candidates;
    // One device can already be inventoried in multiple descendant folders.
    // Honor any blacklist entry, otherwise prefer its direct destination entry.
    const seen = new Set<string>();
    return candidates.sort((a, b) => Number(b.is_blacklisted) - Number(a.is_blacklisted) || Number(b.group_id === selectedGroup.id) - Number(a.group_id === selectedGroup.id))
      .filter(target => { if (seen.has(target.value)) return false; seen.add(target.value); return true; });
  }, [targets, selectedGroup, selectionKey]);
  const eligibleTargets = scopedTargets.filter((target) => !target.is_blacklisted);
  const preset = SCAN_PRESETS.find((item) => item.id === presetId) || SCAN_PRESETS[0];
  const args = mode === 'PRESETS' ? preset.args : normalizeNmapInput(command);
  const validationError = mode === 'COMMAND' ? commandError(command) : null;
  const matchingProfile = profiles.find((profile) => canonicalArgs(profile.nmap_args) === canonicalArgs(args)
    && profile.timeout_seconds === (mode === 'PRESETS' ? preset.timeout : timeout * 60)
    && (!profileName.trim() || mode === 'PRESETS' || profile.name === profileName.trim()));
  const intrusive = isIntrusive(args) || matchingProfile?.invasiveness === 'INTRUSIVE' || matchingProfile?.invasiveness === 'DANGEROUS';
  const timeoutInvalid = mode === 'COMMAND' && (!Number.isInteger(timeout) || timeout < 1 || timeout > 1440);
  const cron = cronPreset === 'CUSTOM' ? customCron.trim() : cronPreset;
  const cronInvalid = scheduled && cron.split(/\s+/).length !== 5;
  useEffect(() => { setConfirmed(false); }, [args, groupId, scheduled, cron, timeout]);
  useEffect(() => { setCompletedScheduleIds(new Set()); }, [args, groupId, cron, timeout, scheduled]);
  useEffect(() => { setCompletedScanIds(new Set()); setConfirmed(false); }, [isOpen, args, groupId, timeout, scheduled, selectionKey]);
  useEffect(() => { setLaunchUncertain(false); }, [isOpen, selectionKey]);
  const disabled = loading || loadFailed || submitting || launchUncertain || !selectedGroup || eligibleTargets.length === 0 || !!validationError || timeoutInvalid || cronInvalid || !confirmed;

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (disabled) return;
    setError(null); setSubmitting(true);
    try {
      let profile = matchingProfile;
      if (!profile) {
        const res = await apiClient.post<ApiResponse<ScanProfileItem>>('/scan-profiles', {
          name: mode === 'PRESETS' ? preset.name : profileName.trim() || `Nmap personalizzato ${new Date().toISOString()}`,
          description: mode === 'PRESETS' ? preset.description : 'Opzioni Nmap definite dall’operatore. Target e output gestiti dalla piattaforma.',
          nmap_args: args, timeout_seconds: mode === 'PRESETS' ? preset.timeout : timeout * 60, category: 'CUSTOM',
        });
        if (!res.data.success || !res.data.data) throw new Error(res.data.error?.message || 'Creazione profilo fallita.');
        profile = res.data.data;
        const created = profile;
        setProfiles((previous) => [...previous, created]);
      }
      if (!scheduled && targetValues !== undefined) {
        const pending = eligibleTargets.filter(target => !completedScanIds.has(target.id));
        const profileId = profile.id;
        const results: PromiseSettledResult<{ targetId: string; scan: ScanItem }>[] = [];
        for (let offset = 0; offset < pending.length; offset += 4) {
          results.push(...await Promise.allSettled(pending.slice(offset, offset + 4).map(async target => {
          const res = await apiClient.post<ApiResponse<ScanItem>>('/scans', { target_id: target.id, profile_id: profileId, authorization: scanAuthorization([target], profile!) });
          if (!res.data.success || !res.data.data) throw new Error('Avvio target fallito.');
          return { targetId: target.id, scan: res.data.data };
          })));
        }
        const succeeded = new Set(completedScanIds);
        const scans: ScanItem[] = [];
        results.forEach(result => { if (result.status === 'fulfilled') { succeeded.add(result.value.targetId); scans.push(result.value.scan); } });
        setCompletedScanIds(succeeded);
        if (scans.length) onScanLaunched(scans);
        const failures = results.filter(result => result.status === 'rejected').length;
        const uncertain = results.some(result => {
          if (result.status !== 'rejected') return false;
          const status = (result.reason as { response?: { status?: number } })?.response?.status;
          return !status || status >= 500;
        });
        if (uncertain) {
          setLaunchUncertain(true);
          throw new Error(`${succeeded.size} scansioni confermate; altri esiti non confermati. Chiudi e verifica lo storico prima di avviare nuovamente: nessun retry automatico.`);
        }
        if (failures) throw new Error(`${succeeded.size} scansioni avviate; ${failures} non riuscite. Riprova solo quelle mancanti.`);
      } else if (!scheduled) {
        const res = await apiClient.post<ApiResponse<ScanItem[] | ScanItem>>('/scans', { group_id: groupId, profile_id: profile.id, authorization: scanAuthorization(eligibleTargets, profile) });
        if (!res.data.success) throw new Error(res.data.error?.message || 'Avvio scansione fallito.');
        onScanLaunched(res.data.data ? Array.isArray(res.data.data) ? res.data.data : [res.data.data] : []);
      } else {
        const pending = eligibleTargets.filter((target) => !completedScheduleIds.has(target.id));
        const profileId = profile.id;
        const results = await Promise.allSettled(pending.map(async (target) => {
          const res = await apiClient.post<ApiResponse<unknown>>('/schedules', { target_id: target.id, profile_id: profileId, cron_expression: cron, is_active: true, authorization: scanAuthorization([target], profile!, cron) });
          if (!res.data.success) throw new Error(res.data.error?.message || 'Creazione pianificazione fallita.');
          return target.id;
        }));
        const succeeded = new Set(completedScheduleIds);
        results.forEach((result) => { if (result.status === 'fulfilled') succeeded.add(result.value); });
        setCompletedScheduleIds(succeeded);
        const failures = results.filter((result) => result.status === 'rejected');
        if (failures.length) {
          onScheduleCreated?.();
          throw new Error(`${succeeded.size} pianificazioni create; ${failures.length} non riuscite. Riprova per completare solo quelle mancanti.`);
        }
        onScheduleCreated?.();
      }
      onClose();
    } catch (err) { setError(errorMessage(err)); }
    finally { setSubmitting(false); }
  }

  if (!isOpen) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-3" onClick={() => { if (!submitting) onClose(); }}>
      <section ref={dialogRef} tabIndex={-1} role="dialog" aria-modal="true" aria-labelledby="new-scan-title" className="flex max-h-[94vh] w-full max-w-2xl flex-col rounded-xl border border-cyber-border bg-slate-900 shadow-xl" onClick={(event) => event.stopPropagation()} onKeyDown={handleDialogKey}>
        <header className="flex items-center justify-between border-b border-cyber-border p-5">
          <div><h2 id="new-scan-title" className="text-lg font-semibold text-white">Nuova scansione Nmap</h2><p className="mt-1 text-xs text-slate-400">Porte, servizi e controlli di vulnerabilità su target autorizzati.</p></div>
          <button type="button" aria-label="Chiudi" disabled={submitting} onClick={onClose} className="rounded p-2 text-slate-400 focus:ring-2 focus:ring-cyan-500"><X className="h-5 w-5" /></button>
        </header>
        <form onSubmit={submit} className="overflow-y-auto p-5 space-y-5">
          {error && <p role="alert" className="rounded-lg border border-rose-500/40 bg-rose-950/30 p-3 text-sm text-rose-300">{error}</p>}
          {loading && <p role="status" className="text-sm text-slate-400">Caricamento target e profili…</p>}
          {targetValues !== undefined && <p role="note" className="text-xs text-cyan-300">Solo i dispositivi selezionati nella discovery e presenti nella cartella scelta. Nessun altro target della cartella viene avviato.</p>}
          <fieldset disabled={submitting || loading || loadFailed} className="space-y-5 disabled:opacity-60">
            <div>
              <label htmlFor="scan-group" className="mb-2 block text-sm text-slate-200">Azienda / reparto</label>
              <select id="scan-group" className={fieldClass} value={groupId} onChange={(event) => setGroupId(event.target.value)}>
                <option value="">Seleziona un gruppo</option>
                {availableGroups.map(({ node, depth }) => <option key={node.id} value={node.id}>{'— '.repeat(depth)}{node.name}</option>)}
              </select>
              <p className="mt-2 text-xs text-slate-400">{eligibleTargets.length} target disponibili · {scopedTargets.length - eligibleTargets.length} esclusi dalla blacklist · reparti inclusi</p>
              {!loading && !eligibleTargets.length && <p className="mt-1 text-xs text-amber-300">Nessun target disponibile nel gruppo selezionato.</p>}
            </div>
            <div className="flex gap-2" role="group" aria-label="Modalità scansione">
              <button type="button" aria-pressed={mode === 'PRESETS'} onClick={() => setMode('PRESETS')} className={`flex-1 rounded-lg border p-2 text-sm ${mode === 'PRESETS' ? 'border-cyan-500 text-cyan-300 bg-cyan-500/10' : 'border-cyber-border text-slate-400'}`}>Profili</button>
              <button type="button" aria-pressed={mode === 'COMMAND'} onClick={() => { if (mode !== 'COMMAND') { setCommand(args); setTimeoutMinutes(preset.timeout / 60); } setMode('COMMAND'); }} className={`flex-1 rounded-lg border p-2 text-sm ${mode === 'COMMAND' ? 'border-cyan-500 text-cyan-300 bg-cyan-500/10' : 'border-cyber-border text-slate-400'}`}>Comando Nmap</button>
            </div>
            {mode === 'PRESETS' ? <div>
              <label htmlFor="scan-preset" className="mb-2 block text-sm text-slate-200">Tipo di scansione</label>
              <select id="scan-preset" className={fieldClass} value={presetId} onChange={(event) => setPresetId(event.target.value)}>{SCAN_PRESETS.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select>
              <p className="mt-2 text-sm text-slate-400">{preset.description}</p>
              <p className="mt-2 text-xs text-cyan-300">Durata indicativa per IP: {profileDurationLabel(preset.id)}</p>
              <p className="mt-1 text-xs text-slate-500">Ipotesi: singolo IP raggiungibile, LAN stabile, T3. Filtri, UDP, NSE, subnet e coda possono allungare i tempi oltre questo intervallo. Non è il timeout.</p>
            </div> : <div className="space-y-3">
              <label htmlFor="nmap-command" className="block text-sm text-slate-200">Comando / opzioni Nmap</label>
              <textarea id="nmap-command" aria-describedby="nmap-command-help" aria-invalid={!!validationError} rows={3} className={`${fieldClass} font-mono`} value={command} onChange={(event) => setCommand(event.target.value)} spellCheck={false} />
              <p id="nmap-command-help" className="text-xs text-slate-400">Esempio: nmap -sS -sV -p 22,80,443 -T3. Inserisci solo opzioni: target, binario e output XML sono gestiti dalla piattaforma. Sono ammessi solo i flag supportati e gli script vuln / vulners.</p>
              {validationError && <p role="alert" className="text-xs text-amber-300">{validationError}</p>}
              <details className="text-sm text-slate-300"><summary className="cursor-pointer">Profili salvati e timeout</summary><div className="mt-3 space-y-3">
                <label htmlFor="saved-profile" className="block">Carica profilo salvato</label>
                <select id="saved-profile" value="" className={fieldClass} onChange={(event) => {
                  const saved = profiles.find((item) => item.id === event.target.value);
                  if (saved) { setCommand(saved.nmap_args); setTimeoutMinutes(saved.timeout_seconds / 60); setProfileName(''); }
                }}><option value="">Seleziona una configurazione</option>{profiles.filter((item) => item.is_custom).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select>
                <label htmlFor="profile-name" className="block">Nome profilo (facoltativo)</label><input id="profile-name" className={fieldClass} value={profileName} onChange={(event) => setProfileName(event.target.value)} maxLength={120} />
                <label htmlFor="scan-timeout" className="block">Timeout per target (minuti)</label><input id="scan-timeout" type="number" min={1} max={1440} className={fieldClass} value={timeout} onChange={(event) => setTimeoutMinutes(Number(event.target.value))} />
                {timeoutInvalid && <p role="alert" className="text-xs text-amber-300">Il timeout deve essere un intero tra 1 e 1440 minuti.</p>}
                <p className="text-xs text-slate-500">Le nuove configurazioni vengono salvate nel tenant per riuso e pianificazione. Il timeout è un limite, non una stima di durata.</p>
              </div></details>
            </div>}
            <div className="rounded-lg border border-cyber-border bg-slate-950 p-3 space-y-2">
              <p className="flex items-center gap-2 text-xs text-slate-400"><Terminal className="h-4 w-4" />Anteprima per ciascun target</p>
              <code data-testid="nmap-preview" className="block break-all text-xs text-cyan-300">nmap -oX - {args} --stats-every 5s -- {'<target>'}</code>
              <p className="text-xs text-slate-500">Validazione definitiva sul server al salvataggio. IPv6: -6 aggiunto automaticamente. Nessuna durata o assenza di impatto garantita.</p>
            </div>
            <label className="flex items-center gap-2 text-sm text-slate-300"><input type="checkbox" checked={scheduled} onChange={(event) => setScheduled(event.target.checked)} />Pianifica scansione ricorrente</label>
            {scheduled && <div className="space-y-2">
              <label htmlFor="scan-frequency" className="block text-sm text-slate-300">Frequenza (fuso orario del server)</label>
              <select id="scan-frequency" className={fieldClass} value={cronPreset} onChange={(event) => setCronPreset(event.target.value)}>{CRON_PRESETS.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select>
              {cronPreset === 'CUSTOM' && <><label htmlFor="scan-cron" className="block text-sm text-slate-300">Espressione cron a 5 campi</label><input id="scan-cron" className={fieldClass} value={customCron} onChange={(event) => setCustomCron(event.target.value)} /></>}
              {cronInvalid && <p role="alert" className="text-xs text-amber-300">Inserisci 5 campi cron; la sintassi verrà verificata dal server.</p>}
            </div>}
            <div className="rounded-lg border border-amber-500/30 p-3 space-y-2">
              <p className="text-sm text-amber-300">{intrusive ? 'Configurazione a impatto elevato o con controlli attivi' : 'Autorizzazione alla scansione attiva'}</p>
              <label className="flex items-start gap-2 text-xs text-slate-300"><input type="checkbox" checked={confirmed} onChange={(event) => setConfirmed(event.target.checked)} />Confermo di possedere esplicita autorizzazione scritta per questa configurazione e questo scope.</label>
              {scheduled && <p className="text-xs text-slate-400">Il consenso vale per le esecuzioni ricorrenti su questi target con queste opzioni, timeout e frequenza. Se cambiano, serve una nuova autorizzazione.</p>}
            </div>
            <details className="text-xs text-slate-400"><summary className="cursor-pointer">Limiti dei risultati</summary><p className="mt-2">La scansione non è un penetration test né una verifica di conformità. Le versioni, le CVE correlate e gli esiti NSE richiedono validazione professionale; risultati vuoti non dimostrano assenza di vulnerabilità. -Pn salta la verifica preliminare degli host, non aggira i firewall. SYN resta rilevabile.</p></details>
          </fieldset>
          <footer className="flex justify-end gap-3 border-t border-cyber-border pt-4">
            <button type="button" disabled={submitting} onClick={onClose} className="rounded-lg px-3 py-2 text-sm text-slate-400">Annulla</button>
            <button type="submit" disabled={disabled} className="flex items-center gap-2 rounded-lg bg-cyan-500 px-4 py-2 text-sm font-semibold text-slate-950 disabled:opacity-40"><Play className="h-4 w-4" />{submitting ? 'Invio…' : scheduled ? 'Crea pianificazione' : 'Avvia scansione'}</button>
          </footer>
        </form>
      </section>
    </div>
  );
};
