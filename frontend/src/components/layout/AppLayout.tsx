import React, { useState, useEffect } from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import axios from 'axios';
import {
  ShieldCheck,
  Activity,
  Crosshair,
  Terminal,
  BarChart3,
  FileText,
  Settings,
  LogOut,
  User,
  CheckCircle2,
  XCircle,
  RefreshCw,
  Menu,
  X,
  Radar,
  Server,
} from 'lucide-react';

interface HealthData {
  status: string;
  app: string;
  version: string;
  timestamp: string;
}

export const AppLayout: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [health, setHealth] = useState<HealthData | null>(null);
  const [healthLoading, setHealthLoading] = useState(false);

  const fetchHealth = async () => {
    setHealthLoading(true);
    try {
      const response = await axios.get<HealthData>('/api/v1/health');
      setHealth(response.data);
    } catch {
      setHealth(null);
    } finally {
      setHealthLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
    const interval = setInterval(fetchHealth, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const navItems = [
    { name: 'Overview', path: '/', icon: Activity },
    { name: 'Target & Asset', path: '/targets', icon: Crosshair },
    { name: 'Network Discovery', path: '/network-discovery', icon: Radar },
    { name: 'Scansioni Nmap', path: '/scans', icon: Terminal },
    { name: 'Server SSH & Firewall', path: '/ssh-servers', icon: Server },
    { name: 'Vulnerabilità (CVE)', path: '/findings', icon: BarChart3 },
    { name: 'Report White-Label', path: '/reports', icon: FileText },
    { name: 'Impostazioni Tenant', path: '/settings', icon: Settings },
  ];

  return (
    <div className="min-h-screen flex flex-col bg-[#0a0d14] text-slate-100">
      {/* Top Navbar */}
      <header className="border-b border-cyber-border bg-[#111726]/90 backdrop-blur-md sticky top-0 z-50 px-4 lg:px-6 py-3 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800"
            aria-label="Toggle menu"
          >
            {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
          <div className="w-9 h-9 rounded-lg bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center text-cyan-400 shadow-[0_0_15px_rgba(6,182,212,0.15)]">
            <ShieldCheck className="w-5 h-5" />
          </div>
          <div>
            <h1 className="text-base lg:text-lg font-bold tracking-tight bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
              NextGen CyberSecurity
            </h1>
            <p className="text-[11px] text-slate-400 hidden sm:block">Vulnerability & Compliance Platform</p>
          </div>
        </div>

        {/* Status Indicator & User Menu */}
        <div className="flex items-center space-x-3 lg:space-x-4">
          {/* Health Pill */}
          <div className="hidden sm:flex items-center space-x-2 px-3 py-1 rounded-full bg-slate-900 border border-cyber-border text-xs">
            {healthLoading ? (
              <RefreshCw className="w-3.5 h-3.5 text-yellow-400 animate-spin" />
            ) : health?.status === 'healthy' ? (
              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
            ) : (
              <XCircle className="w-3.5 h-3.5 text-rose-400" />
            )}
            <span className={health?.status === 'healthy' ? 'text-emerald-400 font-medium' : 'text-slate-400'}>
              {health?.status === 'healthy' ? 'Online' : 'Offline'}
            </span>
          </div>

          {/* User Profile Pill */}
          {user && (
            <div className="flex items-center space-x-2.5 px-3 py-1.5 rounded-lg bg-slate-900/80 border border-cyber-border text-xs">
              <div className="w-6 h-6 rounded-full bg-cyan-950 border border-cyan-500/40 flex items-center justify-center text-cyan-400">
                <User className="w-3.5 h-3.5" />
              </div>
              <div className="hidden md:block text-left">
                <div className="font-medium text-slate-200 leading-tight">{user.full_name || user.email}</div>
                <div className="text-[10px] text-cyan-400 font-mono leading-tight">{user.role}</div>
              </div>
            </div>
          )}

          {/* Logout Button */}
          <button
            onClick={handleLogout}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-rose-900/40 bg-rose-950/20 hover:bg-rose-900/30 text-rose-300 hover:text-rose-200 transition-colors text-xs font-medium"
            title="Esci dalla sessione"
          >
            <LogOut className="w-3.5 h-3.5" />
            <span className="hidden sm:inline">Logout</span>
          </button>
        </div>
      </header>

      {/* Main Workspace Layout */}
      <div className="flex-1 flex">
        {/* Sidebar Nav (Desktop) */}
        <aside className="w-64 border-r border-cyber-border bg-[#0d121f] p-4 hidden md:flex flex-col justify-between">
          <nav className="space-y-1">
            {navItems.map((item) => {
              const Icon = item.icon;
              return (
                <NavLink
                  key={item.path}
                  to={item.path}
                  end={item.path === '/'}
                  className={({ isActive }) =>
                    `flex items-center space-x-3 px-3 py-2 rounded-lg font-medium text-sm transition-colors border ${
                      isActive
                        ? 'bg-cyan-500/10 text-cyan-400 border-cyan-500/30'
                        : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 border-transparent'
                    }`
                  }
                >
                  <Icon className="w-4 h-4" />
                  <span>{item.name}</span>
                </NavLink>
              );
            })}
          </nav>

          <div className="p-3 rounded-lg border border-cyber-border bg-slate-900/60 text-xs space-y-1">
            <div className="text-slate-400">Tenant Context</div>
            <div className="font-semibold text-white truncate" title={user?.tenant_id}>
              {user?.tenant_id ? `ID: ${user.tenant_id.slice(0, 8)}...` : 'Default Organization'}
            </div>
            <div className="text-[10px] text-cyan-400">Multi-Tenant Isolation</div>
          </div>
        </aside>

        {/* Mobile Navigation Drawer */}
        {mobileMenuOpen && (
          <div className="md:hidden fixed inset-0 z-40 bg-black/60 backdrop-blur-sm flex">
            <div className="w-64 bg-[#0d121f] border-r border-cyber-border p-4 flex flex-col justify-between">
              <nav className="space-y-1 mt-14">
                {navItems.map((item) => {
                  const Icon = item.icon;
                  return (
                    <NavLink
                      key={item.path}
                      to={item.path}
                      end={item.path === '/'}
                      onClick={() => setMobileMenuOpen(false)}
                      className={({ isActive }) =>
                        `flex items-center space-x-3 px-3 py-2.5 rounded-lg font-medium text-sm transition-colors border ${
                          isActive
                            ? 'bg-cyan-500/10 text-cyan-400 border-cyan-500/30'
                            : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 border-transparent'
                        }`
                      }
                    >
                      <Icon className="w-4 h-4" />
                      <span>{item.name}</span>
                    </NavLink>
                  );
                })}
              </nav>

              <div className="p-3 rounded-lg border border-cyber-border bg-slate-900/60 text-xs space-y-1">
                <div className="text-slate-400">Tenant Context</div>
                <div className="font-semibold text-white">Multi-Tenant Hybrid Mode</div>
              </div>
            </div>
            <div className="flex-1" onClick={() => setMobileMenuOpen(false)} />
          </div>
        )}

        {/* Content Area */}
        <main className="flex-1 p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto w-full">
          <Outlet />
        </main>
      </div>
    </div>
  );
};
