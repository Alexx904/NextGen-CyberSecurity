import React from 'react';
import { Terminal, AlertTriangle } from 'lucide-react';
import { useDebugConsole } from '../../context/DebugContext';

export const DebugFloatingToggle: React.FC = () => {
  const { toggleDrawer, unreadErrorsCount, isOpen } = useDebugConsole();

  // Non mostrare il pulsante fluttuante se il drawer è già aperto per evitare sovrapposizioni
  if (isOpen) {
    return null;
  }

  return (
    <button
      onClick={toggleDrawer}
      aria-label="Apri Console di Debug (Ctrl+Shift+D)"
      title="Debug & Testing Live Console (Ctrl+Shift+D)"
      className="fixed bottom-5 right-5 z-50 group flex items-center justify-center p-3 rounded-xl bg-slate-900/95 border border-cyan-500/40 text-cyan-400 hover:text-cyan-300 hover:border-cyan-400 shadow-[0_0_20px_rgba(6,182,212,0.25)] hover:shadow-[0_0_25px_rgba(6,182,212,0.45)] backdrop-blur-md transition-all duration-200 transform hover:scale-105 active:scale-95"
    >
      <div className="relative">
        <Terminal className="w-5 h-5 text-cyan-400 group-hover:text-cyan-300 transition-colors" />

        {/* Badge Contatore Errori Non Letti */}
        {unreadErrorsCount > 0 && (
          <span
            data-testid="unread-badge"
            className="absolute -top-3.5 -right-3.5 flex items-center justify-center min-w-[20px] h-5 px-1.5 text-[11px] font-bold text-white bg-red-600 border border-red-400/80 rounded-full shadow-lg shadow-red-500/40 animate-pulse"
          >
            <AlertTriangle className="w-2.5 h-2.5 mr-0.5 inline-block" />
            {unreadErrorsCount > 99 ? '99+' : unreadErrorsCount}
          </span>
        )}
      </div>

      <span className="sr-only">Toggle Debug Console</span>
    </button>
  );
};
