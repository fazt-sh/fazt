import type { ReactNode } from 'react';
import type { CSSProperties } from 'react';

export interface CardProps {
  children: ReactNode;
  className?: string;
  variant?: 'default' | 'bordered' | 'elevated' | 'glass';
  hover?: boolean;
  style?: CSSProperties;
}

export function Card({ children, className = '', variant = 'default', hover = false, style }: CardProps) {
  const variantStyles = {
    default: 'bg-[rgb(var(--bg-elevated))]',
    bordered: 'bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))]',
    elevated: 'bg-[rgb(var(--bg-elevated))] shadow-md',
    glass: 'glass',
  };

  const hoverStyles = hover ? 'hover-lift' : '';

  return (
    <div
      style={style}
      className={`rounded-xl transition-all duration-150 ${variantStyles[variant]} ${hoverStyles} ${className}`}
    >
      {children}
    </div>
  );
}

export interface CardHeaderProps {
  children: ReactNode;
  className?: string;
}

export function CardHeader({ children, className = '' }: CardHeaderProps) {
  return (
    <div className={`px-6 py-4 border-b border-[rgb(var(--border-primary))] ${className}`}>
      {children}
    </div>
  );
}

export interface CardBodyProps {
  children: ReactNode;
  className?: string;
}

export function CardBody({ children, className = '' }: CardBodyProps) {
  return (
    <div className={`px-6 py-4 ${className}`}>
      {children}
    </div>
  );
}

export interface CardFooterProps {
  children: ReactNode;
  className?: string;
}

export function CardFooter({ children, className = '' }: CardFooterProps) {
  return (
    <div className={`px-6 py-4 border-t border-[rgb(var(--border-primary))] ${className}`}>
      {children}
    </div>
  );
}