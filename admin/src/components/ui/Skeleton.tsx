import type { CSSProperties } from 'react';

export interface SkeletonProps {
  className?: string;
  variant?: 'text' | 'circle' | 'rect' | 'card';
  width?: string | number;
  height?: string | number;
  lines?: number;
  animated?: boolean;
}

export function Skeleton({
  className = '',
  variant = 'text',
  width,
  height,
  lines = 1,
  animated = true
}: SkeletonProps) {
  const baseClasses = `bg-[rgb(var(--bg-subtle))] rounded ${animated ? 'animate-pulse' : ''}`;

  const style: CSSProperties = {
    width: width || '100%',
    height: height || (variant === 'text' ? '1rem' : '3rem'),
  };

  switch (variant) {
    case 'text':
      return (
        <div className={`${className} space-y-2`}>
          {Array.from({ length: lines }).map((_, i) => (
            <div
              key={i}
              className={`${baseClasses} rounded-full`}
              style={{
                ...style,
                width: i === lines - 1 && lines > 1 ? '75%' : width,
                height: '1rem',
              }}
            />
          ))}
        </div>
      );

    case 'circle':
      return (
        <div
          className={`${baseClasses} ${className}`}
          style={{
            ...style,
            borderRadius: '50%',
          }}
        />
      );

    case 'rect':
      return (
        <div
          className={`${baseClasses} ${className}`}
          style={style}
        />
      );

    case 'card':
      return (
        <div className={`border border-[rgb(var(--border-primary))] rounded-lg p-4 ${className}`}>
          {/* Image placeholder */}
          <div
            className={`${baseClasses} mb-4 rounded-lg`}
            style={{ height: '12rem' }}
          />
          {/* Title */}
          <div
            className={`${baseClasses} mb-2 rounded-full`}
            style={{ width: '60%', height: '1.25rem' }}
          />
          {/* Text lines */}
          <div className="space-y-2">
            <div
              className={`${baseClasses} rounded-full`}
              style={{ width: '100%', height: '0.875rem' }}
            />
            <div
              className={`${baseClasses} rounded-full`}
              style={{ width: '100%', height: '0.875rem' }}
            />
            <div
              className={`${baseClasses} rounded-full`}
              style={{ width: '40%', height: '0.875rem' }}
            />
          </div>
        </div>
      );

    default:
      return (
        <div
          className={`${baseClasses} ${className}`}
          style={style}
        />
      );
  }
}