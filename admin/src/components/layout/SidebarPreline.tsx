import { useState } from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Globe,
  BarChart3,
  Link2,
  Webhook,
  Settings,
  FileText,
  Palette,
  X,
  Menu,
} from 'lucide-react';

interface NavItem {
  to: string;
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  devOnly?: boolean;
}

const navItems: NavItem[] = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/sites', icon: Globe, label: 'Sites' },
  { to: '/analytics', icon: BarChart3, label: 'Analytics' },
  { to: '/redirects', icon: Link2, label: 'Redirects' },
  { to: '/webhooks', icon: Webhook, label: 'Webhooks' },
  { to: '/logs', icon: FileText, label: 'Logs' },
  { to: '/settings', icon: Settings, label: 'Settings' },
  { to: '/design-system', icon: Palette, label: 'Design System', devOnly: true },
  { to: '/design-system-preline', icon: Palette, label: 'Preline UI', devOnly: true },
];

interface SidebarProps {
  isOpen?: boolean;
  onClose?: () => void;
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const isDev = import.meta.env.DEV;
  const [isCollapsed, setIsCollapsed] = useState(false);

  const NavContent = (
    <div className={`flex flex-col h-full ${isCollapsed ? 'w-16' : 'w-64'} transition-all duration-300 ease-in-out`}>
      {/* Header */}
      <header className="p-4 border-b border-[rgb(var(--border-primary))] flex items-center justify-between">
        {!isCollapsed && (
          <h2 className="text-xl font-bold text-[rgb(var(--text-primary))] font-display">
            Fazt
          </h2>
        )}
        <button
          onClick={() => setIsCollapsed(!isCollapsed)}
          className="p-2 rounded-lg hover:bg-[rgb(var(--bg-hover))] transition-colors duration-200"
        >
          {isCollapsed ? (
            <Menu className="h-5 w-5 text-[rgb(var(--text-secondary))]" />
          ) : (
            <X className="h-5 w-5 text-[rgb(var(--text-secondary))]" />
          )}
        </button>
      </header>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-6 overflow-y-auto">
        <div className="space-y-1">
          {navItems.map((item) => {
            if (item.devOnly && !isDev) return null;

            return (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                onClick={() => onClose?.()}
                className={({ isActive }) =>
                  `group flex items-center ${isCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-xl
                   transition-all duration-200 relative overflow-hidden
                   ${
                     isActive
                       ? 'text-[rgb(var(--accent-text))] bg-[rgb(var(--accent-mid))]'
                       : 'text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]'
                   }`
                }
              >
                {({ isActive }) => (
                  <>
                    <item.icon
                      className={`h-5 w-5 flex-shrink-0 transition-all duration-200
                        ${isActive ? 'text-white scale-110' : 'group-hover:scale-105'}`}
                    />
                    {!isCollapsed && (
                      <span className="text-sm font-medium">{item.label}</span>
                    )}
                    {/* Active indicator */}
                    {isActive && !isCollapsed && (
                      <div className="absolute right-2 top-1/2 -translate-y-1/2 w-1.5 h-1.5 bg-white rounded-full" />
                    )}
                  </>
                )}
              </NavLink>
            );
          })}
        </div>
      </nav>

      {/* Footer */}
      {!isCollapsed && (
        <footer className="p-4 border-t border-[rgb(var(--border-primary))]">
          <div className="bg-[rgb(var(--bg-subtle))] rounded-lg p-3">
            <p className="text-xs text-[rgb(var(--text-secondary))] mb-1">Need help?</p>
            <a
              href="#"
              className="text-sm font-medium text-[rgb(var(--accent))] hover:text-[rgb(var(--accent-mid))] transition-colors"
            >
              View Documentation →
            </a>
          </div>
        </footer>
      )}
    </div>
  );

  // Always render desktop sidebar (hidden on mobile)
  return (
    <>
      {/* Desktop Sidebar */}
      <aside className="hidden lg:flex h-full bg-[rgb(var(--bg-elevated))] border-r border-[rgb(var(--border-primary))]">
        {NavContent}
      </aside>

      {/* Mobile Sidebar Overlay */}
      {isOpen !== undefined && (
        <>
          {/* Backdrop */}
          {isOpen && (
            <div
              className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 lg:hidden"
              onClick={onClose}
            />
          )}

          {/* Mobile sidebar */}
          <div
            className={`
              fixed top-0 left-0 z-50 h-full w-64 transform transition-transform duration-300 ease-in-out
              ${isOpen ? 'translate-x-0' : '-translate-x-full'}
              lg:hidden
            `}
          >
            <div className="h-full bg-[rgb(var(--bg-elevated))] border-r border-[rgb(var(--border-primary))]">
              {NavContent}
            </div>
          </div>
        </>
      )}
    </>
  );
}