import React, { useEffect, useRef, useState } from 'react';
import { getVirtualPatches, type VirtualPatchBundle, type VirtualPatchFile } from '../api/virtualPatching';

interface Props { findingId: string; onClose: () => void }

export function VirtualPatchModal({ findingId, onClose }: Props) {
  const [bundle, setBundle] = useState<VirtualPatchBundle | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [engine, setEngine] = useState('fail2ban');
  const [copied, setCopied] = useState<string | null>(null);
  const [exportError, setExportError] = useState<string | null>(null);
  const dialog = useRef<HTMLElement>(null);

  useEffect(() => {
    let cancelled = false;
    const previous = document.activeElement as HTMLElement | null;
    dialog.current?.focus();
    setBundle(null); setError(null); setCopied(null); setExportError(null); setEngine('fail2ban');
    getVirtualPatches(findingId).then(data => { if (!cancelled) setBundle(data); })
      .catch(() => { if (!cancelled) setError('Impossibile caricare i template. Chiudi e riprova.'); });
    return () => { cancelled = true; previous?.focus(); };
  }, [findingId]);

  function keyDown(event: React.KeyboardEvent<HTMLElement>) {
    if (event.key === 'Escape') { event.stopPropagation(); onClose(); return; }
    if (event.key !== 'Tab') return;
    const controls = Array.from(dialog.current?.querySelectorAll<HTMLElement>('button,select,a[href]') || []).filter(el => !el.matches(':disabled'));
    const first = controls[0], last = controls[controls.length - 1];
    if (!first) return;
    if (event.shiftKey && (document.activeElement === first || document.activeElement === dialog.current)) { event.preventDefault(); last.focus(); }
    else if (!event.shiftKey && (document.activeElement === last || document.activeElement === dialog.current)) { event.preventDefault(); first.focus(); }
  }

  async function copy(file: VirtualPatchFile) {
    setCopied(null); setExportError(null);
    try { await navigator.clipboard.writeText(file.content); setCopied(file.filename); }
    catch { setExportError('Copia non riuscita. Usa il download del file.'); }
  }

  function download(file: VirtualPatchFile) {
    setExportError(null);
    let url: string | undefined;
    try {
      url = URL.createObjectURL(new Blob([file.content], { type: 'text/plain;charset=utf-8' }));
      const link = document.createElement('a');
      link.href = url; link.download = file.filename; link.click();
    } catch { setExportError('Download non riuscito. Riprova o copia il contenuto.'); }
    finally { if (url) URL.revokeObjectURL(url); }
  }

  return <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/80 p-3">
    <section ref={dialog} role="dialog" aria-modal="true" aria-labelledby="virtual-patch-title" tabIndex={-1} onKeyDown={keyDown} className="max-h-[92vh] w-full max-w-2xl space-y-4 overflow-y-auto rounded-xl border border-cyber-border bg-slate-900 p-5 text-slate-200">
      <header className="flex items-center justify-between gap-3"><h2 id="virtual-patch-title" className="text-lg font-semibold">Virtual patch · template</h2><button type="button" onClick={onClose} aria-label="Chiudi virtual patch">Chiudi</button></header>
      {error ? <p role="alert">{error}</p> : !bundle ? <p role="status">Caricamento template…</p> : <>
        <p>{bundle.cve_id || 'Finding senza CVE'} · catalogo {bundle.catalog_revision}</p>
        <p role="status">{bundle.reason}</p>
        {bundle.operational_impact && <p className="rounded border border-amber-700 p-3 text-sm text-amber-200">Impatto operativo · {bundle.product}: {bundle.operational_impact}</p>}
        <p className="text-xs text-slate-400">CVE nel catalogo: {bundle.covered_cves.join(', ')}</p>
        <ul className="list-disc space-y-2 pl-5 text-sm text-amber-200">{bundle.limitations.map(item => <li key={item}>{item}</li>)}</ul>
        {bundle.supported && bundle.files.length > 0 && <>
          <label className="block">Motore di destinazione<select value={engine} onChange={event => { setEngine(event.target.value); setCopied(null); setExportError(null); }} className="mt-2 block w-full rounded bg-slate-950 p-2"><option value="fail2ban">Fail2ban</option><option value="crowdsec">CrowdSec</option><option value="modsecurity">ModSecurity</option><option value="nginx">Nginx</option></select></label>
          {bundle.files.filter(file => file.engine === engine).map(file => <article key={file.filename} className="space-y-2">
            <h3 className="break-all font-mono text-sm">{file.filename}</h3>
            <pre className="max-h-64 overflow-auto whitespace-pre-wrap break-all rounded bg-slate-950 p-3 text-xs">{file.content}</pre>
            <div className="flex flex-wrap gap-3"><button type="button" onClick={() => void copy(file)} className="rounded border border-cyan-700 px-3 py-2">Copia {file.filename}</button><button type="button" onClick={() => download(file)} className="rounded border border-cyan-700 px-3 py-2">Scarica {file.filename}</button></div>
          </article>)}
        </>}
        {copied && <p role="status">Copiato: {copied}</p>}{exportError && <p role="alert">{exportError}</p>}
        {bundle.sources.length > 0 && <ul className="space-y-2 text-sm">{bundle.sources.map(source => <li key={source}><a href={source} target="_blank" rel="noreferrer" className="break-all text-cyan-300 underline">{source}</a></li>)}</ul>}
      </>}
    </section>
  </div>;
}
