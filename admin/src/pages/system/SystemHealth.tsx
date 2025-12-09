import { PageHeader } from '../../components/layout/PageHeader';
import { Card, CardBody, Badge, Button } from '../../components/ui';
import { HeartPulse, Database, Globe, Shield, RefreshCw, CheckCircle, AlertTriangle, XCircle } from 'lucide-react';
import { useMockMode } from '../../context/MockContext';
import { mockData } from '../../lib/mockData';

export function SystemHealth() {
  const { enabled: mockMode } = useMockMode();
  const health = mockMode ? mockData.systemHealth : null;

  const StatusIcon = ({ status }: { status: string }) => {
    switch (status) {
      case 'healthy': return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'degraded': return <AlertTriangle className="w-5 h-5 text-yellow-500" />;
      case 'down': return <XCircle className="w-5 h-5 text-red-500" />;
      default: return <div className="w-5 h-5" />;
    }
  };

  const ServiceRow = ({ name, icon: Icon, status, details }: { name: string, icon: any, status: 'healthy' | 'degraded' | 'down', details?: string }) => (
    <div className="flex items-center justify-between p-4 bg-[rgb(var(--bg-subtle))] rounded-lg border border-[rgb(var(--border-primary))]">
      <div className="flex items-center gap-4">
        <div className="p-2 bg-[rgb(var(--bg-surface))] rounded-lg">
          <Icon className="w-5 h-5 text-[rgb(var(--text-secondary))]" />
        </div>
        <div>
          <h3 className="font-medium text-[rgb(var(--text-primary))]">{name}</h3>
          {details && <p className="text-sm text-[rgb(var(--text-secondary))]">{details}</p>}
        </div>
      </div>
      <div className="flex items-center gap-2">
        <StatusIcon status={status} />
        <span className={`text-sm font-medium capitalize ${
          status === 'healthy' ? 'text-green-500' :
          status === 'degraded' ? 'text-yellow-500' : 'text-red-500'
        }`}>
          {status}
        </span>
      </div>
    </div>
  );

  return (
    <div className="animate-fade-in">
      <PageHeader
        title="System Health"
        description="Monitor system services and operational status."
        action={
          <Button variant="ghost">
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        }
      />

      <div className="grid gap-6">
        <Card variant="bordered">
          <CardBody>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Overall Status</h2>
              <Badge variant={health?.status === 'healthy' ? 'success' : health?.status === 'degraded' ? 'warning' : 'error'} size="lg">
                {health?.status || 'Unknown'}
              </Badge>
            </div>
            
            <div className="grid gap-3">
              <ServiceRow 
                name="Database (SQLite)" 
                icon={Database} 
                status={health?.database.connected ? 'healthy' : 'down'} 
                details={health ? `${(health.database.size / 1024 / 1024).toFixed(2)} MB` : undefined}
              />
              <ServiceRow 
                name="API Server" 
                icon={Globe} 
                status="healthy" 
                details="v0.7.1"
              />
              <ServiceRow 
                name="Authentication" 
                icon={Shield} 
                status="healthy" 
                details="OAuth2 / Session"
              />
              <ServiceRow 
                name="Background Workers" 
                icon={HeartPulse} 
                status="healthy" 
                details="Active"
              />
            </div>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}