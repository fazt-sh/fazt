import { Network } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function Proxy() {
  return (
    <PlaceholderPage
      title="Proxy Configuration"
      description="Configure reverse proxy settings."
      icon={Network}
    />
  );
}
