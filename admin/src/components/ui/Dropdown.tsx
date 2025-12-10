import { useState, useRef, useEffect } from 'react';
import type { ReactNode, ReactElement } from 'react';

export interface DropdownProps {
  trigger: ReactElement;
  children: ReactNode;
  className?: string;
  placement?: 'bottom-left' | 'bottom-right' | 'top-left' | 'top-right';
}

export interface DropdownItemProps {
  children: ReactNode;
  onClick?: () => void;
  icon?: ReactNode;
  disabled?: boolean;
  className?: string;
}

export function Dropdown({ trigger, children, className = '', placement = 'bottom-left' }: DropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const placementClasses = {
    'bottom-left': 'top-full left-0 mt-1',
    'bottom-right': 'top-full right-0 mt-1',
    'top-left': 'bottom-full left-0 mb-1',
    'top-right': 'bottom-full right-0 mb-1',
  };

  return (
    <div className={`relative inline-block ${className}`} ref={dropdownRef}>
      <div onClick={() => setIsOpen(!isOpen)}>
        {trigger}
      </div>

      {isOpen && (
        <div
          className={`absolute z-50 min-w-60 rounded-lg shadow-lg border border-[rgb(var(--border-primary))] bg-[rgb(var(--bg-elevated))] py-1 ${placementClasses[placement]}`}
        >
          {children}
        </div>
      )}
    </div>
  );
}

export function DropdownItem({ children, onClick, icon, disabled = false, className = '' }: DropdownItemProps) {
  const baseStyles = `flex items-center gap-x-3 px-3 py-2 text-sm transition-colors cursor-pointer
    ${disabled
      ? 'text-[rgb(var(--text-tertiary))] cursor-not-allowed'
      : 'text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]'
    } ${className}`;

  return (
    <div
      className={baseStyles}
      onClick={!disabled ? onClick : undefined}
      role="menuitem"
    >
      {icon && (
        <div className="h-4 w-4 flex-shrink-0">
          {icon}
        </div>
      )}
      {children}
    </div>
  );
}

export function DropdownDivider() {
  return (
    <div className="my-1 h-px bg-[rgb(var(--border-primary))]" role="separator" />
  );
}