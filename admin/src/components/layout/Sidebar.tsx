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

export function Sidebar() {
  const isDev = import.meta.env.DEV;

  return (
    <aside className="w-60 flex flex-col border-r border-[rgb(var(--border-primary))] bg-[rgb(var(--bg-elevated))]">
      <nav className="flex-1 px-3 py-6">
        <div className="space-y-0.5">
          {navItems.map((item) => {
            if (item.devOnly && !isDev) return null;

            return (
              <NavLink
                key={item.to}
                to={item.to}
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
  );
}
