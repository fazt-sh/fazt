import { Key } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SecuritySSH() {
  return (
    <PlaceholderPage
      title="SSH Keys"
      description="Manage SSH keys for secure access."
      icon={Key}
    />
  );
}
