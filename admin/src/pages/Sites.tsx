import { PageHeader } from '../components/layout/PageHeader';
import { Card, CardBody } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import { Plus, Globe, FileText, HardDrive } from 'lucide-react';
import { useMockMode } from '../context/MockContext';
import { mockData } from '../lib/mockData';
import type { Site } from '../types/models';

interface SiteCardProps {
  site: Site;
}

function SiteCard({ site }: SiteCardProps) {
  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (date: string): string => {
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  return (
    <Card variant="bordered" className="hover:shadow-md transition-shadow">
      <CardBody>
        <div className="flex items-start justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-50">
              {site.name}
            </h3>
            <div className="flex items-center gap-2 mt-1 text-sm text-gray-500 dark:text-gray-400">
              <Globe className="h-4 w-4" />
              {site.domain}
            </div>
          </div>
          <Badge variant={site.status === 'active' ? 'success' : 'default'}>
            {site.status}
          </Badge>
        </div>

        <div className="grid grid-cols-2 gap-4 mb-4">
          <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
            <FileText className="h-4 w-4" />
            <span>{site.file_count || 0} files</span>
          </div>
          <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
            <HardDrive className="h-4 w-4" />
            <span>{formatBytes(site.total_size || 0)}</span>
          </div>
        </div>

        <div className="text-xs text-gray-500 dark:text-gray-400 mb-4">
          Updated {formatDate(site.updated_at)}
        </div>

        <div className="flex gap-2">
          <Button variant="primary" size="sm" className="flex-1">
            View
          </Button>
          <Button variant="secondary" size="sm">
            Deploy
          </Button>
          <Button variant="ghost" size="sm">
            Delete
          </Button>
        </div>
      </CardBody>
    </Card>
  );
}

export function Sites() {
  const { enabled: mockMode } = useMockMode();
  const sites = mockMode ? mockData.sites : [];

  return (
    <div>
      <PageHeader
        title="Sites"
        description="Manage your hosted sites"
        action={
          <Button variant="primary">
            <Plus className="h-4 w-4 mr-2" />
            Create Site
          </Button>
        }
      />

      {sites.length === 0 ? (
        <Card variant="bordered">
          <CardBody className="text-center py-12">
            <Globe className="h-12 w-12 text-gray-400 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-50 mb-2">
              No sites yet
            </h3>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
              Get started by creating your first site
            </p>
            <Button variant="primary">
              <Plus className="h-4 w-4 mr-2" />
              Create Site
            </Button>
          </CardBody>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {sites.map((site) => (
            <SiteCard key={site.id} site={site} />
          ))}
        </div>
      )}
    </div>
  );
}
