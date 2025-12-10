import { useState, type ReactNode } from 'react';

interface Tab {
  id: string;
  label: string;
  content?: ReactNode;
  disabled?: boolean;
}

interface TabsProps {
  tabs: Tab[];
  defaultTab?: string;
  onChange?: (tabId: string) => void;
  className?: string;
}

export function Tabs({ tabs, defaultTab, onChange, className = '' }: TabsProps) {
  const [activeTab, setActiveTab] = useState(defaultTab || tabs[0]?.id);

  const handleTabClick = (id: string) => {
    setActiveTab(id);
    onChange?.(id);
  };

  return (
    <div className={className}>
      <div className="border-b border-[rgb(var(--border-primary))]">
        <nav className="-mb-px flex space-x-6 overflow-x-auto hide-scrollbar" aria-label="Tabs">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => !tab.disabled && handleTabClick(tab.id)}
              disabled={tab.disabled}
              className={`
                whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors duration-200 outline-none focus-visible:ring-2 focus-visible:ring-[rgb(var(--accent))] focus-visible:ring-offset-2
                ${
                  activeTab === tab.id
                    ? 'border-[rgb(var(--accent))] text-[rgb(var(--accent))]'
                    : tab.disabled
                      ? 'border-transparent text-[rgb(var(--text-tertiary))] cursor-not-allowed'
                      : 'border-transparent text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] hover:border-[rgb(var(--border-secondary))]'
                }
              `}
              aria-current={activeTab === tab.id ? 'page' : undefined}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>
      <div className="mt-6">
        {tabs.map((tab) => (
           <div
             key={tab.id}
             className={`${activeTab === tab.id ? 'block' : 'hidden'} animate-fade-in`}
             role="tabpanel"
             tabIndex={0}
           >
             {tab.content}
           </div>
        ))}
      </div>
    </div>
  );
}
