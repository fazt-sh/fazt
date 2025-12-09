import { Plus } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function CreateSite() {
  return (
    <PlaceholderPage
      title="Create Site"
      description="Deploy a new static site or serverless function."
      icon={Plus}
    />
  );
}
