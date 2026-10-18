import React, { useEffect, useRef, useState } from 'react';
import { attackPathApi } from '../api/attackPath';
import { sshApi } from '../api/ssh';
import type { AttackNode, ChokepointRemediationResult } from '../types/attackPath';
import type { SSHServerItem } from '../types/ssh';

export function supportsChokepointMitigation(node: AttackNode): boolean {
  const parts = node.ip?.split('.') || [];
  const ipv4 = parts.length === 4 && parts.every(part => /^\d{1,3}$/.test(part) && Number(part) <= 255 && String(Number(part)) === part);
  return node.type === 'SERVICE' && ipv4 && node.ip !== '0.0.0.0' && !(Number(parts[0]) >= 224 && Number(parts[0]) <= 239)
    && ['tcp', 'udp'].includes(node.protocol || '') && Number.isInteger(node.port) && node.port! > 0 && node.port! <= 65535;
}

function message(error: unknown): string {
  const value = error as { response?: { data?: { error?: { message?: string } } }; message?: string };
  return value?.response?.data?.error?.message || value?.message || 'Richiesta non completata. Verifica lo stato del firewall.';
}

interface Props {
  node: AttackNode;
  onClose: () => void;
  onCheckFirewall: (server: SSHServerItem) => void;
}

export function ChokepointRemediationModal({ node, onClose, onCheckFirewall }: Props) {
  const [servers, setServers] = useState<SSHServerItem[]>([]);
  const [serverId, setServerId] = useState('');
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [preview, setPreview] = useState<ChokepointRemediationResult | null>(null);
  const [acknowledged, setAcknowledged] = useState(false);
  const [applyUncertain, setApplyUncertain] = useState(false);
  const dialog = useRef<HTMLElement>(null);
  const generation = useRef(0);
  const selected = servers.find(server => server.id === serverId);
  const supported = supportsChokepointMitigation(node);

  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null;
    dialog.current?.focus();
    let cancelled = false;
    setLoading(true); setBusy(false); setServerId(''); setServers([]); setPreview(null); setAcknowledged(false); setError(null); setApplyUncertain(false);
    sshApi.listServers().then(items => {
      if (!cancelled) setServers(items.filter(server => server.host === node.ip && !(node.protocol === 'tcp' && server.port === node.port)));
    }).catch(err => { if (!cancelled) setError(message(err)); }).finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; ++generation.current; previous?.focus(); };
  }, [node.id, node.ip, node.port, node.protocol]);

  const chooseServer = (id: string) => {
    ++generation.current;
    setServerId(id); setPreview(null); setAcknowledged(false); setError(null); setApplyUncertain(false);
  };

  async function loadPreview() {
    if (!selected || busy || !supported || applyUncertain) return;
    const current = ++generation.current;
    setBusy(true); setError(null); setPreview(null); setAcknowledged(false);
    try {
      const result = await attackPathApi.previewRemediation(node.id, selected.id);
      if (result.node_id !== node.id || result.server_id !== selected.id) throw new Error('Anteprima non coerente con il nodo e il server selezionati.');
      if (current === generation.current) setPreview(result);
    } catch (err) { if (current === generation.current) setError(message(err)); }
    finally { if (current === generation.current) setBusy(false); }
  }

  async function apply() {
    if (!selected || !supported || !preview || preview.node_id !== node.id || preview.server_id !== selected.id || !acknowledged || busy || applyUncertain || !selected.host_key_fingerprint) return;
    const current = ++generation.current;
    setBusy(true); setError(null);
    try {
      const result = await attackPathApi.remediate(node.id, selected.id);
      if (current !== generation.current) return;
      if (!result.firewall?.active_safety_timer || result.server_id !== selected.id || result.node_id !== node.id) throw new Error('Esito non confermato. Verifica lo stato del firewall prima di riprovare.');
      onCheckFirewall(selected);
    } catch (err) {
      if (current === generation.current) { setError(message(err)); setApplyUncertain(true); }
    } finally { if (current === generation.current) setBusy(false); }
  }

  function keyDown(event: React.KeyboardEvent<HTMLElement>) {
    if (event.key === 'Escape' && !busy) { onClose(); return; }
    if (event.key !== 'Tab') return;
    const controls = Array.from(dialog.current?.querySelectorAll<HTMLElement>('button,select,input,a[href]') || []).filter(element => !element.matches(':disabled'));
    const first = controls[0], last = controls[controls.length - 1];
    if (!first) return;
    if (event.shiftKey && (document.activeElement === first || document.activeElement === dialog.current)) { event.preventDefault(); last.focus(); }
    else if (!event.shiftKey && (document.activeElement === last || document.activeElement === dialog.current)) { event.preventDefault(); first.focus(); }
  }

  return <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-3">
    <section ref={dialog} role="dialog" aria-modal="true" aria-labelledby="mitigation-title" tabIndex={-1} onKeyDown={keyDown} className="max-h-[92vh] w-full max-w-xl overflow-y-auto rounded-xl border border-cyber-border bg-slate-900 p-5 space-y-4">
      <header className="flex items-center justify-between gap-3"><h2 id="mitigation-title" className="text-lg font-semibold text-white">Neutralizza Chokepoint</h2><button type="button" onClick={onClose} disabled={busy} aria-label="Chiudi mitigazione" className="text-slate-300">Chiudi</button></header>
      <p className="text-sm text-slate-300">Blocco porta su {node.ip}:{node.port}/{node.protocol}. La regola può interrompere il servizio e le connessioni esistenti.</p>
      {!supported && <p role="alert" className="text-amber-300">Sono supportati solo servizi IPv4 TCP/UDP. Questo nodo richiede una verifica manuale.</p>}
      {error && <p role="alert" className="text-rose-300">{error}</p>}
      {loading ? <p role="status">Caricamento server SSH…</p> : servers.length === 0 ? <p role="status" className="text-amber-300">Nessun server SSH compatibile con l’IP del servizio e la porta di gestione. Registra il server nella sezione SSH & Firewall.</p> : <div>
        <label htmlFor="mitigation-server" className="block text-sm text-slate-200">Server SSH da gestire</label>
        <select id="mitigation-server" value={serverId} disabled={busy || applyUncertain} onChange={event => chooseServer(event.target.value)} className="mt-2 w-full rounded border border-slate-700 bg-slate-950 p-2 text-white">
          <option value="">Seleziona esplicitamente un server</option>{servers.map(server => <option key={server.id} value={server.id}>{server.name} · {server.username}@{server.host}:{server.port}</option>)}
        </select>
      </div>}
      {selected && !selected.host_key_fingerprint && <p role="status" className="text-amber-300">Configura il fingerprint verificato del server nella sezione SSH prima di applicare la regola.</p>}
      <button type="button" disabled={!selected || busy || !supported || applyUncertain} onClick={loadPreview} className="rounded border border-cyan-700 px-3 py-2 text-cyan-300 disabled:opacity-40">Anteprima regola</button>
      {preview && <div className="space-y-3 text-sm text-slate-200">
        <p>{preview.rule.chain} · {preview.rule.action} · {preview.rule.destination}:{preview.rule.port}/{preview.rule.protocol}</p>
        <pre className="whitespace-pre-wrap break-all rounded bg-slate-950 p-3 text-xs">{preview.command}</pre>
        <p>Ripristino remoto dopo {preview.safety_timeout_seconds}s se non confermi. Il blocco della porta non dimostra la neutralizzazione dell’intero percorso.</p>
        <label className="flex gap-2"><input type="checkbox" checked={acknowledged} disabled={busy || applyUncertain} onChange={event => setAcknowledged(event.target.checked)} />Confermo l’impatto sul servizio e autorizzo questo blocco porta.</label>
        <button type="button" disabled={!acknowledged || !selected?.host_key_fingerprint || busy || applyUncertain} onClick={apply} className="rounded bg-amber-500 px-3 py-2 font-semibold text-slate-950 disabled:opacity-40">Applica blocco porta</button>
      </div>}
      {selected && <button type="button" disabled={busy} onClick={() => onCheckFirewall(selected)} className="block text-sm text-cyan-300 underline">Verifica stato firewall</button>}
    </section>
  </div>;
}
