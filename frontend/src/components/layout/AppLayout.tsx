import React, { useState, useEffect } from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useCompany } from '../../context/CompanyContext';
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
  Network,
  Building2,
  ChevronDown,
  ArrowRight,
} from 'lucide-react';

interface HealthData {
  status: string;
  app: string;
  version: string;
  timestamp: string;
}

export const AppLayout: React.FC = () => {
  const { user, logout } = useAuth();
  const {
    companies,
    selectedCompanyId,
    selectedCompany,
    setSelectedCompanyId,
  } = useCompany();
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

  interface NavItem {
    name: string;
    path: string;
    icon: React.ElementType;
    badge?: string;
  }

  interface NavSection {
    category?: string;
    items: NavItem[];
  }

  const navSections: NavSection[] = [
    {
      items: [
        { name: 'Dashboard Panoramica', path: '/', icon: Activity },
      ],
    },
    {
      category: 'Organizzazione & Scope',
      items: [
        { name: 'Target & Reparti', path: '/targets', icon: Crosshair, badge: 'Reparti' },
      ],
    },
    {
      category: 'Flusso Operativo',
      items: [
        { name: '1. Network Discovery', path: '/network-discovery', icon: Radar, badge: 'Passo 1' },
        { name: '2. Scansioni Nmap', path: '/scans', icon: Terminal, badge: 'Passo 2' },
        { name: '3. Vulnerabilità (CVE)', path: '/findings', icon: BarChart3, badge: 'Passo 3' },
        { name: '4. Report White-Label', path: '/reports', icon: FileText, badge: 'Passo 4' },
      ],
    },
    {
      category: 'Tool indipendenti dalle aziende',
      items: [
        { name: 'Server SSH & Firewall', path: '/ssh-servers', icon: Server, badge: 'iptables' },
      ],
    },
    {
      category: 'Analisi Minacce',
      items: [
        { name: 'Attack Path Graph', path: '/attack-path', icon: Network, badge: 'Grafo' },
      ],
    },
    {
      category: 'Configurazione',
      items: [
        { name: 'Impostazioni Tenant', path: '/settings', icon: Settings },
      ],
    },
  ];

  const renderCompanySwitcher = (onItemClick?: () => void) => (
    <div className="mb-4 pb-3.5 border-b border-cyber-border space-y-2">
      <div className="flex items-center justify-between text-[11px] font-semibold text-slate-400 uppercase tracking-wider px-1">
        <div className="flex items-center gap-1.5 text-cyan-400">
          <Building2 className="w-3.5 h-3.5" />
          <span>Azienda / Scope</span>
        </div>
        <NavLink
          to="/"
          onClick={onItemClick}
          className="text-[10px] text-slate-400 hover:text-cyan-400 transition flex items-center gap-0.5"
          title="Gestisci Aziende"
        >
          <span>Gestione</span>
          <ArrowRight className="w-2.5 h-2.5" />
        </NavLink>
      </div>

      <div className="relative">
        <select
          value={selectedCompanyId || ''}
          onChange={(e) => {
            setSelectedCompanyId(e.target.value ? e.target.value : null);
            if (onItemClick) onItemClick();
          }}
          className="w-full appearance-none rounded-lg border border-slate-700 bg-slate-800/90 py-2 pl-3 pr-8 text-xs font-medium text-white focus:border-cyan-500 focus:outline-hidden focus:ring-1 focus:ring-cyan-500 cursor-pointer truncate"
          aria-label="Seleziona Azienda Scope"
        >
          <option value="">🌐 Tutte le Aziende (Globale)</option>
          {companies.map((c) => (
            <option key={c.id} value={c.id}>
              🏢 {c.name} ({c.children?.length || 0} rep. · {c.target_count || 0} IP)
            </option>
          ))}
        </select>
        <ChevronDown className="absolute right-2.5 top-2.5 w-3.5 h-3.5 text-slate-400 pointer-events-none" />
      </div>

      {selectedCompany && (
        <div
          className="flex items-center justify-between px-2.5 py-1.5 rounded-lg border text-[11px] font-medium transition-all"
          style={{
            backgroundColor: `${selectedCompany.color_tag || '#3b82f6'}15`,
            borderColor: `${selectedCompany.color_tag || '#3b82f6'}40`,
            color: selectedCompany.color_tag || '#38bdf8',
          }}
        >
          <span className="truncate font-semibold flex items-center gap-1.5">
            <span
              className="w-2 h-2 rounded-full inline-block shrink-0"
              style={{ backgroundColor: selectedCompany.color_tag || '#3b82f6' }}
            />
            <span className="truncate">{selectedCompany.name}</span>
          </span>
          <button
            onClick={() => setSelectedCompanyId(null)}
            className="text-slate-400 hover:text-white p-0.5 ml-1 transition shrink-0"
            title="Disattiva filtro e mostra tutte le aziende"
          >
            <X className="w-3 h-3" />
          </button>
        </div>
      )}
    </div>
  );

  const renderNavLinks = (onItemClick?: () => void) => (
    <div className="space-y-4">
      {navSections.map((section, idx) => (
        <div key={section.category || `sec-${idx}`} className="space-y-1">
          {section.category && (
            <div className="px-3 pb-1 pt-1.5 flex items-center justify-between text-[10px] font-bold uppercase tracking-wider text-slate-400 select-none">
              <span>{section.category}</span>
              <span className="w-8 h-[1px] bg-slate-800" />
            </div>
          )}
          <div className="space-y-0.5">
            {section.items.map((item) => {
              const Icon = item.icon;
              const locked = item.path !== '/' && item.path !== '/ssh-servers' && !companies.some(c => c.id === selectedCompanyId);
              if (locked) return <div key={item.path} aria-disabled="true" title="Seleziona un’azienda dalla Dashboard" className="flex items-center gap-2.5 px-3 py-2 text-xs text-slate-600 cursor-not-allowed"><Icon className="w-4 h-4" />{item.name}</div>;
              return (
                <NavLink
                  key={item.path}
                  to={item.path}
                  end={item.path === '/'}
                  onClick={onItemClick}
                  className={({ isActive }) =>
                    `group flex items-center justify-between px-3 py-2 rounded-lg font-medium text-xs transition-all border ${
                      isActive
                        ? 'bg-cyan-500/15 text-cyan-300 border-cyan-500/40 shadow-[0_0_12px_rgba(6,182,212,0.15)] font-semibold'
                        : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/60 border-transparent'
                    }`
                  }
                >
                  <div className="flex items-center space-x-2.5 truncate">
                    <Icon className="w-4 h-4 shrink-0 transition-transform group-hover:scale-110" />
                    <span className="truncate">{item.name}</span>
                  </div>
                  {item.badge && (
                    <span className="text-[9px] font-mono font-semibold px-1.5 py-0.2 rounded bg-slate-900 border border-slate-750 text-slate-400 group-hover:text-cyan-400 group-hover:border-cyan-500/30 transition-colors shrink-0 ml-1">
                      {item.badge}
                    </span>
                  )}
                </NavLink>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );

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
        <aside className="w-64 border-r border-cyber-border bg-[#0d121f] p-3.5 hidden md:flex flex-col justify-between overflow-y-auto">
          <div>
            {renderCompanySwitcher()}
            <nav className="space-y-1">
              {renderNavLinks()}
            </nav>
          </div>

          <div className="pt-4 mt-auto">
            <div className="p-3 rounded-lg border border-cyber-border bg-slate-900/60 text-xs space-y-1">
              <div className="text-slate-400 text-[11px]">Tenant Context</div>
              <div className="font-semibold text-white truncate text-xs" title={user?.tenant_id}>
                {user?.tenant_id ? `ID: ${user.tenant_id.slice(0, 8)}...` : 'Default Organization'}
              </div>
              <div className="text-[10px] text-cyan-400 font-mono">Multi-Company Ready</div>
            </div>
          </div>
        </aside>

        {/* Mobile Navigation Drawer */}
        {mobileMenuOpen && (
          <div className="md:hidden fixed inset-0 z-40 bg-black/60 backdrop-blur-xs flex">
            <div className="w-72 bg-[#0d121f] border-r border-cyber-border p-4 flex flex-col justify-between overflow-y-auto">
              <div className="mt-14">
                {renderCompanySwitcher(() => setMobileMenuOpen(false))}
                <nav>
                  {renderNavLinks(() => setMobileMenuOpen(false))}
                </nav>
              </div>

              <div className="pt-4 mt-auto">
                <div className="p-3 rounded-lg border border-cyber-border bg-slate-900/60 text-xs space-y-1">
                  <div className="text-slate-400 text-[11px]">Tenant Context</div>
                  <div className="font-semibold text-white text-xs">Multi-Tenant Workstation</div>
                </div>
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
