import { useEffect, useRef, useState } from 'react';
import { useCompany } from '../context/CompanyContext';
import { useAuth } from '../context/AuthContext';
import { complianceApi } from '../api/compliance';
import { AuditExport } from '../components/AuditExport';
import { ComplianceWorkflow } from './ComplianceWorkflow';
import type { ComplianceIssue, ComplianceSnapshot, IssueCommandInput, IssueEvent } from '../types/compliance';

const nil = '00000000-0000-0000-0000-000000000000';
function IssueHistory({ events }: { events: IssueEvent[] }) {
  return <ol>{events.map((event, n) => <li key={n} className="my-3 break-words border-b border-slate-800 pb-2">
    <p>{event.action} · {new Date(event.at).toLocaleString()} · autore {event.actor} · {event.reason}</p>
    {event.observation && <p className="text-sm text-slate-400">Finding {event.observation.finding_id} · scansione {event.observation.scan_id} · {event.observation.source_host} · confidenza {event.observation.confidence} · metodo {event.observation.detection_method}</p>}
    {event.verification && <p className="text-sm text-slate-400">Prova: {event.verification.provenance.reference} · {event.verification.provenance.source} v{event.verification.provenance.version}</p>}
  </li>)}</ol>;
}
const labels: Record<string, string> = { unknown: 'Non valutato', satisfied: 'Soddisfatto con evidenza manuale', partial: 'Parziale', not_satisfied: 'Gap rilevato', not_applicable: 'Escluso con motivazione', open: 'Aperta', mitigation_applied: 'Mitigazione applicata', correction_declared: 'Correzione dichiarata', corrected_verified: 'Correzione verificata' };
function errorMessage(error: unknown) { const e = error as { response?: { data?: { error?: { message?: string } } }; message?: string }; return e.response?.data?.error?.message || e.message || 'Operazione non riuscita'; }

export function Compliance() {
  const { selectedCompanyId, selectedCompany } = useCompany();
  if (!selectedCompanyId) return <p>Seleziona un’azienda per aprire Compliance.</p>;
  return <ComplianceBody key={selectedCompanyId} companyId={selectedCompanyId} companyName={selectedCompany?.name || 'Azienda'} />;
}

function ComplianceBody({ companyId, companyName }: { companyId: string; companyName: string }) {
  const { departments } = useCompany();
  const { user } = useAuth();
  const canWrite = !!user && ['SuperAdmin', 'TenantAdmin', 'SecOps'].includes(user.role);
  const [department, setDepartment] = useState('');
  const [framework, setFramework] = useState('ALL');
  const [data, setData] = useState<ComplianceSnapshot>();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [revision, setRevision] = useState(0);
  const [asset, setAsset] = useState('');
  const [editing, setEditing] = useState<ComplianceIssue>();
  const evidenceDetails = useRef<HTMLDetailsElement>(null);
  const editor = useRef<HTMLDivElement>(null);
  const lastScope = useRef('');
  const request = useRef(0);
  useEffect(() => {
    const id = ++request.current; const controller = new AbortController();
    setLoading(true);
    if (lastScope.current !== `${companyId}/${department}`) { setData(undefined); setEditing(undefined); setAsset(''); lastScope.current = `${companyId}/${department}`; }
    setError('');
    complianceApi.snapshot(companyId, department, controller.signal).then(result => { if (id === request.current) setData(result); }).catch(err => { if (!controller.signal.aborted && id === request.current) setError(errorMessage(err)); }).finally(() => { if (id === request.current) setLoading(false); });
    return () => { controller.abort(); request.current++; };
  }, [companyId, department, revision]);
  useEffect(() => {
    const timer = window.setInterval(() => {
      if (!busy && !loading && !editing && document.visibilityState === 'visible' && !document.activeElement?.closest('form') && !document.querySelector('.compliance-page details[open]')) setRevision(n => n + 1);
    }, 30000);
    return () => window.clearInterval(timer);
  }, [busy, loading, editing]);
  useEffect(() => { if (editing) { editor.current?.scrollIntoView?.({ block: 'nearest' }); editor.current?.focus(); } }, [editing]);
  async function mutate(work: () => Promise<unknown>) { setBusy(true); setError(''); try { await work(); setEditing(undefined); setRevision(n => n + 1); } catch (err) { setError(errorMessage(err)); } finally { setBusy(false); } }
  const activeCard = data?.scorecards.find(c => c.group_id === (department || companyId) && c.framework === framework);
  const controls = data?.scorecards.find(c => c.framework === 'ALL')?.controls || [];
  const inputClass = 'input w-full';
  return <div className="compliance-page space-y-6">
    <header><h1 className="text-2xl font-bold">Compliance · {companyName}</h1><p className="mt-2 text-slate-400">Evidenze e gap di controllo. Il punteggio non è una certificazione né una valutazione di applicabilità legale.</p></header>
    <details className="rounded-xl border border-slate-700 p-4">
      <summary className="cursor-pointer font-semibold">Prima volta? Come usare e interpretare Compliance</summary>
      <div className="mt-4 space-y-3 text-sm text-slate-300">
        <p>Seleziona l’azienda, l’eventuale reparto e il framework. Leggi insieme punteggio, copertura valutata, controlli non valutati ed esclusioni: il risultato riguarda solo i controlli inclusi nel catalogo e non certifica la conformità legale.</p>
        <p>Un controllo soddisfatto vale 1 punto, uno parziale 0,5; un gap o un controllo non valutato valgono 0. Le esclusioni motivate escono dal denominatore. Esempio: un soddisfatto, un parziale e un non valutato danno 50%, con 2 controlli valutati su 3. Non valutato significa che manca una valutazione, non che sia stata accertata una violazione.</p>
        <p>Esamina i finding e registra evidenze con fonte, versione e motivazione. Un gap tecnico aperto prevale su una dichiarazione positiva. Nessun finding o una scansione completata non provano la copertura completa. Il Compliance Score è distinto dal Posture Score tecnico.</p>
        <p>Per seguire una correzione, registra un asset stabile e associa esplicitamente il finding; riusa lo stesso asset se cambia IP. Assegna responsabile e SLA, registra mitigazione e correzione dichiarata, poi una verifica documentata che copra la specifica issue. La verifica è un’attestazione dell’operatore: non avvia test remoti. Una nuova osservazione positiva associata riapre l’issue conservando lo storico.</p>
        <p>Lo SLA decorre dalla prima osservazione positiva del ciclo corrente. L’MTTR medio considera solo i cicli conclusi con correzione verificata: non descrive le issue ancora aperte. Gli Auditor possono consultare ed esportare; SecOps e amministratori possono anche registrare modifiche.</p>
        <p>Esporta il PDF Audit da questa pagina o da Reports. Controlla dati mancanti, esclusioni, fonti e gap prima di condividerlo: è un dossier non firmato sullo stato corrente, non una certificazione. Il periodo dichiarato descrive la revisione e non filtra lo storico.</p>
        <p><strong>I riferimenti normativi non si aggiornano automaticamente.</strong> “Aggiorna evidenze” ricarica i dati usando le regole installate, la cui versione è indicata nella scorecard. Nuove norme e modifiche richiedono revisione delle fonti ufficiali, aggiornamento delle mappature e un rilascio del software; anche le evidenze precedenti devono essere riesaminate. ISO/IEC 27001 è uno standard distinto dai regolamenti UE.</p>
      </div>
    </details>
    <div className="flex flex-wrap gap-4"><label>Scope reparto<select aria-label="Scope reparto" className={inputClass} value={department} disabled={busy} onChange={e => setDepartment(e.target.value)}><option value="">Intera azienda</option>{departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}</select></label>
      <label>Framework<select className={inputClass} value={framework} onChange={e => setFramework(e.target.value)}>{['ALL', 'NIS2', 'DORA', 'ISO27001'].map(f => <option key={f}>{f}</option>)}</select></label><button className="btn-secondary self-end" disabled={busy || loading} onClick={() => setRevision(n => n + 1)}>Aggiorna evidenze</button></div>
    {error && <p role="alert" className="text-red-400">{error}</p>}
    {loading && <p role="status">Caricamento evidenze…</p>}
    {data && <>
      <p className="text-sm text-slate-400">Le attività si aggiornano ogni 30 secondi mentre la pagina è visibile; l’aggiornamento automatico è sospeso durante modifiche e pannelli aperti.</p>
      <ComplianceWorkflow key={`${companyId}/${department}`} data={data} framework={framework} canWrite={canWrite} busy={busy || loading} run={work => void mutate(work)} edit={setEditing} evidence={(id, group) => {
        const details = evidenceDetails.current;
        if (!details) return;
        details.open = true;
        const form = details.querySelector('form');
        if (!form) return;
        (form.elements.namedItem('control') as HTMLSelectElement).value = id;
        (form.elements.namedItem('group') as HTMLSelectElement).value = group;
        details.scrollIntoView?.({ block: 'nearest' });
        (form.elements.namedItem('status') as HTMLSelectElement).focus();
      }} />
      <section className="rounded-xl border border-slate-700 p-4" aria-label="Scorecard">
        <h2 className="font-semibold">Scorecard {framework}: {activeCard?.score == null ? 'Non applicabile' : `${activeCard.score.toFixed(1)}%`}</h2>
        <p>Copertura valutata {activeCard?.assessed ?? 0}/{activeCard?.applicable ?? 0} · Non valutati {activeCard?.unknown ?? 0} · Esclusi {activeCard?.excluded ?? 0}</p>
        <p className="text-sm text-slate-400">Non valutato = zero credito; parziale = mezzo credito. Le esclusioni richiedono evidenza. Cutoff {new Date(data.cutoff).toLocaleString()} · regole {data.rule_version}</p>
        <p>MTTR medio: {data.mean_mttr_hours == null ? 'Non valutato' : `${data.mean_mttr_hours.toFixed(2)} ore`} · cicli verificati {data.completed_cycles}</p>
        <p>Scansioni completate: {data.scans.filter(s => s.status === 'COMPLETED').length}/{Object.values(data.target_counts).reduce((sum, n) => sum + n, 0)} target configurati; {data.scans.length} tentativi più recenti. Nessuna scansione non equivale a copertura completa.</p>
        <div className="mt-4 overflow-x-auto"><table className="w-full text-left text-sm"><thead><tr><th className="p-2">Controllo</th><th className="p-2">Stato</th><th className="p-2">Evidenze</th></tr></thead><tbody>{activeCard?.controls.map(c => <tr key={c.id} className="border-t border-slate-800"><td className="p-2"><a href={c.source} target="_blank" rel="noreferrer" className="text-cyan-400">{c.id}</a><p>{c.title}</p></td><td className="p-2">{labels[c.status]}</td><td className="p-2">{c.finding_ids.length} finding · {c.issue_ids.length} issue{c.evidence_id && <p>Evidenza manuale registrata</p>}</td></tr>)}</tbody></table></div>
      </section>
      {!department && <section aria-label="Scorecard reparti"><h2 className="font-semibold">Reparti</h2>{data.scorecards.filter(c => c.framework === framework && c.group_id !== companyId).map(c => <p key={c.group_id}>{c.group_name}: {c.score == null ? 'N/A' : `${c.score.toFixed(1)}%`} · valutati {c.assessed}/{c.applicable} · non valutati {c.unknown}</p>)}</section>}
      {canWrite && <details ref={evidenceDetails} className="rounded-xl border border-slate-700 p-4"><summary className="cursor-pointer font-semibold">Registra evidenza di controllo</summary><form className="mt-3 grid gap-3" onSubmit={e => { e.preventDefault(); const f = new FormData(e.currentTarget); void mutate(() => complianceApi.evidence({ group_id: String(f.get('group')), control_id: String(f.get('control')), status: String(f.get('status')), reason: String(f.get('reason')), reference: String(f.get('reference')), source_version: String(f.get('version')) })); }}>
        <label>Gruppo dell’evidenza<select name="group" className={inputClass}>{data.groups.map(g => <option value={g.id} key={g.id}>{g.name}</option>)}</select></label>
        <label>Controllo<select name="control" className={inputClass}>{controls.map(c => <option value={c.id} key={c.id}>{c.id} · {c.title}</option>)}</select></label>
        <label>Valutazione<select name="status" className={inputClass}>{['unknown', 'satisfied', 'partial', 'not_satisfied', 'not_applicable'].map(s => <option key={s} value={s}>{labels[s]}</option>)}</select></label>
        <label>Motivazione<textarea name="reason" className={inputClass} required maxLength={4000} /></label><label>Riferimento documentale<input name="reference" className={inputClass} required maxLength={2000} /></label><label>Versione della fonte<input name="version" className={inputClass} required maxLength={100} /></label><button disabled={busy} className="btn-primary justify-self-start">Registra evidenza</button>
      </form></details>}
      <section aria-label="Registro evidenze"><h2 className="font-semibold">Registro evidenze manuali</h2>{data.evidence.length === 0 ? <p>Nessuna evidenza manuale.</p> : data.evidence.map(e => <details key={e.id} className="my-2 rounded border border-slate-700 p-3"><summary>{e.control_id} · {labels[e.status]} · {new Date(e.at).toLocaleString()}</summary><p>{e.reason}</p><p className="break-all">Fonte: {e.reference} · versione {e.source_version}</p><p>Autore {e.actor} · gruppo {data.groups.find(g => g.id === e.group_id)?.name}</p></details>)}</section>
      {canWrite && <details className="rounded-xl border border-slate-700 p-4"><summary className="cursor-pointer font-semibold">Identità asset e osservazioni</summary><p className="my-3 text-sm text-slate-400">Registra un’identità stabile e associa solo i finding di cui hai verificato l’identità. Riusa lo stesso asset dopo un cambio IP; nessuna correlazione automatica per indirizzo.</p>
        <form className="grid gap-3" onSubmit={e => { e.preventDefault(); const f = new FormData(e.currentTarget); void mutate(() => complianceApi.asset(String(f.get('group')), String(f.get('name')))); }}><label>Gruppo asset<select name="group" className={inputClass}>{data.groups.map(g => <option key={g.id} value={g.id}>{g.name}</option>)}</select></label><label>Nome stabile asset<input className={inputClass} name="name" required maxLength={200} /></label><button className="btn-secondary justify-self-start" disabled={busy}>Registra asset</button></form>
        <label className="mt-4 block">Asset verificato per l’associazione<select className={inputClass} value={asset} onChange={e => setAsset(e.target.value)}><option value="">Seleziona asset</option>{data.assets.map(a => <option value={a.id} key={a.id}>{a.name} · {data.groups.find(g => g.id === a.group_id)?.name}</option>)}</select></label>
        {data.findings.filter(f => !f.linked_issue_id).length === 0 && <p>Nessuna osservazione da associare.</p>}
        <ul>{data.findings.filter(f => !f.linked_issue_id).map(f => <li key={f.id} className="my-3 flex flex-wrap items-center justify-between gap-2 border-b border-slate-800 pb-3"><span>{f.host_ip}:{f.port}/{f.protocol} · {f.cve_id || 'Servizio'} · {f.confidence || 'confidenza ignota'}</span><button className="btn-secondary" disabled={busy || !asset || data.assets.find(a => a.id === asset)?.group_id !== f.group_id} onClick={() => void mutate(() => complianceApi.observe(asset, f.id))}>Associa osservazione {f.id.slice(0, 8)}</button></li>)}</ul>
      </details>}
      <section aria-label="Issue persistenti"><h2 className="font-semibold">Remediation, SLA e storico</h2>{data.issues.length === 0 && <p>Nessuna issue correlata: associa osservazioni a un’identità asset verificata per attivare il ciclo di remediation.</p>}
        {data.issues.map(issue => <article key={issue.id} className="my-3 rounded-xl border border-slate-700 p-4"><h3>{issue.asset_name} · {issue.key} · {issue.protocol}/{issue.port}</h3><p>{labels[issue.state]} · versione {issue.version} · {issue.breached ? 'SLA SCADUTO' : issue.deadline ? `Scadenza ${new Date(issue.deadline).toLocaleString()}` : 'SLA non configurato'}</p><p>Responsabile: {issue.owner === nil ? 'Non assegnato' : data.owners.find(o => o.id === issue.owner)?.name || 'Responsabile non più attivo'} · MTTR ciclo corrente: {issue.mttr_hours == null ? 'Non concluso' : `${issue.mttr_hours.toFixed(2)} ore`}</p><details><summary>Storico ({issue.events.length})</summary><IssueHistory events={issue.events} /></details>{canWrite && <button className="btn-secondary mt-3" disabled={busy} onClick={() => setEditing(issue)}>Aggiorna issue {issue.id.slice(0, 8)}</button>}</article>)}
        {editing && canWrite && <div ref={editor} tabIndex={-1}><IssueEditor key={editing.id} issue={editing} busy={busy} owners={data.owners} submit={command => void mutate(() => complianceApi.command(editing.id, command))} cancel={() => setEditing(undefined)} /></div>}
      </section>
      <AuditExport key={`${companyId}-${department}`} companyId={companyId} companyName={companyName} departmentId={department} />
      <details><summary>Metodo e limiti</summary><ul className="list-disc pl-5 text-sm text-slate-400">{data.limitations.map(l => <li key={l}>{l}</li>)}</ul></details>
    </>}
  </div>;
}

function IssueEditor({ issue, busy, owners, submit, cancel }: { issue: ComplianceIssue; busy: boolean; owners: { id: string; name: string }[]; submit: (c: IssueCommandInput) => void; cancel: () => void }) {
  const [action, setAction] = useState('configure');
  return <form className="grid gap-3 rounded-xl border border-cyan-800 p-4" aria-label="Aggiorna remediation" onSubmit={e => {
    e.preventDefault(); const f = new FormData(e.currentTarget);
    const command: IssueCommandInput = { action, expected_version: issue.version, reason: String(f.get('reason')) };
    if (action === 'configure') Object.assign(command, { owner: String(f.get('owner') || nil), policy_id: String(f.get('policy_id')), policy_version: String(f.get('policy_version')), sla_hours: Number(f.get('sla_hours')) });
    if (action === 'verify') command.verification = { provenance: { source: String(f.get('source')), version: String(f.get('version')), reference: String(f.get('reference')) }, complete_coverage: f.get('coverage') === 'on', not_present: f.get('not_present') === 'on' };
    submit(command);
  }}><h3>Aggiorna {issue.asset_name} · {issue.key}</h3><label>Azione<select className="input w-full" value={action} onChange={e => setAction(e.target.value)}><option value="configure">Responsabile e policy SLA</option>{issue.state === 'open' && <option value="mitigation_applied">Dichiara mitigazione applicata</option>}{['open', 'mitigation_applied'].includes(issue.state) && <option value="correction_declared">Dichiara correzione</option>}{issue.state === 'correction_declared' && <option value="verify">Registra verifica della correzione</option>}</select></label>
    {action === 'configure' && <><label>Responsabile<select className="input w-full" name="owner" defaultValue={issue.owner}><option value={nil}>Non assegnato</option>{owners.map(o => <option key={o.id} value={o.id}>{o.name}</option>)}{issue.owner !== nil && !owners.some(o => o.id === issue.owner) && <option value={issue.owner}>Responsabile non più attivo: scegli una nuova assegnazione</option>}</select></label><label>Ore SLA (0 rimuove la policy)<input name="sla_hours" className="input w-full" type="number" min={0} max={87600} defaultValue={issue.sla_hours || 0} required /></label><label>ID policy<input name="policy_id" className="input w-full" defaultValue={issue.policy_id} maxLength={100} /></label><label>Versione policy<input name="policy_version" className="input w-full" defaultValue={issue.policy_version} maxLength={100} /></label><p className="text-sm text-slate-400">La scadenza decorre dall’inizio del ciclo corrente; nessuna scadenza normativa è presunta.</p></>}
    {action === 'verify' && <><p>Attestazione dell’operatore: collega una prova che copra questa specifica issue. Non viene avviato alcun test remoto.</p><label>Fonte della verifica<input name="source" className="input w-full" required maxLength={200} /></label><label>Versione fonte<input name="version" className="input w-full" required maxLength={100} /></label><label>Riferimento prova<input name="reference" className="input w-full" required maxLength={2000} /></label><label><input type="checkbox" name="coverage" required /> Confermo copertura completa della specifica issue</label><label><input type="checkbox" name="not_present" required /> La prova verifica che il problema non è più presente</label></>}
    <label>Motivazione<textarea name="reason" className="input w-full" required maxLength={4000} /></label><div className="flex flex-wrap gap-3"><button disabled={busy} className="btn-primary">Registra aggiornamento</button><button type="button" className="btn-secondary" disabled={busy} onClick={cancel}>Annulla</button></div>
  </form>;
}
