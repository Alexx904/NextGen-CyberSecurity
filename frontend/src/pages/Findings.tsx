import React, { useState, useEffect, useCallback } from 'react';
import { BarChart3, RefreshCw, AlertCircle, Terminal, ExternalLink } from 'lucide-react';
import { Link } from 'react-router-dom';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { ScanItem, ScanListResponseData } from '../types/scan';
import { FindingItem, FindingListResponseData } from '../types/finding';
import { FindingsTable } from '../components/FindingsTable';

export const Findings: React.FC = () => {
  const [scans, setScans] = useState<ScanItem[]>([]);
  const [selectedScanId, setSelectedScanId] = useState<string>('');
  const [findings, setFindings] = useState<FindingItem[]>([]);
  const [loadingScans, setLoadingScans] = useState<boolean>(true);
  const [loadingFindings, setLoadingFindings] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Carica elenco scansioni disponibili
  const fetchScans = useCallback(async () => {
    setLoadingScans(true);
    setError(null);
    try {
      const response = await apiClient.get<ApiResponse<ScanListResponseData>>('/scans?limit=50');
      if (response.data?.data?.items) {
        const items = response.data.data.items;
        setScans(items);
        if (items.length > 0) {
          setSelectedScanId(items[0].id);
        }
      }
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Impossibile recuperare l\'elenco delle scansioni');
    } finally {
      setLoadingScans(false);
    }
  }, []);

  // Carica finding per la scansione selezionata
  const fetchFindings = useCallback(async (scanId: string) => {
    if (!scanId) return;
    setLoadingFindings(true);
    try {
      const response = await apiClient.get<ApiResponse<FindingListResponseData>>(
        `/scans/${scanId}/findings?limit=100`
      );
      if (response.data?.data) {
        setFindings(response.data.data.items || []);
      }
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Impossibile recuperare i finding della scansione');
    } finally {
      setLoadingFindings(false);
    }
  }, []);

  useEffect(() => {
    fetchScans();
  }, [fetchScans]);

  useEffect(() => {
    if (selectedScanId) {
      fetchFindings(selectedScanId);
    }
  }, [selectedScanId, fetchFindings]);

  const selectedScan = scans.find((s) => s.id === selectedScanId);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center space-x-3">
          <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
            <BarChart3 className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-2xl font-bold tracking-tight text-white flex items-center space-x-2">
              <span>Vulnerabilità & Correlazione CVE</span>
            </h2>
            <p className="text-sm text-slate-400 mt-0.5">
              Catalogo vulnerabilità scoperte arricchite con metriche CVSS v3.1 ed evidenze banner
            </p>
          </div>
        </div>

        {selectedScan && (
          <Link
            to={`/scans/${selectedScan.id}`}
            className="self-start sm:self-auto flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-cyan-500/30 bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-400 text-xs font-semibold transition-colors"
          >
            <span>Dettaglio Scansione Completo</span>
            <ExternalLink className="w-3.5 h-3.5" />
          </Link>
        )}
      </div>

      {error && (
        <div className="p-4 rounded-xl border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center space-x-2">
          <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
          <span>{error}</span>
        </div>
      )}

      {/* Selettore Scansione Attiva */}
      <div className="p-4 rounded-xl border border-cyber-border bg-cyber-card/60 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center space-x-3 flex-1">
          <label className="text-xs font-medium text-slate-400 shrink-0">
            Seleziona Scansione di Riferimento:
          </label>
          {loadingScans ? (
            <div className="text-xs text-slate-400 flex items-center space-x-2">
              <RefreshCw className="w-3.5 h-3.5 animate-spin text-cyan-400" />
              <span>Caricamento scansioni...</span>
            </div>
          ) : scans.length === 0 ? (
            <span className="text-xs text-slate-500 italic">Nessuna scansione disponibile</span>
          ) : (
            <select
              value={selectedScanId}
              onChange={(e) => setSelectedScanId(e.target.value)}
              className="bg-slate-900 border border-cyber-border rounded-lg px-3 py-1.5 text-xs text-cyan-400 font-mono focus:outline-none focus:border-cyan-500 max-w-md w-full"
            >
              {scans.map((s) => (
                <option key={s.id} value={s.id}>
                  [{s.status}] ID: {s.id.slice(0, 8)}... - {new Date(s.created_at).toLocaleDateString('it-IT')}{' '}
                  ({s.summary_stats?.critical_count ?? 0} Critiche, {s.summary_stats?.high_count ?? 0} Alte)
                </option>
              ))}
            </select>
          )}
        </div>

        <button
          onClick={() => {
            fetchScans();
            if (selectedScanId) fetchFindings(selectedScanId);
          }}
          className="self-end sm:self-auto p-1.5 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
          title="Aggiorna dati"
        >
          <RefreshCw className={`w-4 h-4 ${loadingFindings ? 'animate-spin text-cyan-400' : ''}`} />
        </button>
      </div>

      {/* Tabella Findings */}
      {scans.length === 0 && !loadingScans ? (
        <div className="py-16 flex flex-col items-center justify-center text-center space-y-3 p-6 rounded-xl border border-cyber-border bg-cyber-card/30">
          <div className="w-12 h-12 rounded-xl bg-slate-900 border border-cyber-border flex items-center justify-center text-slate-500">
            <Terminal className="w-6 h-6" />
          </div>
          <div className="text-sm font-semibold text-slate-200">Nessuna scansione eseguita</div>
          <p className="text-xs text-slate-400 max-w-sm">
            Esegui prima una scansione Nmap per rilevare servizi e correlare vulnerabilità CVE.
          </p>
        </div>
      ) : (
        <FindingsTable findings={findings} loading={loadingFindings} />
      )}
    </div>
  );
};
