import React, { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react';
import { apiClient } from '../api/client';
import { ApiResponse } from '../types/auth';
import { TargetGroupNode, TargetItem, TargetListResponseData } from '../types/target';
import { ScanItem } from '../types/scan';
import { FindingItem } from '../types/finding';

const STORAGE_KEY = 'nextgen_active_company_id';

// Ricava ricorsivamente tutti gli ID di un nodo e dei suoi discendenti (reparti)
export function getAllDescendantGroupIds(node: TargetGroupNode): string[] {
  const ids: string[] = [node.id];
  if (node.children && node.children.length > 0) {
    for (const child of node.children) {
      ids.push(...getAllDescendantGroupIds(child));
    }
  }
  return ids;
}

// Appiattisce ricorsivamente tutti i reparti (figli) di un'azienda
export function flattenDepartments(node: TargetGroupNode): TargetGroupNode[] {
  const result: TargetGroupNode[] = [];
  if (node.children && node.children.length > 0) {
    for (const child of node.children) {
      result.push(child);
      result.push(...flattenDepartments(child));
    }
  }
  return result;
}

export interface CompanyContextType {
  // Azienda attiva selezionata (null = "Tutte le Aziende")
  selectedCompanyId: string | null;
  selectedCompany: TargetGroupNode | null;
  setSelectedCompanyId: (id: string | null) => void;

  // Dati
  companies: TargetGroupNode[]; // Gruppi radice (Aziende)
  allGroupNodes: TargetGroupNode[]; // Albero completo dei gruppi
  departments: TargetGroupNode[]; // Reparti dell'azienda selezionata (appiattiti)
  allTargets: TargetItem[]; // Tutti i bersagli caricati
  loading: boolean;
  error: string | null;

  // Set di ID e valori dello scope attivo
  companyGroupIds: Set<string>; // ID dell'azienda + tutti i suoi reparti
  companyTargetIds: Set<string>; // ID dei target nello scope
  companyTargetValues: Set<string>; // IP, CIDR, hostname dei target nello scope

  // Azioni CRUD
  refreshCompanies: () => Promise<void>;
  createCompany: (input: { name: string; description?: string; color_tag?: string }) => Promise<TargetGroupNode>;
  updateCompany: (id: string, input: { name: string; description?: string; color_tag?: string }) => Promise<void>;
  deleteCompany: (id: string) => Promise<void>;
  createDepartment: (input: { name: string; description?: string; parent_id: string; color_tag?: string }) => Promise<void>;

  // Helper di filtraggio unificati
  filterTargets: (targets: TargetItem[]) => TargetItem[];
  filterScans: (scans: ScanItem[]) => ScanItem[];
  filterFindings: (findings: FindingItem[]) => FindingItem[];
  isTargetInActiveScope: (targetIdOrValue: string) => boolean;
}

const CompanyContext = createContext<CompanyContextType | undefined>(undefined);

export const CompanyProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [selectedCompanyId, setSelectedCompanyIdState] = useState<string | null>(() => {
    return localStorage.getItem(STORAGE_KEY) || null;
  });

  const [allGroupNodes, setAllGroupNodes] = useState<TargetGroupNode[]>([]);
  const [allTargets, setAllTargets] = useState<TargetItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Sincronizzazione persistente in localStorage
  const setSelectedCompanyId = useCallback((id: string | null) => {
    setSelectedCompanyIdState(id);
    if (id) {
      localStorage.setItem(STORAGE_KEY, id);
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  }, []);

  // Ricaricamento gruppi e target da API
  const refreshCompanies = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [groupsRes, targetsRes] = await Promise.all([
        apiClient.get<ApiResponse<TargetGroupNode[]>>('/target-groups'),
        apiClient.get<ApiResponse<TargetListResponseData>>('/targets?limit=500'),
      ]);

      if (groupsRes.data?.data) {
        setAllGroupNodes(groupsRes.data.data);
      }
      if (targetsRes.data?.data?.items) {
        const items = [...targetsRes.data.data.items];
        const total = targetsRes.data.data.total || items.length;
        while (items.length < total) {
          const page = await apiClient.get<ApiResponse<TargetListResponseData>>(`/targets?limit=500&offset=${items.length}`);
          const more = page.data?.data?.items || [];
          if (more.length === 0) throw new Error('Inventario target incompleto: riprova il caricamento');
          items.push(...more);
        }
        setAllTargets(items);
      }
    } catch (err: any) {
      setError(err.response?.data?.error?.message || err.message || 'Errore nel caricamento delle aziende');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshCompanies();
  }, [refreshCompanies]);

  // Le aziende sono i gruppi di livello radice
  const companies = useMemo(() => {
    return allGroupNodes.filter((node) => !node.parent_id);
  }, [allGroupNodes]);

  // Verifica se l'ID selezionato esiste ancora tra le aziende
  useEffect(() => {
    if (!loading && !error && selectedCompanyId) {
      const found = companies.some((c) => c.id === selectedCompanyId);
      if (!found) {
        setSelectedCompanyId(null);
      }
    }
  }, [companies, selectedCompanyId, setSelectedCompanyId, loading, error]);

  // L'oggetto azienda selezionato
  const selectedCompany = useMemo(() => {
    if (!selectedCompanyId) return null;
    return companies.find((c) => c.id === selectedCompanyId) || null;
  }, [companies, selectedCompanyId]);

  // Reparti dell'azienda selezionata (appiattiti per facilità di utilizzo)
  const departments = useMemo(() => {
    if (!selectedCompany) return [];
    return flattenDepartments(selectedCompany);
  }, [selectedCompany]);

  // Insieme di tutti gli ID di gruppo appartenenti all'azienda selezionata (azienda + reparti)
  const companyGroupIds = useMemo(() => {
    if (!selectedCompany) return new Set<string>();
    const ids = getAllDescendantGroupIds(selectedCompany);
    return new Set(ids);
  }, [selectedCompany]);

  // Insieme dei target ID e IP/valori associati all'azienda selezionata
  const { companyTargetIds, companyTargetValues } = useMemo(() => {
    if (!selectedCompanyId) {
      return {
        companyTargetIds: new Set<string>(),
        companyTargetValues: new Set<string>(),
      };
    }

    const targetIds = new Set<string>();
    const targetValues = new Set<string>();

    allTargets.forEach((t) => {
      if (t.group_id && companyGroupIds.has(t.group_id)) {
        targetIds.add(t.id);
        targetValues.add(t.value);
      }
    });

    return {
      companyTargetIds: targetIds,
      companyTargetValues: targetValues,
    };
  }, [selectedCompanyId, allTargets, companyGroupIds]);

  // Azioni CRUD
  const createCompany = async (input: { name: string; description?: string; color_tag?: string }): Promise<TargetGroupNode> => {
    const res = await apiClient.post<ApiResponse<TargetGroupNode>>('/target-groups', {
      name: input.name,
      description: input.description,
      parent_id: null,
      color_tag: input.color_tag || '#3b82f6',
    });
    if (res.data?.success && res.data?.data) {
      await refreshCompanies();
      return res.data.data;
    }
    throw new Error(res.data?.error?.message || 'Impossibile creare l\'azienda');
  };

  const updateCompany = async (id: string, input: { name: string; description?: string; color_tag?: string }) => {
    const res = await apiClient.put<ApiResponse<any>>(`/target-groups/${id}`, {
      name: input.name,
      description: input.description,
      parent_id: null,
      color_tag: input.color_tag,
    });
    if (!res.data?.success) {
      throw new Error(res.data?.error?.message || 'Impossibile aggiornare l\'azienda');
    }
    await refreshCompanies();
  };

  const deleteCompany = async (id: string) => {
    const res = await apiClient.delete<ApiResponse<any>>(`/target-groups/${id}`);
    if (!res.data?.success) {
      throw new Error(res.data?.error?.message || 'Impossibile eliminare l\'azienda');
    }
    if (selectedCompanyId === id) {
      setSelectedCompanyId(null);
    }
    await refreshCompanies();
  };

  const createDepartment = async (input: { name: string; description?: string; parent_id: string; color_tag?: string }) => {
    const res = await apiClient.post<ApiResponse<any>>('/target-groups', {
      name: input.name,
      description: input.description,
      parent_id: input.parent_id,
      color_tag: input.color_tag || selectedCompany?.color_tag || '#10b981',
    });
    if (!res.data?.success) {
      throw new Error(res.data?.error?.message || 'Impossibile creare il reparto');
    }
    await refreshCompanies();
  };

  // Funzioni di filtraggio globale
  const filterTargets = useCallback((targets: TargetItem[]): TargetItem[] => {
    if (!selectedCompanyId) return targets;
    return targets.filter((t) => t.group_id && companyGroupIds.has(t.group_id));
  }, [selectedCompanyId, companyGroupIds]);

  const filterScans = useCallback((scans: ScanItem[]): ScanItem[] => {
    if (!selectedCompanyId) return scans;
    return scans.filter((s) => companyTargetIds.has(s.target_id));
  }, [selectedCompanyId, companyTargetIds]);

  const filterFindings = useCallback((findings: FindingItem[]): FindingItem[] => {
    if (!selectedCompanyId) return findings;
    return findings.filter((f) => companyTargetValues.has(f.host_ip));
  }, [selectedCompanyId, companyTargetValues]);

  const isTargetInActiveScope = useCallback((targetIdOrValue: string): boolean => {
    if (!selectedCompanyId) return true;
    return companyTargetIds.has(targetIdOrValue) || companyTargetValues.has(targetIdOrValue);
  }, [selectedCompanyId, companyTargetIds, companyTargetValues]);

  const value: CompanyContextType = {
    selectedCompanyId,
    selectedCompany,
    setSelectedCompanyId,
    companies,
    allGroupNodes,
    departments,
    allTargets,
    loading,
    error,
    companyGroupIds,
    companyTargetIds,
    companyTargetValues,
    refreshCompanies,
    createCompany,
    updateCompany,
    deleteCompany,
    createDepartment,
    filterTargets,
    filterScans,
    filterFindings,
    isTargetInActiveScope,
  };

  return <CompanyContext.Provider value={value}>{children}</CompanyContext.Provider>;
};

const DEFAULT_COMPANY_CONTEXT: CompanyContextType = {
  selectedCompanyId: null,
  selectedCompany: null,
  setSelectedCompanyId: () => {},
  companies: [],
  allGroupNodes: [],
  departments: [],
  allTargets: [],
  loading: false,
  error: null,
  companyGroupIds: new Set<string>(),
  companyTargetIds: new Set<string>(),
  companyTargetValues: new Set<string>(),
  refreshCompanies: async () => {},
  createCompany: async () => ({ id: '', name: '', depth: 0 } as unknown as TargetGroupNode),
  updateCompany: async () => {},
  deleteCompany: async () => {},
  createDepartment: async () => {},
  filterTargets: (targets) => targets,
  filterScans: (scans) => scans,
  filterFindings: (findings) => findings,
  isTargetInActiveScope: () => true,
};

export const useCompany = (): CompanyContextType => {
  const context = useContext(CompanyContext);
  if (!context) {
    return DEFAULT_COMPANY_CONTEXT;
  }
  return context;
};
