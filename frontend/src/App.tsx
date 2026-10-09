import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { DebugProvider } from './context/DebugContext';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { AppLayout } from './components/layout/AppLayout';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Targets } from './pages/Targets';
import { Scans } from './pages/Scans';
import { ScanDetail } from './pages/ScanDetail';
import { Findings } from './pages/Findings';
import { Reports } from './pages/Reports';
import { NetworkDiscovery } from './pages/NetworkDiscovery';
import { SSHServers } from './pages/SSHServers';
import { Settings } from './pages/Settings';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <DebugProvider>
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
        </DebugProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}
