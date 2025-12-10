import { Gauge } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SystemLimits() {
  return (
    <PlaceholderPage
      title="Resource Limits"
      description="Configure system resource limits and quotas."
      icon={Gauge}
    />
  );
}
