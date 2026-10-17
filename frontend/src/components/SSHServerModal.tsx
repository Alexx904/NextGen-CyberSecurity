import React, { useState, useEffect } from 'react';
import { X, Server, Key, Lock, CheckCircle2, AlertCircle, RefreshCw } from 'lucide-react';
import { sshApi } from '../api/ssh';
import { SSHServerItem, SSHAuthType } from '../types/ssh';

interface SSHServerModalProps {
  isOpen: boolean;
  server: SSHServerItem | null;
  onClose: () => void;
  onSaved: () => void;
}

export const SSHServerModal: React.FC<SSHServerModalProps> = ({
  isOpen,
  server,
  onClose,
  onSaved,
}) => {
  const isEditing = !!server;

  const [name, setName] = useState('');
  const [clientName, setClientName] = useState('');
  const [host, setHost] = useState('');
  const [port, setPort] = useState(22);
  const [username, setUsername] = useState('root');
  const [authType, setAuthType] = useState<SSHAuthType>('KEY');
  const [rawSecret, setRawSecret] = useState('');

  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (server) {
      setName(server.name);
      setClientName(server.client_name || '');
      setHost(server.host);
      setPort(server.port);
      setUsername(server.username);
      setAuthType(server.auth_type);
      setRawSecret('');
    } else {
      setName('');
      setClientName('');
      setHost('');
      setPort(22);
      setUsername('root');
      setAuthType('KEY');
      setRawSecret('');
    }
    setError(null);
  }, [server, isOpen]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('Il nome del server è obbligatorio');
      return;
    }
    if (!host.trim()) {
      setError('Indirizzo host o IP obbligatorio');
      return;
    }
    if (port <= 0 || port > 65535) {
      setError('Numero porta non valido (1-65535)');
      return;
    }
    if (!username.trim()) {
      setError('Username SSH obbligatorio');
      return;
    }
    if (!isEditing && !rawSecret.trim()) {
      setError('La credenziale SSH (password o chiave privata) è obbligatoria');
      return;
    }

    setSaving(true);
    try {
      if (isEditing && server) {
        await sshApi.updateServer(server.id, {
          client_name: clientName.trim(),
          name: name.trim(),
          host: host.trim(),
          port,
          username: username.trim(),
          auth_type: authType,
          raw_secret: rawSecret.trim() ? rawSecret.trim() : undefined,
        });
      } else {
        await sshApi.createServer({
          client_name: clientName.trim(),
          name: name.trim(),
          host: host.trim(),
          port,
          username: username.trim(),
          auth_type: authType,
          raw_secret: rawSecret.trim(),
        });
      }
      onSaved();
      onClose();
    } catch (err: any) {
      setError(err.message || 'Errore durante il salvataggio del server SSH');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-[#0d121f] border border-cyber-border rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-5 animate-in zoom-in-95 duration-150">
        <div className="flex items-center justify-between border-b border-cyber-border pb-3">
          <div className="flex items-center space-x-2.5 text-white font-bold text-base">
            <Server className="w-5 h-5 text-cyan-400" />
            <span>{isEditing ? 'Modifica Server SSH' : 'Registra Nuovo Server SSH'}</span>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-white p-1 rounded-lg"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <label className="block text-xs text-slate-300">Cliente (inventario SSH indipendente)
            <input aria-label="Cliente SSH" maxLength={160} value={clientName} onChange={e => setClientName(e.target.value)} placeholder="Es. Cliente Alfa" className="mt-2 w-full rounded-lg bg-slate-900 border border-slate-700 p-2 text-white" />
          </label>
          {error && (
            <div className="p-3 rounded-lg border border-rose-900/50 bg-rose-950/20 text-rose-300 text-xs flex items-center space-x-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">
              Nome Server / Etichetta
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="es. Web Server Produzione"
              required
              className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-xs text-white placeholder-slate-600 focus:outline-none focus:border-cyan-500"
            />
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <label className="block text-xs font-medium text-slate-400 mb-1">
                Indirizzo Host / IP
              </label>
              <input
                type="text"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                placeholder="es. 192.168.1.100 o srv.local"
                required
                className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-xs text-white placeholder-slate-600 font-mono focus:outline-none focus:border-cyan-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">
                Porta SSH
              </label>
              <input
                type="number"
                value={port}
                onChange={(e) => setPort(parseInt(e.target.value, 10) || 22)}
                min={1}
                max={65535}
                required
                className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-xs text-white font-mono focus:outline-none focus:border-cyan-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">
                Username SSH
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="es. root o ubuntu"
                required
                className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-xs text-white font-mono focus:outline-none focus:border-cyan-500"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">
                Tipo Autenticazione
              </label>
              <div className="flex rounded-lg border border-cyber-border bg-slate-950 p-0.5 text-xs">
                <button
                  type="button"
                  onClick={() => setAuthType('KEY')}
                  className={`flex-1 py-1.5 rounded-md flex items-center justify-center space-x-1.5 transition-colors ${
                    authType === 'KEY'
                      ? 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/40 font-semibold'
                      : 'text-slate-400 hover:text-white'
                  }`}
                >
                  <Key className="w-3.5 h-3.5" />
                  <span>Chiave SSH</span>
                </button>
                <button
                  type="button"
                  onClick={() => setAuthType('PASSWORD')}
                  className={`flex-1 py-1.5 rounded-md flex items-center justify-center space-x-1.5 transition-colors ${
                    authType === 'PASSWORD'
                      ? 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/40 font-semibold'
                      : 'text-slate-400 hover:text-white'
                  }`}
                >
                  <Lock className="w-3.5 h-3.5" />
                  <span>Password</span>
                </button>
              </div>
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">
              {authType === 'KEY' ? 'Chiave Privata RSA / Ed25519 (PEM)' : 'Password SSH'}
              {isEditing && (
                <span className="text-slate-500 font-normal ml-1">
                  (lascia vuoto per mantenere quella esistente)
                </span>
              )}
            </label>
            {authType === 'KEY' ? (
              <textarea
                value={rawSecret}
                onChange={(e) => setRawSecret(e.target.value)}
                placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                rows={5}
                className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-xs text-slate-200 placeholder-slate-700 font-mono focus:outline-none focus:border-cyan-500 resize-none"
              />
            ) : (
              <input
                type="password"
                value={rawSecret}
                onChange={(e) => setRawSecret(e.target.value)}
                placeholder={isEditing ? '••••••••••••••••' : 'Inserisci password SSH'}
                className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-cyber-border text-xs text-slate-200 placeholder-slate-600 focus:outline-none focus:border-cyan-500"
              />
            )}
            <p className="text-[11px] text-cyan-400/80 mt-1 flex items-center space-x-1">
              <span>🔒 Le credenziali vengono cifrate a riposo con algoritmo AES-256-GCM.</span>
            </p>
          </div>

          <div className="flex items-center justify-end space-x-3 pt-3 border-t border-cyber-border">
            <button
              type="button"
              onClick={onClose}
              disabled={saving}
              className="px-4 py-2 rounded-lg border border-slate-700 bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium"
            >
              Annulla
            </button>
            <button
              type="submit"
              disabled={saving}
              className="flex items-center space-x-1.5 px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-bold text-xs shadow-md transition-all"
            >
              {saving ? (
                <>
                  <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                  <span>Salvataggio...</span>
                </>
              ) : (
                <>
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  <span>{isEditing ? 'Salva Modifiche' : 'Registra Server'}</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

