export interface SpinnerProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'primary' | 'success' | 'error' | 'warning';
}

export function Spinner({ className = '', size = 'md', variant = 'default' }: SpinnerProps) {
  const sizeStyles = {
    sm: 'w-4 h-4 border-2',
    md: 'w-6 h-6 border-[3px]',
    lg: 'w-8 h-8 border-[4px]',
  };

  const variantStyles = {
    default: 'border-[rgb(var(--text-tertiary))] border-t-[rgb(var(--text-secondary))]',
    primary: 'border-[rgb(var(--accent-mid))]/30 border-t-[rgb(var(--accent-mid))]',
    success: 'border-[rgb(var(--success))]/30 border-t-[rgb(var(--success))]',
    error: 'border-[rgb(var(--accent-start))]/30 border-t-[rgb(var(--accent-start))]',
    warning: 'border-[rgb(var(--warning))]/30 border-t-[rgb(var(--warning))]',
  };

  return (
    <div
      className={`animate-spin rounded-full ${sizeStyles[size]} ${variantStyles[variant]} ${className}`}
      role="status"
      aria-label="Loading"
    >
      <span className="sr-only">Loading...</span>
    </div>
  );
}