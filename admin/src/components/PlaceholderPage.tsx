import { type LucideIcon, Construction } from 'lucide-react';
import { PageHeader } from './layout/PageHeader';

interface PlaceholderPageProps {
  title: string;
  description: string;
  icon: LucideIcon;
}

export function PlaceholderPage({ title, description, icon: Icon }: PlaceholderPageProps) {
  return (
    <div className="space-y-6">
      <PageHeader title={title} description={description} />

      <div className="bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))] rounded-xl p-12">
        <div className="flex flex-col items-center justify-center text-center">
          <div className="w-16 h-16 rounded-2xl bg-accent/10 flex items-center justify-center mb-4">
            <Icon className="w-8 h-8 text-accent" />
          </div>
          <h2 className="text-xl font-medium text-primary mb-2">{title}</h2>
          <p className="text-secondary max-w-md mb-6">
            This page is under construction. The functionality will be available in a future update.
          </p>
          <div className="flex items-center gap-2 text-sm text-tertiary">
            <Construction className="w-4 h-4" />
            <span>Coming Soon</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))] rounded-lg p-4">
            <div className="w-8 h-8 rounded-lg bg-tertiary/10 mb-3" />
            <div className="h-4 w-24 bg-tertiary/20 rounded mb-2" />
            <div className="h-3 w-32 bg-tertiary/10 rounded" />
          </div>
        ))}
      </div>
    </div>
  );
}
