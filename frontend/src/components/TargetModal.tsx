import React, { useState, useEffect } from 'react';
import { X, Crosshair, AlertCircle, CheckCircle2, ShieldAlert } from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetItem } from '../types/target';
import { validateTargetInput } from '../utils/targetValidator';

interface TargetModalProps {
  isOpen: boolean;
  onClose: () => void;
  onTargetCreated: () => void;
}

export const TargetModal: React.FC<TargetModalProps> = ({
  isOpen,
  onClose,
  onTargetCreated,
}) => {
  const [value, setValue] = useState('');
  const [description, setDescription] = useState('');
  const [isBlacklisted, setIsBlacklisted] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);

  // Reset form quando la modale viene aperta/chiusa
  useEffect(() => {
    if (isOpen) {
      setValue('');
      setDescription('');
      setIsBlacklisted(false);
      setServerError(null);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  // Validazione client-side in tempo reale
  const validation = value ? validateTargetInput(value) : null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setServerError(null);

    const clientValidation = validateTargetInput(value);
    if (!clientValidation.isValid) {
      setServerError(clientValidation.error || 'Formato del target non valido');
      return;
    }

    setSubmitting(true);
    try {
      const response = await apiClient.post<ApiResponse<TargetItem>>('/targets', {
        value: value.trim(),
        description: description.trim() || undefined,
        is_blacklisted: isBlacklisted,
      });

      if (response.data && response.data.success) {
        onTargetCreated();
        onClose();
      } else {
        setServerError(response.data?.error?.message || 'Errore durante la creazione del target');
      }
    } catch (err: any) {
      setServerError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile registrare il target'
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="max-w-md w-full bg-[#111726] border border-cyber-border rounded-2xl shadow-2xl overflow-hidden relative">
        {/* Header Modale */}
        <div className="p-5 border-b border-cyber-border flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Crosshair className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">Nuovo Target di Scansione</h3>
              <p className="text-xs text-slate-400">Aggiungi un IP, CIDR, Range o FQDN allo scope</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {/* Server Error Alert */}
          {serverError && (
            <div className="p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-start space-x-2">
              <AlertCircle className="w-4 h-4 shrink-0 text-rose-400 mt-0.5" />
              <span>{serverError}</span>
            </div>
          )}

          {/* Valore Target */}
          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="text-xs font-medium text-slate-300">
                Bersaglio / Indirizzo <span className="text-red-400">*</span>
              </label>
              {validation && (
                <span
                  className={`text-[11px] font-medium flex items-center space-x-1 ${
                    validation.isValid ? 'text-emerald-400' : 'text-rose-400'
                  }`}
                >
                  {validation.isValid ? (
                    <>
                      <CheckCircle2 className="w-3 h-3" />
                      <span>Formato: {validation.type}</span>
                    </>
                  ) : (
                    <>
                      <AlertCircle className="w-3 h-3" />
                      <span>Non valido</span>
                    </>
                  )}
                </span>
              )}
            </div>
            <input
              type="text"
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder="es: 192.168.1.50, 10.0.0.0/24, scanme.nmap.org"
              required
              className="w-full px-3.5 py-2.5 bg-slate-900 border border-cyber-border rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500 font-mono"
            />
            <p className="text-[11px] text-slate-500 mt-1">
              Supporta IPv4, IPv6, notazione CIDR (es. /24), Range IP (es. 10.0.0.1-50) o FQDN.
            </p>
          </div>

          {/* Descrizione / Note */}
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1.5">
              Descrizione o Etichetta Operativa (Opzionale)
            </label>
            <input
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="es: Server DMZ Web, Gateway Ufficio, Cloud Bastion"
              className="w-full px-3.5 py-2 bg-slate-900 border border-cyber-border rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
            />
          </div>

          {/* Flag Blacklist */}
          <div className="p-3.5 rounded-lg border border-cyber-border bg-slate-900/60 flex items-start space-x-3">
            <input
              type="checkbox"
              id="blacklist-toggle"
              checked={isBlacklisted}
              onChange={(e) => setIsBlacklisted(e.target.checked)}
              className="mt-0.5 rounded border-slate-700 bg-slate-900 text-cyan-500 focus:ring-cyan-500 focus:ring-offset-slate-900"
            />
            <label htmlFor="blacklist-toggle" className="text-xs text-slate-300 cursor-pointer select-none">
              <span className="font-semibold text-slate-200 flex items-center space-x-1.5">
                <span>Includi nella Blacklist di Esclusione</span>
                <ShieldAlert className="w-3.5 h-3.5 text-amber-400" />
              </span>
              <span className="block text-slate-400 text-[11px] mt-0.5">
                I target in blacklist sono bloccati permanentemente da qualsiasi scansione automatica o manuale.
              </span>
            </label>
          </div>

          {/* Bottoni Azione */}
          <div className="pt-2 flex items-center justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
            >
              Annulla
            </button>
            <button
              type="submit"
              disabled={submitting || (!!value && !validation?.isValid)}
              className="px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 text-xs font-semibold transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? 'Registrazione in corso...' : 'Salva Target'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

