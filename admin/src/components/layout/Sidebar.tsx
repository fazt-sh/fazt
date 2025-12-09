import { useState, useEffect } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import {
  LayoutDashboard,
  Globe,
  BarChart3,
  GlobeLock,
  Plus,
  Settings,
  Activity,
  Gauge,
  ScrollText,
  Archive,
  HeartPulse,
  Palette,
  Package,
  Webhook,
  ExternalLink,
  Route,
  Network,
  Bot,
  Shield,
  Key,
  KeyRound,
  Lock,
  Link2,
  Cloud,
  Database,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';

interface MenuItem {
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  path: string;
  exact?: boolean;
  children?: MenuItem[];
}

const menuItems: MenuItem[] = [
  {
    label: 'Dashboard',
    icon: LayoutDashboard,
    path: '/',
    exact: true,
  },
  {
    label: 'Sites',
    icon: Globe,
    path: '/sites',
    children: [
      { label: 'All Sites', path: '/sites', icon: Globe },
      { label: 'Analytics', path: '/sites/analytics', icon: BarChart3 },
      { label: 'Domains', path: '/sites/domains', icon: GlobeLock },
      { label: 'Create Site', path: '/sites/create', icon: Plus },
    ],
  },
  {
    label: 'System',
    icon: Settings,
    path: '/system',
    children: [
      { label: 'Statistics', path: '/system/stats', icon: Activity },
      { label: 'Resource Limits', path: '/system/limits', icon: Gauge },
      { label: 'Logs', path: '/system/logs', icon: ScrollText },
      { label: 'Backup', path: '/system/backup', icon: Archive },
      { label: 'Health', path: '/system/health', icon: HeartPulse },
      { label: 'Settings', path: '/system/settings', icon: Settings },
      { label: 'Design System', path: '/system/design-system', icon: Palette },
    ],
  },
  {
    label: 'Apps',
    icon: Package,
    path: '/apps',
    children: [
      { label: 'All Apps', path: '/apps', icon: Package },
      { label: 'Webhooks', path: '/apps/webhooks', icon: Webhook },
      { label: 'Redirects', path: '/apps/redirects', icon: ExternalLink },
      { label: 'Tunnel', path: '/apps/tunnel', icon: Route },
      { label: 'Proxy', path: '/apps/proxy', icon: Network },
      { label: 'BotDaddy', path: '/apps/botdaddy', icon: Bot },
    ],
  },
  {
    label: 'Security',
    icon: Shield,
    path: '/security',
    children: [
      { label: 'SSH Keys', path: '/security/ssh', icon: Key },
      { label: 'Auth Tokens', path: '/security/tokens', icon: KeyRound },
      { label: 'Password', path: '/security/password', icon: Lock },
    ],
  },
  {
    label: 'External',
    icon: Link2,
    path: '/external',
    children: [
      { label: 'Cloudflare', path: '/external/cloudflare', icon: Cloud },
      { label: 'Litestream', path: '/external/litestream', icon: Database },
    ],
  },
];

interface SidebarProps {
  isOpen?: boolean;
  onClose?: () => void;
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const [expandedItems, setExpandedItems] = useState<string[]>([]);
  const location = useLocation();

  const toggleExpanded = (label: string) => {
    setExpandedItems(prev =>
      prev.includes(label)
        ? prev.filter(item => item !== label)
        : [...prev, label]
    );
  };

  const isActive = (path: string, exact = false) => {
    if (exact) return location.pathname === path;
    return location.pathname.startsWith(path);
  };
  
  // Auto-expand the group of the current page
  useEffect(() => {
    const currentPath = location.pathname;
    const parentGroup = menuItems.find(item => 
      item.children && item.children.some(child => currentPath.startsWith(child.path))
    );
    
    if (parentGroup && !expandedItems.includes(parentGroup.label)) {
      setExpandedItems(prev => [...prev, parentGroup.label]);
    }
  }, [location.pathname]);

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
            {menuItems.map((item) => {
              const hasChildren = item.children && item.children.length > 0;
              const isExpanded = expandedItems.includes(item.label);

              if (hasChildren) {
                return (
                  <div key={item.label} className="mb-1">
                    <button
                      onClick={() => toggleExpanded(item.label)}
                      className={`w-full group flex items-center justify-between gap-3 px-3 py-2 rounded-lg text-[13px] font-medium
                       transition-all duration-150 relative overflow-hidden
                       ${
                         isActive(item.path)
                           ? 'text-[rgb(var(--accent))]'
                           : 'text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]'
                       }`}
                    >
                      <div className="flex items-center gap-3">
                        <item.icon
                          className={`h-[18px] w-[18px] transition-transform duration-150
                            ${isActive(item.path) ? 'scale-100' : 'group-hover:scale-105'}`}
                        />
                        <span>{item.label}</span>
                      </div>
                      {isExpanded ? (
                        <ChevronDown className="h-4 w-4 opacity-50" />
                      ) : (
                        <ChevronRight className="h-4 w-4 opacity-50" />
                      )}
                    </button>

                    {/* Sub-menu */}
                    {isExpanded && (
                      <div className="ml-4 mt-1 space-y-0.5 pl-2 border-l border-[rgb(var(--border-primary))]">
                        {item.children?.map((child) => (
                          <NavLink
                            key={child.path}
                            to={child.path}
                            end
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
                  key={item.path}
                  to={item.path}
                  end={item.exact}
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
