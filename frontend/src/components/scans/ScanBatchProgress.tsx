import React, { useEffect, useState } from 'react';
import { apiClient } from '../../api/client';
import { ApiResponse } from '../../types/auth';
import { ScanItem, ScanListResponseData } from '../../types/scan';
import { formatDuration } from './scanConfiguration';

export const isTerminalScan = (scan: ScanItem) => ['COMPLETED', 'FAILED', 'CANCELLED'].includes(scan.status);

export function summarizeBatch(ids: string[], scans: ScanItem[], now: number) {
  const members = ids.map((id) => scans.find((scan) => scan.id === id));
  const known = members.filter((scan): scan is ScanItem => !!scan);
  const terminal = known.filter(isTerminalScan);
  const running = known.filter((scan) => scan.status === 'RUNNING');
  const queued = known.filter((scan) => scan.status === 'QUEUED');
  const percentage = ids.length ? members.reduce((total, scan) => total + (!scan ? 0 : isTerminalScan(scan) ? 100 : Math.min(99, Math.max(0, Number.isFinite(scan.progress?.percentage) ? scan.progress!.percentage : 0))), 0) / ids.length : 0;
  let remaining: number | undefined;
  const fresh = running.filter((scan) => {
    const timestamp = Date.parse(scan.progress?.observed_at || '');
    return Number.isFinite(timestamp) && now - timestamp <= 30000 && now >= timestamp && Number.isFinite(scan.progress?.remaining_seconds) && (scan.progress?.remaining_seconds || 0) > (now - timestamp) / 1000;
  });
  if (members.every(Boolean) && running.length > 0 && fresh.length === running.length) {
    const times = fresh.map((scan) => Math.max(0, scan.progress!.remaining_seconds! - (now - Date.parse(scan.progress!.observed_at!)) / 1000));
    const durations = fresh.map((scan, i) => {
      const start = Date.parse(scan.started_at || '');
      return times[i] + (Number.isFinite(start) ? Math.max(0, (now - start) / 1000) : times[i]);
    });
    const typical = durations.reduce((sum, duration) => sum + duration, 0) / durations.length;
    remaining = Math.max(...times, (times.reduce((sum, time) => sum + time, 0) + queued.length * typical) / running.length);
  }
  return { percentage, finished: terminal.length, completed: known.filter((scan) => scan.status === 'COMPLETED').length,
    failed: known.filter((scan) => scan.status === 'FAILED').length, cancelled: known.filter((scan) => scan.status === 'CANCELLED').length,
    queued: queued.length, running: running.length, missing: ids.length - known.length, remaining, done: ids.length > 0 && terminal.length === ids.length };
}

export const ScanBatchProgress: React.FC<{ ids: string[]; initialScans?: ScanItem[] }> = ({ ids, initialScans = [] }) => {
  const [scans, setScans] = useState(initialScans);
  const [error, setError] = useState(false);
  const [now, setNow] = useState(Date.now);
  const summary = summarizeBatch(ids, scans, now);
  useEffect(() => {
    let disposed = false;
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout>;
    let latest = initialScans;
    setScans(initialScans); setError(false);
    async function poll() {
      const pending = ids.filter((id) => !latest.some((scan) => scan.id === id && isTerminalScan(scan)));
      if (!pending.length) return;
      try {
        const updated: ScanItem[] = [];
        for (let i = 0; i < pending.length; i += 50) {
          const res = await apiClient.get<ApiResponse<ScanListResponseData>>(`/scans?ids=${pending.slice(i, i + 50).join(',')}`, { signal: controller.signal });
          if (!res.data.success || !res.data.data || res.data.data.items.length !== pending.slice(i, i + 50).length) throw new Error('Snapshot incompleto');
          updated.push(...res.data.data.items);
          if (disposed) return;
        }
        const monotonic = updated.map((scan) => {
          const previous = latest.find((item) => item.id === scan.id)?.progress;
          if (isTerminalScan(scan) || !previous) return scan;
          return { ...scan, progress: scan.progress ? { ...scan.progress, percentage: Math.max(previous.percentage, scan.progress.percentage) } : { ...previous, remaining_seconds: undefined } };
        });
        latest = [...latest.filter((scan) => !updated.some((item) => item.id === scan.id)), ...monotonic];
        if (!disposed) { setScans(latest); setError(false); setNow(Date.now()); }
      } catch { if (!disposed) setError(true); }
      if (!disposed && ids.some((id) => !latest.some((scan) => scan.id === id && isTerminalScan(scan)))) timer = setTimeout(poll, 5000);
    }
    void poll();
    return () => { disposed = true; clearTimeout(timer); controller.abort(); };
    // Batch membership is immutable for this mounted monitor.
  }, [ids]);
  useEffect(() => { if (summary.done) return; const timer = setInterval(() => setNow(Date.now()), 1000); return () => clearInterval(timer); }, [summary.done]);
  return <section className="space-y-3 rounded-xl border border-cyan-500/40 bg-cyan-950/20 p-4" aria-label="Progresso del gruppo di target">
    <div className="flex flex-wrap justify-between gap-2 text-sm text-cyan-300"><h3>Scansione totale del gruppo · ultimo lotto avviato</h3><strong>{summary.percentage.toFixed(0)}%{summary.done ? ' terminato' : ' stimato'}</strong></div>
    <div role="progressbar" aria-label="Avanzamento del gruppo" aria-valuemin={0} aria-valuemax={100} aria-valuenow={summary.percentage} className="h-2 overflow-hidden rounded bg-slate-800"><div className="h-full bg-cyan-400 transition-all" style={{ width: `${summary.percentage}%` }} /></div>
    <p className="text-xs text-slate-300">{summary.finished}/{ids.length} job terminati · {summary.running} in esecuzione · {summary.queued} in coda · {summary.completed} riusciti · {summary.failed} falliti · {summary.cancelled} annullati{summary.missing ? ` · ${summary.missing} in aggiornamento` : ''}</p>
    <p className="text-sm text-cyan-200">{summary.done ? 'Lotto terminato.' : !error && summary.remaining !== undefined ? `Tempo rimanente del gruppo indicativo: ${formatDuration(summary.remaining * 0.5)} – ${formatDuration(summary.remaining * 1.5)}` : 'Tempo rimanente del gruppo: in calcolo'}</p>
    {error && <p role="alert" className="text-xs text-amber-300">Aggiornamento non disponibile: ultimo progresso noto. Riprovo automaticamente.</p>}
    <p className="text-xs text-slate-500">Ogni job corrisponde a un target censito, che può contenere più IP. La stima considera fasi Nmap, coda e parallelismo osservato; rete, NSE e carico condiviso possono modificarla. Il 100% include anche job falliti o annullati.</p>
  </section>;
};
