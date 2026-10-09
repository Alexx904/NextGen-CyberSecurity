import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { DebugProvider } from './context/DebugContext';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { AppLayout } from './components/layout/AppLayout';

// Code-splitting via React.lazy per ottimizzazione bundle e caricamento su richiesta
const Login = lazy(() => import('./pages/Login').then(m => ({ default: m.Login })));
const Dashboard = lazy(() => import('./pages/Dashboard').then(m => ({ default: m.Dashboard })));
const Targets = lazy(() => import('./pages/Targets').then(m => ({ default: m.Targets })));
const Scans = lazy(() => import('./pages/Scans').then(m => ({ default: m.Scans })));
const ScanDetail = lazy(() => import('./pages/ScanDetail').then(m => ({ default: m.ScanDetail })));
const Findings = lazy(() => import('./pages/Findings').then(m => ({ default: m.Findings })));
const Reports = lazy(() => import('./pages/Reports').then(m => ({ default: m.Reports })));
const NetworkDiscovery = lazy(() => import('./pages/NetworkDiscovery').then(m => ({ default: m.NetworkDiscovery })));
const SSHServers = lazy(() => import('./pages/SSHServers').then(m => ({ default: m.SSHServers })));
const Settings = lazy(() => import('./pages/Settings').then(m => ({ default: m.Settings })));

function PageLoadingFallback() {
  return (
    <div className="flex h-64 w-full items-center justify-center">
      <div className="flex items-center gap-3 text-cyan-400 font-mono text-sm">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-cyan-400 border-t-transparent" />
        <span>Caricamento modulo in corso...</span>
      </div>
    </div>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <DebugProvider>
          <Suspense fallback={<PageLoadingFallback />}>
            <Routes>
              {/* Rotta Pubblica: Autenticazione */}
              <Route path="/login" element={<Login />} />

              {/* Rotte Protette: Accessibili solo con sessione valida */}
              <Route element={<ProtectedRoute />}>
                <Route element={<AppLayout />}>
                  <Route path="/" element={<Dashboard />} />
                  <Route path="/targets" element={<Targets />} />
                  <Route path="/network-discovery" element={<NetworkDiscovery />} />
                  <Route path="/scans" element={<Scans />} />
                  <Route path="/scans/:id" element={<ScanDetail />} />
                  <Route path="/ssh-servers" element={<SSHServers />} />
                  <Route path="/findings" element={<Findings />} />
                  <Route path="/reports" element={<Reports />} />
                  <Route path="/settings" element={<Settings />} />
                </Route>
              </Route>

              {/* Fallback per rotte sconosciute */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </DebugProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}
