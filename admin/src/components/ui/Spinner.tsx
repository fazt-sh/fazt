import { Loader2 } from 'lucide-react';

interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function Spinner({ size = 'md', className = '' }: SpinnerProps) {
  const sizeStyles = {
    sm: 'h-4 w-4',
    md: 'h-8 w-8',
    lg: 'h-12 w-12',
  };

  return (
    <Loader2 className={`animate-spin text-primary ${sizeStyles[size]} ${className}`} />
  );
}

export function PageSpinner() {
  return (
    <div className="flex items-center justify-center h-full w-full">
      <Spinner size="lg" />
    </div>
  );
}
