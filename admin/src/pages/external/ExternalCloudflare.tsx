import { Cloud } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function ExternalCloudflare() {
  return (
    <PlaceholderPage
      title="Cloudflare"
      description="Manage Cloudflare DNS and proxy settings."
      icon={Cloud}
    />
  );
}
