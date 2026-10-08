export interface TargetItem {
  id: string;
  tenant_id: string;
  group_id?: string;
  type: string;
  value: string;
  description?: string;
  is_blacklisted: boolean;
  created_at: string;
  updated_at: string;
}

export interface TargetListResponseData {
  items: TargetItem[];
  total: number;
  limit: number;
  offset: number;
}

