import { useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import {
  LayoutDashboard,
  Globe,
  BarChart3,
  Link2,
  Webhook,
  Settings,
  FileText,
  Palette,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';

interface NavItem {
  to?: string;
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  devOnly?: boolean;
  children?: NavItem[];
}

const navItems: NavItem[] = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  {
    to: '/sites',
    icon: Globe,
    label: 'Sites',
    children: [
      { to: '/sites', icon: Globe, label: 'All Sites' },
      { to: '/sites/analytics', icon: BarChart3, label: 'Site Analytics' },
      { to: '/sites/create', icon: Globe, label: 'Create Site' },
    ]
  },
  { to: '/analytics', icon: BarChart3, label: 'Analytics' },
  { to: '/redirects', icon: Link2, label: 'Redirects' },
  { to: '/webhooks', icon: Webhook, label: 'Webhooks' },
  { to: '/logs', icon: FileText, label: 'Logs' },
  { to: '/settings', icon: Settings, label: 'Settings' },
  { to: '/design-system', icon: Palette, label: 'Design System', devOnly: true },
];

interface SidebarProps {
  isOpen?: boolean;
  onClose?: () => void;
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const [expandedItems, setExpandedItems] = useState<string[]>(['sites']);
  const location = useLocation();
  const isDev = import.meta.env.DEV;

  const toggleExpanded = (label: string) => {
    setExpandedItems(prev =>
      prev.includes(label)
        ? prev.filter(item => item !== label)
        : [...prev, label]
    );
  };

  const isActive = (to: string) => {
    if (to === '/') return location.pathname === '/';
    return location.pathname.startsWith(to);
  };

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={onClose}
        />
      )}

      <aside className={`fixed lg:static inset-y-0 left-0 z-50 w-60 flex flex-col border-r border-[rgb(var(--border-primary))] bg-[rgb(var(--bg-elevated))] transform transition-transform duration-300 ease-in-out lg:transform-none
        ${isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}`}>
        <nav className="flex-1 px-3 py-6 overflow-y-auto">
          <div className="space-y-0.5">
            {navItems.map((item) => {
              if (item.devOnly && !isDev) return null;

              const hasChildren = item.children && item.children.length > 0;
              const isExpanded = expandedItems.includes(item.label.toLowerCase());

              if (hasChildren) {
                return (
                  <div key={item.label}>
                    <button
                      onClick={() => toggleExpanded(item.label.toLowerCase())}
                      className={`w-full group flex items-center justify-between gap-3 px-3 py-2 rounded-lg text-[13px] font-medium
                       transition-all duration-150 relative overflow-hidden
                       ${
                         isActive(item.to || '')
                           ? 'text-[rgb(var(--accent))] bg-[rgb(var(--accent-glow))]'
                           : 'text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]'
                       }`}
                    >
                      <div className="flex items-center gap-3">
                        {/* Active indicator */}
                        {isActive(item.to || '') && (
                          <div className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-4 bg-[rgb(var(--accent))] rounded-r" />
                        )}

                        <item.icon
                          className={`h-[18px] w-[18px] transition-transform duration-150
                            ${isActive(item.to || '') ? 'scale-100' : 'group-hover:scale-105'}`}
                        />
                        <span>{item.label}</span>
                      </div>
                      {isExpanded ? (
                        <ChevronDown className="h-4 w-4" />
                      ) : (
                        <ChevronRight className="h-4 w-4" />
                      )}
                    </button>

                    {/* Sub-menu */}
                    {isExpanded && (
                      <div className="ml-6 mt-1 space-y-0.5">
                        {item.children?.map((child) => (
                          <NavLink
                            key={child.to}
                            to={child.to || ''}
                            className={({ isActive }) =>
                              `group flex items-center gap-3 px-3 py-1.5 rounded-lg text-[12px] font-medium
                               transition-all duration-150 relative overflow-hidden
                               ${
                                 isActive
                                   ? 'text-[rgb(var(--accent))] bg-[rgb(var(--accent-glow))]'
                                   : 'text-[rgb(var(--text-tertiary))] hover:text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]'
                               }`
                            }
                          >
                            {({ isActive }) => (
                              <>
                                {/* Active indicator */}
                                {isActive && (
                                  <div className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-3 bg-[rgb(var(--accent))] rounded-r" />
                                )}
                                <child.icon
                                  className={`h-[16px] w-[16px] transition-transform duration-150
                                    ${isActive ? 'scale-100' : 'group-hover:scale-105'}`}
                                />
                                <span>{child.label}</span>
                              </>
                            )}
                          </NavLink>
                        ))}
                      </div>
                    )}
                  </div>
                );
              }

              return (
                <NavLink
                  key={item.to}
                  to={item.to || ''}
                  end={item.to === '/'}
                  className={({ isActive }) =>
                    `group flex items-center gap-3 px-3 py-2 rounded-lg text-[13px] font-medium
                     transition-all duration-150 relative overflow-hidden
                     ${
                       isActive
                         ? 'text-[rgb(var(--accent))] bg-[rgb(var(--accent-glow))]'
                         : 'text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]'
                     }`
                  }
                >
                  {({ isActive }) => (
                    <>
                      {/* Active indicator */}
                      {isActive && (
                        <div className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-4 bg-[rgb(var(--accent))] rounded-r" />
                      )}

                      <item.icon
                        className={`h-[18px] w-[18px] transition-transform duration-150
                          ${isActive ? 'scale-100' : 'group-hover:scale-105'}`}
                      />
                      <span>{item.label}</span>
                    </>
                  )}
                </NavLink>
              );
            })}
          </div>
        </nav>
      </aside>
    </>
  );
}