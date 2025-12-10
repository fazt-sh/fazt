import type { ReactNode } from 'react';

export interface BadgeProps {
  children: ReactNode;
  className?: string;
  variant?: 'default' | 'success' | 'error' | 'warning' | 'info';
  size?: 'sm' | 'md' | 'lg';
  soft?: boolean;
  dot?: boolean;
}

export function Badge({
  children,
  className = '',
  variant = 'default',
  size = 'md',
  soft = false,
  dot = false
}: BadgeProps) {
  const sizeStyles = {
    sm: 'py-0.5 px-2 text-xs',
    md: 'py-1.5 px-3 text-xs',
    lg: 'py-2 px-4 text-sm',
  };

  const variantStyles = soft ? {
    default: 'bg-gray-100 text-gray-800 dark:bg-gray-800/30 dark:text-gray-500',
    success: 'bg-green-100 text-green-800 dark:bg-green-800/30 dark:text-green-500',
    error: 'bg-red-100 text-red-800 dark:bg-red-800/30 dark:text-red-500',
    warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-800/30 dark:text-yellow-500',
    info: 'bg-blue-100 text-blue-800 dark:bg-blue-800/30 dark:text-blue-500',
  } : {
    default: 'bg-gray-800 text-white dark:bg-white dark:text-neutral-800',
    success: 'bg-[rgb(var(--success))] text-white',
    error: 'bg-[rgb(var(--accent-start))] text-white',
    warning: 'bg-[rgb(var(--accent-mid))] text-white',
    info: 'bg-[rgb(var(--info))] text-white',
  };

  const dotStyles = dot ? 'pl-1.5' : '';
  const dotElement = dot ? (
    <span className={`mr-1.5 h-2 w-2 rounded-full ${
      variant === 'success' ? 'bg-current' :
      variant === 'error' ? 'bg-current' :
      variant === 'warning' ? 'bg-current' :
      variant === 'info' ? 'bg-current' :
      'bg-current'
    }`} />
  ) : null;

  return (
    <span
      className={`inline-flex items-center gap-x-1.5 rounded-full font-medium ${sizeStyles[size]} ${variantStyles[variant]} ${dotStyles} ${className}`}
    >
      {dotElement}
      {children}
    </span>
  );
}