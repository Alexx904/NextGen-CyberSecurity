import React, { useState, useEffect } from 'react';
import { X, Folder, AlertCircle } from 'lucide-react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetGroup, TargetGroupNode } from '../types/target';

interface GroupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onGroupSaved: () => void;
  groupToEdit?: TargetGroup | null;
  defaultParentId?: string | null;
  existingGroups: (TargetGroup | TargetGroupNode)[];
}

const COLOR_PRESETS = [
  { name: 'Ciano Cyber', value: '#06b6d4' },
  { name: 'Blu Oceano', value: '#3b82f6' },
  { name: 'Indaco', value: '#6366f1' },
  { name: 'Smeraldo', value: '#10b981' },
  { name: 'Ambra', value: '#f59e0b' },
  { name: 'Rosa Neon', value: '#f43f5e' },
  { name: 'Viola', value: '#a855f7' },
  { name: 'Ardesia', value: '#64748b' },
];

export const GroupModal: React.FC<GroupModalProps> = ({
  isOpen,
  onClose,
  onGroupSaved,
  groupToEdit,
  defaultParentId,
  existingGroups,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [parentId, setParentId] = useState<string>('');
  const [colorTag, setColorTag] = useState<string>('#06b6d4');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      if (groupToEdit) {
        setName(groupToEdit.name || '');
        setDescription(groupToEdit.description || '');
        setParentId(groupToEdit.parent_id || '');
        setColorTag(groupToEdit.color_tag || '#06b6d4');
      } else {
        setName('');
        setDescription('');
        setParentId(defaultParentId || '');
        setColorTag('#06b6d4');
      }
      setError(null);
    }
  }, [isOpen, groupToEdit, defaultParentId]);

  if (!isOpen) return null;

  // Filtra i gruppi selezionabili come genitore (esclude sé stesso in modifica)
  const selectableParents = existingGroups.filter(
    (g) => !groupToEdit || g.id !== groupToEdit.id
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('Il nome della cartella è obbligatorio');
      return;
    }

    setSubmitting(true);
    try {
      const payload = {
        name: name.trim(),
        description: description.trim() || undefined,
        parent_id: parentId ? parentId : null,
        color_tag: colorTag,
      };

      let response;
      if (groupToEdit) {
        response = await apiClient.put<ApiResponse<TargetGroup>>(
          `/target-groups/${groupToEdit.id}`,
          payload
        );
      } else {
        response = await apiClient.post<ApiResponse<TargetGroup>>(
          '/target-groups',
          payload
        );
      }

      if (response.data && response.data.success) {
        onGroupSaved();
        onClose();
      } else {
        setError(response.data?.error?.message || 'Operazione sulla cartella fallita');
      }
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
        err.message ||
        'Impossibile salvare la cartella'
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="max-w-md w-full bg-[#111726] border border-cyber-border rounded-2xl shadow-2xl overflow-hidden relative">
        {/* Header */}
        <div className="p-5 border-b border-cyber-border flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div
              className="p-2 rounded-lg border"
              style={{
                backgroundColor: `${colorTag}15`,
                borderColor: `${colorTag}40`,
                color: colorTag,
              }}
            >
              <Folder className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">
                {groupToEdit ? 'Modifica Cartella Scope' : 'Nuova Cartella Scope'}
              </h3>
              <p className="text-xs text-slate-400">
                Organizza gli asset in gruppi logici gerarchici
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3.5 rounded-lg border border-rose-800/50 bg-rose-950/30 text-rose-300 text-xs flex items-start space-x-2">
              <AlertCircle className="w-4 h-4 shrink-0 text-rose-400 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {/* Nome */}
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1.5">
              Nome Cartella <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="es: Sede Milano, DMZ Produzione, Cloud AWS"
              required
              className="w-full px-3.5 py-2.5 bg-slate-900 border border-cyber-border rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500"
            />
          </div>

          {/* Descrizione */}
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1.5">
              Descrizione o Note (Opzionale)
            </label>
            <input
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="es: Rete perimeteriale esposta al pubblico"
              className="w-full px-3.5 py-2 bg-slate-900 border border-cyber-border rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
            />
          </div>

          {/* Cartella Genitore */}
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1.5">
              Posizione nell'Albero (Cartella Padre)
            </label>
            <select
              value={parentId}
              onChange={(e) => setParentId(e.target.value)}
              className="w-full px-3.5 py-2.5 bg-slate-900 border border-cyber-border rounded-lg text-sm text-slate-100 focus:outline-none focus:border-cyan-500"
            >
              <option value="">📁 Cartella Principale (Radice / Root)</option>
              {selectableParents.map((g) => (
                <option key={g.id} value={g.id}>
                  📂 {g.name}
                </option>
              ))}
            </select>
          </div>

          {/* Badge Colore */}
          <div>
            <label className="block text-xs font-medium text-slate-300 mb-2">
              Colore Etichetta
            </label>
            <div className="flex flex-wrap gap-2.5">
              {COLOR_PRESETS.map((p) => {
                const isSelected = colorTag.toLowerCase() === p.value.toLowerCase();
                return (
                  <button
                    key={p.value}
                    type="button"
                    onClick={() => setColorTag(p.value)}
                    title={p.name}
                    className={`w-7 h-7 rounded-full border-2 transition-transform ${
                      isSelected
                        ? 'border-white scale-110 shadow-lg'
                        : 'border-transparent hover:scale-105 opacity-80 hover:opacity-100'
                    }`}
                    style={{ backgroundColor: p.value }}
                  />
                );
              })}
            </div>
          </div>

          {/* Actions */}
          <div className="pt-3 flex items-center justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 rounded-lg border border-cyber-border bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-medium transition-colors"
            >
              Annulla
            </button>
            <button
              type="submit"
              disabled={submitting || !name.trim()}
              className="px-4 py-2 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 text-xs font-semibold transition-all shadow-[0_0_15px_rgba(6,182,212,0.3)] disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? 'Salvataggio...' : groupToEdit ? 'Aggiorna Cartella' : 'Crea Cartella'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

