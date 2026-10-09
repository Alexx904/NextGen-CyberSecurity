import React from 'react';
import { CheckCircle2, Clock, AlertCircle, XCircle, Radio } from 'lucide-react';
import { ScanStatus } from '../types/scan';

interface ScanProgressBadgeProps {
  status: ScanStatus | string;
  phase?: string;
  percentage?: number;
  showIcon?: boolean;
  className?: string;
}

export const ScanProgressBadge: React.FC<ScanProgressBadgeProps> = ({
  status,
  phase,
  percentage,
  showIcon = true,
  className = '',
}) => {
  switch (status) {
    case 'COMPLETED':
      return (
        <span
          className={`inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 ${className}`}
        >
          {showIcon && <CheckCircle2 className="w-3 h-3 text-emerald-400" />}
          <span>Completata</span>
        </span>
      );

    case 'RUNNING': {
      let label = 'In Esecuzione';
      if (phase && percentage !== undefined) {
        label = `${phase} (${percentage.toFixed(0)}%)`;
      } else if (phase) {
        label = phase;
      } else if (percentage !== undefined) {
        label = `In Esecuzione (${percentage.toFixed(0)}%)`;
      }

      return (
        <span
          className={`inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-cyan-500/15 text-cyan-300 border border-cyan-500/40 shadow-[0_0_10px_rgba(6,182,212,0.2)] animate-pulse ${className}`}
        >
          {showIcon && <Radio className="w-3 h-3 text-cyan-400 animate-spin" />}
          <span>{label}</span>
        </span>
      );
    }

    case 'QUEUED':
      return (
        <span
          className={`inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/30 ${className}`}
        >
          {showIcon && <Clock className="w-3 h-3 text-amber-400" />}
          <span>In Coda</span>
        </span>
      );

    case 'FAILED':
      return (
        <span
          className={`inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-rose-500/10 text-rose-400 border border-rose-500/30 ${className}`}
        >
          {showIcon && <AlertCircle className="w-3 h-3 text-rose-400" />}
          <span>Fallita</span>
        </span>
      );

    case 'CANCELLED':
      return (
        <span
          className={`inline-flex items-center space-x-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-500/10 text-slate-400 border border-slate-500/30 ${className}`}
        >
          {showIcon && <XCircle className="w-3 h-3 text-slate-400" />}
          <span>Annullata</span>
        </span>
      );

    default:
      return (
        <span
          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-800 text-slate-300 ${className}`}
        >
          <span>{status}</span>
        </span>
      );
  }
};
