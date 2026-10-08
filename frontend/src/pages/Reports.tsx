import React from 'react';
import { FileText } from 'lucide-react';

export const Reports: React.FC = () => {
  return (
    <div className="space-y-6">
      <div className="flex items-center space-x-3">
        <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
          <FileText className="w-6 h-6" />
        </div>
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Reportistica White-Label</h2>
          <p className="text-sm text-slate-400">Generazione e download dei report audit PDF e CSV.</p>
        </div>
      </div>
      <div className="p-8 rounded-xl border border-cyber-border bg-cyber-card/30 text-center text-slate-400">
        Modulo Report pronto per l'esportazione documenti white-label.
      </div>
    </div>
  );
};
