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
    <nav className="flex items-center space-x-1 text-sm text-secondary mb-4">
      <Link to="/" className="flex items-center hover:text-primary transition-colors">
        <Home className="w-4 h-4" />
      </Link>
      {breadcrumbs.map((crumb, index) => (
        <div key={crumb.path} className="flex items-center">
          <ChevronRight className="w-4 h-4 mx-1 text-tertiary" />
          {index === breadcrumbs.length - 1 ? (
            <span className="text-primary font-medium">{crumb.label}</span>
          ) : (
            <Link to={crumb.path} className="hover:text-primary transition-colors">
              {crumb.label}
            </Link>
          )}
        </div>
      ))}
    </nav>
  );
}
