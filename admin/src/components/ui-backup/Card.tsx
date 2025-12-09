import type { ReactNode } from 'react';

interface CardProps {
  children: ReactNode;
  className?: string;
  variant?: 'default' | 'bordered' | 'elevated';
  style?: React.CSSProperties;
}

export function Card({ children, className = '', variant = 'default', style }: CardProps) {
  const variantStyles = {
    default: 'bg-[rgb(var(--bg-elevated))] hover-lift',
    bordered: 'glass hover-lift',
    elevated: 'bg-[rgb(var(--bg-elevated))] shadow-[var(--shadow-md)] hover-lift',
  };

  return (
    <div style={style} className={`rounded-xl transition-all duration-150 ${variantStyles[variant]} ${className}`}>
      {children}
    </div>
  );
}

interface CardHeaderProps {
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

interface CardBodyProps {
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

interface CardFooterProps {
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
