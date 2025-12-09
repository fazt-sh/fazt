import { PageHeader } from '../components/layout/PageHeader';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Terminal, TerminalLine } from '../components/ui/Terminal';
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
    <Card variant="bordered" className="p-6 relative overflow-hidden hover-lift radial-glow"
          style={{
            animation: `slideIn 0.4s ease-out ${index * 0.1}s backwards`,
          }}>
      <div className="flex items-start justify-between mb-4">
        <div className="p-2 rounded-lg bg-[rgb(var(--bg-subtle))] hover:bg-[rgb(var(--accent-glow))] transition-colors duration-300">
          <Icon className="h-5 w-5 text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--accent-mid))] transition-colors duration-300" />
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
    </Card>
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

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Quick Actions */}
        <Card variant="bordered" className="p-6"
              style={{
                animation: 'slideIn 0.4s ease-out 0.5s backwards',
              }}>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))] mb-4">
            Quick Actions
          </h2>
          <div className="flex flex-wrap gap-3">
            <Button variant="secondary">Manage Sites</Button>
            <Button variant="secondary">View Analytics</Button>
            <Button variant="secondary">Create Redirect</Button>
            <Button variant="secondary">Settings</Button>
          </div>
        </Card>

        {/* Recent Activity Terminal */}
        <Card variant="bordered" className="p-0 overflow-hidden"
              style={{
                animation: 'slideIn 0.4s ease-out 0.6s backwards',
              }}>
          <Terminal title="Recent Activity" className="border-0 rounded-none">
            <TerminalLine type="input" prefix="$">fazt deploy my-blog --env prod</TerminalLine>
            <TerminalLine type="success">✓ Deployed successfully (42 files, 2.1MB)</TerminalLine>
            <TerminalLine type="output">─</TerminalLine>
            <TerminalLine type="input" prefix="$">fazt logs my-blog --tail 10</TerminalLine>
            <TerminalLine type="output">[2025-12-09 12:00:01] GET / 200 - 1.2ms</TerminalLine>
            <TerminalLine type="output">[2025-12-09 12:00:05] GET /api/posts 200 - 3.4ms</TerminalLine>
            <TerminalLine type="error">[2025-12-09 12:00:12] POST /api/contact 400 - 0.8ms</TerminalLine>
            <TerminalLine type="output">[2025-12-09 12:00:23] GET /static/style.css 304 - 0.3ms</TerminalLine>
          </Terminal>
        </Card>
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
