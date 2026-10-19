export interface ScheduleItem {
  authorization_required?: boolean;
  id: string;
  tenant_id: string;
  target_id: string;
  profile_id: string;
  cron_expression: string;
  is_active: boolean;
  next_run_at?: string;
  last_run_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ScheduleListResponseData {
  items: ScheduleItem[];
  total: number;
}
