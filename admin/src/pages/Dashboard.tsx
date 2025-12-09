import { PageHeader } from '../components/layout/PageHeader';
import { Button, Card, CardBody, Chart, Sparkline, SystemInfo } from '../components/ui';
import { Datamap } from '../components/ui/Datamap';
import { Globe, Zap, Plus, ArrowUpRight, Eye } from 'lucide-react';
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
        description={
          <span>
            Platform overview and quick actions •
            <a
              href="https://fazt-sh.github.io/fazt/docs.html"
              target="_blank"
              rel="noopener noreferrer"
              className="text-[rgb(var(--accent-mid))] hover:text-[rgb(var(--accent))] transition-colors ml-1 hover:underline"
            >
              Documentation
            </a>
          </span>
        }
        action={
          <Button variant="primary">
            <Plus className="h-4 w-4 mr-2" />
            Deploy Site
          </Button>
        }
      />

      {/* Top Row - Stats and Visitor Traffic */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-8">
        <StatCard
          icon={Globe}
          label="Sites"
          value={stats?.total_sites || 0}
          change="+2 this week"
          index={0}
        />
        <StatCard
          icon={Zap}
          label="Events"
          value={(stats?.total_events || 0).toLocaleString()}
          change="+8.3%"
          index={1}
        />

        {/* Visitor Traffic Chart - Same height as stat cards */}
        <Card variant="bordered" className="hover-lift h-full"
              style={{
                animation: 'slideIn 0.4s ease-out 0.3s backwards',
              }}>
          <CardBody className="p-6 h-full flex flex-col">
            <div className="flex items-center justify-between mb-3">
              <div>
                <h2 className="font-display text-sm text-[rgb(var(--text-primary))]">
                  Visitor Traffic
                </h2>
                <p className="text-xs text-[rgb(var(--text-secondary))] mt-0.5">
                  Last 30 days
                </p>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-[rgb(var(--success))]"></div>
                <span className="text-xs text-[rgb(var(--text-secondary))]">+417%</span>
              </div>
            </div>

            <div className="flex-1 flex items-end">
              <Chart
                data={mockMode ? mockData.visitorTraffic || [] : []}
                height={100}
                color="rgb(var(--accent-mid))"
              />
            </div>

            <div className="flex items-center justify-between text-xs mt-3">
              <span className="text-[rgb(var(--text-tertiary))]">30 days ago</span>
              <div className="flex items-center gap-2">
                <Eye className="h-3 w-3 text-[rgb(var(--text-secondary))]" />
                <span className="text-[rgb(var(--text-primary))] font-medium">
                  {mockMode ? (mockData.visitorTraffic?.[mockData.visitorTraffic.length - 1] || 0).toLocaleString() : '0'}
                </span>
              </div>
              <span className="text-[rgb(var(--text-tertiary))]">Today</span>
            </div>
          </CardBody>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* World Map - Now in its own card */}
        <Card variant="bordered" className="p-0 hover-lift"
              style={{
                animation: 'slideIn 0.4s ease-out 0.5s backwards',
              }}>
          <div className="p-6">
            <Datamap />
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