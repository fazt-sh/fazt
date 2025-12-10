import type { ReactNode } from 'react';
import { Copy, Check } from 'lucide-react';
import { useState } from 'react';

interface TerminalProps {
  children: ReactNode;
  className?: string;
  title?: string;
  showCopy?: boolean;
}

export function Terminal({ children, className = '', title = 'Terminal', showCopy = true }: TerminalProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const text = typeof children === 'string' ? children : '';
    if (text) {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className={`terminal ${className}`}>
      <div className="terminal-header">
        <div className="flex items-center gap-2">
          <div className="terminal-dot red"></div>
          <div className="terminal-dot yellow"></div>
          <div className="terminal-dot green"></div>
        </div>
        <div className="flex-1 text-center text-[rgb(var(--text-secondary))] text-xs font-mono">
          {title}
        </div>
        {showCopy && (
          <button
            onClick={handleCopy}
            className="p-1 rounded text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))]
                     hover:bg-[rgba(255,255,255,0.1)] transition-all duration-150"
            aria-label="Copy to clipboard"
          >
            {copied ? (
              <Check className="h-3 w-3" />
            ) : (
              <Copy className="h-3 w-3" />
            )}
          </button>
        )}
      </div>
      <div className="p-4 font-mono text-sm text-[rgb(var(--text-primary))] overflow-x-auto">
        {children}
      </div>
    </div>
  );
}

interface TerminalLineProps {
  children: ReactNode;
  type?: 'input' | 'output' | 'error' | 'success';
  prefix?: string;
}

export function TerminalLine({ children, type = 'output', prefix }: TerminalLineProps) {
  const typeStyles = {
    input: 'text-[rgb(var(--accent-mid))]',
    output: 'text-[rgb(var(--text-primary))]',
    error: 'text-[rgb(var(--accent-start))]',
    success: 'text-[rgb(var(--success))]',
  };

  return (
    <div className={`${typeStyles[type]} mb-1`}>
      {prefix && <span className="text-[rgb(var(--text-secondary))] mr-2">{prefix}</span>}
      {children}
    </div>
  );
}