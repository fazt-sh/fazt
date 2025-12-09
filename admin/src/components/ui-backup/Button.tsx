import { forwardRef } from 'react';
import type { ButtonHTMLAttributes } from 'react';
import { Loader2 } from 'lucide-react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', loading, disabled, children, className = '', ...props }, ref) => {
    const baseStyles = 'inline-flex items-center justify-center rounded-lg font-medium transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-[rgb(var(--accent-mid))]/50 disabled:opacity-50 disabled:cursor-not-allowed';

    const variantStyles = {
      primary: 'bg-gradient-to-r from-[rgb(var(--accent-start))] to-[rgb(var(--accent-mid))] text-white hover:from-[rgb(var(--accent-start))]/90 hover:to-[rgb(var(--accent-mid))]/90 shadow-sm hover:shadow-md hover-glow active:scale-[0.98]',
      secondary: 'bg-[rgb(var(--bg-subtle))] text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))] border border-[rgb(var(--border-primary))] hover:border-[rgb(var(--border-primary))]',
      ghost: 'bg-transparent text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]',
      danger: 'bg-[rgb(var(--accent-start))] text-white hover:bg-[rgb(var(--accent-start))]/90 shadow-sm hover:shadow active:scale-[0.98]',
    };

    const sizeStyles = {
      sm: 'px-3 py-1.5 text-[13px]',
      md: 'px-4 py-2 text-[13px]',
      lg: 'px-5 py-2.5 text-sm',
    };

    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}
        {...props}
      >
        {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';
