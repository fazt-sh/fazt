import { forwardRef } from 'react';
import type { ButtonHTMLAttributes } from 'react';
import { Loader2 } from 'lucide-react';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', loading, disabled, children, className = '', ...props }, ref) => {
    const baseStyles = 'inline-flex items-center justify-center font-medium rounded-lg transition-all duration-150 disabled:opacity-50 disabled:pointer-events-none border-0';

    const sizeStyles = {
      sm: 'py-2 px-3 text-[13px]',
      md: 'py-3 px-4 text-[13px]',
      lg: 'p-4 sm:p-5 text-sm',
    };

    const variantStyles = {
      primary: 'border-transparent bg-[rgb(var(--accent-mid))] text-white hover:bg-[rgb(var(--accent-start))] shadow-sm hover:shadow-md active:scale-[0.98]',
      secondary: 'border-[rgb(var(--border-primary))] bg-[rgb(var(--bg-elevated))] text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))] hover:border-[rgb(var(--border-secondary))]',
      ghost: 'border-transparent text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]',
      danger: 'border-transparent bg-[rgb(var(--accent-start))] text-white hover:bg-[rgb(var(--accent-start))]/90 shadow-sm hover:shadow-md active:scale-[0.98]',
    };

    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={`${baseStyles} ${sizeStyles[size]} ${variantStyles[variant]} ${className}`}
        {...props}
      >
        {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';