/**
 * Admin Routes
 */

export const routes = [
  {
    name: 'dashboard',
    path: '/',
    meta: { title: 'Dashboard', icon: 'layout-grid' }
  },
  {
    name: 'apps',
    path: '/apps',
    meta: { title: 'Apps', icon: 'layers' }
  },
  {
    name: 'app-detail',
    path: '/apps/:id',
    meta: { title: 'App Detail', parent: 'apps' }
  },
  {
    name: 'aliases',
    path: '/aliases',
    meta: { title: 'Aliases', icon: 'link' }
  },
  {
    name: 'logs',
    path: '/logs',
    meta: { title: 'Logs', icon: 'terminal' }
  },
  {
    name: 'system',
    path: '/system',
    meta: { title: 'System', icon: 'heart-pulse' }
  },
  {
    name: 'settings',
    path: '/settings',
    meta: { title: 'Settings', icon: 'settings' }
  },
  {
    name: 'design-system',
    path: '/design-system',
    meta: { title: 'Design System', icon: 'layout' }
  },
  {
    name: '404',
    path: '*',
    meta: { title: 'Not Found' }
  }
]
