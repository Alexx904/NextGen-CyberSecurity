import React, { useCallback, useEffect, useRef, useState } from 'react';
import { alertsApi, type AlertChannel, type AlertConfig, type AlertKind } from '../api/alerts';
import { useAuth } from '../context/AuthContext';
import { useCompany } from '../context/CompanyContext';

const eventLabels: [string, string][] = [
  ['finding.critical', 'Nuova CVE Critical'], ['finding.high', 'Nuova CVE High'],
  ['host.discovered', 'Nuovo host'], ['host.changed', 'Variazioni identità/porte host'],
  ['port.discovered', 'Nuova porta osservata'], ['scan.completed', 'Scansione programmata completata'], ['scan.failed', 'Scansione programmata fallita'],
];

export function AlertChannelsCard() {
  const { user } = useAuth();
  const { companies } = useCompany();
  const canManage = user?.role === 'TenantAdmin' || user?.role === 'SuperAdmin';
  const [channels, setChannels] = useState<AlertChannel[]>([]);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [visible, setVisible] = useState(false);
  const [editing, setEditing] = useState<AlertChannel | null>(null);
  const [name, setName] = useState('');
  const [kind, setKind] = useState<AlertKind>('webhook');
  const [company, setCompany] = useState('');
  const [enabled, setEnabled] = useState(false);
  const [events, setEvents] = useState(['finding.critical']);
  const [config, setConfig] = useState<AlertConfig>({});
  const [replaceConfig, setReplaceConfig] = useState(true);
  const [acknowledged, setAcknowledged] = useState(false);
  const [deleteCandidate, setDeleteCandidate] = useState<AlertChannel | null>(null);
  const generation = useRef(0);
  const listGeneration = useRef(0);
  const opener = useRef<HTMLElement | null>(null);
  const form = useRef<HTMLFormElement>(null);
  const nameInput = useRef<HTMLInputElement>(null);

  const load = useCallback(async () => {
    if (!canManage) return;
    const version = ++listGeneration.current;
    setLoading(true);
    try { const items = await alertsApi.list(); if (version === listGeneration.current) setChannels(items); }
    catch { if (version === listGeneration.current) setError('Impossibile caricare i canali. Ricarica per riprovare.'); }
    finally { if (version === listGeneration.current) setLoading(false); }
  }, [canManage]);
  useEffect(() => { void load(); return () => { ++generation.current; ++listGeneration.current; }; }, [load]);
  useEffect(() => { if (visible) nameInput.current?.focus(); }, [visible, editing]);

  function resetConsent() { ++generation.current; setAcknowledged(false); setMessage(null); }
  function open(channel?: AlertChannel) {
    opener.current = document.activeElement as HTMLElement | null;
    resetConsent(); setError(null); setEditing(channel || null); setName(channel?.name || ''); setKind(channel?.kind || 'webhook');
    setCompany(channel?.company_id || ''); setEnabled(channel?.enabled || false); setEvents(channel?.events || ['finding.critical']);
    setConfig({}); setReplaceConfig(!channel); setDeleteCandidate(null); setVisible(true);
  }
  function close() { resetConsent(); setConfig({}); setVisible(false); setEditing(null); setDeleteCandidate(null); opener.current?.focus(); }
  function updateConfig(patch: Partial<AlertConfig>) { resetConsent(); setConfig(current => ({ ...current, ...patch })); }
  function effectiveConfig(): AlertConfig { return kind === 'smtp' ? { ...config, smtp_port: config.smtp_port || 587 } : config; }

  async function save(event: React.FormEvent) {
    event.preventDefault(); if (busy || !events.length) return;
    setBusy(true); setError(null); setMessage(null);
    try {
      await alertsApi.save({ name, kind, enabled, company_id: company || undefined, events, expected_updated_at: editing?.updated_at, ...(replaceConfig ? { config: effectiveConfig() } : {}) }, editing?.id);
      close(); setMessage('Canale salvato. Le notifiche automatiche seguono lo stato attivo e gli eventi selezionati.'); await load();
    } catch (err) {
      const status = (err as { response?: { status?: number } }).response?.status;
      setError(status === 409 ? 'Il canale è cambiato. Annulla, ricarica e riapri la modifica.' : 'Salvataggio non riuscito. Verifica configurazione, ruolo e azienda.');
    } finally { setBusy(false); }
  }
  async function test() {
    if (busy || !acknowledged || !replaceConfig || !form.current?.reportValidity()) return;
    const version = ++generation.current;
    setBusy(true); setError(null); setMessage(null);
    try {
      const result = await alertsApi.test(kind, effectiveConfig());
      if (version === generation.current) setMessage(result.accepted ? 'Il provider ha accettato il messaggio sintetico. Verifica la ricezione nel canale.' : 'Accettazione non confermata. Verifica il canale.');
    } catch { if (version === generation.current) setError('Test non accettato. Verifica endpoint, TLS e credenziali. Nessun retry automatico.'); }
    finally { setBusy(false); }
  }
  async function remove(channel: AlertChannel) {
    if (busy) return; setBusy(true); setError(null); setMessage(null);
    try { await alertsApi.remove(channel.id); setDeleteCandidate(null); setMessage('Canale rimosso. Un invio già in corso può completarsi.'); await load(); }
    catch { setError('Rimozione non riuscita. Ricarica e riprova.'); }
    finally { setBusy(false); }
  }

  if (!canManage) return <section className="rounded-xl border border-cyber-border bg-slate-900 p-5"><h2 className="font-semibold text-white">Canali di notifica</h2><p className="mt-2 text-sm text-slate-400">La configurazione dei destinatari è riservata agli amministratori del tenant.</p></section>;

  const fieldClass = 'mt-1 block w-full rounded border border-slate-700 bg-slate-950 p-2 text-sm text-white';
  return <section className="space-y-4 rounded-xl border border-cyber-border bg-slate-900 p-5 text-slate-200" aria-labelledby="alert-channels-title">
    <header className="flex flex-wrap items-center justify-between gap-3"><h2 id="alert-channels-title" className="font-semibold text-white">Canali di notifica</h2><div className="flex gap-3"><button type="button" disabled={loading || busy} onClick={() => { setError(null); void load(); }}>Ricarica canali</button><button type="button" disabled={busy} onClick={() => open()}>Aggiungi canale</button></div></header>
    <p className="text-sm text-slate-400">I canali attivi inviano i dati minimi dei nuovi eventi ai destinatari configurati: tipo evento, ID, IP/porta e CVE quando presenti. Le variazioni host escludono latenza e timestamp. La presa in carico del provider non conferma la consegna. Scadenza certificati TLS non ancora disponibile.</p>
    {error && <p role="alert" className="text-rose-300">{error}</p>}{message && <p role="status" className="text-cyan-300">{message}</p>}
    {busy && <p role="status">Operazione canale in corso…</p>}
    {loading ? <p role="status">Caricamento canali…</p> : channels.length === 0 ? <p>Nessun canale configurato.</p> : <ul className="space-y-3">{channels.map(channel => <li key={channel.id} className="flex flex-wrap items-center justify-between gap-3 rounded border border-slate-700 p-3">
      <div><h3 className="font-semibold">{channel.name}</h3><p className="text-xs text-slate-400">{channel.kind} · {channel.enabled ? 'Attivo' : 'Disattivo'} · {channel.company_id ? companies.find(c => c.id === channel.company_id)?.name || 'Azienda configurata' : 'Intero tenant'} · {channel.events.length} eventi</p></div>
      <div className="flex gap-3"><button type="button" disabled={busy} onClick={() => open(channel)} aria-label={`Modifica canale ${channel.name}`}>Modifica</button><button type="button" disabled={busy} onClick={() => { close(); setDeleteCandidate(channel); }} aria-label={`Rimuovi canale ${channel.name}`}>Rimuovi</button></div>
    </li>)}</ul>}
    {deleteCandidate && <div className="space-y-3 rounded border border-amber-700 p-3"><p>Rimuovere {deleteCandidate.name}? La configurazione verrà rimossa. Nuovi invii si fermeranno; richieste già in corso possono completarsi.</p><button type="button" disabled={busy} onClick={() => void remove(deleteCandidate)}>Conferma rimozione canale</button><button type="button" disabled={busy} className="ml-4" onClick={() => setDeleteCandidate(null)}>Annulla rimozione</button></div>}
    {visible && <form ref={form} onSubmit={save} className="space-y-4 rounded border border-cyan-900 p-4">
      <h3 className="font-semibold">{editing ? 'Modifica canale' : 'Nuovo canale'}</h3>
      <fieldset disabled={busy} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2"><label>Nome canale<input ref={nameInput} required maxLength={160} value={name} onChange={event => { resetConsent(); setName(event.target.value); }} className={fieldClass} /></label>
          <label>Tipo canale<select value={kind} onChange={event => { resetConsent(); setKind(event.target.value as AlertKind); setConfig({}); setReplaceConfig(true); }} className={fieldClass}><option value="webhook">Webhook REST con HMAC</option><option value="slack">Slack</option><option value="teams">Teams Workflows</option><option value="telegram">Telegram Bot</option><option value="smtp">Email SMTP TLS</option></select></label>
          <label>Ambito canale<select value={company} onChange={event => { resetConsent(); setCompany(event.target.value); }} className={fieldClass}><option value="">Intero tenant</option>{companies.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}</select></label>
          <label className="flex items-center gap-2"><input type="checkbox" checked={enabled} onChange={event => { resetConsent(); setEnabled(event.target.checked); }} />Attiva notifiche automatiche</label>
        </div>
        <fieldset className="grid gap-2 sm:grid-cols-2"><legend className="mb-2">Eventi da inviare (almeno uno)</legend>{eventLabels.map(([id, label]) => <label key={id} className="flex items-center gap-2 text-sm"><input type="checkbox" checked={events.includes(id)} onChange={event => { resetConsent(); setEvents(current => event.target.checked ? [...current, id] : current.filter(value => value !== id)); }} />{label}</label>)}</fieldset>
        {editing && <label className="flex gap-2 text-sm"><input type="checkbox" disabled={editing.kind !== kind} checked={replaceConfig} onChange={event => { resetConsent(); setConfig({}); setReplaceConfig(event.target.checked); }} />Sostituisci configurazione e credenziali (quelle salvate non vengono mostrate)</label>}
        {replaceConfig && <div className="space-y-3">
          {['slack', 'teams', 'webhook'].includes(kind) && <label className="block">Endpoint HTTPS<input type="password" autoComplete="off" required value={config.endpoint || ''} onChange={event => updateConfig({ endpoint: event.target.value })} className={fieldClass} /><span className="text-xs text-slate-400">Destinazione pubblica su 443; niente redirect, endpoint LAN o localhost. Teams richiede un Workflow con scheda Adaptive Card.</span></label>}
          {kind === 'webhook' && <label className="block">Segreto HMAC (almeno 32 caratteri)<input type="password" autoComplete="new-password" required minLength={32} value={config.hmac_secret || ''} onChange={event => updateConfig({ hmac_secret: event.target.value })} className={fieldClass} /></label>}
          {kind === 'telegram' && <><label className="block">Token Bot Telegram<input type="password" autoComplete="new-password" required value={config.token || ''} onChange={event => updateConfig({ token: event.target.value })} className={fieldClass} /></label><label className="block">Chat ID Telegram<input required value={config.chat_id || ''} onChange={event => updateConfig({ chat_id: event.target.value })} className={fieldClass} /></label></>}
          {kind === 'smtp' && <div className="grid gap-3 sm:grid-cols-2"><label>Host SMTP pubblico<input required value={config.smtp_host || ''} onChange={event => updateConfig({ smtp_host: event.target.value })} className={fieldClass} /></label><label>Porta TLS<select value={config.smtp_port || 587} onChange={event => updateConfig({ smtp_port: Number(event.target.value) })} className={fieldClass}><option value={587}>587 · STARTTLS obbligatorio</option><option value={465}>465 · TLS implicito</option></select></label><label>Mittente email<input type="email" required value={config.smtp_from || ''} onChange={event => updateConfig({ smtp_from: event.target.value })} className={fieldClass} /></label><label>Destinatari email (separati da virgola)<input required value={config.smtp_to?.join(', ') || ''} onChange={event => updateConfig({ smtp_to: event.target.value.split(',').map(item => item.trim()) })} className={fieldClass} /></label><label>Utente SMTP (facoltativo)<input autoComplete="off" value={config.smtp_username || ''} onChange={event => updateConfig({ smtp_username: event.target.value })} className={fieldClass} /></label><label>Password SMTP<input type="password" autoComplete="new-password" value={config.smtp_password || ''} onChange={event => updateConfig({ smtp_password: event.target.value })} className={fieldClass} /></label></div>}
        </div>}
        <label className="flex gap-2 text-sm"><input type="checkbox" disabled={!replaceConfig} checked={acknowledged} onChange={event => setAcknowledged(event.target.checked)} />Autorizzo l’invio di un messaggio sintetico di prova al destinatario configurato.</label>
        <p className="text-xs text-slate-400">Il test usa la configurazione appena inserita, non salva né attiva il canale. Modifiche alla configurazione richiedono un nuovo consenso.</p>
        <div className="flex flex-wrap gap-4"><button type="button" disabled={!acknowledged || !replaceConfig} onClick={() => void test()}>Invia messaggio di prova</button><button type="submit" disabled={!events.length}>Salva canale</button><button type="button" onClick={close}>Annulla modifica canale</button></div>
      </fieldset>
    </form>}
  </section>;
}
