import type { ReactNode } from 'react';

interface PageHeaderProps {
  title: string;
  description?: ReactNode;
  action?: ReactNode;
}

export function PageHeader({ title, description, action }: PageHeaderProps) {
  return (
    <div className="flex items-start justify-between mb-6">
      <div>
        <h1 className="text-2xl font-semibold text-[rgb(var(--text-primary))]">
          {title}
        </h1>
        {description && (
          <div className="text-[rgb(var(--text-secondary))] mt-1">
            {description}
          </div>
        )}
      </div>
      {action && <div>{action}</div>}
    </div>
  );
}
