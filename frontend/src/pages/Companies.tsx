import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Building2,
  Plus,
  Search,
  CheckCircle2,
  FolderTree,
  Crosshair,
  Edit2,
  Trash2,
  ArrowRight,
  Sparkles,
  RefreshCw,
  X,
  ShieldAlert,
} from 'lucide-react';
import { useCompany } from '../context/CompanyContext';
import { TargetGroupNode } from '../types/target';
import { CompanyModal } from '../components/CompanyModal';

export const Companies: React.FC = () => {
  const navigate = useNavigate();
  const {
    companies,
    selectedCompanyId,
    setSelectedCompanyId,
    createCompany,
    updateCompany,
    deleteCompany,
    refreshCompanies,
    loading,
    error,
  } = useCompany();

  const [search, setSearch] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [companyToEdit, setCompanyToEdit] = useState<TargetGroupNode | null>(null);
  const [companyToDelete, setCompanyToDelete] = useState<TargetGroupNode | null>(null);
  const [deleting, setDeleting] = useState(false);

  // Statistiche globali aggregate
  const stats = useMemo(() => {
    let totalDepartments = 0;
    let totalTargets = 0;
    companies.forEach((c) => {
      totalDepartments += c.children?.length || 0;
      totalTargets += c.target_count || 0;
    });
    return {
      totalCompanies: companies.length,
      totalDepartments,
      totalTargets,
    };
  }, [companies]);

  // Filtro di ricerca per nome o descrizione
  const filteredCompanies = useMemo(() => {
    if (!search.trim()) return companies;
    const q = search.toLowerCase();
    return companies.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        (c.description && c.description.toLowerCase().includes(q))
    );
  }, [companies, search]);

  const handleCreateOrEdit = async (data: { name: string; description?: string; color_tag?: string }) => {
    if (companyToEdit) {
      await updateCompany(companyToEdit.id, data);
    } else {
      await createCompany(data);
    }
  };

  const handleDeleteConfirm = async () => {
    if (!companyToDelete) return;
    setDeleting(true);
    try {
      await deleteCompany(companyToDelete.id);
      setCompanyToDelete(null);
    } finally {
      setDeleting(false);
    }
  };

  const getRiskBadgeColor = (risk: string) => {
    switch (risk) {
      case 'CRITICAL':
        return 'bg-red-500/20 text-red-400 border-red-500/40';
      case 'HIGH':
        return 'bg-orange-500/20 text-orange-400 border-orange-500/40';
      case 'MEDIUM':
        return 'bg-amber-500/20 text-amber-400 border-amber-500/40';
      case 'LOW':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/40';
      default:
        return 'bg-slate-700 text-slate-300 border-slate-600';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Pagina */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-cyan-500/20 text-cyan-400">
              <Building2 className="h-6 w-6" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-white">Gestione Aziende & Scope</h1>
              <p className="text-xs text-slate-400">
                Amministra le organizzazioni aziendali, definisci i reparti interni e filtra l'intero software
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => refreshCompanies()}
            disabled={loading}
            className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800/80 px-3 py-2 text-xs font-medium text-slate-300 hover:bg-slate-700"
            title="Ricarica elenco aziende"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            Aggiorna
          </button>
          <button
            onClick={() => {
              setCompanyToEdit(null);
              setIsModalOpen(true);
            }}
            className="flex items-center gap-2 rounded-lg bg-cyan-600 px-4 py-2 text-xs font-medium text-white shadow-lg shadow-cyan-900/30 transition hover:bg-cyan-500"
          >
            <Plus className="h-4 w-4" />
            Nuova Azienda
          </button>
        </div>
      </div>

      {error && (
        <div className="rounded-xl border border-red-500/30 bg-red-500/10 p-4 text-xs text-red-400">
          {error}
        </div>
      )}

      {/* KPI Cards Panoramica */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 backdrop-blur-sm">
          <div className="flex items-center justify-between text-slate-400">
            <span className="text-xs font-medium">Aziende Registrate</span>
            <Building2 className="h-4 w-4 text-cyan-400" />
          </div>
          <div className="mt-2 text-2xl font-bold text-white">{stats.totalCompanies}</div>
          <p className="mt-1 text-[11px] text-slate-500">Perimetri organizzativi configurati</p>
        </div>

        <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 backdrop-blur-sm">
          <div className="flex items-center justify-between text-slate-400">
            <span className="text-xs font-medium">Reparti Aziendali</span>
            <FolderTree className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="mt-2 text-2xl font-bold text-white">{stats.totalDepartments}</div>
          <p className="mt-1 text-[11px] text-slate-500">Divisioni interne censite</p>
        </div>

        <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 backdrop-blur-sm">
          <div className="flex items-center justify-between text-slate-400">
            <span className="text-xs font-medium">Bersagli IP Totali</span>
            <Crosshair className="h-4 w-4 text-purple-400" />
          </div>
          <div className="mt-2 text-2xl font-bold text-white">{stats.totalTargets}</div>
          <p className="mt-1 text-[11px] text-slate-500">Host, subnet e range assegnati</p>
        </div>

        <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 backdrop-blur-sm">
          <div className="flex items-center justify-between text-slate-400">
            <span className="text-xs font-medium">Scope Attivo Globale</span>
            <Sparkles className="h-4 w-4 text-amber-400" />
          </div>
          <div className="mt-2 flex items-center justify-between">
            <span className="truncate text-sm font-semibold text-white">
              {selectedCompanyId
                ? companies.find((c) => c.id === selectedCompanyId)?.name || 'Azienda Selezionata'
                : 'Tutte le Aziende'}
            </span>
            {selectedCompanyId && (
              <button
                onClick={() => setSelectedCompanyId(null)}
                className="text-[11px] text-cyan-400 hover:underline"
              >
                Resetta
              </button>
            )}
          </div>
          <p className="mt-1 text-[11px] text-slate-500">
            {selectedCompanyId ? 'Filtro attivo su tutte le schermate' : 'Visione consolidata globale'}
          </p>
        </div>
      </div>

      {/* Barra di Ricerca & Filtri */}
      <div className="flex flex-col gap-3 rounded-xl border border-slate-800 bg-slate-900/50 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cerca azienda per nome o descrizione..."
            className="w-full rounded-lg border border-slate-700 bg-slate-800/80 pl-9 pr-4 py-2 text-xs text-white placeholder-slate-500 focus:border-cyan-500 focus:outline-none"
          />
        </div>
      </div>

      {/* Griglia Card Aziende */}
      {filteredCompanies.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-800 p-12 text-center">
          <Building2 className="mx-auto h-12 w-12 text-slate-600" />
          <h3 className="mt-3 text-sm font-semibold text-white">Nessuna Azienda Trovata</h3>
          <p className="mt-1 text-xs text-slate-400">
            {search ? 'Nessun risultato corrispondente ai criteri di ricerca' : 'Inizia creando la tua prima azienda cliente o organizzazione'}
          </p>
          {!search && (
            <button
              onClick={() => {
                setCompanyToEdit(null);
                setIsModalOpen(true);
              }}
              className="mt-4 inline-flex items-center gap-2 rounded-lg bg-cyan-600 px-4 py-2 text-xs font-medium text-white hover:bg-cyan-500"
            >
              <Plus className="h-4 w-4" />
              Crea Prima Azienda
            </button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredCompanies.map((company) => {
            const isSelected = selectedCompanyId === company.id;
            const departmentCount = company.children?.length || 0;
            const color = company.color_tag || '#3b82f6';

            return (
              <div
                key={company.id}
                className={`relative flex flex-col justify-between rounded-xl border transition-all ${
                  isSelected
                    ? 'border-cyan-500 bg-slate-900 shadow-lg shadow-cyan-950/40 ring-1 ring-cyan-500'
                    : 'border-slate-800 bg-slate-900/60 hover:border-slate-700 hover:bg-slate-900/90'
                }`}
              >
                {/* Header Card */}
                <div className="p-5">
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <div
                        className="flex h-11 w-11 items-center justify-center rounded-xl shadow-inner"
                        style={{ backgroundColor: `${color}25`, color: color }}
                      >
                        <Building2 className="h-6 w-6" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="font-semibold text-white text-base">{company.name}</h3>
                          {isSelected && (
                            <span className="flex items-center gap-1 rounded-full bg-cyan-500/20 px-2 py-0.5 text-[10px] font-semibold text-cyan-300 border border-cyan-500/30">
                              <CheckCircle2 className="h-3 w-3" />
                              ATTIVA
                            </span>
                          )}
                        </div>
                        <p className="mt-0.5 line-clamp-2 text-xs text-slate-400">
                          {company.description || 'Nessuna descrizione specificata'}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => {
                          setCompanyToEdit(company);
                          setIsModalOpen(true);
                        }}
                        className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-800 hover:text-white"
                        title="Modifica Azienda"
                      >
                        <Edit2 className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => setCompanyToDelete(company)}
                        className="rounded-lg p-1.5 text-slate-400 hover:bg-red-500/20 hover:text-red-400"
                        title="Elimina Azienda"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>

                  {/* Metriche Rapide */}
                  <div className="mt-5 grid grid-cols-3 gap-2 rounded-lg border border-slate-800/80 bg-slate-950/40 p-3 text-center">
                    <div>
                      <div className="text-[11px] text-slate-400">Reparti</div>
                      <div className="mt-1 text-sm font-bold text-white">{departmentCount}</div>
                    </div>
                    <div className="border-x border-slate-800/80">
                      <div className="text-[11px] text-slate-400">Target IP</div>
                      <div className="mt-1 text-sm font-bold text-cyan-400">{company.target_count || 0}</div>
                    </div>
                    <div>
                      <div className="text-[11px] text-slate-400">Rischio</div>
                      <div className="mt-1">
                        <span
                          className={`inline-block rounded px-1.5 py-0.5 text-[10px] font-bold border ${getRiskBadgeColor(
                            company.aggregate_risk
                          )}`}
                        >
                          {company.aggregate_risk || 'NONE'}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Footer Azioni Card */}
                <div className="flex items-center justify-between border-t border-slate-800/80 bg-slate-950/30 p-3 px-5">
                  <button
                    onClick={() => {
                      if (isSelected) {
                        setSelectedCompanyId(null);
                      } else {
                        setSelectedCompanyId(company.id);
                      }
                    }}
                    className={`flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium transition ${
                      isSelected
                        ? 'border border-cyan-500/50 bg-cyan-500/10 text-cyan-300 hover:bg-cyan-500/20'
                        : 'border border-slate-700 bg-slate-800 text-slate-300 hover:bg-slate-700 hover:text-white'
                    }`}
                  >
                    <CheckCircle2 className="h-3.5 w-3.5" />
                    {isSelected ? 'Disattiva Scope' : 'Seleziona come Attiva'}
                  </button>

                  <button
                    onClick={() => {
                      setSelectedCompanyId(company.id);
                      navigate('/targets');
                    }}
                    className="flex items-center gap-1 text-xs font-medium text-cyan-400 hover:text-cyan-300"
                  >
                    <span>Reparti & IP</span>
                    <ArrowRight className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Modale Creazione / Modifica */}
      <CompanyModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateOrEdit}
        companyToEdit={companyToEdit}
      />

      {/* Modale Conferma Eliminazione */}
      {companyToDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4 backdrop-blur-sm">
          <div className="relative w-full max-w-md rounded-xl border border-red-500/40 bg-slate-900 p-6 shadow-2xl">
            <button
              onClick={() => setCompanyToDelete(null)}
              className="absolute right-4 top-4 text-slate-400 hover:text-white"
            >
              <X className="h-5 w-5" />
            </button>

            <div className="flex items-center gap-3 text-red-400">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-red-500/20">
                <ShieldAlert className="h-5 w-5" />
              </div>
              <h3 className="text-base font-semibold text-white">Conferma Eliminazione Azienda</h3>
            </div>

            <p className="mt-3 text-xs text-slate-300 leading-relaxed">
              Sei sicuro di voler eliminare l'azienda{' '}
              <strong className="text-white font-semibold">{companyToDelete.name}</strong>?
              Questa operazione rimuoverà la struttura aziendale, i reparti associati e i collegamenti ai target.
            </p>

            <div className="mt-6 flex justify-end gap-3 border-t border-slate-800 pt-4">
              <button
                type="button"
                onClick={() => setCompanyToDelete(null)}
                disabled={deleting}
                className="rounded-lg border border-slate-700 px-4 py-2 text-xs font-medium text-slate-300 hover:bg-slate-800"
              >
                Annulla
              </button>
              <button
                type="button"
                onClick={handleDeleteConfirm}
                disabled={deleting}
                className="flex items-center gap-2 rounded-lg bg-red-600 px-4 py-2 text-xs font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                <Trash2 className="h-4 w-4" />
                {deleting ? 'Eliminazione...' : 'Conferma Eliminazione'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
