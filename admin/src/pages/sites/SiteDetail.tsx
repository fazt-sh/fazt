import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { PageHeader } from '../../components/layout/PageHeader';
import { Button, Card, CardBody, Badge, Tabs, Modal, Input, Terminal, TerminalLine } from '../../components/ui';
import {
  ArrowLeft, Globe, FileText, HardDrive, Terminal as TerminalIcon, Key, RefreshCw,
  Plus, Trash2, Eye, EyeOff, Copy, Check, AlertTriangle, Link2, Settings
} from 'lucide-react';
import { useMockMode } from '../../context/MockContext';
import { useToast } from '../../context/ToastContext';
import { mockData } from '../../lib/mockData';
import { NotFound } from '../NotFound';

// Mock data for tabs
const mockEnvVars = [
  { id: 'env_1', key: 'DATABASE_URL', value: 'postgres://user:pass@localhost:5432/db', created_at: '2024-12-01T10:00:00Z' },
  { id: 'env_2', key: 'API_SECRET', value: 'sk_live_abc123xyz789', created_at: '2024-12-02T14:00:00Z' },
  { id: 'env_3', key: 'NODE_ENV', value: 'production', created_at: '2024-12-03T09:00:00Z' },
];

const mockApiKeys: Array<{ id: string; name: string; prefix: string; created_at: string; last_used: string | null }> = [
  { id: 'key_1', name: 'Production Deploy', prefix: 'fzt_live_', created_at: '2024-11-15T10:00:00Z', last_used: '2024-12-09T08:30:00Z' },
  { id: 'key_2', name: 'CI/CD Pipeline', prefix: 'fzt_live_', created_at: '2024-11-20T14:00:00Z', last_used: '2024-12-08T16:45:00Z' },
];

const mockLogs = [
  { timestamp: '2024-12-09T10:30:15Z', level: 'info', message: 'Deployment started' },
  { timestamp: '2024-12-09T10:30:16Z', level: 'info', message: 'Uploading 42 files...' },
  { timestamp: '2024-12-09T10:30:18Z', level: 'success', message: 'Files uploaded successfully' },
  { timestamp: '2024-12-09T10:30:19Z', level: 'info', message: 'Invalidating CDN cache...' },
  { timestamp: '2024-12-09T10:30:21Z', level: 'success', message: 'Deployment complete' },
  { timestamp: '2024-12-09T10:30:21Z', level: 'info', message: 'Site is live at https://blog.example.com' },
  { timestamp: '2024-12-09T11:15:03Z', level: 'info', message: 'Incoming request: GET /index.html' },
  { timestamp: '2024-12-09T11:15:03Z', level: 'info', message: 'Served 200 OK (12ms)' },
  { timestamp: '2024-12-09T11:20:45Z', level: 'error', message: 'Failed to load resource: /api/data.json - 404 Not Found' },
  { timestamp: '2024-12-09T11:25:00Z', level: 'info', message: 'Incoming request: GET /about.html' },
];

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

  const EnvTab = () => {
    const { success } = useToast();
    const [envVars, setEnvVars] = useState(mockEnvVars);
    const [visibleValues, setVisibleValues] = useState<Set<string>>(new Set());
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [newKey, setNewKey] = useState('');
    const [newValue, setNewValue] = useState('');

    const toggleVisibility = (id: string) => {
      setVisibleValues(prev => {
        const next = new Set(prev);
        if (next.has(id)) {
          next.delete(id);
        } else {
          next.add(id);
        }
        return next;
      });
    };

    const handleAdd = () => {
      if (newKey && newValue) {
        const newEnv = {
          id: `env_${Date.now()}`,
          key: newKey.toUpperCase().replace(/\s+/g, '_'),
          value: newValue,
          created_at: new Date().toISOString(),
        };
        setEnvVars(prev => [...prev, newEnv]);
        setNewKey('');
        setNewValue('');
        setIsModalOpen(false);
        success('Environment variable added');
      }
    };

    const handleDelete = (id: string) => {
      if (confirm('Are you sure you want to delete this environment variable?')) {
        setEnvVars(prev => prev.filter(e => e.id !== id));
        success('Environment variable deleted');
      }
    };

    const maskValue = (value: string): string => {
      if (value.length <= 8) return '••••••••';
      return value.substring(0, 4) + '••••••••' + value.substring(value.length - 4);
    };

    if (envVars.length === 0) {
      return (
        <Card variant="bordered">
          <CardBody className="py-12 text-center">
            <div className="w-12 h-12 bg-[rgb(var(--bg-subtle))] rounded-full flex items-center justify-center mx-auto mb-4">
              <TerminalIcon className="w-6 h-6 text-[rgb(var(--text-tertiary))]" />
            </div>
            <h3 className="text-lg font-medium text-[rgb(var(--text-primary))]">No Environment Variables</h3>
            <p className="text-[rgb(var(--text-secondary))] mt-1 mb-6">Add environment variables for your serverless functions.</p>
            <Button variant="primary" onClick={() => setIsModalOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Add Variable
            </Button>
          </CardBody>
        </Card>
      );
    }

    return (
      <div className="space-y-4">
        <div className="flex justify-end">
          <Button variant="primary" onClick={() => setIsModalOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add Variable
          </Button>
        </div>

        <Card variant="bordered">
          <div className="divide-y divide-[rgb(var(--border-primary))]">
            {envVars.map((env) => (
              <div key={env.id} className="flex items-center justify-between p-4">
                <div className="flex-1 min-w-0">
                  <div className="font-mono text-sm font-medium text-[rgb(var(--text-primary))]">{env.key}</div>
                  <div className="font-mono text-sm text-[rgb(var(--text-secondary))] mt-1 truncate">
                    {visibleValues.has(env.id) ? env.value : maskValue(env.value)}
                  </div>
                </div>
                <div className="flex items-center gap-2 ml-4">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => toggleVisibility(env.id)}
                    title={visibleValues.has(env.id) ? 'Hide value' : 'Show value'}
                  >
                    {visibleValues.has(env.id) ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
                    onClick={() => handleDelete(env.id)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </Card>

        <Modal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          title="Add Environment Variable"
        >
          <div className="space-y-4">
            <Input
              label="Key"
              placeholder="API_KEY"
              value={newKey}
              onChange={(e) => setNewKey(e.target.value)}
              helperText="Will be converted to UPPERCASE_SNAKE_CASE"
            />
            <Input
              label="Value"
              placeholder="Enter the value..."
              value={newValue}
              onChange={(e) => setNewValue(e.target.value)}
            />
            <div className="flex justify-end gap-3 mt-6">
              <Button variant="ghost" onClick={() => setIsModalOpen(false)}>Cancel</Button>
              <Button variant="primary" onClick={handleAdd} disabled={!newKey || !newValue}>
                Add Variable
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    );
  };

  const KeysTab = () => {
    const { success } = useToast();
    const [keys, setKeys] = useState(mockApiKeys);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [newKeyName, setNewKeyName] = useState('');
    const [generatedKey, setGeneratedKey] = useState<string | null>(null);
    const [copied, setCopied] = useState(false);

    const handleCreate = () => {
      if (newKeyName) {
        const fullKey = `fzt_live_${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`;
        setGeneratedKey(fullKey);
        const newApiKey = {
          id: `key_${Date.now()}`,
          name: newKeyName,
          prefix: 'fzt_live_',
          created_at: new Date().toISOString(),
          last_used: null,
        };
        setKeys(prev => [...prev, newApiKey]);
      }
    };

    const handleCopy = async () => {
      if (generatedKey) {
        await navigator.clipboard.writeText(generatedKey);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
        success('API key copied to clipboard');
      }
    };

    const handleCloseModal = () => {
      setIsModalOpen(false);
      setNewKeyName('');
      setGeneratedKey(null);
      setCopied(false);
    };

    const handleDelete = (id: string) => {
      if (confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) {
        setKeys(prev => prev.filter(k => k.id !== id));
        success('API key revoked');
      }
    };

    if (keys.length === 0 && !isModalOpen) {
      return (
        <Card variant="bordered">
          <CardBody className="py-12 text-center">
            <div className="w-12 h-12 bg-[rgb(var(--bg-subtle))] rounded-full flex items-center justify-center mx-auto mb-4">
              <Key className="w-6 h-6 text-[rgb(var(--text-tertiary))]" />
            </div>
            <h3 className="text-lg font-medium text-[rgb(var(--text-primary))]">No API Keys</h3>
            <p className="text-[rgb(var(--text-secondary))] mt-1 mb-6">Create API keys to deploy programmatically.</p>
            <Button variant="primary" onClick={() => setIsModalOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create API Key
            </Button>
          </CardBody>
        </Card>
      );
    }

    return (
      <div className="space-y-4">
        <div className="flex justify-end">
          <Button variant="primary" onClick={() => setIsModalOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Create API Key
          </Button>
        </div>

        <div className="space-y-3">
          {keys.map((key) => (
            <Card key={key.id} variant="bordered" className="hover:border-[rgb(var(--border-secondary))] transition-colors">
              <CardBody className="flex items-center justify-between p-4">
                <div className="flex items-center gap-4">
                  <div className="p-2 bg-[rgb(var(--bg-subtle))] rounded-lg">
                    <Key className="w-5 h-5 text-[rgb(var(--accent))]" />
                  </div>
                  <div>
                    <div className="font-medium text-[rgb(var(--text-primary))]">{key.name}</div>
                    <div className="flex items-center gap-3 mt-1">
                      <span className="font-mono text-sm text-[rgb(var(--text-secondary))]">{key.prefix}••••••••</span>
                      <span className="text-xs text-[rgb(var(--text-tertiary))]">
                        Created {new Date(key.created_at).toLocaleDateString()}
                      </span>
                      {key.last_used && (
                        <span className="text-xs text-[rgb(var(--text-tertiary))]">
                          Last used {new Date(key.last_used).toLocaleDateString()}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
                  onClick={() => handleDelete(key.id)}
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </CardBody>
            </Card>
          ))}
        </div>

        <Modal
          isOpen={isModalOpen}
          onClose={handleCloseModal}
          title={generatedKey ? 'API Key Created' : 'Create API Key'}
        >
          {generatedKey ? (
            <div className="space-y-4">
              <div className="p-4 bg-[rgb(var(--bg-subtle))] rounded-lg border border-[rgb(var(--border-primary))]">
                <div className="flex items-center justify-between gap-3">
                  <code className="font-mono text-sm text-[rgb(var(--text-primary))] break-all">{generatedKey}</code>
                  <Button variant="ghost" size="sm" onClick={handleCopy}>
                    {copied ? <Check className="w-4 h-4 text-[rgb(var(--success))]" /> : <Copy className="w-4 h-4" />}
                  </Button>
                </div>
              </div>
              <div className="flex items-start gap-3 p-3 bg-amber-50 dark:bg-amber-900/20 rounded-lg border border-amber-200 dark:border-amber-800">
                <AlertTriangle className="w-5 h-5 text-amber-600 dark:text-amber-500 flex-shrink-0 mt-0.5" />
                <div className="text-sm text-amber-800 dark:text-amber-200">
                  <strong>Save this key now!</strong> You won't be able to see it again. Store it securely.
                </div>
              </div>
              <div className="flex justify-end mt-6">
                <Button variant="primary" onClick={handleCloseModal}>Done</Button>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <Input
                label="Key Name"
                placeholder="e.g., Production Deploy"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                helperText="A friendly name to identify this key"
              />
              <div className="flex justify-end gap-3 mt-6">
                <Button variant="ghost" onClick={handleCloseModal}>Cancel</Button>
                <Button variant="primary" onClick={handleCreate} disabled={!newKeyName}>
                  Create Key
                </Button>
              </div>
            </div>
          )}
        </Modal>
      </div>
    );
  };

  const LogsTab = () => {
    const [autoScroll, setAutoScroll] = useState(true);
    const [filter, setFilter] = useState<'all' | 'info' | 'error' | 'success'>('all');

    const filteredLogs = filter === 'all'
      ? mockLogs
      : mockLogs.filter(log => log.level === filter);

    const formatTime = (timestamp: string) => {
      return new Date(timestamp).toLocaleTimeString('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      });
    };

    const getLevelType = (level: string): 'input' | 'output' | 'error' | 'success' => {
      switch (level) {
        case 'error': return 'error';
        case 'success': return 'success';
        default: return 'output';
      }
    };

    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value as typeof filter)}
              className="px-3 py-2 text-sm bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))] rounded-lg focus:outline-none focus:ring-2 focus:ring-[rgb(var(--accent-mid))]/30 text-[rgb(var(--text-primary))]"
            >
              <option value="all">All Logs</option>
              <option value="info">Info</option>
              <option value="success">Success</option>
              <option value="error">Errors</option>
            </select>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 text-sm text-[rgb(var(--text-secondary))]">
              <input
                type="checkbox"
                checked={autoScroll}
                onChange={(e) => setAutoScroll(e.target.checked)}
                className="rounded border-[rgb(var(--border-primary))] text-[rgb(var(--accent-mid))] focus:ring-[rgb(var(--accent-mid))]/30"
              />
              Auto-scroll
            </label>
            <Button variant="secondary" size="sm">
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </div>
        </div>

        <Terminal title={`${site.name} — Logs`} showCopy={false}>
          <div className="space-y-1 max-h-[400px] overflow-y-auto">
            {filteredLogs.length === 0 ? (
              <div className="text-[rgb(var(--text-tertiary))] py-8 text-center">
                No logs matching filter
              </div>
            ) : (
              filteredLogs.map((log, idx) => (
                <TerminalLine key={idx} type={getLevelType(log.level)} prefix={formatTime(log.timestamp)}>
                  <span className={`inline-block w-16 ${
                    log.level === 'error' ? 'text-[rgb(var(--accent-start))]' :
                    log.level === 'success' ? 'text-[rgb(var(--success))]' :
                    'text-[rgb(var(--text-secondary))]'
                  }`}>
                    [{log.level.toUpperCase()}]
                  </span>
                  {log.message}
                </TerminalLine>
              ))
            )}
          </div>
        </Terminal>

        <p className="text-xs text-[rgb(var(--text-tertiary))] text-center">
          Showing last {filteredLogs.length} log entries • Logs are retained for 7 days
        </p>
      </div>
    );
  };

  const SettingsTab = () => {
    const { success } = useToast();
    const [customDomain, setCustomDomain] = useState('');
    const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
    const [deleteConfirmation, setDeleteConfirmation] = useState('');

    const handleAddDomain = () => {
      if (customDomain) {
        success(`Custom domain "${customDomain}" added. Configure your DNS to point to your server.`);
        setCustomDomain('');
      }
    };

    const handleDelete = () => {
      if (deleteConfirmation === site.name) {
        success('Site deleted successfully');
        navigate('/sites');
      }
    };

    return (
      <div className="space-y-6">
        {/* Custom Domain Section */}
        <Card variant="bordered">
          <CardBody>
            <div className="flex items-start gap-4 mb-4">
              <div className="p-2 bg-[rgb(var(--bg-subtle))] rounded-lg">
                <Link2 className="w-5 h-5 text-[rgb(var(--accent))]" />
              </div>
              <div>
                <h3 className="font-medium text-[rgb(var(--text-primary))]">Custom Domain</h3>
                <p className="text-sm text-[rgb(var(--text-secondary))] mt-1">
                  Add a custom domain to serve your site on your own URL.
                </p>
              </div>
            </div>

            <div className="flex gap-3">
              <div className="flex-1">
                <Input
                  placeholder="www.example.com"
                  value={customDomain}
                  onChange={(e) => setCustomDomain(e.target.value)}
                />
              </div>
              <Button variant="secondary" onClick={handleAddDomain} disabled={!customDomain}>
                Add Domain
              </Button>
            </div>

            <div className="mt-4 p-3 bg-[rgb(var(--bg-subtle))] rounded-lg">
              <p className="text-xs text-[rgb(var(--text-secondary))]">
                <strong>Current domains:</strong>
              </p>
              <div className="flex items-center gap-2 mt-2">
                <Badge variant="default">{site.domain}</Badge>
                <span className="text-xs text-[rgb(var(--text-tertiary))]">Primary</span>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* General Settings */}
        <Card variant="bordered">
          <CardBody>
            <div className="flex items-start gap-4 mb-4">
              <div className="p-2 bg-[rgb(var(--bg-subtle))] rounded-lg">
                <Settings className="w-5 h-5 text-[rgb(var(--accent))]" />
              </div>
              <div>
                <h3 className="font-medium text-[rgb(var(--text-primary))]">Site Settings</h3>
                <p className="text-sm text-[rgb(var(--text-secondary))] mt-1">
                  Configure how your site behaves.
                </p>
              </div>
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between py-3 border-b border-[rgb(var(--border-primary))]">
                <div>
                  <div className="font-medium text-sm text-[rgb(var(--text-primary))]">Directory Listing</div>
                  <div className="text-xs text-[rgb(var(--text-secondary))]">Show file browser for directories without index.html</div>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input type="checkbox" className="sr-only peer" />
                  <div className="w-11 h-6 bg-[rgb(var(--bg-subtle))] peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-[rgb(var(--accent-mid))]/30 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[rgb(var(--accent-mid))]"></div>
                </label>
              </div>

              <div className="flex items-center justify-between py-3 border-b border-[rgb(var(--border-primary))]">
                <div>
                  <div className="font-medium text-sm text-[rgb(var(--text-primary))]">Clean URLs</div>
                  <div className="text-xs text-[rgb(var(--text-secondary))]">Remove .html extension from URLs</div>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input type="checkbox" defaultChecked className="sr-only peer" />
                  <div className="w-11 h-6 bg-[rgb(var(--bg-subtle))] peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-[rgb(var(--accent-mid))]/30 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[rgb(var(--accent-mid))]"></div>
                </label>
              </div>

              <div className="flex items-center justify-between py-3">
                <div>
                  <div className="font-medium text-sm text-[rgb(var(--text-primary))]">SPA Mode</div>
                  <div className="text-xs text-[rgb(var(--text-secondary))]">Route all requests to index.html for client-side routing</div>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input type="checkbox" className="sr-only peer" />
                  <div className="w-11 h-6 bg-[rgb(var(--bg-subtle))] peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-[rgb(var(--accent-mid))]/30 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[rgb(var(--accent-mid))]"></div>
                </label>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Danger Zone */}
        <Card variant="bordered" className="border-red-200 dark:border-red-900/50">
          <CardBody>
            <div className="flex items-start gap-4 mb-4">
              <div className="p-2 bg-red-50 dark:bg-red-900/20 rounded-lg">
                <AlertTriangle className="w-5 h-5 text-red-500" />
              </div>
              <div>
                <h3 className="font-medium text-red-600 dark:text-red-400">Danger Zone</h3>
                <p className="text-sm text-[rgb(var(--text-secondary))] mt-1">
                  Irreversible and destructive actions.
                </p>
              </div>
            </div>

            <div className="flex items-center justify-between p-4 bg-red-50 dark:bg-red-900/10 rounded-lg border border-red-200 dark:border-red-900/30">
              <div>
                <div className="font-medium text-sm text-[rgb(var(--text-primary))]">Delete Site</div>
                <div className="text-xs text-[rgb(var(--text-secondary))]">
                  Permanently delete this site and all its files. This cannot be undone.
                </div>
              </div>
              <Button variant="danger" size="sm" onClick={() => setIsDeleteModalOpen(true)}>
                <Trash2 className="w-4 h-4 mr-2" />
                Delete
              </Button>
            </div>
          </CardBody>
        </Card>

        {/* Delete Confirmation Modal */}
        <Modal
          isOpen={isDeleteModalOpen}
          onClose={() => {
            setIsDeleteModalOpen(false);
            setDeleteConfirmation('');
          }}
          title="Delete Site"
        >
          <div className="space-y-4">
            <div className="flex items-start gap-3 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg border border-red-200 dark:border-red-800">
              <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-red-800 dark:text-red-200">
                This action is <strong>permanent</strong> and cannot be undone. All files and configuration will be lost.
              </div>
            </div>

            <p className="text-sm text-[rgb(var(--text-secondary))]">
              To confirm, type <strong className="font-mono text-[rgb(var(--text-primary))]">{site.name}</strong> below:
            </p>

            <Input
              placeholder={site.name}
              value={deleteConfirmation}
              onChange={(e) => setDeleteConfirmation(e.target.value)}
            />

            <div className="flex justify-end gap-3 mt-6">
              <Button variant="ghost" onClick={() => {
                setIsDeleteModalOpen(false);
                setDeleteConfirmation('');
              }}>
                Cancel
              </Button>
              <Button
                variant="danger"
                onClick={handleDelete}
                disabled={deleteConfirmation !== site.name}
              >
                Delete Site
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    );
  };

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
            { id: 'logs', label: 'Logs', content: <LogsTab /> },
            { id: 'settings', label: 'Settings', content: <SettingsTab /> },
        ]}
      />
    </div>
  );
}
