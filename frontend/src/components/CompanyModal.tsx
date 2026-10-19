import React, { useState, useEffect } from 'react';
import { X, Building2, Check } from 'lucide-react';
import { TargetGroupNode } from '../types/target';

interface CompanyModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; description?: string; color_tag?: string }) => Promise<void>;
  companyToEdit?: TargetGroupNode | null;
}

const PRESET_COLORS = [
  '#3b82f6', // Blue
  '#10b981', // Emerald
  '#8b5cf6', // Violet
  '#f59e0b', // Amber
  '#ef4444', // Red
  '#06b6d4', // Cyan
  '#ec4899', // Pink
  '#6366f1', // Indigo
];

export const CompanyModal: React.FC<CompanyModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  companyToEdit,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [colorTag, setColorTag] = useState('#3b82f6');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (companyToEdit) {
      setName(companyToEdit.name);
      setDescription(companyToEdit.description || '');
      setColorTag(companyToEdit.color_tag || '#3b82f6');
    } else {
      setName('');
      setDescription('');
      setColorTag('#3b82f6');
    }
    setError(null);
  }, [companyToEdit, isOpen]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Il nome dell\'azienda è obbligatorio');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      await onSubmit({
        name: name.trim(),
        description: description.trim() || undefined,
        color_tag: colorTag,
      });
      onClose();
    } catch (err: any) {
      setError(err.message || 'Errore durante il salvataggio dell\'azienda');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4 backdrop-blur-xs">
      <div className="relative w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <button
          onClick={onClose}
          className="absolute right-4 top-4 text-slate-400 hover:text-white"
        >
          <X className="h-5 w-5" />
        </button>

        <div className="flex items-center gap-3 border-b border-slate-800 pb-4">
          <div
            className="flex h-10 w-10 items-center justify-center rounded-lg"
            style={{ backgroundColor: `${colorTag}20`, color: colorTag }}
          >
            <Building2 className="h-5 w-5" />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-white">
              {companyToEdit ? 'Modifica Azienda' : 'Registra Nuova Azienda'}
            </h2>
            <p className="text-xs text-slate-400">
              {companyToEdit
                ? 'Aggiorna i parametri anagrafici e il colore identificativo'
                : 'Crea una nuova organizzazione per raggruppare reparti e bersagli IP'}
            </p>
          </div>
        </div>

        {error && (
          <div className="mt-4 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-400">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-300">
              Nome Azienda / Organizzazione <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="es. Acme Security Corp"
              className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white placeholder-slate-500 focus:border-cyan-500 focus:outline-hidden focus:ring-1 focus:ring-cyan-500"
              autoFocus
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-slate-300">
              Descrizione o Note
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              placeholder="es. Perimetro cloud e filiali operative Italia"
              className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white placeholder-slate-500 focus:border-cyan-500 focus:outline-hidden focus:ring-1 focus:ring-cyan-500"
            />
          </div>

          <div>
            <label className="mb-2 block text-xs font-medium text-slate-300">
              Colore Identificativo (Brand & Tag)
            </label>
            <div className="flex flex-wrap items-center gap-2">
              {PRESET_COLORS.map((color) => (
                <button
                  key={color}
                  type="button"
                  onClick={() => setColorTag(color)}
                  className="relative flex h-8 w-8 items-center justify-center rounded-lg border border-slate-700 transition hover:scale-110"
                  style={{ backgroundColor: color }}
                >
                  {colorTag === color && <Check className="h-4 w-4 text-white drop-shadow" />}
                </button>
              ))}
              <input
                type="color"
                value={colorTag}
                onChange={(e) => setColorTag(e.target.value)}
                className="h-8 w-10 cursor-pointer rounded border border-slate-700 bg-slate-800 p-0.5"
                title="Seleziona colore personalizzato"
              />
            </div>
          </div>

          <div className="flex justify-end gap-3 border-t border-slate-800 pt-4">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="rounded-lg border border-slate-700 px-4 py-2 text-xs font-medium text-slate-300 hover:bg-slate-800"
            >
              Annulla
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex items-center gap-2 rounded-lg bg-cyan-600 px-4 py-2 text-xs font-medium text-white hover:bg-cyan-500 disabled:opacity-50"
            >
              <Check className="h-4 w-4" />
              {loading ? 'Salvataggio...' : companyToEdit ? 'Aggiorna Azienda' : 'Crea Azienda'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
