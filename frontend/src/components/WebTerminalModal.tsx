import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import {
  Terminal as TerminalIcon,
  Maximize2,
  Minimize2,
  X,
  RotateCw,
  Trash2,
  CheckCircle2,
  AlertCircle,
  WifiOff,
} from 'lucide-react';
import { SSHServerItem } from '../types/ssh';
import { getStoredAccessToken, getStreamTicket } from '../api/client';
import { debugLogger } from '../utils/debugLogger';

interface WebTerminalModalProps {
  server: SSHServerItem;
  onClose: () => void;
  docked?: boolean;
}

type TerminalConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export const WebTerminalModal: React.FC<WebTerminalModalProps> = ({
  server,
  onClose,
  docked = false,
}) => {
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [status, setStatus] = useState<TerminalConnectionStatus>('connecting');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const terminalContainerRef = useRef<HTMLDivElement>(null);
  const termInstanceRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const resizeObserverRef = useRef<ResizeObserver | null>(null);
  const sessionGenerationRef = useRef(0);

  const initTerminalSession = useCallback(async () => {
    if (!terminalContainerRef.current) return;
    const generation = ++sessionGenerationRef.current;

    resizeObserverRef.current?.disconnect();
    resizeObserverRef.current = null;

    // Teardown precedente se presente
    if (wsRef.current) {
      debugLogger.logWSEvent({ direction: 'out', eventType: 'disconnect-requested', data: { state: 'reconnecting' } });
      wsRef.current.close();
      wsRef.current = null;
    }
    if (termInstanceRef.current) {
      termInstanceRef.current.dispose();
      termInstanceRef.current = null;
    }
    if (terminalContainerRef.current) {
      terminalContainerRef.current.innerHTML = '';
    }

    setStatus('connecting');
    setErrorMessage(null);

    // Inizializza istanza xterm.js
    const term = new Terminal({
      cursorBlink: true,
      cursorStyle: 'block',
      fontFamily: 'Consolas, "Fira Code", Monaco, "Courier New", monospace',
      fontSize: 13,
      lineHeight: 1.2,
      scrollback: 5000,
      theme: {
        background: '#0d1117',
        foreground: '#e6edf3',
        cursor: '#06b6d4',
        cursorAccent: '#0d1117',
        selectionBackground: 'rgba(6, 182, 212, 0.35)',
        black: '#0a0d14',
        red: '#f85149',
        green: '#2ea043',
        yellow: '#d29922',
        blue: '#58a6ff',
        magenta: '#bc8cff',
        cyan: '#39c5cf',
        white: '#b1bac4',
        brightBlack: '#6e7681',
        brightRed: '#ffa198',
        brightGreen: '#56d364',
        brightYellow: '#e3b341',
        brightBlue: '#79c0ff',
        brightMagenta: '#d2a8ff',
        brightCyan: '#56d4dd',
        brightWhite: '#ffffff',
      },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    term.open(terminalContainerRef.current);
    try {
      fitAddon.fit();
    } catch {
      // Ignora calcolo iniziale se container non ancora renderizzato a dimensione piena
    }

    termInstanceRef.current = term;
    fitAddonRef.current = fitAddon;

    term.write(`\x1b[36mConnecting to ${server.username}@${server.host}:${server.port} via NextGen SSH Gateway...\x1b[0m\r\n`);

    // Apertura WebSocket tramite StreamTicket monouso (NFR-SEC-04)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    let wsUrl = '';
    try {
      const ticket = await getStreamTicket();
      wsUrl = `${protocol}//${host}/ws/v1/ssh/servers/${server.id}/terminal?ticket=${encodeURIComponent(ticket)}`;
    } catch {
      if (sessionGenerationRef.current !== generation) return;
      const token = getStoredAccessToken() || '';
      wsUrl = `${protocol}//${host}/ws/v1/ssh/servers/${server.id}/terminal?token=${encodeURIComponent(token)}`;
    }

    if (sessionGenerationRef.current !== generation) return;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;
    const diagnosticUrl = wsUrl.split('?')[0];
    debugLogger.logWSEvent({ direction: 'out', eventType: 'connecting', url: diagnosticUrl, data: { state: 'connecting' } });

    ws.onopen = () => {
      if (wsRef.current !== ws) return;
      debugLogger.logWSEvent({ direction: 'in', eventType: 'open', url: diagnosticUrl, data: { state: 'connected' } });
      setStatus('connected');
      term.focus();
      // Notifica dimensioni iniziali della finestra PTY
      if (term.cols && term.rows) {
        ws.send(
          JSON.stringify({
            type: 'resize',
            cols: term.cols,
            rows: term.rows,
          })
        );
      }
    };

    ws.onmessage = (event: MessageEvent) => {
      if (wsRef.current !== ws) return;
      // Terminal text can contain secrets which cannot be reliably redacted.
      // Record transport metadata only, never terminal input/output content.
      debugLogger.logWSEvent({ direction: 'in', eventType: 'frame', url: diagnosticUrl, data: typeof event.data === 'string' ? { characters: event.data.length } : { bytes: event.data instanceof Blob ? event.data.size : 0 } });
      if (typeof event.data === 'string') {
        term.write(event.data);
      } else if (event.data instanceof Blob) {
        const reader = new FileReader();
        reader.onload = () => {
          if (reader.result && wsRef.current === ws) {
            term.write(reader.result as string);
          }
        };
        reader.readAsText(event.data);
      }
    };

    ws.onerror = () => {
      if (wsRef.current !== ws) return;
      debugLogger.log({ level: 'error', source: 'ws', title: 'WS connection error', url: diagnosticUrl });
      setStatus('error');
      setErrorMessage('Errore di connessione WebSocket al gateway SSH.');
    };

    ws.onclose = (event: CloseEvent) => {
      if (wsRef.current !== ws) return;
      debugLogger.logWSEvent({ direction: 'in', eventType: 'close', url: diagnosticUrl, data: { code: event.code, clean: event.wasClean } });
      setStatus('disconnected');
      term.write('\r\n\x1b[33m[Sessione SSH chiusa dal gateway o dall\'host remoto]\x1b[0m\r\n');
    };

    // Inoltro input tastiera da xterm a WebSocket
    term.onData((data: string) => {
      if (ws.readyState === WebSocket.OPEN) {
        debugLogger.logWSEvent({ direction: 'out', eventType: 'frame', url: diagnosticUrl, data: { characters: data.length } });
        ws.send(data);
      }
    });

    // Inoltro eventi resize da xterm a WebSocket
    term.onResize(({ cols, rows }) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(
          JSON.stringify({
            type: 'resize',
            cols,
            rows,
          })
        );
      }
    });

    // Osservatore ridimensionamento del contenitore
    if (terminalContainerRef.current) {
      const resizeObserver = new ResizeObserver(() => {
        try {
          if (fitAddonRef.current && termInstanceRef.current) {
            fitAddonRef.current.fit();
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
              wsRef.current.send(
                JSON.stringify({
                  type: 'resize',
                  cols: termInstanceRef.current.cols,
                  rows: termInstanceRef.current.rows,
                })
              );
            }
          }
        } catch {
          // Ignora micro-salti durante transizioni CSS
        }
      });
      resizeObserver.observe(terminalContainerRef.current);
      resizeObserverRef.current = resizeObserver;
    }
  }, [server.id, server.host, server.port, server.username]);

  useEffect(() => {
    void initTerminalSession();

    return () => {
      ++sessionGenerationRef.current;
      if (resizeObserverRef.current) {
        resizeObserverRef.current.disconnect();
        resizeObserverRef.current = null;
      }
      if (wsRef.current) {
        debugLogger.logWSEvent({ direction: 'out', eventType: 'disconnect-requested', data: { state: 'closed-by-user' } });
        wsRef.current.close();
        wsRef.current = null;
      }
      if (termInstanceRef.current) {
        termInstanceRef.current.dispose();
        termInstanceRef.current = null;
      }
    };
  }, [initTerminalSession]);

  // Gestione adattamento dimensioni su cambio schermo intero
  useEffect(() => {
    const timer = setTimeout(() => {
      try {
        if (fitAddonRef.current && termInstanceRef.current) {
          fitAddonRef.current.fit();
          if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            wsRef.current.send(
              JSON.stringify({
                type: 'resize',
                cols: termInstanceRef.current.cols,
                rows: termInstanceRef.current.rows,
              })
            );
          }
        }
      } catch {
        // Ignora
      }
    }, 150);
    return () => clearTimeout(timer);
  }, [isFullscreen]);

  const handleClear = () => {
    termInstanceRef.current?.clear();
    termInstanceRef.current?.focus();
  };

  const handleReconnect = () => {
    initTerminalSession();
  };

  return (
    <div role={docked && !isFullscreen ? 'region' : 'dialog'} aria-label={`Terminale SSH ${server.name}`} className={docked && !isFullscreen ? 'relative w-full h-[320px] min-h-[260px] flex mt-3' : 'fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4'}>
      <div
        className={`bg-slate-900 border border-slate-700/80 rounded-xl shadow-2xl flex flex-col overflow-hidden transition-all duration-200 ${
          isFullscreen
            ? 'w-full h-full rounded-none border-none'
            : docked ? 'w-full h-full' : 'w-full max-w-5xl h-[700px] max-h-[92vh]'
        }`}
      >
        {/* Header Terminale */}
        <div className="flex items-center justify-between px-4 py-3 bg-slate-950 border-b border-slate-800 select-none">
          <div className="flex items-center space-x-3">
            <div className="p-1.5 bg-cyan-950/60 border border-cyan-800/50 rounded-lg text-cyan-400">
              <TerminalIcon className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="font-semibold text-slate-100 text-sm">
                  {server.name}
                </span>
                <span className="font-mono text-xs text-slate-400 bg-slate-800 px-2 py-0.5 rounded border border-slate-700">
                  {server.username}@{server.host}:{server.port}
                </span>
              </div>
            </div>

            {/* Badge Stato Connessione */}
            <div className="ml-2 flex items-center">
              {status === 'connected' && (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-950/70 text-emerald-400 border border-emerald-800/60">
                  <CheckCircle2 className="w-3 h-3 mr-1" />
                  Connesso
                </span>
              )}
              {status === 'connecting' && (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-cyan-950/70 text-cyan-400 border border-cyan-800/60 animate-pulse">
                  <RotateCw className="w-3 h-3 mr-1 animate-spin" />
                  Connessione in corso...
                </span>
              )}
              {status === 'disconnected' && (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-amber-950/70 text-amber-400 border border-amber-800/60">
                  <WifiOff className="w-3 h-3 mr-1" />
                  Disconnesso
                </span>
              )}
              {status === 'error' && (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-rose-950/70 text-rose-400 border border-rose-800/60">
                  <AlertCircle className="w-3 h-3 mr-1" />
                  Errore
                </span>
              )}
            </div>
          </div>

          {/* Azioni Toolbar */}
          <div className="flex items-center space-x-2">
            <button
              onClick={handleReconnect}
              title="Riconnetti sessione"
              className="p-1.5 text-slate-400 hover:text-slate-100 hover:bg-slate-800 rounded-lg transition-colors"
            >
              <RotateCw className="w-4 h-4" />
            </button>
            <button
              onClick={handleClear}
              title="Pulisci schermo"
              className="p-1.5 text-slate-400 hover:text-slate-100 hover:bg-slate-800 rounded-lg transition-colors"
            >
              <Trash2 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setIsFullscreen(!isFullscreen)}
              title={isFullscreen ? 'Riduci finestra' : 'Schermo intero'}
              className="p-1.5 text-slate-400 hover:text-slate-100 hover:bg-slate-800 rounded-lg transition-colors"
            >
              {isFullscreen ? (
                <Minimize2 className="w-4 h-4" />
              ) : (
                <Maximize2 className="w-4 h-4" />
              )}
            </button>
            <div className="h-4 w-px bg-slate-800 mx-1" />
            <button
              onClick={onClose}
              title="Chiudi terminale"
              className="p-1.5 text-slate-400 hover:text-rose-400 hover:bg-rose-950/40 rounded-lg transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Notifica di errore se presente */}
        {errorMessage && (
          <div className="px-4 py-2 bg-rose-950/50 border-b border-rose-800/50 text-rose-300 text-xs flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />
              <span>{errorMessage}</span>
            </div>
            <button
              onClick={handleReconnect}
              className="px-2 py-0.5 bg-rose-900/50 hover:bg-rose-800 text-rose-200 rounded text-xs underline"
            >
              Riprova
            </button>
          </div>
        )}

        {/* Contenitore xterm.js */}
        <div className="flex-1 p-3 bg-[#0d1117] overflow-hidden relative">
          <div
            ref={terminalContainerRef}
            className="w-full h-full overflow-hidden"
            style={{ minHeight: '300px' }}
          />
        </div>

        {/* Footer Terminale */}
        <div className="px-4 py-1.5 bg-slate-950 border-t border-slate-800 flex items-center justify-between text-[11px] text-slate-500 font-mono">
          <div className="flex items-center space-x-4">
            <span>NextGen Web Terminal v1.0</span>
            <span>PTY: xterm-256color</span>
            {server.auth_type === 'KEY' ? (
              <span className="text-cyan-400">Auth: Chiave Ed25519/RSA</span>
            ) : (
              <span className="text-slate-400">Auth: Password cifrata AES-GCM</span>
            )}
          </div>
          <div>Premi Ctrl+C o digita &apos;exit&apos; per chiudere la shell</div>
        </div>
      </div>
    </div>
  );
};
