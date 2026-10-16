import React, { useEffect, useState } from 'react';
import { ScanStatus } from '../../types/scan';
import { formatDuration } from './scanConfiguration';

export interface ProgressEstimate {
  percentage: number;
  remainingSeconds?: number;
  receivedAt: number;
  phasePercentage?: number;
}

export function remainingTimeLabel(estimate: ProgressEstimate | null, now: number): string {
  if (!estimate || now - estimate.receivedAt > 30000 || estimate.remainingSeconds === undefined || !Number.isFinite(estimate.remainingSeconds)) return 'Tempo rimanente: in calcolo';
  const remaining = estimate.remainingSeconds - Math.max(0, (now - estimate.receivedAt) / 1000);
  if (remaining <= 0) return 'Tempo rimanente: stima in aggiornamento';
  return `Tempo rimanente indicativo: ${formatDuration(remaining * 0.7)} – ${formatDuration(remaining * 1.3)}`;
}

export const ScanTotalProgress: React.FC<{ status: ScanStatus; estimate: ProgressEstimate | null }> = ({ status, estimate }) => {
  const [now, setNow] = useState(Date.now);
  useEffect(() => {
    if (status !== 'RUNNING') return;
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, [status]);
  const terminal = ['COMPLETED', 'FAILED', 'CANCELLED'].includes(status);
  const percentage = status === 'COMPLETED' ? 100 : Math.max(0, Math.min(99, estimate?.percentage || 0));
  const known = terminal || estimate !== null;
  return <div className="space-y-2 rounded-lg border border-cyan-800/40 bg-slate-950 p-3">
    <div className="flex justify-between gap-2 text-sm text-cyan-300"><span>Avanzamento totale della scansione</span><span>{status === 'QUEUED' ? 'In coda' : known ? `${percentage.toFixed(0)}%${terminal ? '' : ' stimato'}` : 'In attesa di statistiche'}</span></div>
    <div role="progressbar" aria-label="Avanzamento totale della scansione" aria-valuemin={0} aria-valuemax={100} aria-valuenow={known ? percentage : undefined} className="h-2 overflow-hidden rounded bg-slate-800">
      <div className={`h-full bg-cyan-400 transition-all ${!known && status === 'RUNNING' ? 'animate-pulse' : ''}`} style={{ width: `${known ? percentage : status === 'RUNNING' ? 15 : 0}%` }} />
    </div>
    <p className="text-xs text-slate-300">{status === 'QUEUED' ? 'Tempo rimanente non disponibile finché il job è in coda.' : terminal ? status === 'COMPLETED' ? 'Scansione completata.' : 'Scansione interrotta: il lavoro non è stato completato.' : remainingTimeLabel(estimate, now)}</p>
    {!terminal && status !== 'QUEUED' && <p className="text-xs text-slate-500">Stima complessiva basata sulle fasi e sulle statistiche Nmap; può variare, soprattutto durante NSE e correlazione CVE.{estimate?.phasePercentage !== undefined ? ` Fase Nmap corrente: ${estimate.phasePercentage.toFixed(0)}%.` : ''}</p>}
  </div>;
};
