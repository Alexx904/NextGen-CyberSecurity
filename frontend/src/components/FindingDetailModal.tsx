import React, { useState } from 'react';
import {
  X,
  ShieldAlert,
  Server,
  Terminal,
  Wrench,
  ExternalLink,
  Copy,
  Check,
  Globe
} from 'lucide-react';
import { FindingItem } from '../types/finding';
import { VirtualPatchModal } from './VirtualPatchModal';

interface FindingDetailModalProps {
  finding: FindingItem | null;
  onClose: () => void;
}

export const FindingDetailModal: React.FC<FindingDetailModalProps> = ({
  finding,
  onClose,
}) => {
  const [copiedBanner, setCopiedBanner] = useState<boolean>(false);
  const [patchFindingId, setPatchFindingId] = useState<string | null>(null);

  if (!finding) return null;

  const handleCopyBanner = () => {
    if (finding.raw_banner) {
      navigator.clipboard.writeText(finding.raw_banner);
      setCopiedBanner(true);
      setTimeout(() => setCopiedBanner(false), 2000);
    }
  };

  const getSeverityBadge = (severity: string) => {
    switch (severity.toUpperCase()) {
      case 'CRITICAL':
        return (
          <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-red-500/15 text-red-400 border border-red-500/40">
            CRITICAL
          </span>
        );
      case 'HIGH':
        return (
          <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-orange-500/15 text-orange-400 border border-orange-500/40">
            HIGH
          </span>
        );
      case 'MEDIUM':
        return (
          <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-yellow-500/15 text-yellow-400 border border-yellow-500/40">
            MEDIUM
          </span>
        );
      case 'LOW':
        return (
          <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-blue-500/15 text-blue-400 border border-blue-500/40">
            LOW
          </span>
        );
      default:
        return (
          <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-slate-500/15 text-slate-400 border border-slate-500/40">
            INFO
          </span>
        );
    }
  };

  const getCvssColor = (score?: number) => {
    if (score === undefined || score === null) return 'text-slate-400 border-slate-600 bg-slate-800';
    if (score >= 9.0) return 'text-red-400 border-red-500/40 bg-red-950/40';
    if (score >= 7.0) return 'text-orange-400 border-orange-500/40 bg-orange-950/40';
    if (score >= 4.0) return 'text-yellow-400 border-yellow-500/40 bg-yellow-950/40';
    return 'text-blue-400 border-blue-500/40 bg-blue-950/40';
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-xs animate-in fade-in duration-200">
      {patchFindingId === finding.id && <VirtualPatchModal key={finding.id} findingId={finding.id} onClose={() => setPatchFindingId(null)} />}
      <div className="max-w-2xl w-full max-h-[90vh] flex flex-col bg-[#111726] border border-cyber-border rounded-2xl shadow-2xl overflow-hidden relative">
        {/* Header */}
        <div className="p-5 border-b border-cyber-border flex items-center justify-between bg-slate-900/60">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400">
              <ShieldAlert className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h3 className="text-base font-bold text-white font-mono">
                  {finding.cve_id || `${finding.service_name || 'Servizio'} (${finding.port}/${finding.protocol})`}
                </h3>
                {getSeverityBadge(finding.severity)}
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                Rilevamento Nmap & Correlazione Vulnerabilità CVSS v3.1
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Scrollable Body */}
        <div className="p-6 space-y-5 overflow-y-auto flex-1">
          {/* CVSS Score & Target Overview */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="p-4 rounded-xl border border-cyber-border bg-slate-900/80 flex flex-col justify-between">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                CVSS v3.1 Score
              </span>
              <div className="flex items-baseline space-x-2 mt-2">
                <span className={`text-2xl font-bold font-mono px-2 py-0.5 rounded-lg border ${getCvssColor(finding.cvss_score)}`}>
                  {finding.cvss_score !== undefined && finding.cvss_score !== null
                    ? finding.cvss_score.toFixed(1)
                    : 'N/A'}
                </span>
                <span className="text-xs text-slate-500">/ 10.0</span>
              </div>
            </div>

            <div className="p-4 rounded-xl border border-cyber-border bg-slate-900/80 flex flex-col justify-between">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                Host Bersaglio
              </span>
              <div className="mt-2 font-mono text-cyan-400 font-semibold text-sm">
                {finding.host_ip}
              </div>
              {finding.hostname && (
                <div className="text-[11px] text-slate-400 truncate mt-0.5">
                  {finding.hostname}
                </div>
              )}
            </div>

            <div className="p-4 rounded-xl border border-cyber-border bg-slate-900/80 flex flex-col justify-between">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                Porta & Servizio
              </span>
              <div className="mt-2 font-mono text-white text-sm">
                {finding.port}/{finding.protocol.toUpperCase()}
              </div>
              <div className="text-[11px] text-slate-400 truncate mt-0.5">
                {finding.service_name || 'unknown'}{' '}
                {finding.service_version ? `v${finding.service_version}` : ''}
              </div>
            </div>
          </div>

          {/* Servizio Rilevato */}
          <div className="p-4 rounded-xl border border-cyber-border bg-slate-900/40 space-y-2">
            <div className="flex items-center space-x-2 text-xs font-semibold text-slate-300">
              <Server className="w-4 h-4 text-cyan-400" />
              <span>Dettagli Servizio Applicativo</span>
            </div>
            <div className="grid grid-cols-2 gap-2 text-xs pt-1">
              <div>
                <span className="text-slate-500">Prodotto:</span>{' '}
                <span className="text-slate-200 font-medium">{finding.service_name || '-'}</span>
              </div>
              <div>
                <span className="text-slate-500">Versione:</span>{' '}
                <span className="text-slate-200 font-mono">{finding.service_version || '-'}</span>
              </div>
            </div>
          </div>

          {/* Banner Raw / Probing Evidence */}
          {finding.raw_banner && (
            <div className="space-y-2">
              <div className="flex items-center justify-between text-xs font-semibold text-slate-300">
                <div className="flex items-center space-x-2">
                  <Terminal className="w-4 h-4 text-amber-400" />
                  <span>Banner di Risposta Raw (Nmap Evidence)</span>
                </div>
                <button
                  onClick={handleCopyBanner}
                  className="flex items-center space-x-1 text-[11px] text-slate-400 hover:text-white transition-colors"
                >
                  {copiedBanner ? (
                    <>
                      <Check className="w-3 h-3 text-emerald-400" />
                      <span className="text-emerald-400">Copiato</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3 h-3" />
                      <span>Copia</span>
                    </>
                  )}
                </button>
              </div>
              <pre className="p-3.5 rounded-xl bg-slate-950 border border-cyber-border text-xs font-mono text-emerald-400 overflow-x-auto whitespace-pre-wrap max-h-36">
                {finding.raw_banner}
              </pre>
            </div>
          )}

          {/* Remediation & Mitigazione */}
          <div className="p-4 rounded-xl border border-cyan-900/40 bg-cyan-950/20 space-y-2">
            <div className="flex items-center space-x-2 text-xs font-semibold text-cyan-300">
              <Wrench className="w-4 h-4 text-cyan-400" />
              <span>Indicazioni di Remediation & Mitigazione</span>
            </div>
            <p className="text-xs text-slate-300 leading-relaxed">
              {finding.remediation ||
                `Aggiornare il servizio ${finding.service_name || 'rilevato'} all'ultima versione supportata rilasciata dal vendor. Verificare le regole di filtraggio firewall sulla porta ${finding.port} e disabilitare l'esposizione pubblica se il servizio è destinato solo ad uso interno.`}
            </p>
          </div>

          {/* External References (NIST NVD) */}
          {finding.cve_id && (
            <div className="flex flex-wrap items-center gap-3 pt-2">
              <a
                href={`https://nvd.nist.gov/vuln/detail/${finding.cve_id}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-cyan-400 text-xs font-medium transition-colors"
              >
                <Globe className="w-3.5 h-3.5" />
                <span>Scheda NIST NVD ({finding.cve_id})</span>
                <ExternalLink className="w-3 h-3" />
              </a>

              <a
                href={`https://cve.mitre.org/cgi-bin/cvename.cgi?name=${finding.cve_id}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-cyan-400 text-xs font-medium transition-colors"
              >
                <Globe className="w-3.5 h-3.5" />
                <span>MITRE CVE Dictionary</span>
                <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-cyber-border bg-slate-900/60 flex items-center justify-end">
          {finding.cve_id && <button type="button" onClick={() => setPatchFindingId(finding.id)} className="mr-3 rounded-lg border border-cyan-700 px-4 py-2 text-xs text-cyan-300">Genera virtual patch</button>}
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-white text-xs font-semibold transition-colors"
          >
            Chiudi
          </button>
        </div>
      </div>
    </div>
  );
};

