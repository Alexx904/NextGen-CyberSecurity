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

export interface TargetGroup {
  id: string;
  tenant_id: string;
  parent_id?: string | null;
  name: string;
  description?: string | null;
  color_tag?: string | null;
  created_at: string;
  updated_at: string;
}

export interface TargetGroupNode extends TargetGroup {
  children: TargetGroupNode[];
  direct_target_count: number;
  target_count: number;
  aggregate_risk: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'INFO' | 'NONE';
  severity_counts: Record<string, number>;
}

export interface CreateTargetGroupInput {
  name: string;
  description?: string;
  parent_id?: string | null;
  color_tag?: string;
}

export interface UpdateTargetGroupInput {
  name: string;
  description?: string;
  parent_id?: string | null;
  color_tag?: string;
}


