import React from 'react';
import { ShieldCheck, Server, Layers, ShieldAlert } from 'lucide-react';
import { BrandColors } from '../types/tenant';

interface ReportPreviewProps {
  organizationName: string;
  logoUrl?: string | null;
  brandColors: BrandColors;
}

export const ReportPreview: React.FC<ReportPreviewProps> = ({
  organizationName,
  logoUrl,
  brandColors,
}) => {
  return (
    <div className="w-full max-w-md mx-auto aspect-[1/1.414] bg-white text-slate-900 rounded-xl shadow-2xl p-6 flex flex-col justify-between border border-slate-300 relative overflow-hidden select-none">
      {/* Top Banner & Header */}
      <div>
        <div className="flex items-center justify-between pb-4 border-b border-slate-200">
          <div className="flex items-center space-x-3">
            {logoUrl ? (
              <img
                src={logoUrl}
                alt="Logo Aziendale"
                className="h-10 max-w-[120px] object-contain rounded"
              />
            ) : (
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center text-white shadow-sm"
                style={{ backgroundColor: brandColors.primary }}
              >
                <ShieldCheck className="w-6 h-6 text-white" />
              </div>
            )}
            <div>
              <div
                className="text-sm font-black tracking-tight"
                style={{ color: brandColors.primary }}
              >
                {organizationName || 'Nome Organizzazione'}
              </div>
              <div className="text-[9px] text-slate-500 uppercase tracking-wider font-semibold">
                Security Assessment Service
              </div>
            </div>
          </div>

          <div className="text-right">
            <span
              className="px-2 py-0.5 rounded text-[8px] font-bold text-white uppercase tracking-wider"
              style={{ backgroundColor: brandColors.secondary }}
            >
              Riservato / TLP:AMBER
            </span>
            <div className="text-[8px] text-slate-400 mt-1 font-mono">
              {new Date().toLocaleDateString('it-IT')}
            </div>
          </div>
        </div>

        {/* Decorative Gradient Line */}
        <div
          className="h-1.5 w-full mt-3 rounded-full"
          style={{
            background: `linear-gradient(90deg, ${brandColors.primary} 0%, ${brandColors.secondary} 50%, ${brandColors.accent} 100%)`,
          }}
        />

        {/* Report Document Title */}
        <div className="mt-6 space-y-1">
          <div className="text-[10px] uppercase font-bold text-slate-400 tracking-wider">
            Report Ufficiale di Vulnerabilità
          </div>
          <h4
            className="text-base font-extrabold tracking-tight leading-tight"
            style={{ color: brandColors.primary }}
          >
            Vulnerability Assessment & Network Scope Audit
          </h4>
          <p className="text-[10px] text-slate-500 leading-snug">
            Sintesi tecnica dei servizi di rete rilevati, esposizione delle porte e correlazione CVE/CVSS.
          </p>
        </div>

        {/* Mockup Metric Strip */}
        <div className="grid grid-cols-3 gap-2 mt-5">
          <div className="p-2.5 rounded-lg bg-slate-50 border border-slate-200 text-center">
            <div className="flex justify-center text-slate-400 mb-0.5">
              <Server className="w-3.5 h-3.5" />
            </div>
            <div className="text-xs font-black text-slate-800 font-mono">12</div>
            <div className="text-[8px] text-slate-500 font-medium">Host Attivi</div>
          </div>

          <div className="p-2.5 rounded-lg bg-slate-50 border border-slate-200 text-center">
            <div className="flex justify-center text-slate-400 mb-0.5">
              <Layers className="w-3.5 h-3.5" />
            </div>
            <div className="text-xs font-black text-slate-800 font-mono">48</div>
            <div className="text-[8px] text-slate-500 font-medium">Porte Aperte</div>
          </div>

          <div className="p-2.5 rounded-lg bg-slate-50 border border-slate-200 text-center">
            <div className="flex justify-center text-red-500 mb-0.5">
              <ShieldAlert className="w-3.5 h-3.5" />
            </div>
            <div className="text-xs font-black text-red-600 font-mono">3 Critiche</div>
            <div className="text-[8px] text-slate-500 font-medium">CVE High/Crit</div>
          </div>
        </div>

        {/* Mockup Severity Distribution Bar */}
        <div className="mt-5 p-3 rounded-lg bg-slate-50 border border-slate-200 space-y-2">
          <div className="flex items-center justify-between text-[9px] font-bold text-slate-700">
            <span>Distribuzione Rischio Rilevato</span>
            <span style={{ color: brandColors.secondary }}>CVSS v3.1</span>
          </div>

          <div className="h-2 w-full rounded-full bg-slate-200 flex overflow-hidden">
            <div className="w-[20%] bg-red-600" title="Critical" />
            <div className="w-[30%] bg-orange-500" title="High" />
            <div className="w-[25%] bg-yellow-400" title="Medium" />
            <div className="w-[25%] bg-blue-500" title="Low" />
          </div>

          <div className="flex justify-between text-[7px] text-slate-500 font-semibold pt-0.5">
            <span className="text-red-600">● Critical (2)</span>
            <span className="text-orange-500">● High (3)</span>
            <span className="text-yellow-600">● Medium (2)</span>
            <span className="text-blue-600">● Low (2)</span>
          </div>
        </div>
      </div>

      {/* Footer Sheet */}
      <div className="pt-3 border-t border-slate-200 text-[8px] text-slate-400 flex items-center justify-between">
        <div className="truncate max-w-[200px]">
          Doc ID: <span className="font-mono text-slate-600">NXC-SEC-2026-08</span>
        </div>
        <div className="flex items-center space-x-1">
          <div
            className="w-1.5 h-1.5 rounded-full"
            style={{ backgroundColor: brandColors.accent }}
          />
          <span className="font-semibold" style={{ color: brandColors.primary }}>
            {organizationName || 'Tenant'} • White-Label
          </span>
        </div>
      </div>
    </div>
  );
};

