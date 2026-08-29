export interface CommandItem {
  id: string;
  kind: string;
  title: string;
  message: string;
  severity: string;
  status: string;
  project_id?: string;
  project_name?: string;
  pic_user_id?: string;
  due_at?: string;
  source_type?: string;
  source_id?: string;
  aging_days?: number;
}

export interface CommandSummary {
  as_of: string;
  alerts: CommandItem[];
  actions: CommandItem[];
  validations: CommandItem[];
  watchlist: CommandItem[];
  escalations: CommandItem[];
  decisions: CommandItem[];
}