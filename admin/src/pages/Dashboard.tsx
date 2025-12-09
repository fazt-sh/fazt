import { PageHeader } from '../components/layout/PageHeader';
import { Button, Card, CardBody, Chart, Sparkline, SystemInfo } from '../components/ui';
import { Datamap } from '../components/ui/Datamap';
import { Globe, TrendingUp, Zap, Database, Plus, ArrowUpRight, Eye } from 'lucide-react';
import { useMockMode } from '../context/MockContext';
import { mockData } from '../lib/mockData';
import { DashboardSkeleton } from '../components/skeletons';

interface StatCardProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string | number;
  change?: string;
  index: number;
  sparkline?: number[];
}

function StatCard({ icon: Icon, label, value, change, index, sparkline }: StatCardProps) {
  return (
    <Card variant="bordered" className="hover-lift radial-glow"
          style={{
            animation: `slideIn 0.4s ease-out ${index * 0.1}s backwards`,
          }}>
      <CardBody className="p-6 relative overflow-hidden">
        <div className="flex items-start justify-between mb-4">
          <div className="p-2 rounded-lg bg-[rgb(var(--bg-subtle))] hover:bg-[rgb(var(--accent-glow))] transition-colors duration-300">
            <Icon className="h-5 w-5 text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--accent-mid))] transition-colors duration-300" />
          </div>
          <div className="flex items-center gap-2">
            {sparkline && (
              <Sparkline
                data={sparkline}
                width={60}
                height={24}
                color="rgb(var(--success))"
              />
            )}
            {change && (
              <span className="flex items-center gap-1 text-xs font-medium text-green-500">
                <ArrowUpRight className="h-3 w-3" />
                {change}
              </span>
            )}
          </div>
        </div>

        <div className="space-y-1">
          <p className="text-xs font-medium text-[rgb(var(--text-tertiary))] uppercase tracking-wide">
            {label}
          </p>
          <p className="font-mono text-3xl font-bold text-[rgb(var(--text-primary))] tracking-tight">
            {value}
          </p>
        </div>
      </CardBody>
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
  const { enabled: mockMode, loading } = useMockMode();
  const stats = mockMode ? mockData.stats : null;

  if (loading) {
    return <DashboardSkeleton />;
  }

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
          sparkline={mockMode ? mockData.visitorTraffic?.slice(-7) : []}
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
        {/* Visitor Traffic Chart */}
        <Card variant="bordered" className="p-6 hover-lift"
              style={{
                animation: 'slideIn 0.4s ease-out 0.5s backwards',
              }}>
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="font-display text-lg text-[rgb(var(--text-primary))]">
                Visitor Traffic
              </h2>
              <p className="text-sm text-[rgb(var(--text-secondary))] mt-1">
                Last 30 days trend
              </p>
            </div>
            <div className="flex items-center gap-2">
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 rounded-full bg-[rgb(var(--success))]"></div>
                <span className="text-xs text-[rgb(var(--text-secondary))]">+417%</span>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <Chart
              data={mockMode ? mockData.visitorTraffic || [] : []}
              height={120}
              color="rgb(var(--accent-mid))"
            />

            <div className="flex items-center justify-between text-sm">
              <span className="text-[rgb(var(--text-tertiary))]">30 days ago</span>
              <div className="flex items-center gap-3">
                <div className="flex items-center gap-1">
                  <Eye className="h-4 w-4 text-[rgb(var(--text-secondary))]" />
                  <span className="text-[rgb(var(--text-primary))] font-medium">
                    {mockMode ? (mockData.visitorTraffic?.[mockData.visitorTraffic.length - 1] || 0).toLocaleString() : '0'}
                  </span>
                </div>
              </div>
              <span className="text-[rgb(var(--text-tertiary))]">Today</span>
            </div>

            {/* World Map Section */}
            <div className="mt-6 pt-6 border-t border-[rgb(var(--border-primary))]">
              <Datamap />
            </div>
          </div>
        </Card>

        {/* System Information */}
        <SystemInfo />
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
