import { createRouter, createWebHashHistory } from 'vue-router'
import DashboardPage from './pages/DashboardPage.js'
import AppsPage from './pages/AppsPage.js'
import AliasesPage from './pages/AliasesPage.js'
import LogsPage from './pages/LogsPage.js'
import SystemPage from './pages/SystemPage.js'
import SettingsPage from './pages/SettingsPage.js'
import DesignSystemPage from './pages/DesignSystemPage.js'

const routes = [
  { path: '/', name: 'dashboard', component: DashboardPage, meta: { title: 'Dashboard', icon: 'layout-grid' } },
  { path: '/apps', name: 'apps', component: AppsPage, meta: { title: 'Apps', icon: 'layers' } },
  { path: '/apps/:id', name: 'app-detail', component: AppsPage, meta: { title: 'App Detail', parent: 'apps' } },
  { path: '/aliases', name: 'aliases', component: AliasesPage, meta: { title: 'Aliases', icon: 'link' } },
  { path: '/logs', name: 'logs', component: LogsPage, meta: { title: 'Logs', icon: 'terminal' } },
  { path: '/system', name: 'system', component: SystemPage, meta: { title: 'System', icon: 'heart-pulse' } },
  { path: '/settings', name: 'settings', component: SettingsPage, meta: { title: 'Settings', icon: 'settings' } },
  { path: '/design-system', name: 'design-system', component: DesignSystemPage, meta: { title: 'Design System', icon: 'layout' } },
  { path: '/:pathMatch(.*)*', name: 'not-found', redirect: '/' }
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes
})
