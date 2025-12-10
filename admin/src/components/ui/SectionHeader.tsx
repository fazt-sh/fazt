import type { ReactNode } from 'react';

export interface SectionHeaderProps {
  title: string;
  description?: string | ReactNode;
  action?: ReactNode;
  className?: string;
}

export function SectionHeader({ title, description, action, className = '' }: SectionHeaderProps) {
  return (
    <div className={`flex items-start justify-between mb-6 ${className}`}>
      <div>
        <h3 className="text-lg font-medium text-[rgb(var(--text-primary))]">
          {title}
        </h3>
        {description && (
          <div className="text-sm text-[rgb(var(--text-secondary))] mt-1">
            {description}
          </div>
        )}
      </div>
      {action && <div>{action}</div>}
    </div>
  );
}
