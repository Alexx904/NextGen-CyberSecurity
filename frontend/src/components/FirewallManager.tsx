import React, { useState, useEffect, useCallback, useRef } from 'react';
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
  ArrowUp,
  ArrowDown,
  Pencil,
  Terminal,
} from 'lucide-react';
import { sshApi } from '../api/ssh';
import {
  SSHServerItem,
  FirewallStatus,
  ChainName,
  RuleAction,
  AddFirewallRulePayload,
  FirewallRule,
} from '../types/ssh';

const DockedTerminal = React.lazy(() => import('./WebTerminalModal').then(module => ({ default: module.WebTerminalModal })));

interface FirewallManagerProps {
  server: SSHServerItem;
  onClose: () => void;
}

function firewallErrorMessage(error: unknown, fallback: string): string {
  const err = error as { response?: { data?: { error?: { message?: string } } }; message?: string } | null;
  return err?.response?.data?.error?.message || err?.message || fallback;
}

function safetySecondsRemaining(expiry?: string | null): number {
  const remaining = Math.floor((Date.parse(expiry || '') - Date.now()) / 1000);
  return Number.isFinite(remaining) ? Math.max(0, remaining) : 0;
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
  const [newTimeoutSec, setNewTimeoutSec] = useState(60);
  const [newInInterface, setNewInInterface] = useState('');
  const [newOutInterface, setNewOutInterface] = useState('');
  const [editingLine, setEditingLine] = useState<number | null>(null);
  const [policyChoice, setPolicyChoice] = useState<'ACCEPT' | 'DROP'>('ACCEPT');
  const [terminalOpen, setTerminalOpen] = useState(false);
  const ruleOpener = useRef<HTMLButtonElement | null>(null);
  const canMutate = !loading && !actionLoading && !status?.active_safety_timer && Boolean(status?.revision);

  useEffect(() => {
    const policy = status?.chains[activeChain]?.default_policy;
    if (policy === 'ACCEPT' || policy === 'DROP') setPolicyChoice(policy);
  }, [status, activeChain]);

  const closeRuleForm = () => {
    setIsAddRuleOpen(false);
    queueMicrotask(() => ruleOpener.current?.focus());
  };

  const openRuleForm = (button: HTMLButtonElement, rule?: FirewallRule) => {
    ruleOpener.current = button;
    setEditingLine(rule?.line_number ?? null);
    setNewAction(rule?.target ?? 'ACCEPT'); setNewProtocol(rule?.protocol || 'tcp');
    setNewSource(rule?.source || ''); setNewDestination(rule?.destination || '');
    setNewInInterface(rule?.in_interface || ''); setNewOutInterface(rule?.out_interface || '');
    setNewPort(rule?.options.match(/\bdpts?:(\d+(?::\d+)?)/)?.[1] || '');
    setNewComment(rule?.options.match(/\/\* (.*?) \*\//)?.[1] || '');
    setNewPosition(0); setNewTimeoutSec(60); setIsAddRuleOpen(true);
  };

  const handlePolicy = async () => {
    if (!canMutate || !status?.revision) return;
    if (!window.confirm(`Impostare la policy ${policyChoice} sulla catena ${activeChain}?`)) return;
    setActionLoading(true);
    try { setStatus(await sshApi.setFirewallPolicy(server.id, activeChain, policyChoice, status.revision)); }
    catch (err) { setError(firewallErrorMessage(err, 'Cambio policy fallito')); }
    finally { setActionLoading(false); }
  };

  const handleMove = async (from: number, to: number) => {
    if (!canMutate || !status?.revision) return;
    setActionLoading(true);
    try { setStatus(await sshApi.moveFirewallRule(server.id, activeChain, from, to, status.revision)); }
    catch (err) { setError(firewallErrorMessage(err, 'Riordino regola fallito')); }
    finally { setActionLoading(false); }
  };

  const fetchStatus = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await sshApi.getFirewallStatus(server.id);
      setStatus(res);
      if (res.active_safety_timer && res.safety_timer_expires_at) {
        setCountdown(safetySecondsRemaining(res.safety_timer_expires_at));
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
    if (!status?.active_safety_timer || !status.safety_timer_expires_at) return;
    const expiry = status.safety_timer_expires_at;
    setCountdown(safetySecondsRemaining(expiry));
    const interval = setInterval(() => setCountdown(safetySecondsRemaining(expiry)), 1000);
    return () => clearInterval(interval);
  }, [status?.active_safety_timer, status?.safety_timer_expires_at]);

  useEffect(() => {
    if (!status?.active_safety_timer || countdown > 0) return;
    const poll = setTimeout(() => { void fetchStatus(); }, 2000);
    return () => clearTimeout(poll);
  }, [status, countdown, fetchStatus]);

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
    if (!canMutate || !status?.revision) return;
    if (!window.confirm(`Eliminare la regola #${lineNum} dalla catena ${activeChain}?`)) {
      return;
    }
    setActionLoading(true);
    try {
      const updated = await sshApi.deleteFirewallRule(server.id, lineNum, activeChain, true, 60, status.revision);
      setStatus(updated);
      if (updated.active_safety_timer) {
        setCountdown(safetySecondsRemaining(updated.safety_timer_expires_at));
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
    if (!canMutate || !status?.revision) return;
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
      with_safety_timer: true,
      safety_timeout_seconds: newTimeoutSec,
      expected_revision: status.revision,
      in_interface: newInInterface.trim(),
      out_interface: newOutInterface.trim(),
    };

    try {
      const updated = editingLine === null
        ? await sshApi.addFirewallRule(server.id, payload)
        : await sshApi.replaceFirewallRule(server.id, editingLine, payload);
      setStatus(updated);
      if (updated.active_safety_timer) {
        setCountdown(safetySecondsRemaining(updated.safety_timer_expires_at));
      }
      closeRuleForm();
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
        {status?.active_safety_timer && (
          <div className="mt-4 p-4 rounded-xl border border-amber-500/50 bg-amber-950/25 flex flex-col sm:flex-row sm:items-center justify-between gap-3 animate-pulse shadow-lg">
            <div className="flex items-center space-x-3">
              <AlertTriangle className="w-6 h-6 text-amber-400 flex-shrink-0 animate-bounce" />
              <div>
                <div className="text-xs sm:text-sm font-bold text-amber-300 flex items-center gap-2">
                  <span>{countdown > 0 ? 'Safety Rollback Timer ATTIVO!' : 'Scadenza raggiunta: ripristino da verificare.'}</span>
                  <span className="px-2 py-0.5 rounded bg-amber-900/60 border border-amber-500/60 font-mono text-white text-xs">
                    {countdown > 0 ? `${countdown}s rimanenti` : 'Verifica stato remoto'}
                  </span>
                </div>
                <p className="text-[11px] text-amber-200/80 mt-0.5 font-sans">
                  Il server remoto avvia il ripristino delle regole precedenti alla scadenza. Lo stato viene verificato anche dopo la scadenza.
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
                disabled={actionLoading || countdown <= 0}
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

        {status && !status.revision && !loading && (
          <p role="status" className="mt-3 text-xs text-amber-300">Revisione delle regole non disponibile. Aggiorna lo stato prima di modificare il firewall.</p>
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
              onClick={event => openRuleForm(event.currentTarget)}
              disabled={!canMutate}
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
                        <button type="button" aria-label={`Sposta regola ${rule.line_number} su`} disabled={!canMutate || rule.line_number <= 1} onClick={() => handleMove(rule.line_number, rule.line_number - 1)} className="p-1.5 text-slate-400 hover:text-cyan-300 disabled:opacity-30"><ArrowUp className="w-3.5 h-3.5" /></button>
                        <button type="button" aria-label={`Sposta regola ${rule.line_number} giù`} disabled={!canMutate || rule.line_number >= (currentChainData?.rules.length || 0)} onClick={() => handleMove(rule.line_number, rule.line_number + 1)} className="p-1.5 text-slate-400 hover:text-cyan-300 disabled:opacity-30"><ArrowDown className="w-3.5 h-3.5" /></button>
                        <button type="button" aria-label={`Sostituisci regola ${rule.line_number}`} disabled={!canMutate || !['ACCEPT', 'DROP', 'REJECT', 'LOG'].includes(rule.target) || !['all', 'tcp', 'udp', 'icmp'].includes(rule.protocol)} onClick={event => openRuleForm(event.currentTarget, rule)} className="p-1.5 text-slate-400 hover:text-cyan-300 disabled:opacity-30"><Pencil className="w-3.5 h-3.5" /></button>
                        <button
                          type="button"
                          onClick={() => handleDeleteRule(rule.line_number)}
                          disabled={!canMutate}
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
        <div className="mt-4 pt-3 border-t border-cyber-border flex flex-wrap gap-3 items-center justify-between text-xs text-slate-500 font-sans">
          <div className="flex flex-wrap items-center gap-2">
            <span>Catena attiva:</span>
            <strong className="text-white font-mono">{activeChain}</strong>
            <span className="mx-1">•</span>
            <span>Policy default:</span>
            <strong className="text-cyan-400 font-mono">{currentChainData?.default_policy || 'ACCEPT'}</strong>
            <select aria-label={`Nuova policy ${activeChain}`} value={policyChoice} onChange={event => setPolicyChoice(event.target.value as 'ACCEPT' | 'DROP')} disabled={!canMutate} className="bg-slate-950 border border-slate-700 rounded p-1 text-slate-200">
              <option value="ACCEPT">ACCEPT</option><option value="DROP">DROP</option>
            </select>
            <button type="button" onClick={handlePolicy} disabled={!canMutate || policyChoice === currentChainData?.default_policy} className="px-2 py-1 border border-slate-700 rounded text-cyan-300 disabled:opacity-30">Applica policy</button>
            <button type="button" onClick={() => setTerminalOpen(open => !open)} className="px-2 py-1 border border-slate-700 rounded text-cyan-300 inline-flex items-center gap-1"><Terminal className="w-3.5 h-3.5" />{terminalOpen ? 'Chiudi terminale' : 'Apri terminale'}</button>
          </div>

          <div>
            <span>Totale pacchetti: </span>
            <strong className="text-slate-300 font-mono">{currentChainData?.packets || 0}</strong>
          </div>
        </div>

        {/* Modal Inserimento Nuova Regola */}
        {terminalOpen && <React.Suspense fallback={<p role="status">Caricamento terminale...</p>}><DockedTerminal server={server} docked onClose={() => setTerminalOpen(false)} /></React.Suspense>}

        {isAddRuleOpen && (
          <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4">
            <div role="dialog" aria-modal="true" aria-label={editingLine === null ? 'Nuova regola firewall' : `Sostituisci regola ${editingLine}`} onKeyDown={event => {
              if (event.key === 'Escape') { event.preventDefault(); event.stopPropagation(); closeRuleForm(); }
              if (event.key === 'Tab') {
                const controls = Array.from(event.currentTarget.querySelectorAll<HTMLElement>('button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled])'));
                const first = controls[0], last = controls[controls.length - 1];
                if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus(); }
                if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
              }
            }} className="bg-[#0d121f] border border-cyber-border rounded-2xl max-w-lg w-full max-h-[90vh] overflow-y-auto p-6 shadow-2xl space-y-4 animate-in zoom-in-95 duration-150">
              <div className="flex items-center justify-between border-b border-cyber-border pb-3">
                <div className="flex items-center space-x-2 text-white font-bold text-base">
                  <Plus className="w-5 h-5 text-cyan-400" />
                  <span>{editingLine === null ? 'Nuova Regola Firewall' : `Sostituisci regola #${editingLine}`} ({activeChain})</span>
                </div>
                <button
                  type="button"
                  onClick={closeRuleForm}
                  className="text-slate-400 hover:text-white p-1 rounded-lg"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleAddRuleSubmit} className="space-y-3.5 text-xs">
                {editingLine !== null && <p role="note" className="p-2 text-amber-200 border border-amber-900 rounded">La regola viene sostituita con i campi di questo modulo. Eventuali opzioni avanzate mostrate nella tabella non vengono conservate; lo spostamento mantiene invece la regola completa.</p>}
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-slate-400 mb-1 font-medium">Azione Target</label>
                    <select
                      aria-label="Azione Target"
                      autoFocus
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
                      aria-label="Protocollo"
                      onChange={(e) => {
                        setNewProtocol(e.target.value);
                        if (!['tcp', 'udp'].includes(e.target.value)) setNewPort('');
                      }}
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
                      aria-label="IP o CIDR Sorgente"
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
                      aria-label="Porta o Range"
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
                    aria-label="IP o CIDR Destinazione"
                    onChange={(e) => setNewDestination(e.target.value)}
                    placeholder="0.0.0.0/0 oppure lasciare vuoto"
                    className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white font-mono focus:outline-none focus:border-cyan-500"
                  />
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">Posizione nella Catena</label>
                  <select
                    value={newPosition}
                    aria-label="Posizione nella Catena"
                    disabled={editingLine !== null}
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
                    aria-label="Commento Descrittivo"
                    onChange={(e) => setNewComment(e.target.value)}
                    placeholder="es. Allow Web HTTPS from Office"
                    className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-white focus:outline-none focus:border-cyan-500 font-sans"
                  />
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <label className="text-slate-400">Interfaccia ingresso<input aria-label="Interfaccia ingresso" value={newInInterface} maxLength={15} onChange={event => setNewInInterface(event.target.value)} className="block w-full p-2 mt-1 bg-slate-950 border border-slate-700 rounded text-white" /></label>
                  <label className="text-slate-400">Interfaccia uscita<input aria-label="Interfaccia uscita" value={newOutInterface} maxLength={15} onChange={event => setNewOutInterface(event.target.value)} className="block w-full p-2 mt-1 bg-slate-950 border border-slate-700 rounded text-white" /></label>
                </div>

                <div className="p-3 rounded-lg border border-cyan-800/40 bg-cyan-950/20 space-y-2">
                  <p className="font-semibold text-white">Safety Rollback Timer (Anti-Lockout) obbligatorio</p>
                  <p className="text-[11px] text-cyan-300/80 leading-relaxed font-sans">
                    Crea uno snapshot automatico delle regole. Se non confermi entro il tempo prestabilito, la configurazione precedente verrà ripristinata automaticamente.
                  </p>
                  {(
                    <div className="pt-2 flex items-center space-x-2">
                      <label className="text-slate-300 text-xs">Secondi prima del rollback:</label>
                      <input
                        type="number"
                        aria-label="Secondi prima del rollback"
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
                    onClick={closeRuleForm}
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
                        <span>{editingLine === null ? 'Applica Regola' : 'Sostituisci Regola'}</span>
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
