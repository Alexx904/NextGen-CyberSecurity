import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { UserRole } from '../../types/auth';
import { ShieldCheck, RefreshCw } from 'lucide-react';

interface ProtectedRouteProps {
  allowedRoles?: UserRole[];
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ allowedRoles }) => {
  const { user, isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-cyber-dark text-slate-100">
        <div className="flex flex-col items-center space-y-4">
          <div className="w-14 h-14 rounded-2xl bg-cyber-card border border-cyber-border flex items-center justify-center text-cyber-accent shadow-[0_0_25px_rgba(6,182,212,0.2)]">
            <ShieldCheck className="w-8 h-8 animate-pulse text-cyan-400" />
          </div>
          <div className="flex items-center space-x-2 text-slate-400 text-sm">
            <RefreshCw className="w-4 h-4 animate-spin text-cyan-400" />
            <span>Verifica sessione di sicurezza...</span>
          </div>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (allowedRoles && user && !allowedRoles.includes(user.role)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-cyber-dark text-slate-100 p-6">
        <div className="max-w-md w-full p-6 rounded-xl border border-rose-800/40 bg-rose-950/20 text-center space-y-4">
          <h2 className="text-xl font-bold text-rose-400">Accesso Negato</h2>
          <p className="text-sm text-slate-300">
            Il tuo ruolo (<strong>{user.role}</strong>) non dispone dei privilegi necessari per accedere a questa risorsa.
          </p>
          <a
            href="/"
            className="inline-block px-4 py-2 rounded-lg bg-cyber-card border border-cyber-border text-sm text-slate-200 hover:text-white hover:border-cyan-500 transition-colors"
          >
            Torna alla Dashboard
          </a>
        </div>
      </div>
    );
  }

  return <Outlet />;
};
