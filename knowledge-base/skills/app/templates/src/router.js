// Client-Side Router
// Works raw AND built

import { createRouter, createWebHistory } from 'vue-router'

// Import pages
import { HomePage } from './pages/Home.js'
import { SettingsPage } from './pages/Settings.js'

// Define routes
const routes = [
  {
    path: '/',
    name: 'home',
    component: HomePage
  },
  {
    path: '/settings',
    name: 'settings',
    component: SettingsPage
  },
  // Add more routes as needed
  // {
  //   path: '/items',
  //   name: 'items',
  //   component: () => import('./pages/Items.js').then(m => m.ItemsPage)
  // },
  // {
  //   path: '/items/:id',
  //   name: 'item-detail',
  //   component: () => import('./pages/ItemDetail.js').then(m => m.ItemDetailPage)
  // }
]

// Create router
export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    return { top: 0 }
  }
})

// Navigation guards (optional)
router.beforeEach((to, from, next) => {
  // Add auth checks, loading states, etc.
  next()
})
