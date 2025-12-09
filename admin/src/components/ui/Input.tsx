import { forwardRef } from 'react';
import type { InputHTMLAttributes } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, helperText, className = '', ...props }, ref) => {
    return (
      <div className="w-full">
        {label && (
          <label className="block text-[13px] font-medium text-[rgb(var(--text-primary))] mb-2">
            {label}
          </label>
        )}
        <input
          ref={ref}
          className={`
            w-full px-3.5 py-2.5 rounded-lg border text-[13px]
            transition-all duration-150
            ${error
              ? 'border-red-500 focus:ring-2 focus:ring-red-500/20 focus:border-red-500'
              : 'border-[rgb(var(--border-primary))] focus:ring-2 focus:ring-[rgb(var(--accent-glow))] focus:border-[rgb(var(--accent))]'
            }
            bg-[rgb(var(--bg-elevated))]
            text-[rgb(var(--text-primary))]
            placeholder-[rgb(var(--text-tertiary))]
            focus:outline-none
            disabled:opacity-50 disabled:cursor-not-allowed
            ${className}
          `}
          {...props}
        />
        {error && (
          <p className="mt-1.5 text-[13px] text-red-600 dark:text-red-400">{error}</p>
        )}
        {helperText && !error && (
          <p className="mt-1.5 text-[13px] text-[rgb(var(--text-tertiary))]">{helperText}</p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';
