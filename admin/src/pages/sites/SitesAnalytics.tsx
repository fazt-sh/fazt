import { useEffect, useState } from 'react';
import { BarChart3, TrendingUp, Globe, Tag, Calendar, Activity, RefreshCw } from 'lucide-react';
import { PageHeader } from '../../components/layout/PageHeader';
import { Breadcrumbs } from '../../components/ui/Breadcrumbs';
import { Card, CardBody, Button, Chart, Spinner } from '../../components/ui';
import { api } from '../../lib/api';
import type { StatsResponse } from '../../types/models';

interface StatCardProps {
  label: string;
  value: number;
  icon: React.ComponentType<{ className?: string }>;
  description?: string;
}

function StatCard({ label, value, icon: Icon, description }: StatCardProps) {
  return (
    <Card variant="bordered" className="hover-lift">
      <CardBody className="p-5">
        <div className="flex items-center justify-between mb-3">
          <div className="p-2 rounded-lg bg-[rgb(var(--bg-subtle))]">
            <Icon className="h-5 w-5 text-[rgb(var(--accent-mid))]" />
          </div>
        </div>
        <div>
          <p className="text-xs font-medium text-[rgb(var(--text-tertiary))] uppercase tracking-wide mb-1">
            {label}
          </p>
          <p className="font-mono text-2xl font-bold text-[rgb(var(--text-primary))]">
            {value.toLocaleString()}
          </p>
          {description && (
            <p className="text-xs text-[rgb(var(--text-secondary))] mt-1">
              {description}
            </p>
          )}
        </div>
      </CardBody>
    </Card>
  );
}

function TopItemsList({
  title,
  items,
  icon: Icon,
  labelKey,
  valueKey
}: {
  title: string;
  items: Array<Record<string, unknown>>;
  icon: React.ComponentType<{ className?: string }>;
  labelKey: string;
  valueKey: string;
}) {
  const maxValue = Math.max(...items.map(i => Number(i[valueKey]) || 0), 1);

  return (
    <Card variant="bordered" className="hover-lift h-full">
      <CardBody className="p-5">
        <div className="flex items-center gap-2 mb-4">
          <Icon className="h-4 w-4 text-[rgb(var(--accent-mid))]" />
          <h3 className="font-medium text-[rgb(var(--text-primary))]">{title}</h3>
        </div>
        {items.length === 0 ? (
          <p className="text-sm text-[rgb(var(--text-secondary))]">No data yet</p>
        ) : (
          <div className="space-y-3">
            {items.slice(0, 8).map((item, index) => {
              const label = String(item[labelKey] || 'Unknown');
              const value = Number(item[valueKey] || 0);
              const percentage = (value / maxValue) * 100;

              return (
                <div key={index} className="space-y-1">
                  <div className="flex justify-between text-sm">
                    <span className="text-[rgb(var(--text-primary))] truncate max-w-[180px]" title={label}>
                      {label}
                    </span>
                    <span className="text-[rgb(var(--text-secondary))] font-mono">
                      {value.toLocaleString()}
                    </span>
                  </div>
                  <div className="h-1.5 bg-[rgb(var(--bg-subtle))] rounded-full overflow-hidden">
                    <div
                      className="h-full bg-[rgb(var(--accent-mid))] rounded-full transition-all duration-500"
                      style={{ width: `${percentage}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardBody>
    </Card>
  );
}

function SourceTypeChart({ data }: { data: Record<string, number> }) {
  const entries = Object.entries(data).sort((a, b) => b[1] - a[1]);
  const total = entries.reduce((sum, [, count]) => sum + count, 0) || 1;
  const colors = [
    'rgb(var(--accent-mid))',
    'rgb(var(--success))',
    'rgb(var(--warning))',
    'rgb(var(--error))',
    'rgb(var(--text-secondary))',
  ];

  return (
    <Card variant="bordered" className="hover-lift h-full">
      <CardBody className="p-5">
        <div className="flex items-center gap-2 mb-4">
          <Activity className="h-4 w-4 text-[rgb(var(--accent-mid))]" />
          <h3 className="font-medium text-[rgb(var(--text-primary))]">Events by Source</h3>
        </div>
        {entries.length === 0 ? (
          <p className="text-sm text-[rgb(var(--text-secondary))]">No data yet</p>
        ) : (
          <>
            <div className="flex h-4 rounded-full overflow-hidden mb-4">
              {entries.map(([, count], index) => (
                <div
                  key={index}
                  className="h-full transition-all duration-500"
                  style={{
                    width: `${(count / total) * 100}%`,
                    backgroundColor: colors[index % colors.length]
                  }}
                />
              ))}
            </div>
            <div className="space-y-2">
              {entries.map(([source, count], index) => (
                <div key={source} className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <div
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: colors[index % colors.length] }}
                    />
                    <span className="text-[rgb(var(--text-primary))]">{source || 'direct'}</span>
                  </div>
                  <div className="text-[rgb(var(--text-secondary))] font-mono">
                    {count.toLocaleString()} ({Math.round((count / total) * 100)}%)
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </CardBody>
    </Card>
  );
}

export function SitesAnalytics() {
  const [stats, setStats] = useState<StatsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.get<StatsResponse>('/stats');
      setStats(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch stats');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
    // Refresh every 60 seconds
    const interval = setInterval(fetchStats, 60000);
    return () => clearInterval(interval);
  }, []);

  // Convert timeline data to chart format
  const timelineData = stats?.events_timeline?.map(t => t.count) || [];

  return (
    <div>
      <Breadcrumbs />
      <PageHeader
        title="Analytics"
        description="View event tracking and site statistics"
        action={
          <Button variant="secondary" onClick={fetchStats} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

      {error && (
        <Card variant="bordered" className="mb-6 border-[rgb(var(--error))]">
          <CardBody className="p-4">
            <p className="text-[rgb(var(--error))]">{error}</p>
          </CardBody>
        </Card>
      )}

      {loading && !stats ? (
        <div className="flex items-center justify-center py-12">
          <Spinner size="lg" />
        </div>
      ) : stats ? (
        <>
          {/* Stats Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <StatCard
              label="Today"
              value={stats.total_events_today}
              icon={Calendar}
              description="Events today"
            />
            <StatCard
              label="This Week"
              value={stats.total_events_week}
              icon={TrendingUp}
              description="Last 7 days"
            />
            <StatCard
              label="This Month"
              value={stats.total_events_month}
              icon={BarChart3}
              description="Last 30 days"
            />
            <StatCard
              label="All Time"
              value={stats.total_events_all_time}
              icon={Activity}
              description={`${stats.total_unique_domains} unique domains`}
            />
          </div>

          {/* Timeline Chart */}
          <Card variant="bordered" className="mb-6 hover-lift">
            <CardBody className="p-5">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h3 className="font-medium text-[rgb(var(--text-primary))]">Events Timeline</h3>
                  <p className="text-sm text-[rgb(var(--text-secondary))]">Last 24 hours</p>
                </div>
              </div>
              {timelineData.length > 0 ? (
                <Chart
                  data={timelineData}
                  height={180}
                  color="rgb(var(--accent-mid))"
                />
              ) : (
                <div className="h-[180px] flex items-center justify-center text-[rgb(var(--text-secondary))]">
                  No timeline data available
                </div>
              )}
            </CardBody>
          </Card>

          {/* Bottom Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <TopItemsList
              title="Top Domains"
              items={stats.top_domains as unknown as Array<Record<string, unknown>>}
              icon={Globe}
              labelKey="domain"
              valueKey="count"
            />
            <TopItemsList
              title="Top Tags"
              items={stats.top_tags as unknown as Array<Record<string, unknown>>}
              icon={Tag}
              labelKey="tag"
              valueKey="count"
            />
            <SourceTypeChart data={stats.events_by_source_type} />
          </div>

          {/* Additional Stats */}
          {stats.total_redirect_clicks > 0 && (
            <Card variant="bordered" className="mt-6 hover-lift">
              <CardBody className="p-5">
                <div className="flex items-center gap-4">
                  <div className="p-3 rounded-lg bg-[rgb(var(--bg-subtle))]">
                    <TrendingUp className="h-6 w-6 text-[rgb(var(--accent-mid))]" />
                  </div>
                  <div>
                    <p className="text-sm text-[rgb(var(--text-secondary))]">Total Redirect Clicks</p>
                    <p className="font-mono text-2xl font-bold text-[rgb(var(--text-primary))]">
                      {stats.total_redirect_clicks.toLocaleString()}
                    </p>
                  </div>
                </div>
              </CardBody>
            </Card>
          )}
        </>
      ) : null}
    </div>
  );
}
