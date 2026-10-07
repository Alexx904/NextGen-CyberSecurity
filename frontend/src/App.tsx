import { useEffect, useState } from 'react';
import axios from 'axios';
import { 
  ShieldCheck, 
  Server, 
  Terminal, 
  Database, 
  Layers, 
  CheckCircle2, 
  XCircle, 
  RefreshCw,
  Activity,
  BarChart3,
  Crosshair,
  FileText,
  Settings
} from 'lucide-react';

interface HealthData {
  status: string;
  app: string;
  version: string;
  timestamp: string;
}

export default function App() {
  const [health, setHealth] = useState<HealthData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchHealth = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.get<HealthData>('/api/v1/health');
      setHealth(response.data);
    } catch (err: any) {
      setError(err.message || 'Impossibile contattare il backend');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
  }, []);

  return (
    <div className="min-h-screen flex flex-col bg-[#0a0d14] text-slate-100">
      {/* Top Navbar */}
      <header className="border-b border-cyber-border bg-[#111726]/80 backdrop-blur-md sticky top-0 z-50 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 rounded-lg bg-cyber-accent/10 border border-cyber-accent/30 flex items-center justify-center text-cyber-accent shadow-[0_0_15px_rgba(6,182,212,0.15)]">
            <ShieldCheck className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-lg font-bold tracking-tight bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
              NextGen CyberSecurity
            </h1>
            <p className="text-xs text-slate-400">Enterprise Vulnerability & Compliance Platform</p>
          </div>
        </div>

        {/* Status Indicator */}
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-slate-900 border border-cyber-border text-xs">
            {loading ? (
              <RefreshCw className="w-3.5 h-3.5 text-yellow-400 animate-spin" />
            ) : health?.status === 'healthy' ? (
              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
            ) : (
              <XCircle className="w-3.5 h-3.5 text-rose-400" />
            )}
            <span className={health?.status === 'healthy' ? 'text-emerald-400 font-medium' : 'text-slate-400'}>
              {loading ? 'Verifica backend...' : health?.status === 'healthy' ? 'Backend Online' : 'Backend Non Raggiungibile'}
            </span>
          </div>

          <button 
            onClick={fetchHealth} 
            className="p-1.5 rounded-lg border border-cyber-border hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
            title="Aggiorna stato"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </header>

      {/* Main Workspace Layout */}
      <div className="flex-1 flex">
        {/* Sidebar Nav */}
        <aside className="w-64 border-r border-cyber-border bg-[#0d121f] p-4 hidden md:flex flex-col justify-between">
          <nav className="space-y-1">
            <a href="#" className="flex items-center space-x-3 px-3 py-2 rounded-lg bg-cyber-accent/10 text-cyber-accent border border-cyber-accent/30 font-medium text-sm">
              <Activity className="w-4 h-4" />
              <span>Overview</span>
            </a>
            <a href="#" className="flex items-center space-x-3 px-3 py-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 transition-colors text-sm">
              <Crosshair className="w-4 h-4" />
              <span>Target & Asset</span>
            </a>
            <a href="#" className="flex items-center space-x-3 px-3 py-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 transition-colors text-sm">
              <Terminal className="w-4 h-4" />
              <span>Scansioni Nmap</span>
            </a>
            <a href="#" className="flex items-center space-x-3 px-3 py-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 transition-colors text-sm">
              <BarChart3 className="w-4 h-4" />
              <span>Vulnerabilità (CVE)</span>
            </a>
            <a href="#" className="flex items-center space-x-3 px-3 py-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 transition-colors text-sm">
              <FileText className="w-4 h-4" />
              <span>Report White-Label</span>
            </a>
            <a href="#" className="flex items-center space-x-3 px-3 py-2 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 transition-colors text-sm">
              <Settings className="w-4 h-4" />
              <span>Impostazioni Tenant</span>
            </a>
          </nav>

          <div className="p-3 rounded-lg border border-cyber-border bg-slate-900/60 text-xs space-y-1">
            <div className="text-slate-400">Tenant Attivo</div>
            <div className="font-semibold text-white">Default Organization</div>
            <div className="text-[10px] text-cyber-accent">Multi-Tenant Hybrid Mode</div>
          </div>
        </aside>

        {/* Content Area */}
        <main className="flex-1 p-6 lg:p-8 space-y-8 max-w-7xl mx-auto w-full">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">System Infrastructure Overview</h2>
            <p className="text-sm text-slate-400 mt-1">
              Architettura multi-container operativa: Backend Go, Motore Nmap 7.95, PostgreSQL 16 e Frontend React.
            </p>
          </div>

          {/* Infrastructure Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="p-5 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
              <div className="flex items-center justify-between mb-3">
                <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Backend Service</span>
                <Server className="w-5 h-5 text-cyan-400" />
              </div>
              <div className="text-lg font-bold text-white">Golang 1.22+</div>
              <div className="text-xs text-slate-400 mt-1">go-chi/chi v5 REST API</div>
              <div className="mt-4 pt-3 border-t border-cyber-border/60 flex items-center justify-between text-xs">
                <span className="text-slate-400">Status:</span>
                <span className="text-emerald-400 font-medium">Running (Port 8080)</span>
              </div>
            </div>

            <div className="p-5 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
              <div className="flex items-center justify-between mb-3">
                <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Network Scanner</span>
                <Terminal className="w-5 h-5 text-amber-400" />
              </div>
              <div className="text-lg font-bold text-white">Nmap 7.95</div>
              <div className="text-xs text-slate-400 mt-1">CAP_NET_RAW / Linux Alpine</div>
              <div className="mt-4 pt-3 border-t border-cyber-border/60 flex items-center justify-between text-xs">
                <span className="text-slate-400">Capabilities:</span>
                <span className="text-amber-400 font-medium">SYN & Service Ready</span>
              </div>
            </div>

            <div className="p-5 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
              <div className="flex items-center justify-between mb-3">
                <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Database Layer</span>
                <Database className="w-5 h-5 text-blue-400" />
              </div>
              <div className="text-lg font-bold text-white">PostgreSQL 16</div>
              <div className="text-xs text-slate-400 mt-1">JSONB Findings & CVE Store</div>
              <div className="mt-4 pt-3 border-t border-cyber-border/60 flex items-center justify-between text-xs">
                <span className="text-slate-400">Healthcheck:</span>
                <span className="text-emerald-400 font-medium">Healthy (pg_isready)</span>
              </div>
            </div>

            <div className="p-5 rounded-xl border border-cyber-border bg-cyber-card/60 relative overflow-hidden">
              <div className="flex items-center justify-between mb-3">
                <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Web UI Layer</span>
                <Layers className="w-5 h-5 text-violet-400" />
              </div>
              <div className="text-lg font-bold text-white">React 18 + Vite</div>
              <div className="text-xs text-slate-400 mt-1">Tailwind CSS + Nginx Proxy</div>
              <div className="mt-4 pt-3 border-t border-cyber-border/60 flex items-center justify-between text-xs">
                <span className="text-slate-400">Reverse Proxy:</span>
                <span className="text-violet-400 font-medium">Port 3000 &rarr; 8080</span>
              </div>
            </div>
          </div>

          {/* Backend Diagnostics Banner */}
          <div className="p-6 rounded-xl border border-cyber-border bg-cyber-card/40 space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-base font-semibold text-white flex items-center space-x-2">
                <Activity className="w-4 h-4 text-cyber-accent" />
                <span>Backend Health Diagnostic Payload</span>
              </h3>
              <span className="text-xs text-slate-400">Endpoint: /api/v1/health</span>
            </div>

            {loading ? (
              <div className="p-4 rounded-lg bg-slate-900/60 text-slate-400 text-sm flex items-center space-x-2">
                <RefreshCw className="w-4 h-4 animate-spin text-cyber-accent" />
                <span>Interrogazione backend in corso...</span>
              </div>
            ) : error ? (
              <div className="p-4 rounded-lg bg-rose-950/40 border border-rose-800/40 text-rose-300 text-sm">
                Errore: {error}
              </div>
            ) : (
              <pre className="p-4 rounded-lg bg-slate-950 border border-cyber-border/60 text-xs font-mono text-emerald-400 overflow-x-auto">
                {JSON.stringify(health, null, 2)}
              </pre>
            )}
          </div>
        </main>
      </div>
    </div>
  );
}

