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
];

export function Sidebar() {
  const isDev = import.meta.env.DEV;

  return (
    <aside className="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col">
      <nav className="flex-1 px-4 py-6 space-y-1">
        {navItems.map((item) => {
          // Hide dev-only items in production
          if (item.devOnly && !isDev) return null;

          return (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-primary text-white'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                }`
              }
            >
              <item.icon className="h-5 w-5" />
              {item.label}
            </NavLink>
          );
        })}
      </nav>
    </aside>
  );
}
