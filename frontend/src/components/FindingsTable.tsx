import React, { useState, useMemo } from 'react';
import {
  Search,
  Filter,
  Eye,
  ShieldCheck,
  Radio
} from 'lucide-react';
import { FindingItem } from '../types/finding';
import { FindingDetailModal } from './FindingDetailModal';

interface FindingsTableProps {
  findings: FindingItem[];
  loading?: boolean;
}

export const FindingsTable: React.FC<FindingsTableProps> = ({
  findings,
  loading = false,
}) => {
  const [search, setSearch] = useState<string>('');
  const [selectedSeverity, setSelectedSeverity] = useState<string>('ALL');
  const [selectedPort, setSelectedPort] = useState<string>('');
  const [minCvss, setMinCvss] = useState<string>('ALL');
  const [onlyOpenPorts, setOnlyOpenPorts] = useState<boolean>(false);

  const [inspectingFinding, setInspectingFinding] = useState<FindingItem | null>(null);

  // Filtraggio reattivo
  const filteredFindings = useMemo(() => {
    return findings.filter((f) => {
      // Bonifica: scarta record privi di host o non responsivi
      if (!f.host_ip || f.host_ip.trim() === '' || f.host_ip === '0.0.0.0') {
        return false;
      }

      // Toggle client-side: escludi host privi di porte aperte/servizi attivi
      if (onlyOpenPorts && (f.port <= 0 || !f.service_name)) {
        return false;
      }

      const q = search.toLowerCase().trim();
      const matchesSearch =
        q === '' ||
        f.host_ip.toLowerCase().includes(q) ||
        (f.hostname && f.hostname.toLowerCase().includes(q)) ||
        (f.cve_id && f.cve_id.toLowerCase().includes(q)) ||
        (f.service_name && f.service_name.toLowerCase().includes(q)) ||
        (f.raw_banner && f.raw_banner.toLowerCase().includes(q));

      const matchesSeverity =
        selectedSeverity === 'ALL' || f.severity.toUpperCase() === selectedSeverity;

      const matchesPort =
        selectedPort.trim() === '' || f.port.toString() === selectedPort.trim();

      const matchesCvss =
        minCvss === 'ALL' ||
        (minCvss === '9' && (f.cvss_score ?? 0) >= 9.0) ||
        (minCvss === '7' && (f.cvss_score ?? 0) >= 7.0) ||
        (minCvss === '4' && (f.cvss_score ?? 0) >= 4.0);

      return matchesSearch && matchesSeverity && matchesPort && matchesCvss;
    });
  }, [findings, search, selectedSeverity, selectedPort, minCvss, onlyOpenPorts]);

  const getSeverityBadge = (severity: string) => {
    switch (severity.toUpperCase()) {
      case 'CRITICAL':
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-bold bg-red-500/15 text-red-400 border border-red-500/40">
            CRITICAL
          </span>
        );
      case 'HIGH':
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-bold bg-orange-500/15 text-orange-400 border border-orange-500/40">
            HIGH
          </span>
        );
      case 'MEDIUM':
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-yellow-500/15 text-yellow-400 border border-yellow-500/40">
            MEDIUM
          </span>
        );
      case 'LOW':
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-medium bg-blue-500/15 text-blue-400 border border-blue-500/40">
            LOW
          </span>
        );
      default:
        return (
          <span className="px-2 py-0.5 rounded text-[11px] font-medium bg-slate-500/15 text-slate-400 border border-slate-500/40">
            INFO
          </span>
        );
    }
  };

  const getCvssPill = (score?: number) => {
    if (score === undefined || score === null) {
      return <span className="text-slate-500 font-mono text-[11px]">-</span>;
    }
    let color = 'bg-blue-500/10 text-blue-400 border-blue-500/30';
    if (score >= 9.0) color = 'bg-red-500/15 text-red-400 border-red-500/40 font-bold';
    else if (score >= 7.0) color = 'bg-orange-500/15 text-orange-400 border-orange-500/40 font-bold';
    else if (score >= 4.0) color = 'bg-yellow-500/15 text-yellow-400 border-yellow-500/40';

    return (
      <span className={`px-2 py-0.5 rounded border font-mono text-[11px] ${color}`}>
        {score.toFixed(1)}
      </span>
    );
  };

  return (
    <div className="space-y-4">
      {/* Toolbar Filtri */}
      <div className="p-4 rounded-xl border border-cyber-border bg-slate-900/60 flex flex-col md:flex-row md:items-center justify-between gap-3">
        {/* Ricerca Testuale */}
        <div className="relative flex-1 max-w-md">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-500">
            <Search className="w-4 h-4" />
          </div>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Filtra per IP, CVE ID o servizio..."
            className="w-full pl-9 pr-3 py-1.5 bg-slate-950 border border-cyber-border rounded-lg text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
          />
        </div>

        {/* Dropdown Filtri */}
        <div className="flex flex-wrap items-center gap-2.5 text-xs">
          {/* Severità */}
          <div className="flex items-center space-x-1.5">
            <Filter className="w-3.5 h-3.5 text-slate-400" />
            <select
              value={selectedSeverity}
              onChange={(e) => setSelectedSeverity(e.target.value)}
              className="bg-slate-950 border border-cyber-border rounded-lg px-2.5 py-1.5 text-xs text-slate-300 focus:outline-none focus:border-cyan-500"
            >
              <option value="ALL">Tutte le severità</option>
              <option value="CRITICAL">Critical</option>
              <option value="HIGH">High</option>
              <option value="MEDIUM">Medium</option>
              <option value="LOW">Low</option>
              <option value="INFO">Info / Clean</option>
            </select>
          </div>

          {/* Soglia CVSS */}
          <select
            value={minCvss}
            onChange={(e) => setMinCvss(e.target.value)}
            className="bg-slate-950 border border-cyber-border rounded-lg px-2.5 py-1.5 text-xs text-slate-300 focus:outline-none focus:border-cyan-500"
          >
            <option value="ALL">Tutti gli Score CVSS</option>
            <option value="9">CVSS &ge; 9.0 (Critico)</option>
            <option value="7">CVSS &ge; 7.0 (Alto)</option>
            <option value="4">CVSS &ge; 4.0 (Medio)</option>
          </select>

          {/* Filtro Porta */}
          <input
            type="text"
            value={selectedPort}
            onChange={(e) => setSelectedPort(e.target.value)}
            placeholder="Porta (es. 80)"
            className="w-24 px-2.5 py-1.5 bg-slate-950 border border-cyber-border rounded-lg text-xs text-slate-300 placeholder-slate-600 focus:outline-none focus:border-cyan-500 font-mono"
          />

          {/* Toggle Solo Host con Porte Aperte */}
          <button
            type="button"
            onClick={() => setOnlyOpenPorts(!onlyOpenPorts)}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium transition-colors ${
              onlyOpenPorts
                ? 'bg-cyan-500/15 border-cyan-500/40 text-cyan-300'
                : 'bg-slate-950 border-cyber-border text-slate-400 hover:text-slate-200'
            }`}
            title="Mostra esclusivamente host con porte aperte e servizi attivi"
          >
            <span
              className={`w-2 h-2 rounded-full transition-colors ${
                onlyOpenPorts
                  ? 'bg-cyan-400 shadow-[0_0_8px_rgba(6,182,212,0.8)]'
                  : 'bg-slate-600'
              }`}
            />
            <span>Solo Porte Aperte</span>
          </button>
        </div>
      </div>

      {/* Tabella Findings */}
      <div className="rounded-xl border border-cyber-border bg-cyber-card/40 overflow-hidden shadow-xl">
        {loading ? (
          <div className="py-16 flex flex-col items-center justify-center text-slate-400 space-y-2">
            <Radio className="w-6 h-6 animate-spin text-cyan-400" />
            <span className="text-xs">Caricamento vulnerabilità in corso...</span>
          </div>
        ) : filteredFindings.length === 0 ? (
          <div className="py-16 flex flex-col items-center justify-center text-center space-y-3 p-6">
            <div className="w-12 h-12 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-emerald-400">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div className="text-sm font-semibold text-slate-200">
              {findings.length === 0
                ? 'Nessun finding registrato per questa scansione'
                : 'Nessun finding corrisponde ai filtri selezionati'}
            </div>
            <p className="text-xs text-slate-400 max-w-sm">
              {findings.length === 0
                ? 'Nessun servizio vulnerabile è stato rilevato dai banner e dai test Nmap.'
                : 'Modifica o rimuovi i filtri impostati per visualizzare gli altri risultati.'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-cyber-border/80 text-[11px] font-semibold text-slate-400 uppercase tracking-wider bg-slate-900/40">
                  <th className="py-3 px-4">Severità</th>
                  <th className="py-3 px-4">CVSS</th>
                  <th className="py-3 px-4">CVE ID</th>
                  <th className="py-3 px-4">Host IP</th>
                  <th className="py-3 px-4">Porta / Prot</th>
                  <th className="py-3 px-4">Servizio & Versione</th>
                  <th className="py-3 px-4 text-right">Dettagli</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-cyber-border/40 text-xs">
                {filteredFindings.map((finding) => (
                  <tr
                    key={finding.id}
                    className="hover:bg-slate-800/25 transition-colors cursor-pointer"
                    onClick={() => setInspectingFinding(finding)}
                  >
                    <td className="py-3.5 px-4">{getSeverityBadge(finding.severity)}</td>
                    <td className="py-3.5 px-4">{getCvssPill(finding.cvss_score)}</td>
                    <td className="py-3.5 px-4 font-mono font-medium text-cyan-400">
                      {finding.cve_id || <span className="text-slate-500 font-sans italic">-</span>}
                    </td>
                    <td className="py-3.5 px-4 font-mono text-slate-200">
                      <div className="flex items-center space-x-2">
                        <span
                          className="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.8)]"
                          title="Host Operativo (UP)"
                        />
                        <span>{finding.host_ip}</span>
                      </div>
                    </td>
                    <td className="py-3.5 px-4 font-mono text-slate-300">
                      {finding.port}/{finding.protocol.toUpperCase()}
                    </td>
                    <td className="py-3.5 px-4 text-slate-300 max-w-xs truncate">
                      <span className="font-medium text-white">{finding.service_name || '-'}</span>{' '}
                      {finding.service_version && (
                        <span className="text-slate-400 font-mono text-[11px]">
                          v{finding.service_version}
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 px-4 text-right" onClick={(e) => e.stopPropagation()}>
                      <button
                        onClick={() => setInspectingFinding(finding)}
                        className="p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-cyan-400 transition-colors"
                        title="Esamina Scheda Tecnica"
                      >
                        <Eye className="w-3.5 h-3.5" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Drawer / Modale Dettaglio Finding */}
      <FindingDetailModal
        finding={inspectingFinding}
        onClose={() => setInspectingFinding(null)}
      />
    </div>
  );
};

