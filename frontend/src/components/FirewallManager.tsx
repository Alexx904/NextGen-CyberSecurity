import React, { useState, useEffect, useCallback } from 'react';
import {
  ShieldAlert,
  ShieldCheck,
  RefreshCw,
  Plus,
  Trash2,
  AlertTriangle,
  RotateCcw,
  CheckCircle2,
  X,
} from 'lucide-react';
import { sshApi } from '../api/ssh';
import {
  SSHServerItem,
  FirewallStatus,
  ChainName,
  RuleAction,
  AddFirewallRulePayload,
} from '../types/ssh';

interface FirewallManagerProps {
  server: SSHServerItem;
  onClose: () => void;
}

function firewallErrorMessage(error: unknown, fallback: string): string {
  const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string } | null;
  return err?.response?.data?.error?.message || err?.message || fallback;
}

export const FirewallManager: React.FC<FirewallManagerProps> = ({
  server,
  onClose,
}) => {
  const [status, setStatus] = useState<FirewallStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeChain, setActiveChain] = useState<ChainName>('INPUT');

  // Anti-Lockout countdown timer
  const [countdown, setCountdown] = useState<number>(0);
  const [actionLoading, setActionLoading] = useState(false);

  // Modal per inserimento nuova regola
  const [isAddRuleOpen, setIsAddRuleOpen] = useState(false);
  const [newAction, setNewAction] = useState<RuleAction>('ACCEPT');
  const [newProtocol, setNewProtocol] = useState('tcp');
  const [newPort, setNewPort] = useState('');
  const [newSource, setNewSource] = useState('0.0.0.0/0');
  const [newDestination, setNewDestination] = useState('');
  const [newComment, setNewComment] = useState('');
  const [newPosition, setNewPosition] = useState<number>(0); // 0 = append, >0 = insert at position
  const [newWithTimer, setNewWithTimer] = useState(true);
  const [newTimeoutSec, setNewTimeoutSec] = useState(60);

  const fetchStatus = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await sshApi.getFirewallStatus(server.id);
      setStatus(res);
      if (res.active_safety_timer && res.safety_timer_expires_at) {
        const left = Math.max(
          0,
          Math.floor((new Date(res.safety_timer_expires_at).getTime() - Date.now()) / 1000)
        );
        setCountdown(left > 0 ? left : res.safety_timeout_seconds || 60);
      } else {
        setCountdown(0);
      }
    } catch (err: unknown) {
      setError(firewallErrorMessage(err, 'Impossibile caricare le regole firewall dal server remoto'));
    } finally {
      setLoading(false);
    }
  }, [server.id]);

  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  // Gestione conto alla rovescia Safety Rollback Timer
  useEffect(() => {
    if (!status?.active_safety_timer || countdown <= 0) return;

    const interval = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          // Ricarica lo stato dopo lo scadere del timer
          fetchStatus();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [status?.active_safety_timer, countdown, fetchStatus]);

  // Conferma configurazione (disattiva Safety Timer)
  const handleConfirm = async () => {
    setActionLoading(true);
    try {
      await sshApi.confirmFirewall(server.id);
      await fetchStatus();
    } catch (err: unknown) {
      setError(firewallErrorMessage(err, 'Errore durante la conferma della configurazione firewall'));
    } finally {
      setActionLoading(false);
    }
  };

  // Rollback immediato
  const handleRollback = async () => {
    if (!window.confirm('Sei sicuro di voler annullare le modifiche e ripristinare le regole firewall precedenti?')) {
      return;
    }
    setActionLoading(true);
    try {
      await sshApi.rollbackFirewall(server.id);
      await fetchStatus();
    } catch (err: unknown) {
      setError(firewallErrorMessage(err, 'Errore durante il rollback del firewall'));
    } finally {
      setActionLoading(false);
    }
  };

  // Cancellazione regola
  const handleDeleteRule = async (lineNum: number) => {
    if (!window.confirm(`Eliminare la regola #${lineNum} dalla catena ${activeChain}?`)) {
      return;
    }
    setActionLoading(true);
    try {
      const updated = await sshApi.deleteFirewallRule(server.id, lineNum, activeChain, true, 60);
      setStatus(updated);
      if (updated.active_safety_timer) {
        setCountdown(updated.safety_timeout_seconds || 60);
      }
    } catch (err: unknown) {
      setError(firewallErrorMessage(err, 'Errore durante l\'eliminazione della regola firewall'));
    } finally {
      setActionLoading(false);
    }
  };

  // Inserimento nuova regola
  const handleAddRuleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading(true);

    const payload: AddFirewallRulePayload = {
      chain: activeChain,
      action: newAction,
      protocol: newProtocol,
      source: newSource.trim() || '0.0.0.0/0',
      destination: newDestination.trim(),
      port: newPort.trim(),
      comment: newComment.trim(),
      position: newPosition,
      with_safety_timer: newWithTimer,
      safety_timeout_seconds: newTimeoutSec,
    };

    try {
      const updated = await sshApi.addFirewallRule(server.id, payload);
      setStatus(updated);
      if (updated.active_safety_timer) {
        setCountdown(updated.safety_timeout_seconds || 60);
      }
      setIsAddRuleOpen(false);
      // Reset campi
      setNewPort('');
      setNewComment('');
    } catch (err: unknown) {
      setError(firewallErrorMessage(err, 'Errore durante l\'inserimento della regola firewall'));
    } finally {
      setActionLoading(false);
    }
  };

  const currentChainData = status?.chains ? status.chains[activeChain] : null;

  return (
    <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-md flex items-center justify-center p-3 lg:p-6 overflow-y-auto">
      <div className="bg-[#0a0d14] border border-cyber-border rounded-2xl max-w-5xl w-full p-5 lg:p-6 shadow-2xl flex flex-col max-h-[92vh] animate-in zoom-in-95 duration-150">
        {/* Header Modale */}
        <div className="flex items-center justify-between border-b border-cyber-border pb-4">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
              <ShieldAlert className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-base lg:text-lg font-bold text-white flex items-center gap-2">
                Firewall Management (iptables)
                <span className="text-xs font-mono font-normal px-2 py-0.5 rounded bg-slate-900 border border-cyber-border text-cyan-400">
                  {server.username}@{server.host}:{server.port}
                </span>
              </h2>
              <p className="text-xs text-slate-400">
                Amministrazione difensiva host Linux con Safety Rollback Timer anti-lockout
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={fetchStatus}
              disabled={loading || actionLoading}
              className="p-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
              title="Ricarica regole"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              type="button"
              onClick={onClose}
              className="p-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Anti-Lockout Safety Timer Banner */}
        {status?.active_safety_timer && countdown > 0 && (
          <div className="mt-4 p-4 rounded-xl border border-amber-500/50 bg-amber-950/25 flex flex-col sm:flex-row sm:items-center justify-between gap-3 animate-pulse shadow-lg">
            <div className="flex items-center space-x-3">
              <AlertTriangle className="w-6 h-6 text-amber-400 flex-shrink-0 animate-bounce" />
              <div>
                <div className="text-xs sm:text-sm font-bold text-amber-300 flex items-center gap-2">
                  <span>Safety Rollback Timer ATTIVO!</span>
                  <span className="px-2 py-0.5 rounded bg-amber-900/60 border border-amber-500/60 font-mono text-white text-xs">
                    {countdown}s rimanenti
                  </span>
                </div>
                <p className="text-[11px] text-amber-200/80 mt-0.5 font-sans">
                  Se non confermi la sessione, il backend ripristinerà automaticamente le regole precedenti per evitare lockout SSH.
                </p>
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <button
                type="button"
                onClick={handleRollback}
                disabled={actionLoading}
                className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-rose-800 bg-rose-950 hover:bg-rose-900 text-rose-200 font-semibold text-xs transition-colors"
              >
                <RotateCcw className="w-3.5 h-3.5" />
                <span>Annulla & Rollback</span>
              </button>
              <button
                type="button"
                onClick={handleConfirm}
                disabled={actionLoading}
                className="flex items-center space-x-1.5 px-4 py-1.5 rounded-lg bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold text-xs shadow-md transition-all"
              >
                <CheckCircle2 className="w-3.5 h-3.5" />
                <span>Conferma Configurazione</span>
              </button>
            </div>
          </div>
        )}

        {/* Notifica di errore */}
        {error && (
          <div className="mt-4 p-3 rounded-xl border border-rose-900/50 bg-rose-950/20 text-rose-300 text-xs flex items-center justify-between">
            <span>{error}</span>
            <button type="button" onClick={() => setError(null)} className="text-slate-400 hover:text-white">
              <X className="w-4 h-4" />
            </button>
          </div>
        )}

        {/* Barra di Navigazione Catene & Pulsante Nuova Regola */}
        <div className="mt-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-cyber-border pb-3">
          <div className="flex rounded-lg border border-cyber-border bg-[#111726] p-1 text-xs">
            {(['INPUT', 'OUTPUT', 'FORWARD'] as ChainName[]).map((chain) => {
              const chainData = status?.chains ? status.chains[chain] : null;
              const ruleCount = chainData?.rules ? chainData.rules.length : 0;
              const isSelected = activeChain === chain;
              return (
                <button
                  key={chain}
                  type="button"
                  onClick={() => setActiveChain(chain)}
                  className={`flex items-center space-x-2 px-3.5 py-1.5 rounded-md font-semibold transition-all ${
                    isSelected
                      ? 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/40 shadow'
                      : 'text-slate-400 hover:text-white hover:bg-slate-800/40'
                  }`}
                >
                  <span>{chain}</span>
                  <span className="px-1.5 py-0.2 rounded-full bg-slate-900 text-[10px] text-slate-300 font-mono">
                    {ruleCount}
                  </span>
                </button>
              );
            })}
          </div>

          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={handleRollback}
              disabled={actionLoading}
              className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-850 text-slate-300 text-xs font-medium transition-colors"
              title="Ripristina manuale dall'ultimo snapshot salvato"
            >
              <RotateCcw className="w-3.5 h-3.5" />
              <span>Rollback Manuale</span>
            </button>

            <button
              type="button"
              onClick={() => setIsAddRuleOpen(true)}
              className="flex items-center space-x-1.5 px-3.5 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-bold text-xs shadow transition-all"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Aggiungi Regola</span>
            </button>
          </div>
        </div>

        {/* Tabella Regole della Catena Attiva */}
        <div className="flex-1 overflow-y-auto mt-4 rounded-xl border border-cyber-border bg-[#0d121f]">
          <table className="w-full text-left text-xs">
            <thead className="bg-[#111726] border-b border-cyber-border text-slate-400 font-semibold uppercase tracking-wider text-[11px] sticky top-0 z-10">
              <tr>
                <th className="py-2.5 px-3 w-12 text-center">#</th>
                <th className="py-2.5 px-3">Azione</th>
                <th className="py-2.5 px-3">Protocollo</th>
                <th className="py-2.5 px-3">Sorgente</th>
                <th className="py-2.5 px-3">Destinazione</th>
                <th className="py-2.5 px-3">Porta / Match</th>
                <th className="py-2.5 px-3 text-right">Traffico</th>
                <th className="py-2.5 px-3 text-right w-16">Elimina</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-cyber-border/40 font-mono">
              {loading ? (
                <tr>
                  <td colSpan={8} className="py-12 text-center text-slate-500">
                    <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2 text-cyan-500" />
                    Caricamento regole catena {activeChain}...
                  </td>
                </tr>
              ) : !currentChainData || currentChainData.rules.length === 0 ? (
                <tr>
                  <td colSpan={8} className="py-12 text-center text-slate-500">
                    <ShieldCheck className="w-7 h-7 text-slate-600 mx-auto mb-2" />
                    <p className="text-slate-300 font-medium">Nessuna regola definita per la catena {activeChain}</p>
                    <p className="text-slate-500 text-[11px] font-sans mt-0.5">
                      Verrà applicata la policy di default: <strong className="text-cyan-400">{currentChainData?.default_policy || 'ACCEPT'}</strong>
                    </p>
                  </td>
                </tr>
              ) : (
                currentChainData.rules.map((rule) => {
                  const isAccept = rule.target === 'ACCEPT';
                  const isDrop = rule.target === 'DROP' || rule.target === 'REJECT';
                  return (
                    <tr key={rule.line_number} className="hover:bg-slate-900/60 transition-colors">
                      <td className="py-2.5 px-3 text-center text-slate-500 font-bold">
                        {rule.line_number}
                      </td>

                      <td className="py-2.5 px-3">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold ${
                            isAccept
                              ? 'bg-emerald-950 text-emerald-400 border border-emerald-800'
                              : isDrop
                              ? 'bg-rose-950 text-rose-400 border border-rose-800'
                              : 'bg-slate-900 text-slate-300 border border-slate-700'
                          }`}
                        >
                          {rule.target}
                        </span>
                      </td>

                      <td className="py-2.5 px-3 text-slate-300 uppercase">
                        {rule.protocol || 'all'}
                      </td>

                      <td className="py-2.5 px-3 text-cyan-300">
                        {rule.source || '0.0.0.0/0'}
                      </td>

                      <td className="py-2.5 px-3 text-slate-300">
                        {rule.destination || '0.0.0.0/0'}
                      </td>

                      <td className="py-2.5 px-3 text-slate-400 text-[11px]">
                        {rule.options || '-'}
                      </td>

                      <td className="py-2.5 px-3 text-right text-slate-400 text-[11px]">
                        <span title={`${rule.bytes} bytes`}>
                          {rule.packets} pkts
                        </span>
                      </td>

                      <td className="py-2.5 px-3 text-right">
                        <button
                          type="button"
                          onClick={() => handleDeleteRule(rule.line_number)}
                          disabled={actionLoading}
                          className="p-1 rounded text-slate-500 hover:text-rose-400 hover:bg-rose-950/40 transition-colors"
                          title={`Elimina regola #${rule.line_number}`}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Footer Informativo */}
        <div className="mt-4 pt-3 border-t border-cyber-border flex items-center justify-between text-xs text-slate-500 font-sans">
          <div className="flex items-center space-x-2">
            <span>Catena attiva:</span>
            <strong className="text-white font-mono">{activeChain}</strong>
            <span className="mx-1">•</span>
            <span>Policy default:</span>
            <strong className="text-cyan-400 font-mono">{currentChainData?.default_policy || 'ACCEPT'}</strong>
          </div>

          <div>
            <span>Totale pacchetti: </span>
            <strong className="text-slate-300 font-mono">{currentChainData?.packets || 0}</strong>
          </div>
        </div>

        {/* Modal Inserimento Nuova Regola */}
        {isAddRuleOpen && (
          <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4">
            <div className="bg-[#0d121f] border border-cyber-border rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4 animate-in zoom-in-95 duration-150">
              <div className="flex items-center justify-between border-b border-cyber-border pb-3">
                <div className="flex items-center space-x-2 text-white font-bold text-base">
                  <Plus className="w-5 h-5 text-cyan-400" />
                  <span>Nuova Regola Firewall ({activeChain})</span>
                </div>
                <button
                  type="button"
                  onClick={() => setIsAddRuleOpen(false)}
                  className="text-slate-400 hover:text-white p-1 rounded-lg"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleAddRuleSubmit} className="space-y-3.5 text-xs">
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-slate-400 mb-1 font-medium">Azione Target</label>
                    <select
                      value={newAction}
                      onChange={(e) => setNewAction(e.target.value as RuleAction)}
                      className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white focus:outline-none focus:border-cyan-500"
                    >
                      <option value="ACCEPT">ACCEPT (Consenti)</option>
                      <option value="DROP">DROP (Scarta silenziosamente)</option>
                      <option value="REJECT">REJECT (Rifiuta con ICMP)</option>
                      <option value="LOG">LOG (Registra nei log)</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-slate-400 mb-1 font-medium">Protocollo</label>
                    <select
                      value={newProtocol}
                      onChange={(e) => setNewProtocol(e.target.value)}
                      className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white focus:outline-none focus:border-cyan-500 uppercase font-mono"
                    >
                      <option value="tcp">TCP</option>
                      <option value="udp">UDP</option>
                      <option value="icmp">ICMP</option>
                      <option value="all">ALL (Tutti)</option>
                    </select>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-slate-400 mb-1 font-medium">IP o CIDR Sorgente</label>
                    <input
                      type="text"
                      value={newSource}
                      onChange={(e) => setNewSource(e.target.value)}
                      placeholder="0.0.0.0/0 oppure 192.168.1.0/24"
                      className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white font-mono focus:outline-none focus:border-cyan-500"
                    />
                  </div>

                  <div>
                    <label className="block text-slate-400 mb-1 font-medium">Porta (o Range)</label>
                    <input
                      type="text"
                      value={newPort}
                      onChange={(e) => setNewPort(e.target.value)}
                      disabled={newProtocol !== 'tcp' && newProtocol !== 'udp'}
                      placeholder="es. 22 oppure 80:443"
                      className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white font-mono focus:outline-none focus:border-cyan-500 disabled:opacity-50"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">IP o CIDR Destinazione (opzionale)</label>
                  <input
                    type="text"
                    value={newDestination}
                    onChange={(e) => setNewDestination(e.target.value)}
                    placeholder="0.0.0.0/0 oppure lasciare vuoto"
                    className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white font-mono focus:outline-none focus:border-cyan-500"
                  />
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">Posizione nella Catena</label>
                  <select
                    value={newPosition}
                    onChange={(e) => setNewPosition(parseInt(e.target.value, 10))}
                    className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white focus:outline-none focus:border-cyan-500 font-sans"
                  >
                    <option value={0}>In coda (Append alla fine)</option>
                    <option value={1}>In cima (#1 - Massima priorità)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">Commento Descrittivo</label>
                  <input
                    type="text"
                    value={newComment}
                    onChange={(e) => setNewComment(e.target.value)}
                    placeholder="es. Allow Web HTTPS from Office"
                    className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white focus:outline-none focus:border-cyan-500 font-sans"
                  />
                </div>

                <div className="p-3 rounded-lg border border-cyan-800/40 bg-cyan-950/20 space-y-2">
                  <label className="flex items-center space-x-2 cursor-pointer select-none">
                    <input
                      type="checkbox"
                      checked={newWithTimer}
                      onChange={(e) => setNewWithTimer(e.target.checked)}
                      className="rounded border-slate-700 bg-slate-950 text-cyan-500 focus:ring-cyan-500/20 w-4 h-4"
                    />
                    <span className="font-semibold text-white">Attiva Safety Rollback Timer (Anti-Lockout)</span>
                  </label>
                  <p className="text-[11px] text-cyan-300/80 leading-relaxed font-sans">
                    Crea uno snapshot automatico delle regole. Se non confermi entro il tempo prestabilito, la configurazione precedente verrà ripristinata automaticamente.
                  </p>
                  {newWithTimer && (
                    <div className="pt-2 flex items-center space-x-2">
                      <label className="text-slate-300 text-xs">Secondi prima del rollback:</label>
                      <input
                        type="number"
                        min={10}
                        max={300}
                        value={newTimeoutSec}
                        onChange={(e) => setNewTimeoutSec(parseInt(e.target.value, 10) || 60)}
                        className="w-20 px-2 py-1 rounded bg-slate-950 border border-cyan-700/50 text-white font-mono text-xs focus:outline-none focus:border-cyan-400"
                      />
                    </div>
                  )}
                </div>

                <div className="flex items-center justify-end space-x-3 pt-3 border-t border-cyber-border">
                  <button
                    type="button"
                    onClick={() => setIsAddRuleOpen(false)}
                    className="px-4 py-2 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium"
                  >
                    Annulla
                  </button>
                  <button
                    type="submit"
                    disabled={actionLoading}
                    className="flex items-center space-x-1.5 px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-bold text-xs shadow"
                  >
                    {actionLoading ? (
                      <>
                        <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                        <span>Applicazione...</span>
                      </>
                    ) : (
                      <>
                        <CheckCircle2 className="w-3.5 h-3.5" />
                        <span>Applica Regola</span>
                      </>
                    )}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
