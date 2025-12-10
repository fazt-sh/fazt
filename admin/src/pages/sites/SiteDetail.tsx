import { useParams, useNavigate } from 'react-router-dom';
import { PageHeader } from '../../components/layout/PageHeader';
import { Button, Card, CardBody, Badge, Tabs } from '../../components/ui';
import { ArrowLeft, Globe, FileText, HardDrive, Terminal, Key, RefreshCw } from 'lucide-react';
import { useMockMode } from '../../context/MockContext';
import { mockData } from '../../lib/mockData';
import { NotFound } from '../NotFound';

export function SiteDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { enabled: mockMode } = useMockMode();
  
  const site = mockMode ? mockData.sites.find(s => s.id === id) : null;

  if (!site) {
    return <NotFound />;
  }

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const OverviewTab = () => (
    <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card variant="bordered">
                <CardBody>
                    <h3 className="text-sm font-medium text-[rgb(var(--text-secondary))] mb-2">Status</h3>
                    <Badge variant={site.status === 'active' ? 'success' : 'default'}>{site.status}</Badge>
                </CardBody>
            </Card>
             <Card variant="bordered">
                <CardBody>
                    <h3 className="text-sm font-medium text-[rgb(var(--text-secondary))] mb-2">Files</h3>
                    <div className="flex items-center gap-2">
                        <FileText className="w-5 h-5 text-[rgb(var(--text-tertiary))]" />
                        <span className="text-xl font-mono font-semibold">{site.file_count || 0}</span>
                    </div>
                </CardBody>
            </Card>
             <Card variant="bordered">
                <CardBody>
                    <h3 className="text-sm font-medium text-[rgb(var(--text-secondary))] mb-2">Total Size</h3>
                    <div className="flex items-center gap-2">
                        <HardDrive className="w-5 h-5 text-[rgb(var(--text-tertiary))]" />
                        <span className="text-xl font-mono font-semibold">{formatBytes(site.total_size || 0)}</span>
                    </div>
                </CardBody>
            </Card>
        </div>

        <Card variant="bordered">
            <CardBody>
                <h3 className="font-medium text-[rgb(var(--text-primary))] mb-4">Deployment Info</h3>
                <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
                    <div>
                        <dt className="text-[rgb(var(--text-secondary))]">Domain</dt>
                        <dd className="font-mono mt-1 text-[rgb(var(--text-primary))]">{site.domain}</dd>
                    </div>
                    <div>
                        <dt className="text-[rgb(var(--text-secondary))]">Site ID</dt>
                        <dd className="font-mono mt-1 text-[rgb(var(--text-primary))]">{site.id}</dd>
                    </div>
                    <div>
                        <dt className="text-[rgb(var(--text-secondary))]">Created At</dt>
                        <dd className="mt-1 text-[rgb(var(--text-primary))]">{new Date(site.created_at).toLocaleString()}</dd>
                    </div>
                    <div>
                        <dt className="text-[rgb(var(--text-secondary))]">Last Updated</dt>
                        <dd className="mt-1 text-[rgb(var(--text-primary))]">{new Date(site.updated_at).toLocaleString()}</dd>
                    </div>
                </dl>
            </CardBody>
        </Card>
    </div>
  );

  const EnvTab = () => (
      <Card variant="bordered">
          <CardBody className="py-12 text-center">
              <div className="w-12 h-12 bg-[rgb(var(--bg-subtle))] rounded-full flex items-center justify-center mx-auto mb-4">
                  <Terminal className="w-6 h-6 text-[rgb(var(--text-tertiary))]" />
              </div>
              <h3 className="text-lg font-medium text-[rgb(var(--text-primary))]">Environment Variables</h3>
              <p className="text-[rgb(var(--text-secondary))] mt-1 mb-6">Manage environment variables for your function deployments.</p>
              <Button variant="secondary">Add Variable</Button>
          </CardBody>
      </Card>
  );

  const KeysTab = () => (
    <Card variant="bordered">
        <CardBody className="py-12 text-center">
            <div className="w-12 h-12 bg-[rgb(var(--bg-subtle))] rounded-full flex items-center justify-center mx-auto mb-4">
                <Key className="w-6 h-6 text-[rgb(var(--text-tertiary))]" />
            </div>
            <h3 className="text-lg font-medium text-[rgb(var(--text-primary))]">API Keys</h3>
            <p className="text-[rgb(var(--text-secondary))] mt-1 mb-6">Manage API keys for accessing your site programmatically.</p>
            <Button variant="secondary">Create Key</Button>
        </CardBody>
    </Card>
  );

  return (
    <div className="animate-fade-in">
      <div className="mb-6">
        <button 
            onClick={() => navigate('/sites')}
            className="flex items-center text-sm text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] transition-colors mb-2"
        >
            <ArrowLeft className="w-4 h-4 mr-1" />
            Back to Sites
        </button>
        <PageHeader
            title={site.name}
            description={site.domain}
            action={
                <div className="flex gap-2">
                    <Button variant="secondary" onClick={() => window.open(`http://${site.domain}`, '_blank')}>
                        <Globe className="w-4 h-4 mr-2" />
                        Visit Site
                    </Button>
                    <Button variant="primary">
                        <RefreshCw className="w-4 h-4 mr-2" />
                        Redeploy
                    </Button>
                </div>
            }
        />
      </div>

      <Tabs 
        tabs={[
            { id: 'overview', label: 'Overview', content: <OverviewTab /> },
            { id: 'env', label: 'Environment', content: <EnvTab /> },
            { id: 'keys', label: 'API Keys', content: <KeysTab /> },
            { id: 'logs', label: 'Logs', content: <div className="p-8 text-center text-[rgb(var(--text-secondary))]">Logs coming soon</div> },
            { id: 'settings', label: 'Settings', content: <div className="p-8 text-center text-[rgb(var(--text-secondary))]">Settings coming soon</div> },
        ]}
      />
    </div>
  );
}
