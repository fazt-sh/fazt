import { Link, useLocation } from 'react-router-dom';
import { ChevronRight, Home } from 'lucide-react';

const routeLabels: Record<string, string> = {
  sites: 'Sites',
  analytics: 'Analytics',
  domains: 'Domains',
  create: 'Create',
  system: 'System',
  stats: 'Statistics',
  limits: 'Limits',
  logs: 'Logs',
  backup: 'Backup',
  health: 'Health',
  settings: 'Settings',
  'design-system': 'Design System',
  apps: 'Apps',
  webhooks: 'Webhooks',
  redirects: 'Redirects',
  tunnel: 'Tunnel',
  proxy: 'Proxy',
  botdaddy: 'BotDaddy',
  security: 'Security',
  ssh: 'SSH Keys',
  tokens: 'Auth Tokens',
  password: 'Password',
  external: 'External',
  cloudflare: 'Cloudflare',
  litestream: 'Litestream',
};

export function Breadcrumbs() {
  const location = useLocation();
  const pathSegments = location.pathname.split('/').filter(Boolean);

  if (pathSegments.length === 0) return null;

  const breadcrumbs = pathSegments.map((segment, index) => {
    const path = '/' + pathSegments.slice(0, index + 1).join('/');
    const label = routeLabels[segment] || segment.charAt(0).toUpperCase() + segment.slice(1);
    return { label, path };
  });

  return (
    <nav className="flex items-center space-x-2 text-sm mb-6 animate-fade-in">
      <Link 
        to="/" 
        className="group relative flex items-center text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] transition-colors py-1"
      >
        <Home className="w-4 h-4" />
        <span className="absolute bottom-0 left-0 w-full h-[2px] bg-[rgb(var(--accent))] transform scale-x-0 transition-transform duration-300 group-hover:scale-x-100 origin-left rounded-full" />
      </Link>
      
      {breadcrumbs.map((crumb, index) => (
        <div key={crumb.path} className="flex items-center">
          <ChevronRight className="w-4 h-4 mx-2 text-[rgb(var(--text-tertiary))]" />
          {index === breadcrumbs.length - 1 ? (
            <span className="text-[rgb(var(--accent))] font-medium tracking-wide py-1">
              {crumb.label}
            </span>
          ) : (
            <Link 
              to={crumb.path} 
              className="group relative text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] transition-colors font-medium py-1"
            >
              {crumb.label}
              <span className="absolute bottom-0 left-0 w-full h-[2px] bg-[rgb(var(--accent))] transform scale-x-0 transition-transform duration-300 group-hover:scale-x-100 origin-left rounded-full" />
            </Link>
          )}
        </div>
      ))}
    </nav>
  );
}