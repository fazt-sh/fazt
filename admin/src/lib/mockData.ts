import type { User, Site, AnalyticsStats, SystemHealth, SystemConfig } from '../types/models';

export const mockData = {
  user: {
    username: 'admin',
    created_at: '2024-01-15T10:30:00Z',
  } as User,

  sites: [
    {
      id: 'site_1',
      name: 'my-blog',
      domain: 'blog.example.com',
      status: 'active' as const,
      created_at: '2024-01-15T10:30:00Z',
      updated_at: '2024-12-09T10:00:00Z',
      file_count: 42,
      total_size: 2048000,
    },
    {
      id: 'site_2',
      name: 'portfolio',
      domain: 'portfolio.example.com',
      status: 'active' as const,
      created_at: '2024-02-20T14:20:00Z',
      updated_at: '2024-12-08T15:30:00Z',
      file_count: 28,
      total_size: 1536000,
    },
    {
      id: 'site_3',
      name: 'docs',
      domain: 'docs.example.com',
      status: 'active' as const,
      created_at: '2024-03-10T09:15:00Z',
      updated_at: '2024-12-07T11:45:00Z',
      file_count: 156,
      total_size: 5120000,
    },
  ] as Site[],

  stats: {
    total_views: 12543,
    total_events: 8932,
    unique_visitors: 3421,
    storage_used: 8704000, // ~8.7 MB
    total_sites: 3,
  } as AnalyticsStats,

  systemHealth: {
    status: 'healthy' as const,
    uptime: 345600, // 4 days in seconds
    database: {
      size: 10485760, // 10 MB
      connected: true,
    },
  } as SystemHealth,

  systemConfig: {
    domain: 'localhost',
    port: 8080,
    version: 'v0.8.0-dev',
  } as SystemConfig,
};
