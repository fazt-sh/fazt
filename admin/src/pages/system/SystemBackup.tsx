import { Archive } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SystemBackup() {
  return (
    <PlaceholderPage
      title="Backup Management"
      description="Manage system backups and restores."
      icon={Archive}
    />
  );
}
