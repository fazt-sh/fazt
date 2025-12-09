import { PageHeader } from '../components/layout/PageHeader';
import { Button } from '../components/ui/Button';
import { Globe, TrendingUp, Zap, Database, Plus, ArrowUpRight } from 'lucide-react';
import { useMockMode } from '../context/MockContext';
import { mockData } from '../lib/mockData';

interface StatCardProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string | number;
  change?: string;
  index: number;
}

function StatCard({ icon: Icon, label, value, change, index }: StatCardProps) {
  return (
    <div
      className="group relative rounded-xl border border-[rgb(var(--border-primary))] bg-[rgb(var(--bg-elevated))] p-6
                 hover:border-[rgb(var(--accent))] transition-all duration-300 hover:shadow-[var(--shadow-md)]
                 overflow-hidden"
      style={{
        animation: `slideIn 0.4s ease-out ${index * 0.1}s backwards`,
      }}
    >
      {/* Subtle gradient overlay on hover */}
      <div className="absolute inset-0 bg-gradient-to-br from-[rgb(var(--accent-glow))] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

      <div className="relative">
        <div className="flex items-start justify-between mb-4">
          <div className="p-2 rounded-lg bg-[rgb(var(--bg-subtle))] group-hover:bg-[rgb(var(--accent-glow))] transition-colors duration-300">
            <Icon className="h-5 w-5 text-[rgb(var(--text-secondary))] group-hover:text-[rgb(var(--accent))] transition-colors duration-300" strokeWidth={2} />
          </div>
          {change && (
            <span className="flex items-center gap-1 text-xs font-medium text-green-500">
              <ArrowUpRight className="h-3 w-3" />
              {change}
            </span>
          )}
        </div>

        <div className="space-y-1">
          <p className="text-xs font-medium text-[rgb(var(--text-tertiary))] uppercase tracking-wide">
            {label}
          </p>
          <p className="font-mono text-3xl font-bold text-[rgb(var(--text-primary))] tracking-tight">
            {value}
          </p>
        </div>
      </div>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export function Dashboard() {
  const { enabled: mockMode } = useMockMode();
  const stats = mockMode ? mockData.stats : null;

  return (
    <div>
      <PageHeader
        title="Dashboard"
        description="Platform overview and quick actions"
        action={
          <Button variant="primary">
            <Plus className="h-4 w-4 mr-2" />
            Deploy Site
          </Button>
        }
      />

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard
          icon={Globe}
          label="Sites"
          value={stats?.total_sites || 0}
          change="+2 this week"
          index={0}
        />
        <StatCard
          icon={TrendingUp}
          label="Views"
          value={(stats?.total_views || 0).toLocaleString()}
          change="+12.5%"
          index={1}
        />
        <StatCard
          icon={Zap}
          label="Events"
          value={(stats?.total_events || 0).toLocaleString()}
          change="+8.3%"
          index={2}
        />
        <StatCard
          icon={Database}
          label="Storage"
          value={formatBytes(stats?.storage_used || 0)}
          index={3}
        />
      </div>

      {/* Quick Actions */}
      <div
        className="rounded-xl border border-[rgb(var(--border-primary))] bg-[rgb(var(--bg-elevated))] p-6"
        style={{
          animation: 'slideIn 0.4s ease-out 0.5s backwards',
        }}
      >
        <h2 className="font-display text-lg text-[rgb(var(--text-primary))] mb-4">
          Quick Actions
        </h2>
        <div className="flex flex-wrap gap-3">
          <Button variant="secondary">Manage Sites</Button>
          <Button variant="secondary">View Analytics</Button>
          <Button variant="secondary">Create Redirect</Button>
          <Button variant="secondary">Settings</Button>
        </div>
      </div>

      <style>{`
        @keyframes slideIn {
          from {
            opacity: 0;
            transform: translateY(10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
}
