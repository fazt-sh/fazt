// Domain models matching the Fazt API

export interface User {
  username: string;
  created_at: string;
}

export interface Site {
  id: string;
  name: string;
  domain: string;
  status: 'active' | 'inactive';
  created_at: string;
  updated_at: string;
  file_count?: number;
  total_size?: number;
}

export interface AnalyticsStats {
  total_views: number;
  total_events: number;
  unique_visitors: number;
  storage_used: number;
  total_sites: number;
}

export interface Event {
  id: string;
  site_id: string;
  event_type: string;
  timestamp: string;
  metadata?: Record<string, unknown>;
}

export interface Redirect {
  id: string;
  short_code: string;
  target_url: string;
  click_count: number;
  created_at: string;
}

export interface Webhook {
  id: string;
  endpoint: string;
  method: string;
  created_at: string;
}

export interface SystemHealth {
  status: 'healthy' | 'degraded' | 'down';
  uptime: number;
  database: {
    size: number;
    connected: boolean;
  };
}

export interface SystemConfig {
  domain: string;
  port: number;
  version: string;
}
