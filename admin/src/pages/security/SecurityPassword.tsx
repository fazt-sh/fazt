import { Lock } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SecurityPassword() {
  return (
    <PlaceholderPage
      title="Password Settings"
      description="Change your account password."
      icon={Lock}
    />
  );
}
