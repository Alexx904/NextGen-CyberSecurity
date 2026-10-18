import type { ReportFilterParams, ReportItem, ReportListResponseData } from '../types/report';

export async function loadReportArchive(
  list: (params: ReportFilterParams) => Promise<ReportListResponseData>,
  pageSize: number,
  isCurrent: () => boolean,
): Promise<ReportItem[]> {
  const items: ReportItem[] = [];
  const seen = new Set<string>();
  let total: number | undefined;
  do {
    const page = await list({ limit: pageSize, offset: items.length });
    if (!isCurrent()) return [];
    if (!Number.isSafeInteger(page.total) || page.total < 0) throw new Error('Conteggio dello storico non valido.');
    if (total !== undefined && page.total !== total) throw new Error('Lo storico è cambiato durante il caricamento. Aggiorna per riprovare.');
    total = page.total;
    if (!Array.isArray(page.items) || (!page.items.length && items.length < total)) {
      throw new Error('Storico incompleto: impossibile caricare tutte le pagine. Aggiorna per riprovare.');
    }
    for (const item of page.items) {
      if (seen.has(item.scan_id)) throw new Error('Storico cambiato o pagina duplicata. Aggiorna per riprovare.');
      seen.add(item.scan_id);
      items.push(item);
    }
  } while (items.length < total);
  return items;
}

export interface ArchiveFilters {
  dateFrom: string;
  dateTo: string;
  profile: string;
  target: string;
  group: string;
  severity: string;
  search: string;
}

export function filterReportArchive(items: ReportItem[], filters: ArchiveFilters): ReportItem[] {
  if (filters.dateFrom && filters.dateTo && filters.dateFrom > filters.dateTo) return [];
  const query = filters.search.trim().toLowerCase();
  return items.filter(item => {
    const observedAt = new Date(item.completed_at || item.created_at);
    if (Number.isNaN(observedAt.getTime())) return false;
    const day = observedAt.toISOString().slice(0, 10);
    return (!filters.dateFrom || day >= filters.dateFrom) && (!filters.dateTo || day <= filters.dateTo)
      && (!filters.profile || item.profile_id === filters.profile)
      && (!filters.target || item.target_id === filters.target)
      && (!filters.group || item.group_id === filters.group)
      && (!filters.severity || item.max_severity === filters.severity)
      && (!query || [item.target_value, item.group_name, item.profile_name].some(value => value?.toLowerCase().includes(query)));
  });
}
