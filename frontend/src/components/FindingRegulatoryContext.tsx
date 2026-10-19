import type { FindingItem } from '../types/finding';

export function FindingRegulatoryContext({ finding }: { finding: FindingItem }) {
  const context = finding.regulatory_context;
  return <section aria-label="Implicazioni normative del risultato" className="space-y-3 rounded-xl border border-cyan-900/40 bg-cyan-950/20 p-4 text-sm break-words">
    <h4 className="font-semibold text-cyan-300">Implicazioni normative del risultato</h4>
    {!context ? <p>Contesto automatico non disponibile da questa versione del server. Ricarica i risultati dopo l’aggiornamento del backend.</p> : <>
      <p className="text-xs text-slate-400">Generato automaticamente dai dati della scansione · regole {context.rule_version} · confidenza {context.confidence || 'non disponibile'} · metodo {context.detection_method || 'non disponibile'}</p>
      {context.implications.length === 0 && <p>Nessuna implicazione specifica nelle regole automatiche installate per questo risultato. Questo non dimostra sicurezza o conformità.</p>}
      {context.implications.map(implication => <article key={implication.kind} className="space-y-2 border-t border-cyan-900/40 pt-3">
        <p><strong>Dato rilevato:</strong> {implication.observation}</p>
        <p><strong>Perché conta:</strong> {implication.meaning}</p>
        <p><strong>Cosa fare:</strong> {implication.action}</p>
        <ul className="space-y-1">{implication.references.map(reference => <li key={reference.id}><a href={reference.source} target="_blank" rel="noreferrer" className="text-cyan-300 underline focus-visible:outline focus-visible:outline-2">{reference.id} · {reference.title}</a></li>)}</ul>
      </article>)}
      <p className="text-xs text-slate-400">{context.limit}</p>
    </>}
  </section>;
}
