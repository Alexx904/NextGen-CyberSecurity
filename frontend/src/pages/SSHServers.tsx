import React, { useState, useEffect, useCallback } from 'react';
import {
  Server,
  Plus,
  RefreshCw,
  Terminal,
  Shield,
  Edit2,
  Trash2,
  Key,
  Lock,
  CheckCircle2,
  XCircle,
  HelpCircle,
  Search,
  AlertTriangle,
  Cpu,
} from 'lucide-react';
import { sshApi } from '../api/ssh';
import {
  SSHServerItem,
  SSHConnectionTestResult,
} from '../types/ssh';
import { SSHServerModal } from '../components/SSHServerModal';
import { FirewallManager } from '../components/FirewallManager';
import { WebTerminalModal } from '../components/WebTerminalModal';

export const SSHServers: React.FC = () => {
  const [servers, setServers] = useState<SSHServerItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [clientFilter, setClientFilter] = useState('');

  // Test connessione live
  const [testingId, setTestingId] = useState<string | null>(null);
  const [testResults, setTestResults] = useState<Record<string, SSHConnectionTestResult>>({});

  // Modali
  const [isServerModalOpen, setIsServerModalOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<SSHServerItem | null>(null);

  const [firewallServer, setFirewallServer] = useState<SSHServerItem | null>(null);
  const [terminalServer, setTerminalServer] = useState<SSHServerItem | null>(null);

  // Eliminazione
  const [serverToDelete, setServerToDelete] = useState<SSHServerItem | null>(null);
  const [deleting, setDeleting] = useState(false);

  const loadServers = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await sshApi.listServers();
      setServers(data || []);
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      setError(e.response?.data?.error?.message || e.message || 'Errore durante il caricamento dei server SSH');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadServers();
  }, [loadServers]);

  const handleTestConnection = async (server: SSHServerItem) => {
    try {
      setTestingId(server.id);
      const result = await sshApi.testConnection(server.id);
      setTestResults((prev) => ({
        ...prev,
        [server.id]: result,
      }));
      // Aggiorna lo stato inline
      setServers((prev) =>
        prev.map((s) =>
          s.id === server.id
            ? {
                ...s,
                status: result.online ? 'ONLINE' : 'OFFLINE',
                last_connected_at: result.online
                  ? new Date().toISOString()
                  : s.last_connected_at,
              }
            : s
        )
      );
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      setTestResults((prev) => ({
        ...prev,
        [server.id]: {
          online: false,
          latency_ms: 0,
          os_info: '',
          is_root: false,
          error: e.response?.data?.error?.message || e.message || 'Connessione fallita',
        },
      }));
    } finally {
      setTestingId(null);
    }
  };

  const handleDeleteServer = async () => {
    if (!serverToDelete) return;
    try {
      setDeleting(true);
      await sshApi.deleteServer(serverToDelete.id);
      setServerToDelete(null);
      await loadServers();
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: { message?: string } } } };
      alert(e.response?.data?.error?.message || 'Errore durante la cancellazione del server');
    } finally {
      setDeleting(false);
    }
  };

  const filteredServers = servers.filter(s => !clientFilter || (s.client_name?.trim() || 'Senza cliente') === clientFilter).filter(
    (s) =>
      s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.host.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (s.client_name || '').toLowerCase().includes(searchQuery.toLowerCase())
  ).sort((a, b) => (a.client_name?.trim() || 'Senza cliente').localeCompare(b.client_name?.trim() || 'Senza cliente') || a.name.localeCompare(b.name));

  const totalServers = servers.length;
  const onlineServers = servers.filter((s) => s.status === 'ONLINE').length;
  const keyAuthServers = servers.filter((s) => s.auth_type === 'KEY').length;

  return (
    <div className="space-y-6">
      {/* Intestazione */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-100 flex items-center space-x-2">
            <Server className="w-7 h-7 text-cyan-400" />
            <span>Server SSH & Firewall iptables</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Gestisci l&apos;inventario host remoti con crittografia AES-256-GCM, shell Web Terminal e regole firewall con Anti-Lockout.
          </p>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={() => loadServers()}
            disabled={loading}
            className="flex items-center space-x-1.5 px-3 py-2 bg-slate-800/80 hover:bg-slate-700/80 text-slate-300 rounded-lg text-sm transition-colors border border-slate-700"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            <span>Aggiorna</span>
          </button>
          <button
            onClick={() => {
              setEditingServer(null);
              setIsServerModalOpen(true);
            }}
            className="flex items-center space-x-1.5 px-4 py-2 bg-cyan-600 hover:bg-cyan-500 text-white rounded-lg text-sm font-medium transition-colors shadow-lg shadow-cyan-950/40"
          >
            <Plus className="w-4 h-4" />
            <span>Aggiungi Server</span>
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center space-x-4">
          <div className="p-3 bg-cyan-950/50 border border-cyan-800/50 rounded-lg text-cyan-400">
            <Server className="w-6 h-6" />
          </div>
          <div>
            <div className="text-2xl font-bold text-slate-100">{totalServers}</div>
            <div className="text-xs text-slate-400">Server Registrati</div>
          </div>
        </div>

        <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center space-x-4">
          <div className="p-3 bg-emerald-950/50 border border-emerald-800/50 rounded-lg text-emerald-400">
            <CheckCircle2 className="w-6 h-6" />
          </div>
          <div>
            <div className="text-2xl font-bold text-emerald-400">{onlineServers}</div>
            <div className="text-xs text-slate-400">Host Online Raggiungibili</div>
          </div>
        </div>

        <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center space-x-4">
          <div className="p-3 bg-purple-950/50 border border-purple-800/50 rounded-lg text-purple-400">
            <Key className="w-6 h-6" />
          </div>
          <div>
            <div className="text-2xl font-bold text-purple-300">{keyAuthServers}</div>
            <div className="text-xs text-slate-400">Autenticazione a Chiave Privata</div>
          </div>
        </div>
      </div>

      {/* Barra di ricerca e filtro */}
      <div className="flex items-center bg-slate-900/60 border border-slate-800 rounded-lg px-3 py-2">
        <Search className="w-4 h-4 text-slate-400 mr-2 shrink-0" />
        <input
          type="text"
          placeholder="Cerca per nome, IP o username..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="bg-transparent border-none text-slate-200 placeholder-slate-500 text-sm focus:outline-hidden w-full"
        />
        {searchQuery && (
          <button
            onClick={() => setSearchQuery('')}
            className="text-xs text-slate-400 hover:text-slate-200 ml-2"
          >
            Azzera
          </button>
        )}
      </div>

      {/* Gestione errori generali */}
      {error && (
        <div className="bg-rose-950/30 border border-rose-800/50 rounded-xl p-4 text-rose-300 text-sm flex items-center space-x-3">
          <AlertTriangle className="w-5 h-5 text-rose-400 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      <label className="flex items-center gap-3 text-sm text-slate-300">Cliente SSH
        <select aria-label="Filtro cliente SSH" value={clientFilter} onChange={e => setClientFilter(e.target.value)} className="rounded-lg bg-slate-900 border border-slate-700 p-2">
          <option value="">Tutti i clienti SSH</option>
          {[...new Set(servers.map(s => s.client_name?.trim() || 'Senza cliente'))].sort().map(c => <option key={c} value={c}>{c}</option>)}
        </select>
        <span className="text-xs text-slate-500">Inventario autonomo, indipendente dall’azienda selezionata</span>
      </label>
      {/* Tabella Server */}
      <div className="bg-slate-900/60 border border-slate-800 rounded-xl overflow-hidden shadow-xl">
        {loading && servers.length === 0 ? (
          <div className="p-12 text-center text-slate-400 flex flex-col items-center justify-center">
            <RefreshCw className="w-8 h-8 animate-spin text-cyan-400 mb-3" />
            <p>Caricamento inventario server SSH in corso...</p>
          </div>
        ) : filteredServers.length === 0 ? (
          <div className="p-12 text-center text-slate-400 flex flex-col items-center justify-center">
            <Server className="w-12 h-12 text-slate-600 mb-3" />
            <p className="text-base text-slate-300 font-medium">Nessun server trovato</p>
            <p className="text-sm text-slate-500 mt-1">
              {searchQuery
                ? 'Nessun risultato corrisponde alla ricerca corrente.'
                : 'Aggiungi il tuo primo server remoto con "Aggiungi Server".'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300">
              <thead className="bg-slate-950/80 text-xs font-semibold text-slate-400 uppercase tracking-wider border-b border-slate-800">
                <tr>
                  <th className="px-5 py-3">Server / Host</th>
                  <th className="px-4 py-3">Utente & Auth</th>
                  <th className="px-4 py-3">Stato</th>
                  <th className="px-4 py-3">Test Live & Info</th>
                  <th className="px-5 py-3 text-right">Azioni</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {filteredServers.map((server, index) => {
                  const test = testResults[server.id];
                  const isTesting = testingId === server.id;

                  return (
                    <React.Fragment key={server.id}>
                    {(index === 0 || filteredServers[index - 1].client_name !== server.client_name) && <tr className="bg-cyan-950/20"><th colSpan={5} scope="rowgroup" className="px-5 py-3 text-cyan-300 text-sm">{server.client_name?.trim() || 'Senza cliente'}</th></tr>}
                    <tr
                      key={server.id}
                      className="hover:bg-slate-800/30 transition-colors"
                    >
                      {/* Nome e Host */}
                      <td className="px-5 py-4">
                        <div className="flex items-center space-x-3">
                          <div className="p-2 bg-slate-800 rounded-lg text-cyan-400 shrink-0">
                            <Server className="w-5 h-5" />
                          </div>
                          <div>
                            <div className="font-semibold text-slate-100 flex items-center space-x-2">
                              <span>{server.name}</span>
                            </div>
                            <div className="font-mono text-xs text-slate-400 mt-0.5">
                              {server.host}:{server.port}
                            </div>
                          </div>
                        </div>
                      </td>

                      {/* Utente & Auth */}
                      <td className="px-4 py-4">
                        <div className="space-y-1">
                          <div className="font-mono text-xs text-slate-200">
                            {server.username}
                          </div>
                          <div className="flex items-center space-x-1 text-xs">
                            {server.auth_type === 'KEY' ? (
                              <span className="inline-flex items-center text-purple-400 bg-purple-950/40 px-2 py-0.5 rounded border border-purple-800/40">
                                <Key className="w-3 h-3 mr-1" />
                                Chiave SSH
                              </span>
                            ) : (
                              <span className="inline-flex items-center text-amber-400 bg-amber-950/40 px-2 py-0.5 rounded border border-amber-800/40">
                                <Lock className="w-3 h-3 mr-1" />
                                Password AES
                              </span>
                            )}
                          </div>
                        </div>
                      </td>

                      {/* Stato */}
                      <td className="px-4 py-4 whitespace-nowrap">
                        {server.status === 'ONLINE' && (
                          <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-950/60 text-emerald-400 border border-emerald-800/50">
                            <CheckCircle2 className="w-3.5 h-3.5 mr-1.5" />
                            ONLINE
                          </span>
                        )}
                        {server.status === 'OFFLINE' && (
                          <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-rose-950/60 text-rose-400 border border-rose-800/50">
                            <XCircle className="w-3.5 h-3.5 mr-1.5" />
                            OFFLINE
                          </span>
                        )}
                        {server.status === 'UNCHECKED' && (
                          <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-slate-800 text-slate-400 border border-slate-700">
                            <HelpCircle className="w-3.5 h-3.5 mr-1.5" />
                            NON VERIFICATO
                          </span>
                        )}
                      </td>

                      {/* Test Live & Info */}
                      <td className="px-4 py-4">
                        <div className="flex flex-col space-y-1">
                          <button
                            onClick={() => handleTestConnection(server)}
                            disabled={isTesting}
                            className="inline-flex items-center space-x-1 text-xs text-cyan-400 hover:text-cyan-300 transition-colors w-fit"
                          >
                            <RefreshCw
                              className={`w-3.5 h-3.5 ${isTesting ? 'animate-spin' : ''}`}
                            />
                            <span>{isTesting ? 'Test in corso...' : 'Verifica connessione'}</span>
                          </button>

                          {test && (
                            <div className="text-xs space-y-0.5">
                              {test.online ? (
                                <div className="flex items-center space-x-2 text-slate-300">
                                  <span className="text-emerald-400 font-mono">
                                    {test.latency_ms} ms
                                  </span>
                                  {test.is_root && (
                                    <span className="px-1.5 py-0.2 bg-rose-950/60 text-rose-400 text-[10px] font-bold rounded border border-rose-800/50">
                                      ROOT
                                    </span>
                                  )}
                                  {test.os_info && (
                                    <span className="text-slate-400 truncate max-w-[180px] flex items-center" title={test.os_info}>
                                      <Cpu className="w-3 h-3 mr-1 text-slate-500" />
                                      {test.os_info}
                                    </span>
                                  )}
                                </div>
                              ) : (
                                <div className="text-rose-400 text-[11px] truncate max-w-[200px]" title={test.error}>
                                  {test.error || 'Host irraggiungibile'}
                                </div>
                              )}
                            </div>
                          )}
                        </div>
                      </td>

                      {/* Azioni */}
                      <td className="px-5 py-4 text-right whitespace-nowrap">
                        <div className="flex items-center justify-end space-x-2">
                          {/* Terminale Web */}
                          <button
                            onClick={() => setTerminalServer(server)}
                            title="Apri Terminale Web SSH"
                            className="p-1.5 bg-slate-800 hover:bg-cyan-900/50 text-cyan-400 hover:text-cyan-300 rounded-lg transition-colors border border-slate-700 hover:border-cyan-700/60"
                          >
                            <Terminal className="w-4 h-4" />
                          </button>

                          {/* Firewall iptables */}
                          <button
                            onClick={() => setFirewallServer(server)}
                            title="Gestisci Firewall iptables"
                            className="p-1.5 bg-slate-800 hover:bg-emerald-900/50 text-emerald-400 hover:text-emerald-300 rounded-lg transition-colors border border-slate-700 hover:border-emerald-700/60"
                          >
                            <Shield className="w-4 h-4" />
                          </button>

                          {/* Modifica */}
                          <button
                            onClick={() => {
                              setEditingServer(server);
                              setIsServerModalOpen(true);
                            }}
                            title="Modifica impostazioni server"
                            className="p-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-slate-100 rounded-lg transition-colors border border-slate-700"
                          >
                            <Edit2 className="w-4 h-4" />
                          </button>

                          {/* Elimina */}
                          <button
                            onClick={() => setServerToDelete(server)}
                            title="Elimina server"
                            className="p-1.5 bg-slate-800 hover:bg-rose-950/60 text-slate-400 hover:text-rose-400 rounded-lg transition-colors border border-slate-700 hover:border-rose-800/60"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Modal Aggiungi / Modifica Server */}
      <SSHServerModal
        isOpen={isServerModalOpen}
        server={editingServer}
        onClose={() => {
          setIsServerModalOpen(false);
          setEditingServer(null);
        }}
        onSaved={() => {
          loadServers();
        }}
      />

      {/* Modal Gestione Firewall iptables */}
      {firewallServer && (
        <FirewallManager
          server={firewallServer}
          onClose={() => setFirewallServer(null)}
        />
      )}

      {/* Modal Terminale Web SSH */}
      {terminalServer && (
        <WebTerminalModal
          server={terminalServer}
          onClose={() => setTerminalServer(null)}
        />
      )}

      {/* Modal Conferma Eliminazione */}
      {serverToDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xs p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-xl max-w-md w-full p-6 shadow-2xl">
            <div className="flex items-center space-x-3 text-rose-400 mb-4">
              <AlertTriangle className="w-6 h-6 shrink-0" />
              <h3 className="text-lg font-semibold text-slate-100">
                Elimina Server SSH
              </h3>
            </div>
            <p className="text-sm text-slate-300 mb-6">
              Sei sicuro di voler rimuovere il server{' '}
              <strong className="text-cyan-400">{serverToDelete.name}</strong> ({serverToDelete.host})?
              Tutti i segreti cifrati associati verranno cancellati dal database del tenant.
            </p>
            <div className="flex justify-end space-x-3">
              <button
                type="button"
                onClick={() => setServerToDelete(null)}
                disabled={deleting}
                className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-sm transition-colors"
              >
                Annulla
              </button>
              <button
                type="button"
                onClick={handleDeleteServer}
                disabled={deleting}
                className="px-4 py-2 bg-rose-600 hover:bg-rose-500 text-white rounded-lg text-sm font-medium transition-colors"
              >
                {deleting ? 'Eliminazione in corso...' : 'Conferma Eliminazione'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
