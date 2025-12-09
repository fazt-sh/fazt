import { PageHeader } from '../components/layout/PageHeader';
import { Card, CardBody } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Globe, Eye, Activity, HardDrive, Plus } from 'lucide-react';
import { useMockMode } from '../context/MockContext';
import { mockData } from '../lib/mockData';

interface StatCardProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string | number;
  description?: string;
}

function StatCard({ icon: Icon, label, value, description }: StatCardProps) {
  return (
    <Card variant="bordered">
      <CardBody>
        <div className="flex items-center gap-4">
          <div className="p-3 rounded-lg bg-primary/10">
            <Icon className="h-6 w-6 text-primary" />
          </div>
          <div className="flex-1">
            <p className="text-sm text-gray-500 dark:text-gray-400">{label}</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-50 mt-1">
              {value}
            </p>
            {description && (
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                {description}
              </p>
            )}
          </div>
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
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function Dashboard() {
  const { enabled: mockMode } = useMockMode();
  const stats = mockMode ? mockData.stats : null;

  return (
    <div>
      <PageHeader
        title="Dashboard"
        description="Overview of your Fazt platform"
        action={
          <Button variant="primary">
            <Plus className="h-4 w-4 mr-2" />
            Create Site
          </Button>
        }
      />

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard
          icon={Globe}
          label="Total Sites"
          value={stats?.total_sites || 0}
        />
        <StatCard
          icon={Eye}
          label="Total Views"
          value={stats?.total_views.toLocaleString() || 0}
          description="Last 30 days"
        />
        <StatCard
          icon={Activity}
          label="Total Events"
          value={stats?.total_events.toLocaleString() || 0}
          description="Last 30 days"
        />
        <StatCard
          icon={HardDrive}
          label="Storage Used"
          value={formatBytes(stats?.storage_used || 0)}
        />
      </div>

      {/* Quick Actions */}
      <Card variant="bordered">
        <CardBody>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-50 mb-4">
            Quick Actions
          </h2>
          <div className="flex flex-wrap gap-3">
            <Button variant="secondary">View Sites</Button>
            <Button variant="secondary">View Analytics</Button>
            <Button variant="secondary">Create Redirect</Button>
            <Button variant="secondary">System Settings</Button>
          </div>
        </CardBody>
      </Card>
    </div>
  );
}
