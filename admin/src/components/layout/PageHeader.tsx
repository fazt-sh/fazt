import type { ReactNode } from 'react';

interface PageHeaderProps {
  title: string;
  description?: ReactNode;
  action?: ReactNode;
}

export function PageHeader({ title, description, action }: PageHeaderProps) {
  return (
    <div className="flex items-center justify-between mb-8">
      <div>
        <h1 className="font-display text-3xl text-[rgb(var(--text-primary))] tracking-tight">
          {title}
        </h1>
        {description && (
          <p className="mt-2 text-[13px] text-[rgb(var(--text-secondary))]">
            {description}
          </p>
        )}
      </div>
      {action && <div>{action}</div>}
    </div>
  );
}
