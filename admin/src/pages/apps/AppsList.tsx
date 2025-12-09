import { Package } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AppsList() {
  return (
    <PlaceholderPage
      title="Installed Apps"
      description="Manage installed applications and extensions."
      icon={Package}
    />
  );
}
