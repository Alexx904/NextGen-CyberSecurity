import React, { useState, useEffect, useCallback } from 'react';
import {
  Settings as SettingsIcon,
  Upload,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  ShieldCheck,
  Palette,
  Eye,
  Trash2,
  Lock
} from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { Tenant, BrandColors, UpdateBrandingPayload } from '../types/tenant';
import { useAuth } from '../context/AuthContext';
import { ReportPreview } from '../components/ReportPreview';
import { AlertChannelsCard } from '../components/AlertChannelsCard';

const COLOR_PRESETS = [
  {
    name: 'NextGen Cyber',
    colors: { primary: '#0a0d14', secondary: '#06b6d4', accent: '#3b82f6' },
  },
  {
    name: 'Enterprise Navy',
    colors: { primary: '#0f172a', secondary: '#2563eb', accent: '#38bdf8' },
  },
  {
    name: 'Emerald Trust',
    colors: { primary: '#064e3b', secondary: '#059669', accent: '#34d399' },
  },
  {
    name: 'Midnight Crimson',
    colors: { primary: '#1a1016', secondary: '#dc2626', accent: '#f87171' },
  },
];

export const Settings: React.FC = () => {
  const { user } = useAuth();

  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Form State
  const [name, setName] = useState<string>('');
  const [logoUrl, setLogoUrl] = useState<string>('');
  const [brandColors, setBrandColors] = useState<BrandColors>({
    primary: '#0f172a',
    secondary: '#2563eb',
    accent: '#38bdf8',
  });

  const canEdit = user?.role === 'SuperAdmin' || user?.role === 'TenantAdmin';

  const fetchTenantSettings = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<ApiResponse<Tenant>>('/tenants/current');
      if (response.data?.data) {
        const t = response.data.data;
        setTenant(t);
        setName(t.name);
        setLogoUrl(t.logo_url || '');
        if (t.brand_colors) {
          setBrandColors({
            primary: t.brand_colors.primary || '#0f172a',
            secondary: t.brand_colors.secondary || '#2563eb',
            accent: t.brand_colors.accent || '#38bdf8',
          });
        }
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile recuperare le impostazioni del tenant'
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTenantSettings();
  }, [fetchTenantSettings]);

  // Gestione Upload File Locale -> Base64
  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      alert('Seleziona un file immagine valido (PNG, JPG, SVG)');
      return;
    }

    if (file.size > 2 * 1024 * 1024) {
      alert('La dimensione massima consentita per il logo è 2MB');
      return;
    }

    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === 'string') {
        setLogoUrl(reader.result);
      }
    };
    reader.readAsDataURL(file);
  };

  const handleApplyPreset = (preset: BrandColors) => {
    setBrandColors(preset);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canEdit) return;

    setError(null);
    setSuccessMessage(null);
    setSaving(true);

    try {
      const payload: UpdateBrandingPayload = {
        name: name.trim(),
        logo_url: logoUrl.trim() || undefined,
        brand_colors: brandColors,
      };

      const response = await apiClient.put<ApiResponse<Tenant>>('/tenants/current/branding', payload);
      if (response.data && response.data.success) {
        setTenant(response.data.data);
        setSuccessMessage('Impostazioni di branding aggiornate con successo!');
        setTimeout(() => setSuccessMessage(null), 4000);
      } else {
        setError(response.data?.error?.message || 'Errore durante l\'aggiornamento del branding');
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile salvare le impostazioni di branding'
      );
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center space-x-3">
          <div className="p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
            <SettingsIcon className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-2xl font-bold tracking-tight text-white flex items-center space-x-2">
              <span>Impostazioni Tenant & Branding White-Label</span>
            </h2>
            <p className="text-sm text-slate-400 mt-0.5">
              Personalizzazione dell'identità visiva, loghi e palette per i report audit PDF e CSV
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          {!canEdit && (
            <span className="flex items-center space-x-1 px-2.5 py-1 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-400 text-xs font-medium">
              <Lock className="w-3.5 h-3.5" />
              <span>Sola Lettura ({user?.role})</span>
            </span>
          )}

          <button
            onClick={fetchTenantSettings}
            disabled={loading}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-cyber-border bg-slate-900/60 hover:bg-slate-800 text-slate-400 hover:text-white text-xs font-medium transition-colors"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin text-cyan-400' : ''}`} />
            <span>Ricarica</span>
          </button>
        </div>
      </div>

      {/* Success Alert */}
      {successMessage && (
        <div className="p-4 rounded-xl border border-emerald-800/50 bg-emerald-950/30 text-emerald-300 text-xs flex items-center space-x-2">
          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          <span>{successMessage}</span>
        </div>
      )}

      {/* Error Alert */}
      {error && (
        <div className="p-4 rounded-xl border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />
            <span>{error}</span>
          </div>
          <button
            onClick={fetchTenantSettings}
            className="px-2.5 py-1 rounded bg-rose-900/50 hover:bg-rose-900 text-rose-200 text-xs font-medium"
          >
            Riprova
          </button>
        </div>
      )}

      {/* Main Grid: Form + Live Preview */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Form Column */}
        <div className="lg:col-span-7 space-y-6">
          <form onSubmit={handleSubmit} className="p-6 rounded-2xl border border-cyber-border bg-cyber-card/60 shadow-xl space-y-6">
            <div className="border-b border-cyber-border/60 pb-3">
              <h3 className="text-base font-semibold text-white">Configurazione Organizzazione</h3>
              <p className="text-xs text-slate-400">Parametri anagrafici del tenant isolato</p>
            </div>

            {/* Ragione Sociale */}
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1.5">
                Ragione Sociale / Nome Azienda <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={!canEdit || loading}
                required
                placeholder="es: Acme Cyber Defense Corp"
                className="w-full px-3.5 py-2.5 bg-slate-900 border border-cyber-border rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-hidden focus:border-cyan-500 disabled:opacity-60"
              />
              {tenant && (
                <div className="mt-1 flex items-center space-x-2 text-[11px] text-slate-500 font-mono">
                  <span>Tenant ID: {tenant.id}</span>
                  <span>•</span>
                  <span>Slug: {tenant.slug}</span>
                </div>
              )}
            </div>

            {/* Logo Upload Section */}
            <div className="space-y-3 pt-2 border-t border-cyber-border/60">
              <label className="block text-xs font-medium text-slate-300">
                Logo Aziendale (White-Label PDF & Web Header)
              </label>

              {logoUrl ? (
                <div className="p-4 rounded-xl border border-cyber-border bg-slate-900/60 flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <div className="w-16 h-16 rounded-lg bg-slate-950 border border-cyber-border flex items-center justify-center p-2 overflow-hidden">
                      <img src={logoUrl} alt="Logo" className="max-h-full max-w-full object-contain" />
                    </div>
                    <div>
                      <div className="text-xs font-semibold text-white">Logo Attualmente Impostato</div>
                      <div className="text-[11px] text-slate-400 mt-0.5">
                        Utilizzato come intestazione grafica nei report PDF
                      </div>
                    </div>
                  </div>

                  {canEdit && (
                    <button
                      type="button"
                      onClick={() => setLogoUrl('')}
                      className="p-2 rounded-lg border border-rose-900/40 bg-rose-950/20 text-rose-400 hover:bg-rose-900/30 transition-colors"
                      title="Rimuovi Logo"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  )}
                </div>
              ) : (
                <div className="p-6 rounded-xl border-2 border-dashed border-cyber-border hover:border-cyan-500/50 bg-slate-900/40 text-center transition-colors">
                  <div className="w-10 h-10 rounded-lg bg-cyan-500/10 text-cyan-400 mx-auto flex items-center justify-center mb-2">
                    <Upload className="w-5 h-5" />
                  </div>
                  <div className="text-xs font-medium text-slate-200">
                    Trascina o carica il file del logo aziendale
                  </div>
                  <p className="text-[11px] text-slate-500 mt-0.5">
                    Formati supportati: PNG, JPG, SVG (Max 2MB)
                  </p>
                  {canEdit && (
                    <label className="mt-3 inline-block px-3 py-1.5 rounded-lg bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 border border-cyan-500/30 text-xs font-semibold cursor-pointer transition-colors">
                      <span>Sfoglia File Locale</span>
                      <input
                        type="file"
                        accept="image/*"
                        onChange={handleFileUpload}
                        className="hidden"
                      />
                    </label>
                  )}
                </div>
              )}

              {/* URL Alternativo */}
              <div>
                <input
                  type="url"
                  value={logoUrl.startsWith('data:') ? '' : logoUrl}
                  onChange={(e) => setLogoUrl(e.target.value)}
                  disabled={!canEdit || loading}
                  placeholder="Oppure inserisci URL diretto del logo (https://...)"
                  className="w-full px-3 py-2 bg-slate-900/80 border border-cyber-border rounded-lg text-xs text-slate-300 placeholder-slate-500 focus:outline-hidden focus:border-cyan-500 disabled:opacity-60"
                />
              </div>
            </div>

            {/* Palette Colori Brand */}
            <div className="space-y-4 pt-2 border-t border-cyber-border/60">
              <div className="flex items-center justify-between">
                <div>
                  <label className="block text-xs font-medium text-slate-300">
                    Palette Colori White-Label
                  </label>
                  <p className="text-[11px] text-slate-500">
                    Applicata a bordi, grafici e intestazioni dei report esportati
                  </p>
                </div>
                <Palette className="w-4 h-4 text-cyan-400" />
              </div>

              {/* Preset Rapidi */}
              <div>
                <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2">
                  Preset Consigliati
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                  {COLOR_PRESETS.map((preset) => (
                    <button
                      key={preset.name}
                      type="button"
                      disabled={!canEdit}
                      onClick={() => handleApplyPreset(preset.colors)}
                      className="p-2 rounded-lg border border-cyber-border bg-slate-900/60 hover:border-slate-700 text-left transition-colors disabled:opacity-60"
                    >
                      <div className="flex space-x-1 mb-1.5">
                        <span className="w-3.5 h-3.5 rounded-full" style={{ backgroundColor: preset.colors.primary }} />
                        <span className="w-3.5 h-3.5 rounded-full" style={{ backgroundColor: preset.colors.secondary }} />
                        <span className="w-3.5 h-3.5 rounded-full" style={{ backgroundColor: preset.colors.accent }} />
                      </div>
                      <div className="text-[11px] font-medium text-slate-300">{preset.name}</div>
                    </button>
                  ))}
                </div>
              </div>

              {/* Selettori Colori */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2">
                {/* Primario */}
                <div className="p-3 rounded-xl border border-cyber-border bg-slate-900/80 space-y-1.5">
                  <span className="text-[11px] font-semibold text-slate-400 uppercase">Colore Primario</span>
                  <div className="flex items-center space-x-2">
                    <input
                      type="color"
                      value={brandColors.primary}
                      onChange={(e) => setBrandColors({ ...brandColors, primary: e.target.value })}
                      disabled={!canEdit}
                      className="w-8 h-8 rounded border border-cyber-border bg-transparent cursor-pointer"
                    />
                    <input
                      type="text"
                      value={brandColors.primary}
                      onChange={(e) => setBrandColors({ ...brandColors, primary: e.target.value })}
                      disabled={!canEdit}
                      className="w-full px-2 py-1 bg-slate-950 border border-cyber-border rounded text-xs font-mono text-white focus:outline-hidden focus:border-cyan-500 uppercase"
                    />
                  </div>
                </div>

                {/* Secondario */}
                <div className="p-3 rounded-xl border border-cyber-border bg-slate-900/80 space-y-1.5">
                  <span className="text-[11px] font-semibold text-slate-400 uppercase">Colore Secondario</span>
                  <div className="flex items-center space-x-2">
                    <input
                      type="color"
                      value={brandColors.secondary}
                      onChange={(e) => setBrandColors({ ...brandColors, secondary: e.target.value })}
                      disabled={!canEdit}
                      className="w-8 h-8 rounded border border-cyber-border bg-transparent cursor-pointer"
                    />
                    <input
                      type="text"
                      value={brandColors.secondary}
                      onChange={(e) => setBrandColors({ ...brandColors, secondary: e.target.value })}
                      disabled={!canEdit}
                      className="w-full px-2 py-1 bg-slate-950 border border-cyber-border rounded text-xs font-mono text-white focus:outline-hidden focus:border-cyan-500 uppercase"
                    />
                  </div>
                </div>

                {/* Accento */}
                <div className="p-3 rounded-xl border border-cyber-border bg-slate-900/80 space-y-1.5">
                  <span className="text-[11px] font-semibold text-slate-400 uppercase">Colore Accento</span>
                  <div className="flex items-center space-x-2">
                    <input
                      type="color"
                      value={brandColors.accent}
                      onChange={(e) => setBrandColors({ ...brandColors, accent: e.target.value })}
                      disabled={!canEdit}
                      className="w-8 h-8 rounded border border-cyber-border bg-transparent cursor-pointer"
                    />
                    <input
                      type="text"
                      value={brandColors.accent}
                      onChange={(e) => setBrandColors({ ...brandColors, accent: e.target.value })}
                      disabled={!canEdit}
                      className="w-full px-2 py-1 bg-slate-950 border border-cyber-border rounded text-xs font-mono text-white focus:outline-hidden focus:border-cyan-500 uppercase"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Footer Form */}
            {canEdit && (
              <div className="pt-4 border-t border-cyber-border/60 flex items-center justify-end">
                <button
                  type="submit"
                  disabled={saving || loading}
                  className="flex items-center space-x-2 px-5 py-2.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-semibold text-xs transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {saving ? (
                    <>
                      <RefreshCw className="w-4 h-4 animate-spin" />
                      <span>Salvataggio modifiche...</span>
                    </>
                  ) : (
                    <>
                      <ShieldCheck className="w-4 h-4" />
                      <span>Salva Modifiche Branding</span>
                    </>
                  )}
                </button>
              </div>
            )}
          </form>
        </div>

        {/* Live Preview Column */}
        <div className="lg:col-span-5 space-y-3">
          <div className="flex items-center justify-between px-1">
            <div className="flex items-center space-x-2 text-xs font-semibold text-slate-300">
              <Eye className="w-4 h-4 text-cyan-400" />
              <span>Anteprima Live Report PDF White-Label</span>
            </div>
            <span className="text-[10px] text-slate-500 font-mono">Formato A4</span>
          </div>

          <ReportPreview
            organizationName={name}
            logoUrl={logoUrl}
            brandColors={brandColors}
          />
        </div>
      </div>
      <AlertChannelsCard />
    </div>
  );
};
