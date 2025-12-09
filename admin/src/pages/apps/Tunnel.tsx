import { Route } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function Tunnel() {
  return (
    <PlaceholderPage
      title="Tunnel Management"
      description="Manage secure tunnels to local services."
      icon={Route}
    />
  );
}
