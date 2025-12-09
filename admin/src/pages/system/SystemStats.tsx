import { PageHeader } from '../../components/layout/PageHeader';
import { SystemInfo } from '../../components/ui';
import { Button } from '../../components/ui';
import { RefreshCw } from 'lucide-react';

export function SystemStats() {
  return (
    <div className="animate-fade-in">
      <PageHeader
        title="System Statistics"
        description="Real-time performance metrics and system status."
        action={
          <Button variant="ghost">
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        }
      />
      
      <div className="max-w-4xl">
        <SystemInfo />
      </div>
    </div>
  );
}