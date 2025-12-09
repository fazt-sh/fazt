import { forwardRef } from 'react';
import type { InputHTMLAttributes } from 'react';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
  icon?: React.ReactNode;
  iconPosition?: 'left' | 'right';
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, helperText, icon, iconPosition = 'right', className = '', ...props }, ref) => {
    const inputBaseStyles = `block w-full rounded-lg text-sm transition-all duration-150
      focus:outline-none disabled:opacity-50 disabled:pointer-events-none
      ${error
        ? 'border-[rgb(var(--accent-start))] focus:ring-2 focus:ring-[rgb(var(--accent-start))]/20 focus:border-[rgb(var(--accent-start))]'
        : 'border-[rgb(var(--border-primary))] focus:ring-2 focus:ring-[rgb(var(--accent-mid))]/30 focus:border-[rgb(var(--accent-mid))] hover:border-[rgb(var(--border-secondary))]'
      }
      bg-[rgb(var(--bg-elevated))]
      text-[rgb(var(--text-primary))]
      placeholder-[rgb(var(--text-tertiary))]
      ${icon ? (iconPosition === 'left' ? 'pl-10' : 'pr-10') : 'px-3.5'}
      py-3`;

    return (
      <div className="w-full">
        {label && (
          <label className="block text-[13px] font-medium text-[rgb(var(--text-primary))] mb-2">
            {label}
          </label>
        )}

        <div className="relative">
          {icon && iconPosition === 'left' && (
            <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
              <div className="h-5 w-5 text-[rgb(var(--text-tertiary))]">
                {icon}
              </div>
            </div>
          )}

          <input
            ref={ref}
            className={`${inputBaseStyles} ${className}`}
            {...props}
          />

          {error && iconPosition === 'right' && (
            <div className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
              <svg className="h-5 w-5 text-[rgb(var(--accent-start))]" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
            </div>
          )}

          {icon && !error && iconPosition === 'right' && (
            <div className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
              <div className="h-5 w-5 text-[rgb(var(--text-tertiary))]">
                {icon}
              </div>
            </div>
          )}
        </div>

        {error && (
          <p className="mt-1.5 text-[13px] text-[rgb(var(--accent-start))] flex items-center gap-1">
            <svg className="h-4 w-4 flex-shrink-0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            {error}
          </p>
        )}

        {helperText && !error && (
          <p className="mt-1.5 text-[13px] text-[rgb(var(--text-tertiary))]">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';