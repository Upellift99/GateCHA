import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { title: 'Sign in' },
    },
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { requiresAuth: true, title: 'Dashboard' },
    },
    {
      path: '/keys',
      name: 'keys',
      component: () => import('../views/ApiKeysView.vue'),
      meta: { requiresAuth: true, title: 'API Keys' },
    },
    {
      path: '/keys/new',
      name: 'keys-new',
      component: () => import('../views/ApiKeyFormView.vue'),
      meta: { requiresAuth: true, title: 'Create API Key' },
    },
    {
      path: '/keys/:id',
      name: 'keys-detail',
      component: () => import('../views/ApiKeyDetailView.vue'),
      meta: { requiresAuth: true, title: 'API Key' },
    },
    {
      path: '/keys/:id/edit',
      name: 'keys-edit',
      component: () => import('../views/ApiKeyFormView.vue'),
      meta: { requiresAuth: true, title: 'Edit API Key' },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue'),
      meta: { requiresAuth: true, title: 'Settings' },
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('gatecha_token')
  if (to.meta.requiresAuth && !token) {
    return { name: 'login' }
  }
  if (to.name === 'login' && token) {
    return { name: 'dashboard' }
  }
})

// Each page names itself in the tab (WCAG 2.4.2), so several open tabs and
// screen-reader users can tell where they are.
router.afterEach((to) => {
  const title = typeof to.meta.title === 'string' ? to.meta.title : ''
  document.title = title ? `${title} · GateCHA` : 'GateCHA'
})

export default router
