import { useState } from 'react';
import { complianceApi } from '../api/compliance';
import type { AssessmentAsset, ComplianceFinding, ComplianceIssue, ComplianceSnapshot } from '../types/compliance';

export function assetCandidates(data: ComplianceSnapshot, finding: ComplianceFinding): AssessmentAsset[] {
  const ids = new Set(data.issues.filter(issue => issue.group_id === finding.group_id && issue.events.some(event => event.observation?.source_host === finding.host_ip)).map(issue => issue.asset_id));
  return data.assets.filter(asset => asset.group_id === finding.group_id && ids.has(asset.id));
}

function Association({ data, finding, busy, run }: { data: ComplianceSnapshot; finding: ComplianceFinding; busy: boolean; run: (work: () => Promise<void>) => void }) {
  const candidates = assetCandidates(data, finding);
  const [assetId, setAssetId] = useState(candidates.length === 1 ? candidates[0].id : '');
  const [created, setCreated] = useState<AssessmentAsset>();
  const assets = data.assets.filter(a => a.group_id === finding.group_id);
  if (created && !assets.some(a => a.id === created.id)) assets.push(created);
  return <form aria-label={`Associa ${finding.id}`} className="mt-3 grid gap-3" onSubmit={event => {
    event.preventDefault(); const form = new FormData(event.currentTarget);
    run(async () => {
      let id = assetId;
      if (id === 'new') {
        const result = await complianceApi.asset(finding.group_id, String(form.get('asset_name')).trim());
        if (!result) throw new Error('Asset non restituito: aggiorna i dati prima di riprovare.');
        setCreated(result); setAssetId(result.id); id = result.id;
      }
      await complianceApi.observe(id, finding.id);
    });
  }}>
    <p>{candidates.length === 1 ? 'Proposta dallo storico dello stesso reparto; conferma che sia lo stesso dispositivo.' : candidates.length > 1 ? 'Identità ambigua: più asset nello storico hanno usato questo indirizzo.' : 'Identità da confermare: scegli un asset esistente oppure registrane uno.'} L’IP da solo non prova l’identità.</p>
    <label>Asset per {finding.host_ip}:{finding.port}<select required className="input w-full" value={assetId} disabled={busy} onChange={e => setAssetId(e.target.value)}><option value="">Scegli asset verificato</option>{assets.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}<option value="new">Registra nuovo asset</option></select></label>
    {assetId === 'new' && <label>Nome del nuovo asset<input name="asset_name" className="input w-full" required maxLength={200} defaultValue={`Host ${finding.host_ip}`} /></label>}
    {created && <p>Asset {created.name} registrato. Se l’associazione fallisce, riprova con questo asset.</p>}
    <label><input type="checkbox" required disabled={busy} /> Confermo l’identità del dispositivo e l’associazione a questa osservazione</label>
    <button className="btn-primary justify-self-start" disabled={busy}>Conferma associazione</button>
  </form>;
}

export function ComplianceWorkflow({ data, framework, canWrite, busy, run, edit, evidence }: { data: ComplianceSnapshot; framework: string; canWrite: boolean; busy: boolean; run: (work: () => Promise<void>) => void; edit: (issue: ComplianceIssue) => void; evidence: (id: string, group: string) => void }) {
  const [selected, setSelected] = useState<string[]>([]);
  const [results, setResults] = useState<string[]>([]);
  const card = data.scorecards.find(c => c.group_id === (data.scope.department_id || data.scope.company_id) && c.framework === framework);
  const findings = data.findings.filter(f => !f.linked_issue_id && (!f.port_state || ['open', 'open|filtered'].includes(f.port_state)));
  const issues = data.issues.filter(i => i.state !== 'corrected_verified').sort((a, b) => Number(b.breached) - Number(a.breached) || a.first_seen.localeCompare(b.first_seen));
  const controls = card?.controls.filter(c => ['unknown', 'partial', 'not_satisfied'].includes(c.status)) || [];
  return <section aria-label="Attività guidate" className="space-y-4 rounded-xl border border-cyan-800 p-4">
    <h2 className="text-xl font-semibold">Cosa devo fare adesso?</h2>
    <p>Le attività sono preparate dai dati registrati. Risolvi prima gli SLA scaduti, conferma le identità e completa le evidenze mancanti. Il framework filtra i controlli; le attività tecniche restano visibili per tutto lo scope.</p>
    <ol className="list-decimal pl-5"><li><a className="text-cyan-400" href="/scans">Apri scansioni per raccogliere dati o verificare una correzione</a>. Seleziona target e profilo e conferma il consenso nel flusso esistente; qui non viene avviata alcuna scansione.</li><li>Conferma le osservazioni da seguire: {findings.length}.</li><li>Gestisci le correzioni aperte: {issues.length}.</li><li>Completa i controlli da valutare o rivedere: {controls.length}. La scorecard e il PDF usano i dati salvati.</li></ol>
    {findings.map(f => <details key={f.id} className="rounded border border-slate-700 p-3"><summary className="cursor-pointer">Conferma dispositivo · {f.host_ip}:{f.port}/{f.protocol} · {f.cve_id || f.service_name || 'Servizio rilevato'}</summary><p>Reparto: {data.groups.find(g => g.id === f.group_id)?.name}. Controlli collegati: {card?.controls.filter(c => c.finding_ids.includes(f.id)).map(c => c.id).join(', ') || 'nessuna mappatura nel framework selezionato'}.</p>{canWrite ? <Association data={data} finding={f} busy={busy} run={run} /> : <p>Un operatore SecOps deve confermare l’identità.</p>}</details>)}
    {issues.map(issue => <article key={issue.id} className="rounded border border-slate-700 p-3"><h3>{issue.breached ? 'Priorità: SLA scaduto · ' : ''}{issue.asset_name} · {issue.key}</h3><p>{issue.state === 'correction_declared' ? 'Prossimo passo: raccogli una prova completa e registra la verifica.' : issue.state === 'mitigation_applied' ? 'Prossimo passo: completa e dichiara la correzione.' : 'Prossimo passo: assegna il responsabile e pianifica mitigazione o correzione.'}</p><p>Responsabile: {data.owners.find(o => o.id === issue.owner)?.name || 'Da assegnare'} · SLA: {issue.sla_hours ? `${issue.sla_hours} ore` : 'Da configurare'}</p>{canWrite && <div className="mt-2 flex flex-wrap gap-3"><label><input type="checkbox" checked={selected.includes(issue.id)} disabled={busy} onChange={e => setSelected(ids => e.target.checked ? [...ids, issue.id] : ids.filter(id => id !== issue.id))} /> Includi {issue.asset_name} · {issue.key} nella policy</label><button className="btn-secondary" disabled={busy} onClick={() => edit(issue)}>Gestisci correzione {issue.id.slice(0, 8)}</button></div>}</article>)}
    {canWrite && issues.length > 0 && <details><summary>Applica responsabile e SLA alle attività selezionate</summary><form aria-label="Policy attività" className="mt-3 grid gap-3" onSubmit={e => {
      e.preventDefault(); const f = new FormData(e.currentTarget); const chosen = issues.filter(i => selected.includes(i.id));
      run(async () => {
        const outcomes: string[] = []; let failed = false;
        for (const issue of chosen) {
          try { await complianceApi.command(issue.id, { action: 'configure', expected_version: issue.version, reason: String(f.get('reason')), owner: String(f.get('owner')), sla_hours: Number(f.get('hours')), policy_id: String(f.get('policy')), policy_version: String(f.get('version')) }); outcomes.push(`${issue.asset_name} · ${issue.key}: salvato`); }
          catch { failed = true; outcomes.push(`${issue.asset_name} · ${issue.key}: non salvato; aggiorna i dati e riprova`); }
        }
        setResults(outcomes); setSelected([]);
        if (failed) throw new Error('Policy applicata solo in parte: consulta gli esiti e aggiorna le evidenze prima di riprovare.');
      });
    }}><p>La policy sostituisce responsabile e SLA delle sole attività selezionate. La scadenza decorre dalla prima osservazione del ciclo, non da ora.</p><label>Responsabile delle attività<select name="owner" required className="input w-full" disabled={busy}><option value="">Scegli responsabile</option>{data.owners.map(o => <option key={o.id} value={o.id}>{o.name}</option>)}</select></label><label>Ore della policy<input name="hours" type="number" min={1} max={87600} required className="input w-full" disabled={busy} /></label><label>Nome policy<input name="policy" maxLength={100} required className="input w-full" disabled={busy} /></label><label>Versione della policy<input name="version" maxLength={100} required className="input w-full" disabled={busy} /></label><label>Motivo dell’assegnazione<textarea name="reason" required maxLength={4000} className="input w-full" disabled={busy} /></label><button className="btn-primary justify-self-start" disabled={busy || !selected.length}>Applica a {selected.length} attività</button></form></details>}
    {results.length > 0 && <ul aria-label="Esiti policy" role="status">{results.map((r, n) => <li key={n}>{r}</li>)}</ul>}
    {controls.map(control => <article key={control.id} className="rounded border border-slate-700 p-3"><h3>{control.id} · {control.title}</h3><p>{control.status === 'unknown' ? 'Manca una valutazione: raccogli documenti o prove pertinenti.' : 'Rivedi il gap e le prove disponibili; una dichiarazione positiva non nasconde problemi tecnici aperti.'}</p>{canWrite && <button className="btn-secondary mt-2" disabled={busy} onClick={() => evidence(control.id, card!.group_id)}>Valuta {control.id}</button>}</article>)}
    {!findings.length && !issues.length && !controls.length && <p>Nessuna attività in questo riepilogo. Controlla comunque copertura ed esclusioni: non è una certificazione.</p>}
  </section>;
}
